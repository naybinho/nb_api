package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func (s *server) handleGetGroups(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	groups, err := sess.client.GetJoinedGroups(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"groups": groups})
}

func (s *server) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req struct {
		Name         string   `json:"name"`
		Participants []string `json:"participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	jids := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		jids[i] = resolveJID(p)
	}

	info, err := sess.client.CreateGroup(r.Context(), whatsmeow.ReqCreateGroup{
		Name:         req.Name,
		Participants: jids,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"group": info})
}

func (s *server) handleGetGroupInfo(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	info, err := sess.client.GetGroupInfo(r.Context(), gid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"group": info})
}

// handleUpdateGroup updates group name, description (topic), locked and/or announce settings.
//
// Body (all fields optional):
//
//	{
//	  "name":     "New group name",
//	  "topic":    "New description",
//	  "locked":   true,    // restrict metadata changes to admins
//	  "announce": true     // restrict sending messages to admins
//	}
func (s *server) handleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req struct {
		Name     *string `json:"name"`
		Topic    *string `json:"topic"`
		Locked   *bool   `json:"locked"`
		Announce *bool   `json:"announce"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	var lastErr error

	if req.Name != nil {
		if err := sess.client.SetGroupName(r.Context(), gid, *req.Name); err != nil {
			lastErr = err
		}
	}
	if req.Topic != nil {
		// SetGroupTopic(ctx, jid, previousID, previousSetAt, topic)
		// Pass empty strings / zero values for the previous-state args to let the server handle it.
		if err := sess.client.SetGroupTopic(r.Context(), gid, "", "", *req.Topic); err != nil {
			lastErr = err
		}
	}
	if req.Locked != nil {
		if err := sess.client.SetGroupLocked(r.Context(), gid, *req.Locked); err != nil {
			lastErr = err
		}
	}
	if req.Announce != nil {
		if err := sess.client.SetGroupAnnounce(r.Context(), gid, *req.Announce); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": lastErr.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleLeaveGroup(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	err := sess.client.LeaveGroup(r.Context(), gid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleGetGroupInvite(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	link, err := sess.client.GetGroupInviteLink(r.Context(), gid, false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"link": link})
}

func (s *server) handleRevokeGroupInvite(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	link, err := sess.client.GetGroupInviteLink(r.Context(), gid, true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"link": link})
}

func (s *server) handleJoinGroup(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	var req struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	gid, err := sess.client.JoinGroupWithLink(r.Context(), req.Link)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"groupId": gid.String()})
}

// handleUpdateGroupParticipants handles add/remove actions for group participants.
func (s *server) handleUpdateGroupParticipants(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	var req struct {
		Action       string   `json:"action"`       // "add" or "remove"
		Participants []string `json:"participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	jids := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		jids[i] = resolveJID(p)
	}

	participants, err := sess.client.UpdateGroupParticipants(r.Context(), gid, jids, whatsmeow.ParticipantChange(req.Action))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"participants": participants})
}

// handlePromoteDemoteGroupParticipants handles promote/demote of a single participant.
//
// Body: { "action": "promote" | "demote" }
func (s *server) handlePromoteDemoteGroupParticipants(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	jid := resolveJID(r.PathValue("jid"))

	var req struct {
		Action string `json:"action"` // "promote" or "demote"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Action != "promote" && req.Action != "demote") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'promote' or 'demote'"})
		return
	}

	participants, err := sess.client.UpdateGroupParticipants(r.Context(), gid, []types.JID{jid}, whatsmeow.ParticipantChange(req.Action))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"participants": participants})
}

// handleGetGroupRequests returns the list of users who have requested to join a group.
func (s *server) handleGetGroupRequests(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	participants, err := sess.client.GetGroupRequestParticipants(r.Context(), gid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"participants": participants})
}

// handleApproveGroupRequests approves or rejects membership requests.
//
// Body:
//
//	{
//	  "participants": ["5511...@s.whatsapp.net"],
//	  "action": "approve" | "reject"
//	}
func (s *server) handleApproveGroupRequests(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	var req struct {
		Participants []string `json:"participants"`
		Action       string   `json:"action"` // "approve" or "reject"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Action != "approve" && req.Action != "reject" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'approve' or 'reject'"})
		return
	}

	jids := make([]types.JID, len(req.Participants))
	for i, p := range req.Participants {
		jids[i] = resolveJID(p)
	}

	participants, err := sess.client.UpdateGroupParticipants(r.Context(), gid, jids, whatsmeow.ParticipantChange(req.Action))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"participants": participants})
}

// handleSetGroupPhoto sets the group profile picture.
//
// Body (JSON): { "data": "<base64 JPEG bytes>" }
// Or multipart/form-data with a "photo" file field.
func (s *server) handleSetGroupPhoto(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	imgBytes, err := readImageBytes(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pictureID, err := sess.client.SetGroupPhoto(r.Context(), gid, imgBytes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pictureId": pictureID})
}

// handleDeleteGroupPhoto removes the group profile picture.
func (s *server) handleDeleteGroupPhoto(w http.ResponseWriter, r *http.Request) {
	sess := s.sessionByID(w, r.PathValue("sid"))
	if sess == nil || sess.client == nil {
		return
	}

	gid, _ := types.ParseJID(r.PathValue("gid"))
	_, err := sess.client.SetGroupPhoto(r.Context(), gid, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// readImageBytes reads image bytes from the request.
// It supports:
//  1. multipart/form-data with a "photo" field
//  2. JSON body with { "data": "<base64>" }
//  3. Raw binary body
func readImageBytes(r *http.Request) ([]byte, error) {
	ct := r.Header.Get("Content-Type")

	if isMultipart(ct) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
		f, _, err := r.FormFile("photo")
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Try JSON with base64
	var jsonReq struct {
		Data string `json:"data"`
	}
	if jsonErr := json.Unmarshal(body, &jsonReq); jsonErr == nil && jsonReq.Data != "" {
		decoded, err := base64.StdEncoding.DecodeString(jsonReq.Data)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(jsonReq.Data)
		}
		if err != nil {
			return nil, err
		}
		return decoded, nil
	}

	// Assume raw bytes
	return body, nil
}

func isMultipart(ct string) bool {
	for _, part := range []string{"multipart/form-data", "multipart/mixed"} {
		if len(ct) >= len(part) && ct[:len(part)] == part {
			return true
		}
	}
	return false
}
