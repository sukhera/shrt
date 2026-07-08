package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

// cacheTTLNoExpiry is the cache lifetime for links that never expire.
const cacheTTLNoExpiry = 24 * time.Hour

// Link is the store-level view of a shortened link used by the redirect path.
type Link struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// slugCacheKey returns the Redis key for a slug's cached link payload.
func slugCacheKey(slug string) string {
	return "slug:" + slug
}

// GetBySlug resolves a slug to its link using a cache-aside strategy:
//
//  1. Look up slug:<slug> in Redis. On hit, return the cached link.
//  2. On miss, query Postgres. If found, populate the cache with a TTL of
//     (expires_at - now) or 24h when there is no expiry.
//
// Redis failures never block the lookup — they are logged and the call falls
// through to Postgres. A link past its expiry is returned wrapped in
// ErrExpired so the handler can respond 410. An absent slug yields ErrNotFound.
func (s *Store) GetBySlug(ctx context.Context, slug string) (*Link, error) {
	link, ok := s.readSlugCache(ctx, slug)
	if !ok {
		var err error
		link, err = s.loadSlugFromDB(ctx, slug)
		if err != nil {
			return nil, err
		}
		s.writeSlugCache(ctx, link)
	}

	if link.ExpiresAt != nil && !link.ExpiresAt.After(time.Now()) {
		return nil, ErrExpired
	}
	return link, nil
}

// readSlugCache returns the cached link for slug, or ok=false on miss or any
// cache error (errors are logged, not propagated).
func (s *Store) readSlugCache(ctx context.Context, slug string) (*Link, bool) {
	raw, err := s.rdb.Get(ctx, slugCacheKey(slug)).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			slog.Warn("redis get failed, falling through to db", "slug", slug, "err", err)
		}
		return nil, false
	}
	var link Link
	if err := json.Unmarshal(raw, &link); err != nil {
		slog.Warn("corrupt cache entry, falling through to db", "slug", slug, "err", err)
		return nil, false
	}
	return &link, true
}

// loadSlugFromDB reads an active link from Postgres, translating a no-rows
// result into ErrNotFound.
func (s *Store) loadSlugFromDB(ctx context.Context, slug string) (*Link, error) {
	row, err := s.queries.GetLinkBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get link by slug: %w", err)
	}

	link := &Link{ID: row.ID.String(), Slug: row.Slug, OriginalURL: row.OriginalUrl}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		link.ExpiresAt = &t
	}
	return link, nil
}

// writeSlugCache stores link in Redis with a TTL derived from its expiry.
// Failures are logged and ignored — caching is best-effort.
func (s *Store) writeSlugCache(ctx context.Context, link *Link) {
	ttl := cacheTTLNoExpiry
	if link.ExpiresAt != nil {
		ttl = time.Until(*link.ExpiresAt)
		if ttl <= 0 {
			return // already expired; nothing worth caching
		}
	}

	payload, err := json.Marshal(link)
	if err != nil {
		slog.Warn("marshal cache entry failed", "slug", link.Slug, "err", err)
		return
	}
	if err := s.rdb.Set(ctx, slugCacheKey(link.Slug), payload, ttl).Err(); err != nil {
		slog.Warn("redis set failed", "slug", link.Slug, "err", err)
	}
}
