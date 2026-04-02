//go:build sqlite

package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/conall/outalator/internal/config"
	_ "modernc.org/sqlite"
)

// SQLiteStorage implements the Storage interface using SQLite.
// Build with -tags sqlite to enable this backend.
type SQLiteStorage struct {
	db *sql.DB
}

// schema is the DDL applied automatically on first open (all statements are
// idempotent via CREATE IF NOT EXISTS).
//
//go:embed schema.sql
var schema string

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
	// Embed the foreign_keys pragma in the DSN so it is applied on every
	// connection the pool creates, not just the first one.
	dsn := buildDSN(path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite supports only one writer at a time; cap the pool to avoid
	// SQLITE_BUSY contention.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate sqlite database: %w", err)
	}

	return s, nil
}

// buildDSN appends the _pragma=foreign_keys(ON) query parameter so that every
// new pool connection has foreign-key enforcement enabled.
func buildDSN(path string) string {
	pragma := "_pragma=foreign_keys(ON)"
	if path == ":memory:" {
		// Use mode=memory&cache=shared so that all pool connections share the
		// same in-memory database. Without cache=shared each new connection
		// gets its own isolated empty DB — safe with MaxOpenConns(1) today,
		// but fragile if the cap is ever removed.
		return "file::memory:?mode=memory&cache=shared&" + pragma
	}
	if strings.HasPrefix(path, "file:") {
		if strings.Contains(path, "?") {
			return path + "&" + pragma
		}
		return path + "?" + pragma
	}
	// Bare file path — convert to SQLite file URI.
	return "file:" + path + "?" + pragma
}

// migrate runs the embedded schema DDL (idempotent CREATE IF NOT EXISTS).
func (s *SQLiteStorage) migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

// Close closes the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// marshalJSONMap safely marshals a string map, returning {} for nil maps.
func marshalJSONMap(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// marshalJSONAny safely marshals any value with sensible nil handling:
//   - untyped nil            → {}
//   - typed-nil map/ptr/etc. → {}
//   - typed-nil slice        → []  (avoids storing {} which can't deserialise back into a slice)
func marshalJSONAny(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return []byte("[]"), nil
		}
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		if rv.IsNil() {
			return []byte("{}"), nil
		}
	}
	return json.Marshal(v)
}
