package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func (s *server) handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jidStrs := r.URL.Query()["jid"]
	if len(jidStrs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provide at least one jid"})
		return
	}

	jids := make([]types.JID, len(jidStrs))
	for i, j := range jidStrs {
		jids[i], _ = types.ParseJID(j)
	}

	info, err := sess.client.GetUserInfo(r.Context(), jids)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"info": info})
}

func (s *server) handleSubscribePresence(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jid, _ := types.ParseJID(r.PathValue("jid"))
	err := sess.client.SubscribePresence(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleCheckIsOnWhatsApp(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req struct {
		Phones []string `json:"phones"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Normaliza números brasileiros: tenta com e sem o 9 extra
	for i := range req.Phones {
		p := normalizePhone(req.Phones[i])
		switch {
		case len(p) == 12 && strings.HasPrefix(p, "55"):
			p = p[:4] + "9" + p[4:]
		case len(p) == 13 && strings.HasPrefix(p, "55"):
			p = p[:4] + p[5:]
		}
		req.Phones[i] = p
	}

	resp, err := sess.client.IsOnWhatsApp(r.Context(), req.Phones)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": resp})
}

func (s *server) handleGetContacts(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	contacts, err := sess.client.Store.Contacts.GetAllContacts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"contacts": contacts})
}

func (s *server) handleGetContactInfo(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jid, _ := types.ParseJID(r.PathValue("jid"))
	contact, err := sess.client.Store.Contacts.GetContact(r.Context(), jid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (s *server) handleGetContactAvatar(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	jid, _ := types.ParseJID(r.PathValue("jid"))
	preview := r.URL.Query().Get("preview") == "true"
	
	pic, err := sess.client.GetProfilePictureInfo(r.Context(), jid, &whatsmeow.GetProfilePictureParams{
		Preview: preview,
		ExistingID: r.URL.Query().Get("existingId"),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"picture": pic})
}
