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

	// ErrAliasTaken is returned when a requested custom alias is already in use,
	// or when random slug generation exhausts its retries (a near-impossible
	// signal that the keyspace is saturated).
	ErrAliasTaken = errors.New("store: alias already taken")

	// ErrInvalidURL is returned when a submitted URL is malformed or uses an
	// unsupported scheme.
	ErrInvalidURL = errors.New("store: invalid url")

	// ErrUnsafeURL is returned when a URL is flagged by Google Safe Browsing.
	ErrUnsafeURL = errors.New("store: unsafe url")

	// ErrForbidden is returned when a caller is authenticated but does not own
	// the targeted resource.
	ErrForbidden = errors.New("store: forbidden")

	// ErrEmailTaken is returned when registering with an email that already
	// exists.
	ErrEmailTaken = errors.New("store: email already registered")

	// ErrInvalidCredentials is returned when an email/password pair does not
	// match. It is intentionally vague (same error for unknown email and wrong
	// password) to avoid leaking which accounts exist.
	ErrInvalidCredentials = errors.New("store: invalid credentials")

	// ErrInvalidToken is returned when a JWT or refresh token is missing,
	// malformed, expired, revoked, or otherwise fails validation.
	ErrInvalidToken = errors.New("store: invalid token")
)
