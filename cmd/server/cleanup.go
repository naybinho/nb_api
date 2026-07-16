package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// startCleanupScheduler starts a background goroutine that periodically
// deletes old recordings from S3 based on RECORDING_RETENTION_DAYS env var.
// The cleanup runs immediately on startup and then every 6 hours.
// Set RECORDING_RETENTION_DAYS to 0 or empty to disable automatic cleanup.
func startCleanupScheduler(log *slog.Logger, historyStore *callHistoryStore) {
	daysStr := os.Getenv("RECORDING_RETENTION_DAYS")
	if daysStr == "" {
		log.Info("RECORDING_RETENTION_DAYS not set, automatic recording cleanup disabled")
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		log.Warn("invalid RECORDING_RETENTION_DAYS value, cleanup disabled", "value", daysStr)
		return
	}

	// Create S3 client for cleanup
	client, bucket, err := newS3ClientForCleanup()
	if err != nil {
		log.Warn("failed to create S3 client for recording cleanup", "err", err)
		return
	}

	log.Info("recording cleanup scheduled",
		"retention_days", days,
		"bucket", bucket,
		"interval", "6h",
	)

	go func() {
		// Run immediately on startup
		runCleanup(log, historyStore, client, bucket, days)

		// Then run every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			runCleanup(log, historyStore, client, bucket, days)
		}
	}()
}

// newS3ClientForCleanup creates a standalone S3 client for the cleanup job.
func newS3ClientForCleanup() (*minio.Client, string, error) {
	rawEndpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")
	pathStyleStr := os.Getenv("S3_PATH_STYLE")

	if rawEndpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, "", nil
	}

	if region == "" {
		region = "us-east-1"
	}

	endpoint, useSSL := cleanEndpoint(rawEndpoint)

	pathStyle := true
	if pathStyleStr == "false" {
		pathStyle = false
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       useSSL,
		Region:       region,
		BucketLookup: getBucketLookup(pathStyle),
	})
	if err != nil {
		return nil, "", err
	}

	return client, bucket, nil
}

// runCleanup performs a single cleanup cycle:
// 1. Queries the DB for recordings older than retentionDays
// 2. Deletes each recording from S3
// 3. Clears the recording_url in the DB
func runCleanup(log *slog.Logger, historyStore *callHistoryStore, client *minio.Client, bucket string, retentionDays int) {
	ctx := context.Background()
	cutoff := time.Now().AddDate(0, 0, -retentionDays).UnixMilli()

	log.Info("running recording cleanup",
		"retention_days", retentionDays,
		"cutoff_before", time.UnixMilli(cutoff).Format(time.RFC3339),
	)

	rows, err := historyStore.listExpiredRecordings(ctx, cutoff)
	if err != nil {
		log.Error("failed to query expired recordings for cleanup", "err", err)
		return
	}

	if len(rows) == 0 {
		log.Debug("no expired recordings to clean up")
		return
	}

	log.Info("found expired recordings to clean up", "count", len(rows))

	deleted := 0
	for _, row := range rows {
		if row.RecordingURL == "" {
			continue
		}

		// Extract object key from the S3 URL
		objectKey := extractObjectKey(row.RecordingURL)
		if objectKey == "" {
			log.Warn("could not extract object key from recording URL, skipping",
				"call_id", row.CallID, "url", row.RecordingURL)
			continue
		}

		// Delete the object from S3
		if err := client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
			log.Error("failed to delete recording from S3",
				"call_id", row.CallID,
				"key", objectKey,
				"err", err,
			)
			continue
		}

		// Clear the recording URL in the database
		if err := historyStore.clearRecordingURL(ctx, row.CallID); err != nil {
			log.Error("failed to clear recording URL in database",
				"call_id", row.CallID,
				"err", err,
			)
			continue
		}

		deleted++
		log.Info("deleted expired recording",
			"call_id", row.CallID,
			"key", objectKey,
			"age_days", retentionDays,
		)
	}

	log.Info("recording cleanup completed",
		"total_expired", len(rows),
		"deleted", deleted,
		"failed", len(rows)-deleted,
	)
}

// extractObjectKey extracts the S3 object key from a recording URL.
// Expected URL format (path-style):
//
//	https://{endpoint}/{bucket}/recordings/{sessionId}/{callId}_{timestamp}.wav
//
// Returns the object key (e.g. "recordings/{sessionId}/{callId}_{timestamp}.wav").
func extractObjectKey(recordingURL string) string {
	parsed, err := url.Parse(recordingURL)
	if err != nil {
		return ""
	}

	// Path is: /{bucket}/recordings/{sessionId}/{callId}_{timestamp}.wav
	path := strings.TrimPrefix(parsed.Path, "/")

	// Remove the first path segment (bucket name) to get the object key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}
