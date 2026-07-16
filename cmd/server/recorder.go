package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// AudioRecorder manages the recording of a call's audio to S3-compatible storage.
// It captures PCM audio from both directions (mic and peer), mixes them into a
// single WAV file, and uploads it to S3 when the call ends.
type AudioRecorder struct {
	log       *slog.Logger
	client    *minio.Client
	bucket    string
	region    string
	sessionID string
	callID    string

	mu          sync.Mutex
	buf         *bytes.Buffer
	peerBuf     *bytes.Buffer // peer audio (WhatsApp -> browser)
	micBuf      *bytes.Buffer // mic audio (browser -> WhatsApp)
	startedAt   time.Time
	closed      bool
	totalFrames int64

	// S3 config
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
}

// NewRecorder creates a new AudioRecorder for a call. It returns nil if S3 is
// not configured (env vars missing) so callers can safely ignore recording.
func NewRecorder(log *slog.Logger, sessionID, callID string) *AudioRecorder {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		log.Debug("S3 not configured, recording disabled",
			"endpoint", endpoint != "",
			"accessKey", accessKey != "",
			"secretKey", secretKey != "",
			"bucket", bucket != "",
		)
		return nil
	}

	useSSL := true
	if os.Getenv("S3_SSL") == "false" {
		useSSL = false
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		log.Warn("failed to create S3 client, recording disabled", "err", err)
		return nil
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil || !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
			log.Warn("failed to create S3 bucket, recording disabled", "err", err)
			return nil
		}
	}

	return &AudioRecorder{
		log:       log.With("component", "recorder", "call_id", callID),
		client:    client,
		bucket:    bucket,
		region:    region,
		sessionID: sessionID,
		callID:    callID,
		peerBuf:   &bytes.Buffer{},
		micBuf:    &bytes.Buffer{},
		startedAt: time.Now(),
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
		useSSL:    useSSL,
	}
}

// WritePeerAudio records PCM audio received from the WhatsApp peer (incoming).
// The audio is 16 kHz mono float32 PCM.
func (r *AudioRecorder) WritePeerAudio(pcm []float32) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	writePCMFloat32(r.peerBuf, pcm)
	r.totalFrames += int64(len(pcm))
}

// WriteMicAudio records PCM audio captured from the browser microphone (outgoing).
// The audio is 16 kHz mono float32 PCM.
func (r *AudioRecorder) WriteMicAudio(pcm []float32) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	writePCMFloat32(r.micBuf, pcm)
	r.totalFrames += int64(len(pcm))
}

// writePCMFloat32 converts float32 PCM samples to Int16 LE and writes to buffer.
func writePCMFloat32(buf *bytes.Buffer, samples []float32) {
	for _, s := range samples {
		// Clamp to [-1, 1] and convert to int16
		if s > 1.0 {
			s = 1.0
		} else if s < -1.0 {
			s = -1.0
		}
		v := int16(s * 32767)
		_ = binary.Write(buf, binary.LittleEndian, v)
	}
}

// Close finalizes the recording, creates a WAV file from the captured audio,
// uploads it to S3, and returns the public URL of the recording.
// Returns empty string if recording was not enabled or failed.
func (r *AudioRecorder) Close() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ""
	}
	r.closed = true
	peerData := r.peerBuf.Bytes()
	micData := r.micBuf.Bytes()
	r.mu.Unlock()

	// If there's no audio data, skip upload
	if len(peerData) == 0 && len(micData) == 0 {
		r.log.Debug("no audio data recorded, skipping upload")
		return ""
	}

	// Mix both streams: if both have data, we average them
	// If only one stream has data, use it directly
	var finalData []byte
	if len(peerData) > 0 && len(micData) > 0 {
		finalData = mixInt16Streams(peerData, micData)
	} else if len(peerData) > 0 {
		finalData = peerData
	} else {
		finalData = micData
	}

	// Build WAV file
	wavBuf := buildWAV(finalData)
	duration := time.Since(r.startedAt)

	// Generate object name: recordings/{sessionId}/{callId}_{timestamp}.wav
	objectName := fmt.Sprintf("recordings/%s/%s_%d.wav",
		r.sessionID, r.callID, r.startedAt.UnixMilli())

	// Upload to S3
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := r.client.PutObject(ctx, r.bucket, objectName,
		bytes.NewReader(wavBuf.Bytes()),
		int64(wavBuf.Len()),
		minio.PutObjectOptions{
			ContentType: "audio/wav",
		},
	)
	if err != nil {
		r.log.Error("failed to upload recording to S3", "err", err)
		return ""
	}

	// Build public URL
	var url string
	if r.useSSL {
		url = fmt.Sprintf("https://%s/%s/%s", r.endpoint, r.bucket, objectName)
	} else {
		url = fmt.Sprintf("http://%s/%s/%s", r.endpoint, r.bucket, objectName)
	}

	r.log.Info("recording uploaded",
		"url", url,
		"duration", duration,
		"size_bytes", wavBuf.Len(),
	)
	return url
}

