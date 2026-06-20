package auth

import "context"

// WithUser returns a copy of ctx with u stored under the auth context key.
// This is intended for use in tests that need to inject an authenticated user
// without a live session store.
func WithUser(ctx context.Context, u *UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}
