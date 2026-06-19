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

// scanFunc is the common signature for (*sql.Row).Scan and (*sql.Rows).Scan,
// used by all per-entity scan helpers in this package.
type scanFunc func(dest ...any) error

// schema is the DDL applied automatically on first open (all statements are
// idempotent via CREATE IF NOT EXISTS).
//
//go:embed schema.sql
var schema string

// NewStorage creates an SQLiteStorage from an application DatabaseConfig.
// cfg.Path is the file path (e.g. "outalator.db" or ":memory:").
func NewStorage(ctx context.Context, cfg config.DatabaseConfig) (*SQLiteStorage, error) {
	path := cfg.Path
	if path == "" {
		path = "outalator.db"
	}
	return New(ctx, path)
}

// New opens (or creates) the SQLite database at path and runs schema migrations.
// ctx is used for the ping and schema migration; cancelling it aborts startup.
func New(ctx context.Context, path string) (*SQLiteStorage, error) {
	// Reject bare paths containing '?' or '&' to prevent accidental DSN
	// parameter injection from operator-supplied configuration (e.g. DB_PATH).
	// Callers that need advanced URI options should pass a "file:" URI directly.
	if !strings.HasPrefix(path, "file:") && path != ":memory:" {
		if strings.ContainsAny(path, "?&") {
			return nil, fmt.Errorf(
				"sqlite path %q must not contain '?' or '&'; use a file: URI for advanced DSN options",
				path,
			)
		}
	}

	dsn := buildDSN(path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite supports only one writer at a time; cap the pool to avoid
	// SQLITE_BUSY contention.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.migrate(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate sqlite database: %w", err)
	}

	return s, nil
}

// buildDSN builds the DSN for modernc.org/sqlite.
// The _pragma=... query-parameter syntax is specific to this driver; swapping
// the driver will require revisiting this format.
func buildDSN(path string) string {
	if path == ":memory:" {
		// WAL mode is not supported for in-memory databases.
		// Use mode=memory&cache=shared so that all pool connections share the
		// same in-memory database. Without cache=shared each new connection
		// gets its own isolated empty DB — safe with MaxOpenConns(1) today,
		// but fragile if the cap is ever removed.
		return "file::memory:?mode=memory&cache=shared&_pragma=foreign_keys(ON)"
	}
	// Enable foreign-key enforcement and WAL journal mode for file databases.
	// WAL improves concurrent read latency and is the standard recommendation
	// for any SQLite database accessed by more than one goroutine.
	pragmas := "_pragma=foreign_keys(ON)&_pragma=journal_mode(WAL)"
	if strings.HasPrefix(path, "file:") {
		if strings.Contains(path, "?") {
			return path + "&" + pragmas
		}
		return path + "?" + pragmas
	}
	// Bare file path — convert to SQLite file URI.
	return "file:" + path + "?" + pragmas
}

// migrate runs the embedded schema DDL (idempotent CREATE IF NOT EXISTS).
// Statements are executed one at a time because database/sql's ExecContext does
// not guarantee multi-statement support across all drivers.
func (s *SQLiteStorage) migrate(ctx context.Context) error {
	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("schema migration: %w", err)
		}
	}
	return nil
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
//
// Chan and Func are handled defensively (they cannot appear in domain structs)
// to prevent a panic from reflect.Value.IsNil on an unexpected type.
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
