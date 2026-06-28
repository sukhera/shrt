package store

import (
	"context"
	"log/slog"
	"time"
)

// RateLimit reports whether a request under the given key is allowed within a
// fixed window, using the Redis INCR+EXPIRE pattern: the first request in a
// window sets the counter's TTL, and the window resets when the key expires.
//
// limit is the maximum number of requests permitted per window. The boolean
// return is true when the request is allowed. Redis failures fail open — the
// request is allowed and the error logged — so an outage degrades to no limiting
// rather than a hard outage.
func (s *Store) RateLimit(ctx context.Context, key string, limit int, window time.Duration) bool {
	redisKey := "ratelimit:" + key

	count, err := s.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		slog.Warn("rate limit incr failed, allowing request", "key", key, "err", err)
		return true
	}

	// Only the first request in a window (count == 1) sets the expiry, so the
	// window is anchored to the first hit rather than sliding on every request.
	if count == 1 {
		if err := s.rdb.Expire(ctx, redisKey, window).Err(); err != nil {
			slog.Warn("rate limit expire failed", "key", key, "err", err)
		}
	}

	return count <= int64(limit)
}
