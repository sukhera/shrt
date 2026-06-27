# shrt — Backend Context

The golang-expert skill/agent covers all Go best practices and coding standards.
This file covers project-specific decisions only — not general Go knowledge.

## Structure

```
backend/
├── cmd/shrt/main.go        # entry point — wire deps, call server.New().Start()
├── server/                 # HTTP handlers, route registration, middleware
│   ├── server.go           # Server struct, NewServer(), Start(), routes()
│   ├── auth.go             # register/login/refresh/logout handlers
│   ├── link.go             # link CRUD handlers
│   ├── redirect.go         # GET /:slug handler
│   ├── middleware.go        # auth middleware, rate limiting, CORS
│   └── response.go         # respondJSON(), respondError(), error→HTTP mapping
├── store/                  # DB + Redis + all business logic
│   ├── store.go            # Store struct, NewStore()
│   ├── link.go             # CreateLink, GetBySlug, ListByUser, Update, Delete
│   ├── user.go             # CreateUser, GetByEmail
│   ├── token.go            # CreateRefreshToken, Revoke, DeleteExpired
│   └── errors.go           # ErrNotFound, ErrAliasTaken, ErrUnsafeURL, etc.
├── internal/config/        # env loading only — nothing else goes here
├── db/migrations/          # golang-migrate SQL files
└── db/queries/             # sqlc .sql files (never edit generated output)
```

## Libraries

| Purpose | Library |
|---------|---------|
| Router | `go-chi/chi/v5` |
| DB driver | `jackc/pgx/v5` (pgxpool) |
| DB queries | `sqlc-dev/sqlc` (generated — never hand-edit) |
| Redis | `redis/go-redis/v9` |
| Migrations | `golang-migrate/migrate/v4` |
| JWT | `golang-jwt/jwt/v5` RS256 |
| Validation | `go-playground/validator/v10` |
| Logging | `log/slog` (stdlib) |

## Key decisions

- **No service layer.** `store/` owns business logic AND data access. Handlers call store directly.
- **Slug generation** lives in `store/link.go` — 7-char base62, retry up to 5x on collision.
- **Redis cache** key: `slug:<slug>` → original URL. TTL = `expires_at - now()` or 24h if no expiry.
- **Cache failures** must never block redirects — log warning, fall through to Postgres.
- **Rate limiting** via Redis INCR+EXPIRE. On Redis failure, allow request and log.
- **Default redirect** is HTTP 302 (not 301) — preserves future analytics tracking.
- **Safe Browsing** check in `store/link.go` before insert. If API key is empty, log warning and skip.
- **Sentinel errors** defined in `store/errors.go`, mapped to HTTP codes in `server/response.go`.
- **HTTP server timeouts** must be set: ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout.
- **Health endpoint** at `GET /health` — checks DB + Redis connectivity.

## Config

All env vars loaded in `internal/config/config.go`. App panics at startup on missing required vars.
See `.env.example` for all variables. Never call `os.Getenv` outside `internal/config/`.

## Running locally

```bash
make docker-up      # start Postgres + Redis
make migrate-up     # run DB migrations
make dev            # hot reload via air
go test -race ./... # always run with race detector
```

## CI must pass

```bash
go fmt ./...
go vet ./...
golangci-lint run ./...
go test -race ./...
go build ./cmd/shrt
```
