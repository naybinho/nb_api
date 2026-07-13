package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// resolveJID accepts either a full JID string or a phone number.
func resolveJID(raw string) types.JID {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "@") {
		jid, _ := types.ParseJID(raw)
		return jid
	}
	return types.NewJID(normalizePhone(raw), types.DefaultUserServer)
}

// resolvePhoneJID normaliza o número e, para números brasileiros (código 55),
// tenta ambas as versões (com e sem o 9 extra) via IsOnWhatsApp, usando a que
// existir.
func (s *server) resolvePhoneJID(ctx context.Context, sess *Session, raw string) (types.JID, error) {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "@") {
		return types.ParseJID(raw)
	}

	phone := normalizePhone(raw)

	if strings.HasPrefix(phone, "55") {
		phoneWith9 := phone[:4] + "9" + phone[4:]

		switch len(phone) {
		case 12:
			// 55 + DDD + 8 dígitos → tenta sem e com o 9
			resp, err := sess.client.IsOnWhatsApp(ctx, []string{phone, phoneWith9})
			if err == nil {
				for _, r := range resp {
					if r.IsIn && !r.JID.IsEmpty() {
						return r.JID, nil
					}
				}
			}
			return types.NewJID(phoneWith9, types.DefaultUserServer), nil

		case 13:
			// 55 + DDD + 9 + 8 dígitos → tenta com e sem o 9
			phoneWithout9 := phone[:4] + phone[5:]
			resp, err := sess.client.IsOnWhatsApp(ctx, []string{phone, phoneWithout9})
			if err == nil {
				for _, r := range resp {
					if r.IsIn && !r.JID.IsEmpty() {
						return r.JID, nil
					}
				}
			}
			return types.NewJID(phone, types.DefaultUserServer), nil
		}
	}

	return types.NewJID(phone, types.DefaultUserServer), nil
}

// handleSendMessage is a generic dispatcher that accepts a `type` field.
func (s *server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	if sess.client == nil || !sess.client.IsConnected() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not connected"})
		return
	}

	// Read body once and re-use.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}

	var envelope struct {
		Type string `json:"type"` // "text", "image", "video", "audio", "document", "sticker", "location", "contact"
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	switch envelope.Type {
	case "text", "":
		var req struct {
			To   string `json:"to"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(body, &req); err != nil || strings.TrimSpace(req.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to and text required"})
			return
		}
		toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
			return
		}
		msg := &waE2E.Message{Conversation: proto.String(req.Text)}
		resp2, err := sess.client.SendMessage(r.Context(), toJID, msg)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"messageId": resp2.ID, "timestamp": resp2.Timestamp})

	case "image", "video", "audio", "document", "sticker":
		s.sendMediaFromBody(w, r, sess, body, envelope.Type)

	case "location":
		s.sendLocationFromBody(w, r, sess, body)

	case "contact":
		s.sendContactFromBody(w, r, sess, body)

	case "list":
		s.handleSendList(w, r)

	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown type: " + envelope.Type})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Text
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleSendText(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	if sess.client == nil || !sess.client.IsConnected() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not connected"})
		return
	}

	var req struct {
		To   string `json:"to"`
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}
	msg := &waE2E.Message{Conversation: proto.String(req.Text)}
	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

// ──────────────────────────────────────────────────────────────────────────────
// Media
// ──────────────────────────────────────────────────────────────────────────────

type sendMediaReq struct {
	To       string `json:"to"`
	Type     string `json:"type"`     // image, video, audio, document, sticker
	Data     string `json:"data"`     // base64-encoded file bytes
	URL      string `json:"url"`      // alternative: public URL (not yet supported for upload)
	Caption  string `json:"caption"`
	Filename string `json:"filename"` // for documents
	Mimetype string `json:"mimetype"`
}

func (s *server) handleSendMedia(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}
	var typeEnv struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(body, &typeEnv)
	s.sendMediaFromBody(w, r, sess, body, typeEnv.Type)
}

func (s *server) sendMediaFromBody(w http.ResponseWriter, r *http.Request, sess *Session, body []byte, mediaType string) {
	var req sendMediaReq
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Type == "" {
		req.Type = mediaType
	}
	if strings.TrimSpace(req.To) == "" || strings.TrimSpace(req.Data) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to and data (base64) are required"})
		return
	}

	fileBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		// Try raw URL encoding fallback
		fileBytes, err = base64.URLEncoding.DecodeString(req.Data)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data must be valid base64"})
			return
		}
	}

	// Determine mimetype
	mimetype := req.Mimetype
	if mimetype == "" {
		mimetype = http.DetectContentType(fileBytes)
	}

	// Map type to whatsmeow MediaType
	var wmMediaType whatsmeow.MediaType
	switch strings.ToLower(req.Type) {
	case "image":
		wmMediaType = whatsmeow.MediaImage
	case "video":
		wmMediaType = whatsmeow.MediaVideo
	case "audio":
		wmMediaType = whatsmeow.MediaAudio
	case "document":
		wmMediaType = whatsmeow.MediaDocument
	case "sticker":
		wmMediaType = whatsmeow.MediaImage
	default:
		wmMediaType = whatsmeow.MediaDocument
	}

	uploadResp, err := sess.client.Upload(r.Context(), fileBytes, wmMediaType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "upload failed: " + err.Error()})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}
	var waMsg *waE2E.Message

	switch strings.ToLower(req.Type) {
	case "image":
		waMsg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				Mimetype:      proto.String(mimetype),
				Caption:       proto.String(req.Caption),
			},
		}
	case "video":
		waMsg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				Mimetype:      proto.String(mimetype),
				Caption:       proto.String(req.Caption),
			},
		}
	case "audio":
		waMsg = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				Mimetype:      proto.String(mimetype),
			},
		}
	case "sticker":
		waMsg = &waE2E.Message{
			StickerMessage: &waE2E.StickerMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				Mimetype:      proto.String(mimetype),
			},
		}
	default: // document
		filename := req.Filename
		if filename == "" {
			exts, _ := mime.ExtensionsByType(mimetype)
			if len(exts) > 0 {
				filename = "file" + exts[0]
			} else {
				filename = "file"
			}
		}
		waMsg = &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uploadResp.FileLength),
				Mimetype:      proto.String(mimetype),
				FileName:      proto.String(filename),
				Caption:       proto.String(req.Caption),
			},
		}
	}

	resp, err := sess.client.SendMessage(r.Context(), toJID, waMsg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

// ──────────────────────────────────────────────────────────────────────────────
// Location
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleSendLocation(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}
	s.sendLocationFromBody(w, r, sess, body)
}

func (s *server) sendLocationFromBody(w http.ResponseWriter, r *http.Request, sess *Session, body []byte) {
	var req struct {
		To      string  `json:"to"`
		Lat     float64 `json:"lat"`
		Lng     float64 `json:"lng"`
		Name    string  `json:"name"`
		Address string  `json:"address"`
		URL     string  `json:"url"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.To) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to is required"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}
	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(req.Lat),
			DegreesLongitude: proto.Float64(req.Lng),
			Name:             proto.String(req.Name),
			Address:          proto.String(req.Address),
			URL:              proto.String(req.URL),
		},
	}
	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

