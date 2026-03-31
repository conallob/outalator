//go:build sqlite

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/conall/outalator/internal/config"
	_ "modernc.org/sqlite"
)

// SQLiteStorage implements the Storage interface using SQLite.
// Build with -tags sqlite to enable this backend.
type SQLiteStorage struct {
	db *sql.DB
}

// schema is the SQLite DDL executed on first open.
const schema = `
CREATE TABLE IF NOT EXISTS outages (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    status      TEXT NOT NULL,
    severity    TEXT NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    resolved_at DATETIME,
    metadata    TEXT NOT NULL DEFAULT '{}',
    custom_fields TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS alerts (
    id              TEXT PRIMARY KEY,
    outage_id       TEXT NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    external_id     TEXT NOT NULL,
    source          TEXT NOT NULL,
    team_name       TEXT,
    title           TEXT NOT NULL,
    description     TEXT,
    severity        TEXT,
    triggered_at    DATETIME NOT NULL,
    acknowledged_at DATETIME,
    resolved_at     DATETIME,
    created_at      DATETIME NOT NULL,
    source_metadata TEXT NOT NULL DEFAULT '{}',
    metadata        TEXT NOT NULL DEFAULT '{}',
    custom_fields   TEXT NOT NULL DEFAULT '{}',
    UNIQUE(external_id, source)
);

CREATE TABLE IF NOT EXISTS notes (
    id            TEXT PRIMARY KEY,
    outage_id     TEXT NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    content       TEXT NOT NULL,
    format        TEXT NOT NULL DEFAULT 'plaintext',
    author        TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    metadata      TEXT NOT NULL DEFAULT '{}',
    custom_fields TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS tags (
    id            TEXT PRIMARY KEY,
    outage_id     TEXT NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    key           TEXT NOT NULL,
    value         TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    custom_fields TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_outages_created_at ON outages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_outages_status     ON outages(status);
CREATE INDEX IF NOT EXISTS idx_outages_severity   ON outages(severity);

CREATE INDEX IF NOT EXISTS idx_alerts_outage_id    ON alerts(outage_id);
CREATE INDEX IF NOT EXISTS idx_alerts_external_id  ON alerts(external_id, source);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alerts(triggered_at DESC);

CREATE INDEX IF NOT EXISTS idx_notes_outage_id  ON notes(outage_id);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tags_outage_id  ON tags(outage_id);
CREATE INDEX IF NOT EXISTS idx_tags_key_value  ON tags(key, value);
`

// NewStorage creates an SQLiteStorage from an application DatabaseConfig.
// cfg.Path is the file path (e.g. "outalator.db" or ":memory:").
func NewStorage(cfg config.DatabaseConfig) (*SQLiteStorage, error) {
	path := cfg.Path
	if path == "" {
		path = "outalator.db"
	}
	return New(path)
}

// New opens (or creates) the SQLite database at path and runs schema migrations.
func New(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite foreign-key enforcement is off by default; enable it.
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Limit to a single writer connection to avoid SQLITE_BUSY errors.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate sqlite database: %w", err)
	}

	return s, nil
}

// migrate runs the embedded schema DDL (idempotent CREATE IF NOT EXISTS).
func (s *SQLiteStorage) migrate() error {
	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// marshalJSONMap safely marshals a map, returning {} for nil maps.
func marshalJSONMap(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// marshalJSONAny safely marshals any value, returning {} for nil.
func marshalJSONAny(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	if m, ok := v.(map[string]any); ok && m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(v)
}
