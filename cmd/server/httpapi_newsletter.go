package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func (s *server) handleCreateNewsletter(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}
	
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Picture     string `json:"picture"` // base64 encoded
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var pictureBytes []byte
	if req.Picture != "" {
		var err error
		pictureBytes, err = base64.StdEncoding.DecodeString(req.Picture)
		if err != nil {
			pictureBytes, _ = base64.URLEncoding.DecodeString(req.Picture)
		}
	}
	
	meta, err := sess.client.CreateNewsletter(r.Context(), whatsmeow.CreateNewsletterParams{
		Name:        req.Name,
		Description: req.Description,
		Picture:     pictureBytes,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"newsletter": meta})
}

func (s *server) handleGetNewsletterInfo(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jid, _ := types.ParseJID(r.PathValue("jid"))
	
	info, err := sess.client.GetNewsletterInfo(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"info": info})
}

func (s *server) handleUnfollowNewsletter(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jid, _ := types.ParseJID(r.PathValue("jid"))
	
	err := sess.client.UnfollowNewsletter(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateNewsletter updates newsletter name, description, or picture.
//
// Body:
//
//	{
//	  "name": "New Name",
//	  "description": "New description",
//	  "picture": "<base64>"
//	}
func (s *server) handleUpdateNewsletter(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented: whatsmeow does not support newsletter update in this version"})
}

func (s *server) handleMuteNewsletter(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jid, _ := types.ParseJID(r.PathValue("jid"))
	
	var req struct {
		Mute bool `json:"mute"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req) // Ignore err, default to false if invalid
	
	err := sess.client.NewsletterToggleMute(r.Context(), jid, req.Mute)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
