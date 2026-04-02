//go:build !sqlite

package storage

import (
	"fmt"

	"github.com/conall/outalator/internal/config"
	"github.com/conall/outalator/internal/storage/postgres"
)

// New creates a Storage instance based on cfg.Driver.
// By default only "postgres" is supported. Build with -tags sqlite to also
// enable "sqlite".
func New(cfg config.DatabaseConfig) (Storage, error) {
	switch cfg.Driver {
	case DriverSQLite:
		return nil, fmt.Errorf(
			"sqlite support is not compiled in; rebuild with: go build -tags sqlite",
		)
	case DriverPostgres, "":
		return postgres.NewStorage(cfg)
	default:
		return nil, fmt.Errorf("unknown storage driver %q; supported: %s, %s", cfg.Driver, DriverPostgres, DriverSQLite)
	}
}
