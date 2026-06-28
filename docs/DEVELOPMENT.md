# Development Guide

A deeper reference for working in the shrt codebase: environment, make targets,
migrations, code generation, testing, and common tasks. For first-time setup see
the [README quick start](../README.md#quick-start); for contribution rules see
[CONTRIBUTING.md](../CONTRIBUTING.md).

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.25+ | Backend |
| Node | 20+ | Frontend |
| Docker | recent | Postgres + Redis via Compose |
| `make` | — | Task runner |
| `openssl` | — | Generate the JWT key pair |

`make tools` installs the Go dev tooling: [air](https://github.com/air-verse/air)
(hot reload), [sqlc](https://sqlc.dev), [migrate](https://github.com/golang-migrate/migrate),
and [golangci-lint](https://golangci-lint.run).

## Environment

Configuration is via environment variables, all documented in
[`.env.example`](../.env.example). Copy it once:

```bash
cp .env.example .env
```

The Makefile loads `.env` automatically (`-include .env; export`), so `make`
targets see these values. The backend **panics at startup** if a required
variable is missing — there are no silent defaults for secrets.

Required: `BASE_URL`, `DATABASE_URL`, `REDIS_URL`. Everything else has a sensible
default (see `backend/internal/config/config.go`, the only file permitted to read
the environment).

### JWT keys

```bash
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem
```

`backend/keys/` is gitignored — never commit keys.

## Make targets

| Target | What it does |
|--------|--------------|
| `make dev` | Start API (air) + frontend together; Ctrl+C stops both |
| `make dev-api` | Start the Go API only (hot reload) |
| `make dev-web` | Start the Next.js frontend only |
| `make build` | Build the Go binary to `backend/bin/shrt` |
| `make test` | `go test -race ./...` |
| `make lint` | `golangci-lint run ./...` |
| `make sqlc` | Regenerate sqlc code from `db/queries/*.sql` |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back the last migration |
| `make docker-up` | Start Postgres + Redis |
| `make docker-down` | Stop them |
| `make tools` | Install Go dev tools |

> **Running without air:** if `air` isn't installed, run the API directly:
> ```bash
> cd backend && set -a && . ../.env && set +a && go run ./cmd/shrt
> ```

## Database migrations

Migrations live in `backend/db/migrations/` as `NNNNNN_name.{up,down}.sql` pairs,
applied with [golang-migrate](https://github.com/golang-migrate/migrate).

```bash
make migrate-up           # apply all pending
make migrate-down         # roll back one

# Create a new migration pair
migrate create -ext sql -dir backend/db/migrations -seq <name>
```

Every `up` must have a matching `down`. After changing the schema, regenerate
sqlc (below) so the Go types stay in sync.

## Code generation (sqlc)

Queries are written by hand in `backend/db/queries/*.sql`; sqlc generates
type-safe Go from them against the migration schema. Configuration is in
`backend/sqlc.yaml`.

```bash
# 1. Edit or add a query in db/queries/*.sql
# 2. Regenerate
make sqlc
```

**Never hand-edit** the generated files (`db/*.sql.go`, `db/models.go`,
`db/querier.go`) — change the `.sql` source and regenerate. Generated code **is**
committed so the repo builds without sqlc installed.

## Testing

### Backend

```bash
make test                 # go test -race ./...
```

Unit tests (slug generation, etc.) run anywhere. **Integration tests** in
`server/` need Postgres and Redis; they connect to `TEST_DATABASE_URL` /
`TEST_REDIS_URL` (falling back to the local Compose defaults) and **skip
gracefully** if the infrastructure isn't reachable. So with `make docker-up`
running, `make test` exercises the full suite; without it, the integration tests
skip rather than fail.

The test harness creates its own schema and an ephemeral RSA key pair, so it runs
in a clean checkout without `make migrate-up` or committed keys.

### Frontend

```bash
cd frontend
npm run type-check        # tsc --noEmit
npm run lint              # eslint .
npm run build             # production build
```

### End-to-end (Playwright)

E2E tests in `frontend/e2e/` drive the full stack and need it running.

```bash
make dev                  # in one terminal: backend + frontend + infra

cd frontend
npx playwright install chromium   # first run only
npm run e2e                       # headless
npm run e2e:ui                    # interactive inspector
```

They cover: anonymous shorten → copy → redirect; register → login → create →
edit → delete; and the expired-link 410 path.

## Before opening a PR

Run the full gate locally (GitHub Actions CI is temporarily disabled — see the
[README CI note](../README.md#ci-checks)):

```bash
# Backend
cd backend && go vet ./... && golangci-lint run ./... && go test -race ./... && go build ./cmd/shrt

# Frontend
cd frontend && npm run type-check && npm run lint && npm run build
```

## Common tasks

**Add a new API endpoint**
1. Write the handler in `server/` (parse → call store → respond). No business
   logic here.
2. Put the logic in `store/`; add any sentinel errors to `store/errors.go` and
   map them in `server/response.go`.
3. Register the route in `server/server.go`.
4. Add an integration test alongside the handler.
5. Update [`docs/API.md`](API.md) and [`openapi.yaml`](../openapi.yaml).

**Add a new query**
1. Write the SQL in `db/queries/*.sql` with a sqlc annotation.
2. `make sqlc`.
3. Call the generated method from `store/`.

**Debugging**
- Check service health: `curl localhost:8080/health`.
- The API logs structured request lines (method, path, status, duration) via
  chi's logger; watch the `make dev` output.
- Inspect Redis: `docker compose exec redis redis-cli` then `KEYS slug:*`.
- Inspect Postgres: `docker compose exec postgres psql -U shrt -d shrt`.

## Project conventions

These are enforced in review — see [`backend/CLAUDE.md`](../backend/CLAUDE.md)
and [`frontend/CLAUDE.md`](../frontend/CLAUDE.md) for the full standards.

- No service layer; `store/` owns logic, `server/` owns HTTP.
- `os.Getenv` only in `internal/config/`.
- Wrap errors with context; match with `errors.Is`/`errors.As`.
- Frontend: all backend calls through `lib/api.ts`; all shared types in
  `types/api.ts`; no `any`; semantic Tailwind tokens only.
