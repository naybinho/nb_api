package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *server) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}
	whs, err := s.webhookStore.list(r.Context(), sess.id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"webhooks": whs})
}

func (s *server) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	var req struct {
		URL     string `json:"url"`
		Events  string `json:"events"`
		Enabled *bool  `json:"enabled"`
		Secret  string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
		return
	}
	events := strings.TrimSpace(req.Events)
	if events == "" {
		events = "*"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	now := time.Now()
	wh := &webhookRow{
		ID:        newWebhookID(),
		SessionID: sess.id,
		URL:       strings.TrimSpace(req.URL),
		Events:    events,
		Enabled:   enabled,
		Secret:    req.Secret,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.webhookStore.insert(r.Context(), wh); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, wh)
}

func (s *server) handleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	wid := r.PathValue("wid")
	wh, err := s.webhookStore.get(r.Context(), wid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}
	if wh.SessionID != sess.id {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}

	var req struct {
		URL     *string `json:"url"`
		Events  *string `json:"events"`
		Enabled *bool   `json:"enabled"`
		Secret  *string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.URL != nil {
		wh.URL = strings.TrimSpace(*req.URL)
	}
	if req.Events != nil {
		wh.Events = strings.TrimSpace(*req.Events)
	}
	if req.Enabled != nil {
		wh.Enabled = *req.Enabled
	}
	if req.Secret != nil {
		wh.Secret = *req.Secret
	}
	wh.UpdatedAt = time.Now()

	if err := s.webhookStore.update(r.Context(), wh); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, wh)
}

func (s *server) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	wid := r.PathValue("wid")
	wh, err := s.webhookStore.get(r.Context(), wid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}
	if wh.SessionID != sess.id {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}

	if err := s.webhookStore.delete(r.Context(), wid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleTestWebhook(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil {
		return
	}

	wid := r.PathValue("wid")
	wh, err := s.webhookStore.get(r.Context(), wid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}
	if wh.SessionID != sess.id {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "webhook not found"})
		return
	}

	testPayload := map[string]any{
		"event":     "test",
		"sessionId": sess.id,
		"timestamp": time.Now().UnixMilli(),
		"data": map[string]any{
			"message": "This is a test webhook event from NB_API",
		},
	}
	body, _ := json.Marshal(testPayload)

	if err := s.webhookDispatcher.Send(r.Context(), *wh, body); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "webhook test failed",
			"detail":  err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "Test webhook sent successfully"})
}
