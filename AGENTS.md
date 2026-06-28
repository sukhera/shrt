# shrt — Agent Context

Read this file first, then read `IMPLEMENTATION-PLAN.md` for the full milestone breakdown, API contract, and your task list.

## What is this project?

`shrt` is an open-source, self-hostable URL shortener. MIT licensed.

- **Backend:** Go 1.25+, chi router, sqlc, pgx/v5, go-redis/v9, JWT RS256
- **Frontend:** Next.js (latest stable), TypeScript, Tailwind, shadcn/ui (New York + Zinc)
- **Database:** PostgreSQL 15
- **Cache:** Redis 7 / Upstash

## Key documents

| Document | Purpose |
|----------|---------|
| `IMPLEMENTATION-PLAN.md` | Design system, folder structure, API contract, milestone task lists, agent briefs |
| `URL-Shortener-PRD.md` | Full product requirements |
| `backend/CLAUDE.md` | Backend project-specific decisions (read before touching any Go code) |
| `frontend/CLAUDE.md` | Frontend project-specific decisions (read before touching any TS/TSX code) |

## Critical rules

- **No service layer.** `store/` owns all business logic. Handlers parse and respond only.
- **No `os.Getenv` outside `internal/config/`.**
- **No `any` in TypeScript.** All types in `types/api.ts`.
- **No tokens in `localStorage`.** Access token in memory; refresh token in httpOnly cookie.
- **Never edit sqlc-generated files** — run `make sqlc` to regenerate.
- **Never commit `.env`** — copy `.env.example` to `.env` and fill in values.

## Running locally

```bash
cp .env.example .env          # fill in values
make docker-up                # start Postgres + Redis
make migrate-up               # run DB migrations
make dev                      # start API + frontend together (Ctrl+C stops both)
                              # or: make dev-api / make dev-web to start separately
```
