package main

import (
	"encoding/json"
	"net/http"

	"go.mau.fi/whatsmeow/types"
)

func (s *server) handleSetProfileName(w http.ResponseWriter, r *http.Request) {
	// The current version of whatsmeow does not expose a high-level API to change the
	// profile display name (push name) for regular accounts.
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "not implemented: whatsmeow does not provide a high-level API for changing the profile display name",
	})
}

func (s *server) handleSetProfileStatus(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := sess.client.SetStatusMessage(r.Context(), req.Status)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSetProfilePhoto sets the current user's profile picture.
//
// Body (JSON): { "data": "<base64 JPEG bytes>" }
// Or multipart/form-data with a "photo" file field.
func (s *server) handleSetProfilePhoto(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}
	if sess.client.Store.ID == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not paired"})
		return
	}

	imgBytes, err := readImageBytes(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// For the own profile picture, use SetGroupPhoto but target our own JID.
	ownJID := *sess.client.Store.ID
	pictureID, err := sess.client.SetGroupPhoto(r.Context(), ownJID.ToNonAD(), imgBytes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pictureId": pictureID})
}

// handleSetPrivacySettings updates privacy settings.
//
// Body:
//
//	{
//	  "group_add": "all" | "contacts" | "contact_blacklist",
//	  "last":      "all" | "contacts" | "contact_blacklist" | "none",
//	  "status":    "all" | "contacts" | "contact_blacklist" | "none",
//	  "profile":   "all" | "contacts" | "contact_blacklist" | "none",
//	  "readreceipts": "all" | "none"
//	}
func (s *server) handleSetPrivacySettings(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var lastErr error
	for category, value := range req {
		// category strings: group_add, last, status, profile, readreceipts
		_, err := sess.client.SetPrivacySetting(r.Context(), types.PrivacySettingType(category), types.PrivacySetting(value))
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": lastErr.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleGetPrivacySettings(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}
	resp := sess.client.GetPrivacySettings(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{"settings": resp})
}
