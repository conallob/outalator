package testutil

import (
	"context"

	"github.com/conall/outalator/internal/auth"
)

// WithUser returns a copy of ctx with u stored under the auth context key.
// Intended for tests that need to inject an authenticated user without a
// live session store.
func WithUser(ctx context.Context, u *auth.UserInfo) context.Context {
	return context.WithValue(ctx, auth.UserContextKey, u)
}
