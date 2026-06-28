# shrt

A clean, self-hostable URL shortener. MIT licensed.

**Live instance:** coming soon  
**Repository:** https://github.com/sukhera/shrt

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.25+, chi, sqlc, pgx/v5, go-redis/v9 |
| Database | PostgreSQL 15 |
| Cache | Redis 7 / Upstash |
| Frontend | Next.js (latest stable), TypeScript, Tailwind, shadcn/ui |
| Auth | JWT RS256 (access token 1h, refresh token 30d) |

## Features (v1)

- Shorten any URL — no account required
- Custom aliases and expiry dates for registered users
- User dashboard to manage, edit, and delete links
- Dark mode
- Self-hostable with Docker Compose or any VPS

## Screenshots

| Home | Dashboard |
|------|-----------|
| ![Home page](docs/screenshots/home.png) | ![Dashboard](docs/screenshots/dashboard.png) |

## Local development

### Prerequisites

- Go 1.25+
- Node 20+
- Docker (for Postgres + Redis)

### Setup

```bash
# 1. Clone
git clone https://github.com/sukhera/shrt.git
cd shrt

# 2. Environment
cp .env.example .env
# Edit .env and fill in values

# 3. Generate RSA key pair for JWT
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem

# 4. Install Go dev tools
make tools

# 5. Start Postgres + Redis
make docker-up

# 6. Run migrations
make migrate-up

# 7. Start everything
make dev
```

- API: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Health check: `http://localhost:8080/health`

`make dev` starts both the Go API (with hot reload via air) and Next.js together. Ctrl+C stops both. Use `make dev-api` or `make dev-web` to start them individually.

> `make dev` and `make dev-api` use [air](https://github.com/air-verse/air) for hot reload — `make tools` installs it. Without air, run the API directly:
> ```bash
> cd backend && set -a && . ../.env && set +a && go run ./cmd/shrt
> ```

### Short URLs and `BASE_URL`

Generated short URLs are built from `BASE_URL`. In local dev this defaults to the
API origin, so links render as `http://localhost:8080/<slug>`. In production, set
`BASE_URL` to the public short domain (e.g. `https://shrt.example.com`) so links
read correctly. The redirect server (`GET /<slug>`) must be reachable at that
domain.

## Testing

```bash
# Backend — unit + integration tests (needs Postgres + Redis running)
make test                       # go test -race ./...

# Frontend — type-check and lint
cd frontend
npm run type-check              # tsc --noEmit
npm run lint                    # eslint .

# End-to-end (Playwright) — needs the full stack running (make dev)
cd frontend
npx playwright install chromium # first run only
npm run e2e                     # headless
npm run e2e:ui                  # interactive inspector
```

The E2E suite covers the critical paths: anonymous shorten → copy → redirect;
register → login → create → edit → delete; and the expired-link 410 response.

## CI checks

Run these locally before opening a PR (both must pass):

```bash
# Backend
cd backend && go vet ./... && golangci-lint run ./... && go test -race ./... && go build ./cmd/shrt

# Frontend
cd frontend && npm run type-check && npm run lint && npm run build
```

> GitHub Actions CI is temporarily disabled (billing); restore the workflows from
> git history once resolved. Until then these local checks are the gate.

## Deployment options

### Railway (simplest)

Deploy the backend as a Railway service with a Railway Postgres and Upstash Redis add-on. Set all env vars in Railway's dashboard. Estimated cost: ~$5/month.

### Hetzner VPS + Docker

Run the backend in Docker on a Hetzner CX11 (~€4/month). Use Neon (free tier) for Postgres and Upstash (free tier) for Redis. Use Caddy as a reverse proxy for automatic HTTPS.

### Frontend

Deploy the Next.js frontend to Vercel (free tier). Set `NEXT_PUBLIC_API_URL` to
your backend URL. The auth route handlers run server-side, so also set `API_URL`
(it falls back to `NEXT_PUBLIC_API_URL`, then `http://localhost:8080`).

### Production checklist

- Run migrations against the production database: `migrate -path backend/db/migrations -database "$DATABASE_URL" up`
- Set `ENV=production` (enables `Secure` cookies and stricter behaviour)
- Set `BASE_URL` to your short domain and `FRONTEND_URL` to the deployed frontend origin (CORS allowlist)
- Provide the JWT key pair via `JWT_PRIVATE_KEY_PATH` / `JWT_PUBLIC_KEY_PATH` (never commit `backend/keys/`)
- Terminate HTTPS at the reverse proxy (Caddy/Vercel handle this automatically); the app expects to run behind a trusted proxy that sets `X-Forwarded-For`
- Optionally set `SAFE_BROWSING_API_KEY` to enable Google Safe Browsing checks

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
