package store

import "errors"

// Sentinel errors returned by the store. Handlers match these with errors.Is
// and map them to HTTP responses — never compare error strings.
var (
	// ErrNotFound is returned when a requested record does not exist (or is
	// soft-deleted).
	ErrNotFound = errors.New("store: not found")

	// ErrExpired is returned when a link exists but its expires_at is in the past.
	ErrExpired = errors.New("store: link expired")
)
