package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSessionStoreRoundtrip(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "sessions_test.db")
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	st, err := newSessionStore(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	id := newSessionID()
	if len(id) != 32 {
		t.Fatalf("session id should be 32 hex chars, got %d", len(id))
	}
	apiKey := newAPIKey()
	if len(apiKey) != 44 || apiKey[:4] != "wac_" {
		t.Fatalf("api key should start with wac_ and be 44 chars, got %q (%d)", apiKey, len(apiKey))
	}
	if err := st.insert(ctx, id, "Account A", apiKey); err != nil {
		t.Fatal(err)
	}

	rows, err := st.list(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != id || rows[0].Name != "Account A" || rows[0].JID != "" || rows[0].APIKey != apiKey {
		t.Fatalf("unexpected rows after insert: %+v", rows)
	}

	fetched, err := st.getByKey(ctx, apiKey)
	if err != nil {
		t.Fatal(err)
	}
	if fetched.ID != id || fetched.Name != "Account A" {
		t.Fatalf("getByKey returned wrong row: %+v", fetched)
	}

	if err := st.setJID(ctx, id, "5511999999999:1@s.whatsapp.net"); err != nil {
		t.Fatal(err)
	}
	rows, _ = st.list(ctx)
	if rows[0].JID != "5511999999999:1@s.whatsapp.net" {
		t.Fatalf("jid not persisted: %+v", rows[0])
	}

	newKey := "wac_custom_key_12345"
	if err := st.updateAPIKey(ctx, id, newKey); err != nil {
		t.Fatal(err)
	}
	fetched, _ = st.getByKey(ctx, newKey)
	if fetched == nil || fetched.ID != id {
		t.Fatalf("updateAPIKey did not persist: %+v", fetched)
	}

	if err := st.updateName(ctx, id, "Account B"); err != nil {
		t.Fatal(err)
	}
	rows, _ = st.list(ctx)
	if rows[0].Name != "Account B" {
		t.Fatalf("updateName did not persist: %+v", rows[0])
	}

	if err := st.delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	rows, _ = st.list(ctx)
	if len(rows) != 0 {
		t.Fatalf("expected empty after delete, got %+v", rows)
	}
}
