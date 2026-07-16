package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

type webhookRow struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	URL       string    `json:"url"`
	Events    string    `json:"events"`
	Enabled   bool      `json:"enabled"`
	Secret    string    `json:"secret,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type webhookStore struct {
	db *sql.DB
}

func newWebhookStore(ctx context.Context, db *sql.DB) (*webhookStore, error) {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS webhooks (
		id         VARCHAR(255) PRIMARY KEY,
		session_id VARCHAR(255) NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		url        TEXT NOT NULL,
		events     TEXT NOT NULL DEFAULT '*',
		enabled    BOOLEAN NOT NULL DEFAULT true,
		secret     VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		return nil, err
	}
	return &webhookStore{db: db}, nil
}

func newWebhookID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return "wh_" + hex.EncodeToString(b)
}

func (s *webhookStore) list(ctx context.Context, sessionID string) ([]webhookRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, url, events, enabled, secret, created_at, updated_at
		 FROM webhooks WHERE session_id = $1 ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []webhookRow
	for rows.Next() {
		var r webhookRow
		if err := rows.Scan(&r.ID, &r.SessionID, &r.URL, &r.Events, &r.Enabled, &r.Secret, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *webhookStore) get(ctx context.Context, id string) (*webhookRow, error) {
	var r webhookRow
	err := s.db.QueryRowContext(ctx,
		`SELECT id, session_id, url, events, enabled, secret, created_at, updated_at
		 FROM webhooks WHERE id = $1`, id).
		Scan(&r.ID, &r.SessionID, &r.URL, &r.Events, &r.Enabled, &r.Secret, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *webhookStore) insert(ctx context.Context, r *webhookRow) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO webhooks (id, session_id, url, events, enabled, secret, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		r.ID, r.SessionID, r.URL, r.Events, r.Enabled, r.Secret, r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *webhookStore) update(ctx context.Context, r *webhookRow) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE webhooks SET url = $1, events = $2, enabled = $3, secret = $4, updated_at = $5 WHERE id = $6`,
		r.URL, r.Events, r.Enabled, r.Secret, r.UpdatedAt, r.ID)
	return err
}

func (s *webhookStore) delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	return err
}

// listBySessionAndEvent returns all enabled webhooks for a session that
// subscribe to the given event type (or '*' for all events).
func (s *webhookStore) listBySessionAndEvent(ctx context.Context, sessionID, evtType string) ([]webhookRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, url, events, enabled, secret, created_at, updated_at
		 FROM webhooks
		 WHERE session_id = $1 AND enabled = true
		   AND (events = '*' OR events LIKE $2 OR events LIKE $3 OR events LIKE $4)`,
		sessionID,
		evtType+",%",   // starts with event
		"%,"+evtType,   // ends with event
		"%,"+evtType+",%", // middle of list
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []webhookRow
	for rows.Next() {
		var r webhookRow
		if err := rows.Scan(&r.ID, &r.SessionID, &r.URL, &r.Events, &r.Enabled, &r.Secret, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// listAllEnabled returns all enabled webhooks across all sessions.
// Used by the dispatcher to check which session webhooks are active.
func (s *webhookStore) listAllEnabled(ctx context.Context) ([]webhookRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, url, events, enabled, secret, created_at, updated_at
		 FROM webhooks WHERE enabled = true ORDER BY session_id, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []webhookRow
	for rows.Next() {
		var r webhookRow
		if err := rows.Scan(&r.ID, &r.SessionID, &r.URL, &r.Events, &r.Enabled, &r.Secret, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
