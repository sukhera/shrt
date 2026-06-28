# shrt

A clean, self-hostable URL shortener. MIT licensed.

**Live instance:** coming soon  
**Repository:** https://github.com/sukhera/shrt

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.22+, chi, sqlc, pgx/v5, go-redis/v9 |
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

## Local development

### Prerequisites

- Go 1.22+
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

## Deployment options

### Railway (simplest)

Deploy the backend as a Railway service with a Railway Postgres and Upstash Redis add-on. Set all env vars in Railway's dashboard. Estimated cost: ~$5/month.

### Hetzner VPS + Docker

Run the backend in Docker on a Hetzner CX11 (~€4/month). Use Neon (free tier) for Postgres and Upstash (free tier) for Redis. Use Caddy as a reverse proxy for automatic HTTPS.

### Frontend

Deploy the Next.js frontend to Vercel (free tier). Set `NEXT_PUBLIC_API_URL` to your backend URL.

## Contributing

See `CONTRIBUTING.md` (added in M5).

## License

MIT