// mixInt16Streams averages two Int16 LE audio streams sample-by-sample.
// If one stream is longer, the remaining samples are passed through at half volume.
func mixInt16Streams(a, b []byte) []byte {
	samplesA := len(a) / 2
	samplesB := len(b) / 2
	maxSamples := samplesA
	if samplesB > maxSamples {
		maxSamples = samplesB
	}

	out := make([]byte, maxSamples*2)
	for i := 0; i < maxSamples; i++ {
		var sa, sb int16
		if i < samplesA {
			sa = int16(binary.LittleEndian.Uint16(a[i*2:]))
		}
		if i < samplesB {
			sb = int16(binary.LittleEndian.Uint16(b[i*2:]))
		}
		// Mix: average the two samples
		mixed := (int32(sa) + int32(sb)) / 2
		// Clamp to int16 range
		if mixed > 32767 {
			mixed = 32767
		} else if mixed < -32768 {
			mixed = -32768
		}
		binary.LittleEndian.PutUint16(out[i*2:], uint16(int16(mixed)))
	}
	return out
}

// buildWAV creates a WAV file from Int16 LE PCM data.
// Sample rate: 16000 Hz, mono, 16-bit.
func buildWAV(pcmData []byte) *bytes.Buffer {
	buf := &bytes.Buffer{}
	sampleRate := int32(16000)
	bitsPerSample := int16(16)
	numChannels := int16(1)
	byteRate := sampleRate * int32(numChannels*bitsPerSample/8)
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := len(pcmData)
	totalSize := 36 + dataSize

	// RIFF header
	_, _ = buf.Write([]byte("RIFF"))
	_ = binary.Write(buf, binary.LittleEndian, int32(totalSize))
	_, _ = buf.Write([]byte("WAVE"))

	// fmt subchunk
	_, _ = buf.Write([]byte("fmt "))
	_ = binary.Write(buf, binary.LittleEndian, int32(16))          // Subchunk1Size
	_ = binary.Write(buf, binary.LittleEndian, int16(1))           // AudioFormat (PCM)
	_ = binary.Write(buf, binary.LittleEndian, numChannels)        // NumChannels
	_ = binary.Write(buf, binary.LittleEndian, sampleRate)         // SampleRate
	_ = binary.Write(buf, binary.LittleEndian, byteRate)           // ByteRate
	_ = binary.Write(buf, binary.LittleEndian, blockAlign)         // BlockAlign
	_ = binary.Write(buf, binary.LittleEndian, bitsPerSample)      // BitsPerSample

	// data subchunk
	_, _ = buf.Write([]byte("data"))
	_ = binary.Write(buf, binary.LittleEndian, int32(dataSize))
	_, _ = buf.Write(pcmData)

	return buf
}

// RecordingURL returns the S3 URL for a given object path.
// This can be used to generate URLs for existing recordings.
func RecordingURL(endpoint, bucket, objectPath string) string {
	useSSL := true
	if os.Getenv("S3_SSL") == "false" {
		useSSL = false
	}
	if useSSL {
		return fmt.Sprintf("https://%s/%s/%s", endpoint, bucket, objectPath)
	}
	return fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, objectPath)
}


