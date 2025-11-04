package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresStorage implements the Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// Config holds PostgreSQL connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// New creates a new PostgreSQL storage instance
func New(cfg Config) (*PostgresStorage, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// marshalJSONMap safely marshals a map, returning {} for nil maps instead of null
func marshalJSONMap(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// marshalJSONAny safely marshals any value, returning {} for nil maps instead of null
func marshalJSONAny(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	// Check if it's a nil map
	if m, ok := v.(map[string]any); ok && m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(v)
}
