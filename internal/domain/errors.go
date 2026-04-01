package domain

import "errors"

// ErrNotFound is returned by storage implementations when a requested entity
// does not exist. Callers may use errors.Is to distinguish not-found from
// other storage errors (e.g. to return HTTP 404 vs 500).
var ErrNotFound = errors.New("not found")
