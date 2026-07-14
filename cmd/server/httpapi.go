package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"nb_api/internal/voip/core"
)

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/sessions", s.handleSessionList)
	mux.HandleFunc("POST /api/sessions", s.handleSessionCreate)
	mux.HandleFunc("DELETE /api/sessions/{sid}", s.handleSessionDelete)
	mux.HandleFunc("POST /api/sessions/{sid}/logout", s.handleSessionLogout)
	mux.HandleFunc("POST /api/sessions/{sid}/pair", s.handleSessionPair)
	mux.HandleFunc("PUT /api/sessions/{sid}/apikey", s.handleUpdateAPIKey)
	mux.HandleFunc("PUT /api/sessions/{sid}/name", s.handleUpdateName)
	mux.HandleFunc("POST /api/sessions/{sid}/calls", s.handleStartCall)
	mux.HandleFunc("POST /api/sessions/{sid}/calls/{id}/webrtc", s.handleWebRTC)
	mux.HandleFunc("POST /api/sessions/{sid}/calls/{id}/accept", s.handleAccept)
	mux.HandleFunc("POST /api/sessions/{sid}/calls/{id}/reject", s.handleReject)
	mux.HandleFunc("DELETE /api/sessions/{sid}/calls/{id}", s.handleEndCall)
	mux.HandleFunc("GET /api/sessions/{sid}/history", s.handleHistory)
	
	// Messages
	mux.HandleFunc("POST /api/sessions/{sid}/messages", s.handleSendMessage) // Accepts form-data or JSON (for text, media, etc)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/text", s.handleSendText)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/media", s.handleSendMedia) // image, video, audio, document, sticker
	mux.HandleFunc("POST /api/sessions/{sid}/messages/location", s.handleSendLocation)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/contact", s.handleSendContact)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/list", s.handleSendList)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/list-interactive", s.handleSendListInteractive)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/poll", s.handleSendPoll)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/react", s.handleSendReaction)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/edit", s.handleEditMessage)
	mux.HandleFunc("DELETE /api/sessions/{sid}/messages/{id}", s.handleRevokeMessage)
	mux.HandleFunc("POST /api/sessions/{sid}/messages/{id}/read", s.handleMarkRead)
	mux.HandleFunc("GET /api/sessions/{sid}/media/{id}", s.handleDownloadMedia)

	// Groups
	mux.HandleFunc("GET /api/sessions/{sid}/groups", s.handleGetGroups)
	mux.HandleFunc("POST /api/sessions/{sid}/groups", s.handleCreateGroup)
	mux.HandleFunc("GET /api/sessions/{sid}/groups/{gid}", s.handleGetGroupInfo)
	mux.HandleFunc("PUT /api/sessions/{sid}/groups/{gid}", s.handleUpdateGroup)
	mux.HandleFunc("DELETE /api/sessions/{sid}/groups/{gid}", s.handleLeaveGroup)
	mux.HandleFunc("GET /api/sessions/{sid}/groups/{gid}/invite", s.handleGetGroupInvite)
	mux.HandleFunc("POST /api/sessions/{sid}/groups/{gid}/invite/revoke", s.handleRevokeGroupInvite)
	mux.HandleFunc("POST /api/sessions/{sid}/groups/join", s.handleJoinGroup)
	mux.HandleFunc("POST /api/sessions/{sid}/groups/{gid}/participants", s.handleUpdateGroupParticipants)
	mux.HandleFunc("PUT /api/sessions/{sid}/groups/{gid}/participants/{jid}", s.handlePromoteDemoteGroupParticipants)
	mux.HandleFunc("GET /api/sessions/{sid}/groups/{gid}/requests", s.handleGetGroupRequests)
	mux.HandleFunc("POST /api/sessions/{sid}/groups/{gid}/requests", s.handleApproveGroupRequests)
	mux.HandleFunc("PUT /api/sessions/{sid}/groups/{gid}/photo", s.handleSetGroupPhoto)
	mux.HandleFunc("DELETE /api/sessions/{sid}/groups/{gid}/photo", s.handleDeleteGroupPhoto)

	// Users / Contacts
	mux.HandleFunc("GET /api/sessions/{sid}/users", s.handleGetUserInfo)
	mux.HandleFunc("GET /api/sessions/{sid}/users/{jid}/presence", s.handleSubscribePresence)
	mux.HandleFunc("POST /api/sessions/{sid}/users/check", s.handleCheckIsOnWhatsApp)
	mux.HandleFunc("GET /api/sessions/{sid}/contacts", s.handleGetContacts)
	mux.HandleFunc("GET /api/sessions/{sid}/contacts/{jid}", s.handleGetContactInfo)
	mux.HandleFunc("GET /api/sessions/{sid}/contacts/{jid}/avatar", s.handleGetContactAvatar)

	// Profile
	mux.HandleFunc("PUT /api/sessions/{sid}/profile/name", s.handleSetProfileName)
	mux.HandleFunc("PUT /api/sessions/{sid}/profile/status", s.handleSetProfileStatus)
	mux.HandleFunc("PUT /api/sessions/{sid}/profile/photo", s.handleSetProfilePhoto)
	mux.HandleFunc("PUT /api/sessions/{sid}/profile/privacy", s.handleSetPrivacySettings)
	mux.HandleFunc("GET /api/sessions/{sid}/profile/privacy", s.handleGetPrivacySettings)

	// Misc
	mux.HandleFunc("GET /api/sessions/{sid}/blocklist", s.handleGetBlocklist)
	mux.HandleFunc("POST /api/sessions/{sid}/blocklist", s.handleUpdateBlocklist)

	// Newsletters (Channels)
	mux.HandleFunc("POST /api/sessions/{sid}/newsletters", s.handleCreateNewsletter)
	mux.HandleFunc("GET /api/sessions/{sid}/newsletters/{jid}", s.handleGetNewsletterInfo)
	mux.HandleFunc("DELETE /api/sessions/{sid}/newsletters/{jid}", s.handleUnfollowNewsletter)
	mux.HandleFunc("PUT /api/sessions/{sid}/newsletters/{jid}", s.handleUpdateNewsletter)
	mux.HandleFunc("POST /api/sessions/{sid}/newsletters/{jid}/mute", s.handleMuteNewsletter)

	mux.HandleFunc("GET /api/events", s.handleEvents)

	// Swagger
	mux.HandleFunc("GET /swagger", s.handleSwaggerUI)
	mux.HandleFunc("GET /swagger.json", s.handleSwaggerJSON)

	if s.staticDir != "" {
		if _, err := os.Stat(s.staticDir); err == nil {
			mux.Handle("/", http.FileServer(http.Dir(s.staticDir)))
		}
	}
	return withCORS(s.withAuth(mux))
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Client-Id, X-Api-Key, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (s *server) withAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rotas públicas (Swagger UI e spec)
		if r.URL.Path == "/swagger" || r.URL.Path == "/swagger.json" {
			h.ServeHTTP(w, r)
			return
		}
		// Se não configurou credenciais, permite acesso livre
		if s.authUsername == "" {
			h.ServeHTTP(w, r)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != s.authUsername || pass != s.authPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="NB_Api"`)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func clientID(r *http.Request) string {
	if id := r.Header.Get("X-Client-Id"); id != "" {
		return id
	}
	return r.URL.Query().Get("clientId")
}

func (s *server) sessionByID(w http.ResponseWriter, sid string) *Session {
	sess, ok := s.sessions.Get(sid)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such session"})
		return nil
	}
	return sess
}

func (s *server) handleEvents(w http.ResponseWriter, r *http.Request) {
	s.broker.serveSSE(w, r, clientID(r))
}

func (s *server) handleSessionList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"sessions": s.sessions.infos()})
}

func (s *server) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		APIKey string `json:"apiKey"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "Session"
	}
	apiKey := strings.TrimSpace(body.APIKey)
	id, err := s.sessions.Create(name, apiKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	sess, _ := s.sessions.Get(id)
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "apiKey": sess.apiKey})
}

