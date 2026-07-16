package main

import (
	"context"
	"database/sql"
	"time"
)

// CallHistoryRow represents a persisted call history record.
type CallHistoryRow struct {
	CallID       string     `json:"callId"`
	SessionID    string     `json:"sessionId"`
	Peer         string     `json:"peer"`
	Direction    string     `json:"direction"`
	StartedAt    int64      `json:"startedAt"`
	EndedAt      *int64     `json:"endedAt,omitempty"`
	EndReason    string     `json:"endReason,omitempty"`
	Recorded     bool       `json:"recorded"`
	RecordingURL string     `json:"recordingUrl,omitempty"`
}

type callHistoryStore struct{ db *sql.DB }

func newCallHistoryStore(ctx context.Context, db *sql.DB) (*callHistoryStore, error) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS call_history (
			call_id      VARCHAR(255) PRIMARY KEY,
			session_id   VARCHAR(255) NOT NULL,
			peer         VARCHAR(255) NOT NULL,
			direction    VARCHAR(20) NOT NULL,
			started_at   BIGINT NOT NULL,
			ended_at     BIGINT,
			end_reason   VARCHAR(100) NOT NULL DEFAULT '',
			recorded     BOOLEAN NOT NULL DEFAULT false,
			recording_url TEXT NOT NULL DEFAULT ''
		)
	`)
	if err != nil {
		return nil, err
	}

	// Create index for faster queries by session
	_, _ = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_call_history_session
		ON call_history(session_id, started_at DESC)
	`)

	// Add columns if missing (for upgrades)
	_, _ = db.ExecContext(ctx, `ALTER TABLE call_history ADD COLUMN IF NOT EXISTS recorded BOOLEAN NOT NULL DEFAULT false`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE call_history ADD COLUMN IF NOT EXISTS recording_url TEXT NOT NULL DEFAULT ''`)

	return &callHistoryStore{db: db}, nil
}

func (s *callHistoryStore) insert(ctx context.Context, row *CallHistoryRow) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO call_history (call_id, session_id, peer, direction, started_at, ended_at, end_reason, recorded, recording_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (call_id) DO UPDATE SET
			ended_at = COALESCE($6, call_history.ended_at),
			end_reason = CASE WHEN $7 <> '' THEN $7 ELSE call_history.end_reason END,
			recorded = $8,
			recording_url = CASE WHEN $9 <> '' THEN $9 ELSE call_history.recording_url END
	`,
		row.CallID, row.SessionID, row.Peer, row.Direction,
		row.StartedAt, row.EndedAt, row.EndReason,
		row.Recorded, row.RecordingURL,
	)
	return err
}

func (s *callHistoryStore) updateEnded(ctx context.Context, callID string, endedAt int64, endReason string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE call_history SET ended_at = $1, end_reason = $2
		WHERE call_id = $3
	`, endedAt, endReason, callID)
	return err
}

func (s *callHistoryStore) updateRecordingURL(ctx context.Context, callID string, url string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE call_history SET recording_url = $1 WHERE call_id = $2
	`, url, callID)
	return err
}

func (s *callHistoryStore) listBySession(ctx context.Context, sessionID string, limit int) ([]CallHistoryRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT call_id, session_id, peer, direction, started_at, ended_at,
		       COALESCE(end_reason, ''), recorded, COALESCE(recording_url, '')
		FROM call_history
		WHERE session_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CallHistoryRow
	for rows.Next() {
		var r CallHistoryRow
		var endedAt sql.NullInt64
		if err := rows.Scan(&r.CallID, &r.SessionID, &r.Peer, &r.Direction,
			&r.StartedAt, &endedAt, &r.EndReason, &r.Recorded, &r.RecordingURL); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			r.EndedAt = &endedAt.Int64
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// getByID retrieves a single history record by call ID.
func (s *callHistoryStore) getByID(ctx context.Context, callID string) (*CallHistoryRow, error) {
	var r CallHistoryRow
	var endedAt sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT call_id, session_id, peer, direction, started_at, ended_at,
		       COALESCE(end_reason, ''), recorded, COALESCE(recording_url, '')
		FROM call_history
		WHERE call_id = $1
	`, callID).Scan(&r.CallID, &r.SessionID, &r.Peer, &r.Direction,
		&r.StartedAt, &endedAt, &r.EndReason, &r.Recorded, &r.RecordingURL)
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		r.EndedAt = &endedAt.Int64
	}
	return &r, nil
}

// cleanupOldEntries removes history older than the specified duration.
// Useful for periodic cleanup if desired.
func (s *callHistoryStore) cleanupOldEntries(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan).UnixMilli()
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM call_history WHERE started_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
