package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sukhera/shrt/backend/internal/config"
	"github.com/sukhera/shrt/backend/store"
)

// testDSN returns the Postgres/Redis URLs for integration tests, falling back
// to the local docker-compose defaults.
func testDSN() (string, string) {
	pg := os.Getenv("TEST_DATABASE_URL")
	if pg == "" {
		pg = "postgres://shrt:shrt@localhost:5432/shrt?sslmode=disable"
	}
	rd := os.Getenv("TEST_REDIS_URL")
	if rd == "" {
		rd = "redis://localhost:6379"
	}
	return pg, rd
}

// newTestServer connects to the local infra and returns a Server plus a raw
// pool for seeding. It skips the test if the infra is not reachable.
func newTestServer(t *testing.T) (*Server, *pgxpool.Pool) {
	t.Helper()
	pgURL, redisURL := testDSN()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pgURL)
	if err != nil || pool.Ping(ctx) != nil {
		t.Skip("skipping integration test: Postgres not reachable at " + pgURL)
	}

	cfg := &config.Config{
		Port:                "8080",
		Env:                 "test",
		BaseURL:             "http://localhost:8080",
		FrontendURL:         "http://localhost:3000",
		DatabaseURL:         pgURL,
		RedisURL:            redisURL,
		DefaultRedirectCode: 302,
		SlugLength:          7,
		RateLimitAnon:       1000, // generous by default; rate-limit tests lower it
		RateLimitUser:       1000,
	}

	st, err := store.New(ctx, cfg)
	if err != nil {
		pool.Close()
		t.Skip("skipping integration test: store init failed: " + err.Error())
	}
	t.Cleanup(st.Close)

	ensureSchema(t, pool)
	flushRedis(t, redisURL)

	return New(cfg, st), pool
}

// ensureSchema creates the links table if it is absent so the test can run
// against a fresh database.
func ensureSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS links (
			id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id      UUID,
			slug         TEXT        NOT NULL,
			original_url TEXT        NOT NULL,
			is_custom    BOOLEAN     NOT NULL DEFAULT FALSE,
			expires_at   TIMESTAMPTZ,
			deleted_at   TIMESTAMPTZ,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_links_slug_active ON links (slug) WHERE deleted_at IS NULL;`)
	if err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
}

func flushRedis(t *testing.T, redisURL string) {
	t.Helper()
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return
	}
	rdb := redis.NewClient(opt)
	defer func() { _ = rdb.Close() }()
	rdb.FlushDB(context.Background())
}

// seedLink inserts a link row and returns a cleanup func.
func seedLink(t *testing.T, pool *pgxpool.Pool, slug, url string, expiresAt *time.Time, deleted bool) {
	t.Helper()
	var deletedAt *time.Time
	if deleted {
		now := time.Now()
		deletedAt = &now
	}
	_, err := pool.Exec(context.Background(),
		`INSERT INTO links (slug, original_url, expires_at, deleted_at) VALUES ($1, $2, $3, $4)`,
		slug, url, expiresAt, deletedAt)
	if err != nil {
		t.Fatalf("seed link %q: %v", slug, err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM links WHERE slug = $1`, slug)
	})
}

func TestRedirect(t *testing.T) {
	srv, pool := newTestServer(t)

	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	seedLink(t, pool, "valid01", "https://example.com/permanent", nil, false)
	seedLink(t, pool, "valid02", "https://example.com/temporary", &future, false)
	seedLink(t, pool, "gone001", "https://example.com/expired", &past, false)
	seedLink(t, pool, "del0001", "https://example.com/deleted", nil, true)

	tests := []struct {
		name       string
		slug       string
		wantStatus int
		wantLoc    string
	}{
		{"permanent link → 302 (default code)", "valid01", http.StatusFound, "https://example.com/permanent"},
		{"link with expiry → 302", "valid02", http.StatusFound, "https://example.com/temporary"},
		{"expired link → 410", "gone001", http.StatusGone, ""},
		{"deleted link → 404", "del0001", http.StatusNotFound, ""},
		{"unknown slug → 404", "nosuch1", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.slug, nil)
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantLoc != "" {
				if loc := rec.Header().Get("Location"); loc != tt.wantLoc {
					t.Errorf("Location: got %q, want %q", loc, tt.wantLoc)
				}
			}
		})
	}
}

func TestRedirect_PermanentUsesConfiguredCode(t *testing.T) {
	srv, pool := newTestServer(t)
	// Override the redirect code to 301 for a permanent link.
	srv.cfg.DefaultRedirectCode = http.StatusMovedPermanently

	seedLink(t, pool, "perm301", "https://example.com/p", nil, false)

	req := httptest.NewRequest(http.MethodGet, "/perm301", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("status: got %d, want 301", rec.Code)
	}
}