func (s *server) handleSessionDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Delete(r.Context(), r.PathValue("sid")); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleSessionLogout(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Logout(r.Context(), r.PathValue("sid")); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleSessionPair(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Pair(r.PathValue("sid")); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleUpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	sid := r.PathValue("sid")
	var body struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	apiKey := strings.TrimSpace(body.APIKey)
	if apiKey == "" {
		apiKey = newAPIKey()
	}
	if err := s.sessions.UpdateAPIKey(sid, apiKey); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"apiKey": apiKey})
}

func (s *server) handleUpdateName(w http.ResponseWriter, r *http.Request) {
	sid := r.PathValue("sid")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	if err := s.sessions.UpdateName(sid, name); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleStartCall(w http.ResponseWriter, r *http.Request) {
	if sess := s.sessionByID(w, r.PathValue("sid")); sess != nil {
		s.doStartCall(sess, w, r)
	}
}

func (s *server) handleWebRTC(w http.ResponseWriter, r *http.Request) {
	if sess := s.sessionByID(w, r.PathValue("sid")); sess != nil {
		s.doWebRTC(sess, w, r)
	}
}

func (s *server) handleAccept(w http.ResponseWriter, r *http.Request) {
	if sess := s.sessionByID(w, r.PathValue("sid")); sess != nil {
		s.doAccept(sess, w, r)
	}
}

func (s *server) handleReject(w http.ResponseWriter, r *http.Request) {
	if sess := s.sessionByID(w, r.PathValue("sid")); sess != nil {
		s.doReject(sess, w, r)
	}
}

func (s *server) handleEndCall(w http.ResponseWriter, r *http.Request) {
	if sess := s.sessionByID(w, r.PathValue("sid")); sess != nil {
		s.doEndCall(sess, w, r)
	}
}

func (s *server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if sess := s.sessionByID(w, r.PathValue("sid")); sess != nil {
		writeJSON(w, http.StatusOK, map[string]any{"rows": s.broker.historyRows(sess.id, 50)})
	}
}

func (s *server) doStartCall(sess *Session, w http.ResponseWriter, r *http.Request) {
	if sess.client.Store.ID == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not paired"})
		return
	}
	var body struct {
		Phone      string `json:"phone"`
		DurationMs int    `json:"duration_ms"`
		Record     bool   `json:"record"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Phone) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone required"})
		return
	}
	owner := clientID(r)
	if other := s.broker.ownerActiveCall(owner); other != "" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "operator already on a call"})
		return
	}
	if max := s.sessions.maxCalls; max > 0 && sess.reg.count() >= max {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "max concurrent calls"})
		return
	}
	peer, err := s.resolvePhoneJID(r.Context(), sess, body.Phone)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid phone"})
		return
	}

	callID, err := sess.startOutgoing(r.Context(), peer, false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.broker.upsertCall(CallRecord{
		SessionID: sess.id, CallID: callID, Owner: &owner, Direction: "outbound", Peer: peer.String(),
		StartedAt: time.Now().UnixMilli(), Status: StatusRinging,
	})
	writeJSON(w, http.StatusOK, map[string]any{"call": map[string]string{"callId": callID}})
}

func (s *server) doWebRTC(sess *Session, w http.ResponseWriter, r *http.Request) {
	callID := r.PathValue("id")
	ac, ok := sess.reg.get(callID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such call"})
		return
	}
	var body struct {
		SDPOffer string `json:"sdp_offer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.SDPOffer == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sdp_offer required"})
		return
	}
	bridge, answer, err := NewBridge(body.SDPOffer, s.log, s.natIP)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	bridge.OnBrowserPCM = func(pcm []float32) {
		ac.cm.FeedCapturedPCM(pcm)
	}
	bridge.OnTerminalICE = func() {
		go sess.terminateCall(callID, core.EndCallReasonUserEnded)
	}
	sess.setBridge(callID, bridge)
	writeJSON(w, http.StatusOK, map[string]string{"sdp_answer": answer})
}

func (s *server) doAccept(sess *Session, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ac, ok := sess.reg.get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such call"})
		return
	}
	owner := clientID(r)
	if other := s.broker.ownerActiveCall(owner); other != "" && other != id {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "operator already on a call"})
		return
	}
	if !s.broker.setOwner(id, owner) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "claimed by another client"})
		return
	}
	s.broker.emitIncomingClaimed(sess.id, id, owner)
	if err := ac.cm.AcceptCall(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"call": map[string]string{"callId": id}})
}

func (s *server) doReject(sess *Session, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if ac, ok := sess.reg.get(id); ok {
		_ = ac.cm.RejectCall(r.Context(), id, core.EndCallReasonDeclined)
	}
	sess.removeCall(id)
	s.broker.endCall(id, string(core.EndCallReasonDeclined))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) doEndCall(sess *Session, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if ac, ok := sess.reg.get(id); ok {
		_ = ac.cm.EndCall(r.Context(), core.EndCallReasonUserEnded)
	}
	sess.removeCall(id)
	s.broker.endCall(id, string(core.EndCallReasonUserEnded))
	w.WriteHeader(http.StatusNoContent)
}

func normalizePhone(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "+")
	var b strings.Builder
	for _, c := range p {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	return b.String()
}
