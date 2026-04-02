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
	case DriverSQLite:
		return sqlite.NewStorage(cfg)
	case DriverPostgres, "":
		return postgres.NewStorage(cfg)
	default:
		return nil, fmt.Errorf("unknown storage driver %q; supported: %s, %s", cfg.Driver, DriverPostgres, DriverSQLite)
	}
}
