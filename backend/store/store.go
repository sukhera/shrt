// Package store owns all data access and business logic for shrt: Postgres
// queries, Redis caching, slug generation, Safe Browsing, and auth. The HTTP
// layer (server/) calls into this package and never touches the database or
// cache directly.
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sukhera/shrt/backend/db"
	"github.com/sukhera/shrt/backend/internal/config"
)

// Store holds the database pool and Redis client and exposes the application's
// business operations.
type Store struct {
	cfg     *config.Config
	pool    *pgxpool.Pool
	rdb     *redis.Client
	queries *db.Queries
}

// New connects to Postgres and Redis and returns a ready Store. The caller is
// responsible for calling Close when done.
func New(ctx context.Context, cfg *config.Config) (*Store, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		pool.Close()
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Store{
		cfg:     cfg,
		pool:    pool,
		rdb:     rdb,
		queries: db.New(pool),
	}, nil
}

// Close releases the database and Redis connections.
func (s *Store) Close() {
	s.pool.Close()
	_ = s.rdb.Close()
}

// Ping checks connectivity to Postgres and Redis. Used by the health endpoint.
func (s *Store) Ping(ctx context.Context) error {
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	return nil
}
