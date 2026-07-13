package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

type sessionRow struct {
	ID     string
	Name   string
	JID    string
	APIKey string
}

type sessionStore struct{ db *sql.DB }

func newSessionStore(ctx context.Context, db *sql.DB) (*sessionStore, error) {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS sessions (
		id      VARCHAR(255) PRIMARY KEY,
		name    VARCHAR(255) NOT NULL,
		jid     VARCHAR(255),
		api_key VARCHAR(255) NOT NULL DEFAULT ''
	)`)
	if err != nil {
		return nil, err
	}
	_, _ = db.ExecContext(ctx, `ALTER TABLE sessions ADD COLUMN IF NOT EXISTS api_key VARCHAR(255) NOT NULL DEFAULT ''`)
	st := &sessionStore{db: db}
	if err := st.fillMissingAPIKeys(ctx); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *sessionStore) fillMissingAPIKeys(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM sessions WHERE api_key = '' OR api_key IS NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range ids {
		key := newAPIKey()
		if _, err := s.db.ExecContext(ctx, `UPDATE sessions SET api_key = $1 WHERE id = $2`, key, id); err != nil {
			return err
		}
	}
	return nil
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func newAPIKey() string {
	b := make([]byte, 20)
	rand.Read(b)
	return "wac_" + hex.EncodeToString(b)
}

func (s *sessionStore) list(ctx context.Context) ([]sessionRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, COALESCE(jid, ''), COALESCE(api_key, '') FROM sessions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []sessionRow
	for rows.Next() {
		var r sessionRow
		if err := rows.Scan(&r.ID, &r.Name, &r.JID, &r.APIKey); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *sessionStore) insert(ctx context.Context, id, name, apiKey string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sessions (id, name, jid, api_key) VALUES ($1, $2, NULL, $3)`, id, name, apiKey)
	return err
}

func (s *sessionStore) getByKey(ctx context.Context, apiKey string) (*sessionRow, error) {
	var r sessionRow
	err := s.db.QueryRowContext(ctx, `SELECT id, name, COALESCE(jid, ''), COALESCE(api_key, '') FROM sessions WHERE api_key = $1`, apiKey).Scan(&r.ID, &r.Name, &r.JID, &r.APIKey)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *sessionStore) updateAPIKey(ctx context.Context, id, apiKey string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET api_key = $1 WHERE id = $2`, apiKey, id)
	return err
}

func (s *sessionStore) updateName(ctx context.Context, id, name string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET name = $1 WHERE id = $2`, name, id)
	return err
}

func (s *sessionStore) setJID(ctx context.Context, id, jid string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET jid = $1 WHERE id = $2`, jid, id)
	return err
}

func (s *sessionStore) delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}
