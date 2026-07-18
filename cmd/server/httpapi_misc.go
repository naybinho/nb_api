package main

import (
	"encoding/json"
	"net/http"
	
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (s *server) handleGetBlocklist(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	blocklist, err := sess.client.GetBlocklist(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"blocklist": blocklist})
}

func (s *server) handleUpdateBlocklist(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req struct {
		Action string   `json:"action"` // "block" or "unblock"
		JIDs   []string `json:"jids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	
	action := events.BlocklistChangeAction(req.Action)
	var blocklist *types.Blocklist
	
	for _, j := range req.JIDs {
		jid, _ := types.ParseJID(j)
		var err error
		blocklist, err = sess.client.UpdateBlocklist(r.Context(), jid, action)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	
	if blocklist == nil {
		blocklist, _ = sess.client.GetBlocklist(r.Context())
	}
	
	writeJSON(w, http.StatusOK, map[string]any{"blocklist": blocklist})
}
