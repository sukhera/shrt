# shrt — Agent & Contributor Guide

## What is this project?

`shrt` is an open-source URL shortener built with Go (backend) and Next.js (frontend). It is self-hostable and MIT licensed. The author runs a public live instance. This is not a SaaS product — the goal is a clean, well-structured codebase that anyone can clone and run.

**Repository:** https://github.com/sukhera/shrt

## Key documents — read these before writing any code

- `IMPLEMENTATION-PLAN.md` — milestone breakdown, ordered task lists, agent briefs, API contract, design system
- `URL-Shortener-PRD.md` — full product requirements
- `backend/CLAUDE.md` — Go coding standards (read before touching any backend code)
- `frontend/CLAUDE.md` — Next.js/TypeScript coding standards (read before touching any frontend code)

## Project structure

```
shrt/
├── backend/        # Go API + redirect server (single binary)
├── frontend/       # Next.js App Router
├── docker-compose.yml
├── Makefile
├── .env.example
└── .github/workflows/
```

Full folder layout is in `IMPLEMENTATION-PLAN.md § 2. Project Structure`.

## Running locally

```bash
# 1. Start Postgres + Redis
make docker-up

# 2. Run DB migrations
make migrate-up

# 3. Start Go API (hot reload)
make dev

# 4. Start Next.js (separate terminal)
cd frontend && npm run dev
```

Backend runs on `http://localhost:8080`. Frontend runs on `http://localhost:3000`.

Copy `.env.example` to `.env` and fill in values before running. The app panics at startup if required env vars are missing.

## Tech stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.22+, chi router, sqlc, pgx/v5, go-redis/v9 |
| Database | PostgreSQL 15+ |
| Cache | Redis 7+ / Upstash |
| Frontend | Next.js (latest stable), TypeScript, Tailwind CSS, shadcn/ui |
| Auth | JWT RS256 (access token 1h, refresh token 30d) |
| CI | GitHub Actions |

## Git workflow

- `main` — always deployable; direct pushes blocked; requires CI pass
- `dev` — integration branch; PRs merge here first
- Feature branches: `feat/slug-generation`, `fix/cache-invalidation`
- Every PR must pass CI (lint + test + build) before merge

Commit messages follow Conventional Commits:
```
feat(redirect): add Redis cache with TTL fallback
fix(auth): correct bcrypt cost factor
chore(ci): add golangci-lint to backend workflow
test(links): integration test for expired link 410 response
```

## CI

- `backend-ci.yml` — golangci-lint, go test ./..., go build
- `frontend-ci.yml` — tsc --noEmit, next lint, next build

Both must pass before any PR is merged.

## Environment variables

All required variables are documented in `.env.example`. Never commit `.env` or any file containing real secrets. Never hardcode secrets in source code. The RSA key pair lives in `backend/keys/` which is gitignored.

## What not to do

- Do not add new dependencies without a clear justification
- Do not introduce global state
- Do not write business logic in HTTP handlers or React components
- Do not use `any` in TypeScript
- Do not call `os.Getenv` outside of `backend/internal/config/`
- Do not store tokens in `localStorage`
