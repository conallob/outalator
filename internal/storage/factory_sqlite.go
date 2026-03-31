//go:build sqlite

package storage

import (
	"fmt"

	"github.com/conall/outalator/internal/config"
	"github.com/conall/outalator/internal/storage/postgres"
	"github.com/conall/outalator/internal/storage/sqlite"
)

// New creates a Storage instance based on cfg.Driver.
// "postgres" is the production backend; "sqlite" is for local testing.
func New(cfg config.DatabaseConfig) (Storage, error) {
	switch cfg.Driver {
	case "sqlite":
		return sqlite.NewStorage(cfg)
	case "postgres", "":
		return postgres.NewStorage(cfg)
	default:
		return nil, fmt.Errorf("unknown storage driver %q; supported: postgres, sqlite", cfg.Driver)
	}
}
