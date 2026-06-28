package server

import "context"

// ctxKey is an unexported type for context keys defined in this package, so they
// never collide with keys from other packages.
type ctxKey int

const userIDKey ctxKey = iota

// withUserID returns a copy of ctx carrying the authenticated user's ID. The
// auth middleware (added in M3) sets this; link handlers read it.
func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// userIDFromContext returns the authenticated user's ID and whether one is
// present. Handlers that require auth treat a missing ID as unauthenticated.
func userIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}
