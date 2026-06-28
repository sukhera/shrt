package store

import (
	"context"
	"crypto/rand"
	"fmt"
)

// base62Alphabet is the slug character set: digits, uppercase, lowercase. Its
// length (62) divides 248 evenly (62 × 4), so rejection sampling against 248
// introduces no modulo bias.
const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// maxSlugAttempts bounds collision retries during random slug generation.
const maxSlugAttempts = 5

// randomSlug returns a cryptographically random base62 string of length n.
func randomSlug(n int) (string, error) {
	out := make([]byte, n)
	// Read more bytes than needed up front to amortise rejection sampling.
	buf := make([]byte, n+n/2+1)
	bi := len(buf) // force an initial fill

	for i := 0; i < n; {
		if bi >= len(buf) {
			if _, err := rand.Read(buf); err != nil {
				return "", fmt.Errorf("read random bytes: %w", err)
			}
			bi = 0
		}
		b := buf[bi]
		bi++
		// Reject the top 8 values (248–255) to keep the distribution uniform
		// across the 62-symbol alphabet.
		if b >= 248 {
			continue
		}
		out[i] = base62Alphabet[b%62]
		i++
	}
	return string(out), nil
}

// generateSlug produces a unique base62 slug of the configured length, retrying
// on collision up to maxSlugAttempts times. It returns ErrAliasTaken if every
// attempt collides — in practice this only happens if the keyspace is saturated.
func (s *Store) generateSlug(ctx context.Context) (string, error) {
	for attempt := 0; attempt < maxSlugAttempts; attempt++ {
		slug, err := randomSlug(s.cfg.SlugLength)
		if err != nil {
			return "", err
		}
		exists, err := s.queries.SlugExists(ctx, slug)
		if err != nil {
			return "", fmt.Errorf("check slug existence: %w", err)
		}
		if !exists {
			return slug, nil
		}
	}
	return "", ErrAliasTaken
}