// ──────────────────────────────────────────────────────────────────────────────
// Contact (vCard)
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleSendContact(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}
	s.sendContactFromBody(w, r, sess, body)
}

func (s *server) sendContactFromBody(w http.ResponseWriter, r *http.Request, sess *Session, body []byte) {
	var req struct {
		To          string `json:"to"`
		DisplayName string `json:"displayName"`
		VCard       string `json:"vcard"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.To) == "" || strings.TrimSpace(req.VCard) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to and vcard are required"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}
	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(req.DisplayName),
			Vcard:       proto.String(req.VCard),
		},
	}
	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

// ──────────────────────────────────────────────────────────────────────────────
// Reaction
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleSendReaction(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	var req struct {
		To        string `json:"to"`
		MessageId string `json:"messageId"`
		Reaction  string `json:"reaction"` // Emoji or empty to remove
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}
	msg := sess.client.BuildReaction(toJID, types.EmptyJID, req.MessageId, req.Reaction)

	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID})
}

// ──────────────────────────────────────────────────────────────────────────────
// Edit
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleEditMessage(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	var req struct {
		To        string `json:"to"`
		MessageId string `json:"messageId"`
		Text      string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}
	newMsg := &waE2E.Message{Conversation: proto.String(req.Text)}
	msg := sess.client.BuildEdit(toJID, req.MessageId, newMsg)

	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID})
}

// ──────────────────────────────────────────────────────────────────────────────
// Revoke
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleRevokeMessage(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	msgID := r.PathValue("id")
	toJID, err := s.resolvePhoneJID(r.Context(), sess, r.URL.Query().Get("to"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}

	msg := sess.client.BuildRevoke(toJID, types.EmptyJID, msgID)
	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID})
}

// ──────────────────────────────────────────────────────────────────────────────
// Mark Read
// ──────────────────────────────────────────────────────────────────────────────

func (s *server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	msgID := r.PathValue("id")
	toJID, err := s.resolvePhoneJID(r.Context(), sess, r.URL.Query().Get("to"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}

	err = sess.client.MarkRead(r.Context(), []types.MessageID{types.MessageID(msgID)}, time.Now(), toJID, types.EmptyJID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ──────────────────────────────────────────────────────────────────────────────
// Download Media
// ──────────────────────────────────────────────────────────────────────────────

// handleDownloadMedia downloads a WhatsApp media file using the metadata provided as query params.
//
// Query params:
//
//	url        — WhatsApp CDN URL (from the received message)
//	directPath — direct path
//	mediaKey   — base64-encoded media key
//	type       — image | video | audio | document | sticker
//	encSha256  — base64-encoded encrypted SHA256
//	sha256     — base64-encoded SHA256
func (s *server) handleDownloadMedia(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	q := r.URL.Query()
	mediaKeyB64 := q.Get("mediaKey")
	encSha256B64 := q.Get("encSha256")
	sha256B64 := q.Get("sha256")
	directPath := q.Get("directPath")
	mediaTypeStr := q.Get("type")

	if mediaKeyB64 == "" || directPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mediaKey and directPath are required"})
		return
	}

	mediaKey, err := base64.StdEncoding.DecodeString(mediaKeyB64)
	if err != nil {
		mediaKey, err = base64.URLEncoding.DecodeString(mediaKeyB64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mediaKey base64"})
			return
		}
	}
	encSha256, _ := base64.StdEncoding.DecodeString(encSha256B64)
	sha256, _ := base64.StdEncoding.DecodeString(sha256B64)

	// Determine whatsmeow MediaType
	var wmMediaType whatsmeow.MediaType
	switch strings.ToLower(mediaTypeStr) {
	case "image":
		wmMediaType = whatsmeow.MediaImage
	case "video":
		wmMediaType = whatsmeow.MediaVideo
	case "audio":
		wmMediaType = whatsmeow.MediaAudio
	case "sticker":
		wmMediaType = whatsmeow.MediaImage
	default:
		wmMediaType = whatsmeow.MediaDocument
	}

	// Build a minimal DownloadableMessage using ImageMessage as the carrier for the metadata.
	// All downloadable messages share the same fields via the interface.
	var dlMsg whatsmeow.DownloadableMessage
	switch wmMediaType {
	case whatsmeow.MediaVideo:
		dlMsg = &waE2E.VideoMessage{
			DirectPath:    proto.String(directPath),
			MediaKey:      mediaKey,
			FileEncSHA256: encSha256,
			FileSHA256:    sha256,
		}
	case whatsmeow.MediaAudio:
		dlMsg = &waE2E.AudioMessage{
			DirectPath:    proto.String(directPath),
			MediaKey:      mediaKey,
			FileEncSHA256: encSha256,
			FileSHA256:    sha256,
		}
	case whatsmeow.MediaDocument:
		dlMsg = &waE2E.DocumentMessage{
			DirectPath:    proto.String(directPath),
			MediaKey:      mediaKey,
			FileEncSHA256: encSha256,
			FileSHA256:    sha256,
		}
	default: // image & sticker
		dlMsg = &waE2E.ImageMessage{
			DirectPath:    proto.String(directPath),
			MediaKey:      mediaKey,
			FileEncSHA256: encSha256,
			FileSHA256:    sha256,
		}
	}

	data, err := sess.client.Download(r.Context(), dlMsg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "download failed: " + err.Error()})
		return
	}

	// Determine content-type for the response
	contentType := http.DetectContentType(data)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *server) handleSendList(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	if sess.client == nil || !sess.client.IsConnected() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not connected"})
		return
	}

	var req struct {
		To          string `json:"to"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ButtonText  string `json:"buttonText"`
		FooterText  string `json:"footerText"`
		Sections    []struct {
			Title string `json:"title"`
			Rows  []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				RowID       string `json:"rowId"`
			} `json:"rows"`
		} `json:"sections"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.To) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to is required"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}

	buttons := make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, 0)
	for _, sec := range req.Sections {
		for _, row := range sec.Rows {
			id := row.RowID
			if id == "" {
				id = row.Title
			}
			params := map[string]string{
				"display_text": row.Title,
				"id":           id,
			}
			paramsJSON, _ := json.Marshal(params)
			paramsStr := string(paramsJSON)
			buttons = append(buttons, &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
				Name:             proto.String("quick_reply"),
				ButtonParamsJSON: &paramsStr,
			})
		}
	}
	if len(buttons) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one section with at least one row is required"})
		return
	}

	msgVersion := int32(3)
	interactive := &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String(req.Description),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons:        buttons,
				MessageVersion: &msgVersion,
			},
		},
	}
	if req.Title != "" {
		interactive.Header = &waE2E.InteractiveMessage_Header{
			Title: proto.String(req.Title),
		}
	}
	if req.FooterText != "" {
		interactive.Footer = &waE2E.InteractiveMessage_Footer{
			Text: proto.String(req.FooterText),
		}
	}

	msgSecret := make([]byte, 32)
	_, _ = rand.Read(msgSecret)

	msg := &waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{
			DeviceListMetadata:        &waE2E.DeviceListMetadata{},
			DeviceListMetadataVersion: proto.Int32(2),
			MessageSecret:             msgSecret,
		},
		InteractiveMessage: interactive,
	}

	bizNode := binary.Node{
		Tag: "biz",
		Content: []binary.Node{{
			Tag: "interactive",
			Attrs: binary.Attrs{
				"type": "native_flow",
				"v":    "1",
			},
			Content: []binary.Node{{
				Tag: "native_flow",
				Attrs: binary.Attrs{
					"v":    "9",
					"name": "mixed",
				},
			}},
		}},
	}

	resp, err := sess.client.SendMessage(r.Context(), toJID, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &[]binary.Node{bizNode},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

func (s *server) handleSendListInteractive(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	if sess.client == nil || !sess.client.IsConnected() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not connected"})
		return
	}

	var req struct {
		To          string `json:"to"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ButtonText  string `json:"buttonText"`
		FooterText  string `json:"footerText"`
		Sections    []struct {
			Title string `json:"title"`
			Rows  []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				RowID       string `json:"rowId"`
			} `json:"rows"`
		} `json:"sections"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.To) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to is required"})
		return
	}
	if len(req.Sections) == 0 || len(req.Sections[0].Rows) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one section with at least one row is required"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}

	sections := make([]*waE2E.ListMessage_Section, 0, len(req.Sections))
	for _, sec := range req.Sections {
		rows := make([]*waE2E.ListMessage_Row, 0, len(sec.Rows))
		for _, row := range sec.Rows {
			r := &waE2E.ListMessage_Row{
				Title:       proto.String(row.Title),
				Description: proto.String(row.Description),
				RowID:       proto.String(row.RowID),
			}
			if r.RowID == nil || *r.RowID == "" {
				r.RowID = proto.String(row.Title)
			}
			rows = append(rows, r)
		}
		sections = append(sections, &waE2E.ListMessage_Section{
			Title: proto.String(sec.Title),
			Rows:  rows,
		})
	}

	msgSecret := make([]byte, 32)
	_, _ = rand.Read(msgSecret)

	listMsg := &waE2E.ListMessage{
		Title:       proto.String(req.Title),
		Description: proto.String(req.Description),
		ButtonText:  proto.String(req.ButtonText),
		ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		Sections:    sections,
		FooterText:  proto.String(req.FooterText),
	}

	inner := &waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{
			DeviceListMetadata:        &waE2E.DeviceListMetadata{},
			DeviceListMetadataVersion: proto.Int32(2),
			MessageSecret:             msgSecret,
		},
		ListMessage: listMsg,
	}

	msg := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{
			Message: inner,
		},
	}

	bizNode := binary.Node{
		Tag: "biz",
		Content: []binary.Node{{
			Tag: "list",
			Attrs: binary.Attrs{
				"type": "product_list",
				"v":    "2",
			},
		}},
	}

	resp, err := sess.client.SendMessage(r.Context(), toJID, msg, whatsmeow.SendRequestExtra{
		AdditionalNodes: &[]binary.Node{bizNode},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

func (s *server) handleSendPoll(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	if sess.client == nil || !sess.client.IsConnected() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not connected"})
		return
	}

	var req struct {
		To              string   `json:"to"`
		Name            string   `json:"name"`
		Options         []string `json:"options"`
		SelectableCount int      `json:"selectableCount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.To) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to is required"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if len(req.Options) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least 2 options are required"})
		return
	}

	toJID, err := s.resolvePhoneJID(r.Context(), sess, req.To)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JID"})
		return
	}

	selectable := req.SelectableCount
	if selectable < 1 || selectable > len(req.Options) {
		selectable = 1
	}

	msg := sess.client.BuildPollCreation(req.Name, req.Options, selectable)
	resp, err := sess.client.SendMessage(r.Context(), toJID, msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messageId": resp.ID, "timestamp": resp.Timestamp})
}

