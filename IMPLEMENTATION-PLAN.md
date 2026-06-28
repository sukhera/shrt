---
title: shrt — Implementation Plan
tags:
  - implementation
  - plan
  - agents
status: active
created: 2026-06-27
updated: 2026-06-27
version: "1.0"
repository: https://github.com/sukhera/shrt
owner: Ahmed Sukhera
---

# shrt — Implementation Plan

## Overview

This document is the single source of truth for building `shrt`. It defines design decisions, project conventions, folder structure, milestone breakdown, and agent briefs. Every agent working on this project reads this document first.

Related: [[URL-Shortener-PRD]]

---

## Table of Contents

- [[#1. Design System]]
- [[#2. Project Structure]]
- [[#3. Conventions]]
- [[#4. API Contract]]
- [[#5. Milestones]]
- [[#6. Agent Briefs]]
- [[#7. Git Workflow]]

---

## 1. Design System

Design decisions are locked for v1. The frontend agent implements these exactly — no deviations.

### 1.1 Component Library

**shadcn/ui** — initialized with the `zinc` base color and the New York style variant. All UI primitives come from shadcn (Button, Input, Card, Table, Dialog, Badge, etc.). Do not introduce other UI libraries.

```bash
npx shadcn@latest init
# Style: New York
# Base color: Zinc
# CSS variables: yes
```

### 1.2 Fonts

**Geist Sans** for all body text and UI. **Geist Mono** for slugs, short URLs, and any code-like strings (copy buttons, slug display). Both ship via the `geist` package (`geist/font/sans`, `geist/font/mono`).

```ts
import { GeistSans } from 'geist/font/sans'
import { GeistMono } from 'geist/font/mono'
```

### 1.3 Color Palette

The palette uses shadcn's CSS variable system. Override the defaults in `globals.css` with the values below.

**Primary accent: Violet**

```css
/* globals.css — :root (light mode) */
--background: 0 0% 100%;
--foreground: 240 10% 3.9%;
--card: 0 0% 100%;
--card-foreground: 240 10% 3.9%;
--popover: 0 0% 100%;
--popover-foreground: 240 10% 3.9%;
--primary: 262 83% 58%;          /* violet-600 */
--primary-foreground: 0 0% 98%;
--secondary: 240 4.8% 95.9%;
--secondary-foreground: 240 5.9% 10%;
--muted: 240 4.8% 95.9%;
--muted-foreground: 240 3.8% 46.1%;
--accent: 262 83% 96%;           /* violet-50 */
--accent-foreground: 262 83% 58%;
--destructive: 0 84.2% 60.2%;
--destructive-foreground: 0 0% 98%;
--border: 240 5.9% 90%;
--input: 240 5.9% 90%;
--ring: 262 83% 58%;
--radius: 0.5rem;

/* .dark */
--background: 240 10% 3.9%;
--foreground: 0 0% 98%;
--card: 240 10% 3.9%;
--card-foreground: 0 0% 98%;
--primary: 263 70% 65%;          /* violet-400 — lighter for dark mode */
--primary-foreground: 240 10% 3.9%;
--secondary: 240 3.7% 15.9%;
--secondary-foreground: 0 0% 98%;
--muted: 240 3.7% 15.9%;
--muted-foreground: 240 5% 64.9%;
--accent: 240 3.7% 15.9%;
--accent-foreground: 0 0% 98%;
--border: 240 3.7% 15.9%;
--input: 240 3.7% 15.9%;
--ring: 263 70% 65%;
```

### 1.4 Icons

**Lucide React** only (ships with shadcn). Do not add Heroicons, Font Awesome, or any other icon library.

### 1.5 Motion & Animation

Subtle only. Use Tailwind's `transition-colors`, `transition-opacity`, and shadcn's built-in animation for dialogs/dropdowns. No page transitions, no scroll animations, no heavy motion libraries.

### 1.6 UI Patterns

- **Slug display:** monospace font (Geist Mono), truncated with a copy button inline
- **Short URLs:** shown as full URL (`https://shrt.io/abc1234`), monospace, violet text, copy button
- **Status badges:** `active` (green), `expires soon` (amber, < 7 days), `expired` (zinc/muted)
- **Empty states:** centered illustration area with a short message and a primary CTA button
- **Destructive actions:** always require a confirmation Dialog before executing
- **Loading states:** skeleton loaders for table rows; spinner on buttons during submission
- **Toasts:** use shadcn's `Sonner` integration for success/error feedback (e.g. "Link copied", "Link deleted")

### 1.7 Page Layout

- Max content width: `max-w-4xl` centered with `mx-auto px-4`
- Navigation: a minimal top nav with the `shrt` wordmark (Geist Mono, violet) on the left, auth links on the right
- No sidebar. No complex navigation. Home → Dashboard is the full user journey.

---

## 2. Project Structure

### 2.1 Repository Layout

```
shrt/                          # github.com/sukhera/shrt
├── backend/                   # Go API + redirect server
│   ├── cmd/
│   │   └── shrt/
│   │       └── main.go        # entry point — wire deps, call server.New().Start()
│   ├── server/                # HTTP handlers, route registration, middleware
│   │   ├── server.go          # Server struct, NewServer(), Start(), routes()
│   │   ├── auth.go            # register/login/refresh/logout handlers
│   │   ├── link.go            # link CRUD handlers
│   │   ├── redirect.go        # GET /:slug handler
│   │   ├── middleware.go      # auth middleware, rate limiting, CORS
│   │   └── response.go        # respondJSON(), respondError(), error→HTTP mapping
│   ├── store/                 # DB + Redis + all business logic
│   │   ├── store.go           # Store struct, NewStore() — holds pgxpool + redis client
│   │   ├── link.go            # CreateLink, GetBySlug, ListByUser, Update, Delete
│   │   ├── user.go            # CreateUser, GetByEmail
│   │   ├── token.go           # CreateRefreshToken, Revoke, DeleteExpired
│   │   └── errors.go          # ErrNotFound, ErrAliasTaken, ErrUnsafeURL, etc.
│   ├── internal/
│   │   └── config/            # env var loading only — nothing else goes here
│   ├── db/
│   │   ├── migrations/        # golang-migrate SQL files
│   │   └── queries/           # sqlc .sql query files (never edit generated output)
│   ├── sqlc.yaml
│   ├── go.mod                 # module: github.com/sukhera/shrt/backend
│   └── go.sum
├── frontend/                  # Next.js app (App Router)
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   └── register/
│   │   ├── dashboard/
│   │   │   ├── loading.tsx    # required — skeleton UI while data loads
│   │   │   └── error.tsx      # required — error boundary for dashboard
│   │   ├── globals.css        # CSS variables (design system overrides)
│   │   ├── layout.tsx         # root layout — fonts, providers, nav
│   │   └── page.tsx           # home page — URL shortener form
│   ├── components/
│   │   ├── ui/                # shadcn generated — never hand-edit
│   │   └── app/               # project components (max ~300 lines each)
│   ├── lib/
│   │   ├── api.ts             # typed fetch wrapper — only place fetch is called
│   │   ├── auth.ts            # token storage (in-memory access, httpOnly refresh)
│   │   └── utils.ts           # shadcn cn() utility + shared helpers
│   ├── hooks/
│   │   ├── use-links.ts       # all TanStack Query hooks for link data (linkKeys factory)
│   │   └── use-auth.ts        # auth state hook
│   ├── providers/             # React context wrappers (QueryProvider, ThemeProvider)
│   ├── types/
│   │   └── api.ts             # all shared API types — no inline types in components
│   ├── mocks/                 # MSW handlers for component tests
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── package.json
├── docker-compose.yml         # Postgres + Redis for local dev
├── Makefile                   # dev shortcuts
├── .env.example               # all required env vars documented
├── .github/
│   └── workflows/
│       ├── backend-ci.yml     # Go: lint + test + build
│       └── frontend-ci.yml    # Next.js: lint + type-check + build
├── IMPLEMENTATION-PLAN.md     # this document
├── URL-Shortener-PRD.md
├── AGENTS.md                  # agent context file (symlink-friendly)
└── README.md
```

### 2.2 Go Package Conventions

- **No service layer.** `store/` owns business logic AND data access. Handlers call store directly — no intermediary service structs.
- `server/` contains only HTTP concerns: parse request → call store method → write response. No business logic in handlers.
- `store/` owns everything below the HTTP layer: slug generation, Safe Browsing checks, cache reads/writes, DB queries, password hashing, token issuance. All sentinel errors are defined in `store/errors.go`.
- sqlc generates type-safe query code from `db/queries/` — do not hand-edit generated files. Run `make sqlc` after changing `.sql` files.
- `internal/config/` loads env vars once at startup into a typed `Config` struct. Nothing else goes in `internal/`. Never call `os.Getenv` outside of `internal/config/`.
- Errors: use `fmt.Errorf("context: %w", err)` wrapping throughout; use `errors.Is` / `errors.As` for matching — never string comparison.

### 2.3 Next.js Conventions

- App Router only — no Pages Router
- Server Components by default; add `"use client"` only when hooks or browser APIs are required
- API calls from the frontend go through `lib/api.ts` — a typed fetch wrapper that handles auth headers and error parsing
- All shared TypeScript types live in `types/` — no inline `any`
- shadcn components go in `components/ui/` (generated); custom app components go in `components/app/`

---

## 3. Conventions

### 3.1 Environment Variables

Every required variable is documented in `.env.example`. The app panics at startup if a required variable is missing. No silent defaults for secrets.

```bash
# .env.example

# ── Server ────────────────────────────────────
PORT=8080
ENV=development                    # development | production
BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000

# ── Database ──────────────────────────────────
DATABASE_URL=postgres://shrt:shrt@localhost:5432/shrt?sslmode=disable

# ── Redis / Upstash ───────────────────────────
REDIS_URL=redis://localhost:6379

# ── JWT ───────────────────────────────────────
# Generate: openssl genrsa -out private.pem 2048
# and:      openssl rsa -in private.pem -pubout -out public.pem
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_ACCESS_TTL=1h
JWT_REFRESH_TTL=720h             # 30 days

# ── Google Safe Browsing ──────────────────────
SAFE_BROWSING_API_KEY=

# ── Rate Limiting ─────────────────────────────
RATE_LIMIT_ANON=10               # shortens per hour for anonymous users
RATE_LIMIT_USER=200              # shortens per hour for registered users

# ── App ───────────────────────────────────────
SLUG_LENGTH=7
DEFAULT_REDIRECT_CODE=302
```

### 3.2 Makefile Targets

```makefile
make dev          # start Go server with hot reload (air)
make test         # run Go tests
make lint         # golangci-lint
make migrate-up   # run pending migrations
make migrate-down # roll back one migration
make sqlc         # regenerate sqlc code
make docker-up    # docker-compose up -d (Postgres + Redis)
make docker-down  # docker-compose down
```

### 3.3 Git Workflow

- `main` — always deployable; direct pushes blocked
- `dev` — integration branch; PRs merge here first
- Feature branches: `feat/slug-generation`, `fix/cache-invalidation`, `chore/ci-setup`
- Every PR must pass CI (lint + test + build) before merge
- Commit messages follow Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`, `test:`

### 3.4 Testing

- **Backend:** table-driven unit tests for service logic; integration tests for handlers using `net/http/httptest` and a real test DB (separate from dev DB)
- **Frontend:** component tests with React Testing Library + MSW (Mock Service Worker) for key components (copy button, form validation)
- **E2E:** Playwright, added in the QA milestone; covers the critical path (shorten → copy → visit short URL → redirect)
- Test files live alongside the code they test (`link_service_test.go` next to `link_service.go`)

---

## 4. API Contract

The frontend agent builds against this contract. It is locked — any change requires updating this document first.

Base URL (local): `http://localhost:8080/api/v1`
Base URL (production): `https://api.shrt.io/v1` *(or your domain)*

All timestamps are ISO 8601 UTC. All IDs are UUIDs.

### 4.1 Link Object (shared response shape)

```json
{
  "id": "018f4a2b-...",
  "slug": "abc1234",
  "short_url": "https://shrt.io/abc1234",
  "original_url": "https://www.example.com/long/path",
  "is_custom": false,
  "expires_at": null,
  "created_at": "2026-06-27T10:00:00Z",
  "updated_at": "2026-06-27T10:00:00Z"
}
```

### 4.2 Endpoints

#### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/register` | No | Create account |
| `POST` | `/auth/login` | No | Login |
| `POST` | `/auth/refresh` | No (refresh token) | Refresh access token |
| `POST` | `/auth/logout` | Yes | Revoke refresh token |

**POST /auth/register**
```json
// Request
{ "email": "user@example.com", "password": "minlength8" }

// Response 201
{ "user": { "id": "...", "email": "user@example.com" }, "access_token": "...", "refresh_token": "..." }
```

**POST /auth/login**
```json
// Request
{ "email": "user@example.com", "password": "..." }

// Response 200
{ "access_token": "...", "refresh_token": "...", "expires_in": 3600 }
```

**POST /auth/refresh**
```json
// Request
{ "refresh_token": "..." }

// Response 200
{ "access_token": "...", "expires_in": 3600 }
```

#### Links

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/links` | Optional | Shorten a URL |
| `GET` | `/links` | Yes | List user's links |
| `GET` | `/links/:slug` | Yes | Get link details |
| `PATCH` | `/links/:slug` | Yes | Update a link |
| `DELETE` | `/links/:slug` | Yes | Delete a link |

**POST /links**
```json
// Request (alias and expires_at optional)
{ "url": "https://example.com/long", "alias": "my-link", "expires_at": "2026-12-31T23:59:59Z" }

// Response 201 — Link object
```

**GET /links**
```
Query params: page (default 1), limit (default 20, max 100), sort (created_at | expires_at), order (asc | desc), q (search term)
```
```json
// Response 200
{
  "data": [ /* Link objects */ ],
  "pagination": { "page": 1, "limit": 20, "total": 47 }
}
```

**PATCH /links/:slug**
```json
// Request — any subset
{ "url": "https://new-destination.com", "alias": "new-alias", "expires_at": null }

// Response 200 — updated Link object
```

**DELETE /links/:slug** → `204 No Content`

#### Error Envelope (all errors)

```json
{
  "error": {
    "code": "ALIAS_TAKEN",
    "message": "That alias is already in use.",
    "status": 409
  }
}
```

#### Error Codes

| Code | Status | Meaning |
|------|--------|---------|
| `INVALID_URL` | 422 | Malformed URL |
| `UNSAFE_URL` | 422 | Flagged by Safe Browsing |
| `ALIAS_TAKEN` | 409 | Custom alias already exists |
| `LINK_NOT_FOUND` | 404 | Slug doesn't exist or is deleted |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Authenticated but not the owner |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL` | 500 | Server error |

#### Redirect Endpoint (not under `/api/v1`)

```
GET /:slug
→ 302  Found           (valid link with expiry, or DEFAULT_REDIRECT_CODE=302)
→ 301  Moved Permanently (valid link, no expiry, DEFAULT_REDIRECT_CODE=301)
→ 404  Not Found       (unknown slug)
→ 410  Gone            (expired link)
```

---

## 5. Milestones

### M0 — Foundation *(agent: Foundation Agent)*

**Goal:** Working repo skeleton with local dev environment. No business logic yet.

Tasks (in order):
1. Clone repo from `https://github.com/sukhera/shrt`
2. Create `backend/` Go module (`github.com/sukhera/shrt/backend`) with entry point at `cmd/shrt/main.go`
3. Create `frontend/` Next.js app with TypeScript, Tailwind, App Router
4. Initialize shadcn/ui with New York style + Zinc base
5. Apply design system: install Geist font, override CSS variables per [[#1.3 Color Palette]]
6. Create `docker-compose.yml` for Postgres (port 5432) and Redis (port 6379)
7. Create `Makefile` with targets from [[#3.2 Makefile Targets]]
8. Create `.env.example` with all variables from [[#3.1 Environment Variables]]
9. Create `backend/db/migrations/` directory and write first migration: `001_create_users.sql`, `002_create_links.sql`, `003_create_refresh_tokens.sql` per [[#7. Data Models]] in [[URL-Shortener-PRD]]
10. Set up `golang-migrate` and verify migrations run cleanly
11. Set up GitHub Actions: `backend-ci.yml` (lint + test + build) and `frontend-ci.yml` (lint + type-check + build)
12. Create `AGENTS.md` at repo root (copy of agent context summary)
13. Create `README.md` (project description, local dev setup instructions, deployment options)

**Done when:** `make docker-up && make migrate-up && make dev` runs without errors; `cd frontend && npm run dev` starts the Next.js dev server.

---

### M1 — Redirect Service *(agent: Backend Agent)*

**Goal:** Core redirect logic working end-to-end locally.

Dependencies: M0 complete.

Tasks (in order):
1. Set up `internal/config/config.go` — load `.env` into typed `Config` struct, panic on missing required vars
2. Write sqlc query files in `db/queries/` for: `GetLinkBySlug`, `GetLinksByUserID`, `CreateLink`, `UpdateLink`, `SoftDeleteLink`; run `make sqlc` to generate (do not hand-edit generated output)
3. Implement `store/store.go` — `Store` struct holding `pgxpool.Pool` + `redis.Client`; `NewStore(cfg)` initialises both connections
4. Implement `store/link.go` — `GetBySlug(ctx, slug)` with Redis cache-aside: check `slug:<slug>` key first; on miss query Postgres, populate cache with TTL = `expires_at - now()` or 24h if no expiry; cache failures must never block the lookup (log warning, fall through)
5. Implement `server/server.go` — `Server` struct; `NewServer(cfg, store)`; `Start()` with all HTTP timeouts set (ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout); basic middleware: logger, recoverer, CORS
6. Wire `cmd/shrt/main.go` — load config, create store, create server, call `server.Start()`
7. Implement `server/redirect.go` — `GET /:slug` handler: call `store.GetBySlug`; not found → 404; expired → 410; valid → 302 (or value of `DEFAULT_REDIRECT_CODE`)
8. Write integration tests for the redirect handler covering all four cases (found, not found, deleted, expired)

**Done when:** `curl localhost:8080/abc1234` returns 302 for a valid slug seeded in the DB, 404 for an unknown slug, 410 for an expired slug.

---

### M2 — Link API *(agent: Backend Agent)*

**Goal:** Full CRUD API for links, with Safe Browsing and rate limiting.

Dependencies: M1 complete.

Tasks (in order):
1. Implement slug generation in `store/link.go` — 7-char base62 random string (`crypto/rand`); retry up to 5x on collision; return `ErrAliasTaken` if all collide
2. Implement Google Safe Browsing check in `store/link.go` — call the API before insert; return `ErrUnsafeURL` if flagged; if API key is empty, log warning and skip (allows local dev without a key)
3. Implement `store/link.go` methods: `CreateLink`, `ListByUser` (paginated, search, sort), `GetBySlug`, `UpdateLink` (validate ownership, invalidate Redis cache), `DeleteLink` (soft delete, invalidate cache)
4. Implement `server/link.go` handlers: `POST /api/v1/links`, `GET /api/v1/links`, `GET /api/v1/links/:slug`, `PATCH /api/v1/links/:slug`, `DELETE /api/v1/links/:slug` — parse request, call store, return response envelope; no business logic in handlers
5. Implement rate limiting in `server/middleware.go` using Redis INCR+EXPIRE pattern; on Redis failure, allow request and log; apply to link creation route
6. Write unit tests for slug generation (collision handling, base62 alphabet); write integration tests for all five link endpoints

**Done when:** All link endpoints respond correctly per [[#4. API Contract]]; rate limiting returns 429 after threshold exceeded.

---

### M3 — Authentication *(agent: Backend Agent)*

**Goal:** Working register, login, token refresh, and logout. All link endpoints require auth except `POST /links` (optional) and `GET /:slug` (redirect, no auth).

Dependencies: M2 complete.

Tasks (in order):
1. Generate RSA key pair; add `backend/keys/` to `.gitignore`; document key generation command in README
2. Add sqlc query files for refresh tokens: `CreateRefreshToken`, `GetRefreshTokenByHash`, `RevokeRefreshToken`, `DeleteExpiredTokens`; run `make sqlc`
3. Implement JWT helpers inline in `store/token.go` — RS256 signing (`CreateAccessToken`, `CreateRefreshToken`), token parsing and validation using `golang-jwt/jwt/v5`
4. Implement `store/user.go` — `CreateUser` (bcrypt hash password, insert, issue tokens), `GetByEmail`
5. Implement `store/token.go` auth methods — `Login` (verify password, issue tokens), `RefreshToken` (validate refresh token, issue new access token), `Logout` (revoke refresh token)
6. Implement `server/auth.go` handlers: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout` — parse request, call store, return tokens
7. Implement `AuthMiddleware` in `server/middleware.go` — parse Bearer token, validate, attach user to context; apply to all `/api/v1/links` routes (optional on `POST /links`)
8. Write integration tests for all four auth endpoints

**Done when:** Can register, login, create a link as an authenticated user, refresh the token, and logout. Unauthenticated requests to protected endpoints return 401.

---

### M4 — Frontend *(agent: Frontend Agent)*

**Goal:** Full working UI — home page, auth flows, dashboard.

Dependencies: M3 complete (backend fully running locally). Agent must have the backend running at `http://localhost:8080` to test against.

Tasks (in order):
1. Install and configure Sonner for toast notifications
2. Create `lib/api.ts` — typed fetch wrapper: attaches `Authorization` header from stored access token, handles 401 (trigger token refresh), parses error envelope into typed `ApiError` objects
3. Create `lib/auth.ts` — token storage primitives: access token in a module-scoped variable (never localStorage); refresh token sent/received via httpOnly cookie through a Next.js API route at `/api/refresh`
4. Create auth state management — `providers/` React Context + `useReducer` in `hooks/use-auth.ts`; on app load, attempt silent refresh to restore session
5. Implement `/login` page — email + password form, shadcn `Card` layout, call `POST /auth/login`, redirect to `/dashboard` on success
6. Implement `/register` page — same layout as login, call `POST /auth/register`, auto-login on success
7. Implement home page (`/`) — URL input (full width, prominent), "Shorten" button, advanced toggle revealing alias + expiry fields, result card with short URL in Geist Mono + copy button, Sonner toast on copy; works for both anonymous and authenticated users
8. Implement dashboard (`/dashboard`) — auth-guarded; fetch `GET /links` via `hooks/use-links.ts`; render as a table (shadcn `Table`) with columns: Slug (monospace, copy button), Destination (truncated), Created, Status badge; search input (debounced, 300ms); sort controls; pagination
9. Implement Edit modal (shadcn `Dialog`) — pre-filled form with current URL, alias, expiry; `PATCH /links/:slug` on submit; optimistic UI update via TanStack Query cache invalidation
10. Implement Delete confirmation (shadcn `AlertDialog`) — confirm before calling `DELETE /links/:slug`; remove row from table on success with Sonner toast
11. Implement `/login` redirect guard — unauthenticated users hitting `/dashboard` are redirected to `/login`
12. Implement 404 page and 410 page (branded, consistent with design system)
13. Implement dark mode toggle (shadcn `ThemeProvider` + next-themes); persist preference via `next-themes`

**Done when:** Can shorten a URL anonymously on the home page; register; log in; view, edit, and delete links in the dashboard; copy short URLs; dark mode works.

---

### M5 — QA & Polish *(agent: QA Agent)*

**Goal:** Tested, clean, ready to ship.

Dependencies: M4 complete.

Tasks (in order):
1. Run `golangci-lint run ./...` on backend; fix all warnings
2. Run `tsc --noEmit` and `next lint` on frontend; fix all warnings
3. Write Playwright E2E tests for the critical path:
   - Anonymous user shortens a URL, copies it, visits it, gets redirected
   - User registers, logs in, creates a link, edits it, deletes it
   - Expired link returns the 410 branded page
4. Manually verify the security checklist:
   - HTTPS enforced (test with HTTP request)
   - Rate limiting triggers correctly
   - Unauthenticated access to dashboard redirects to login
   - CORS rejects requests from unknown origins
   - Attempting to edit/delete another user's link returns 403
5. Update `README.md`: add screenshots of home page and dashboard, complete deployment guide for all options in [[URL-Shortener-PRD#12. Deployment Options]]
6. Verify `.env.example` is complete and all variables are documented
7. Add `CONTRIBUTING.md` — how to run locally, how to submit a PR, code style notes

**Done when:** All Playwright tests pass; lint is clean; README is complete enough for a stranger to self-host the project.

---

### M6 — Deploy *(author, not an agent)*

**Goal:** Live instance at custom domain.

Tasks:
1. Register domain (Cloudflare Registrar)
2. Deploy Go backend to Railway (or Hetzner VPS with Docker)
3. Set up Neon Postgres (or Railway Postgres); run migrations against production DB
4. Set up Upstash Redis; point `REDIS_URL` at Upstash endpoint
5. Deploy Next.js frontend to Vercel; set `NEXT_PUBLIC_API_URL` to production backend URL
6. Configure DNS: `shrt.yourdomain.com` → backend, `yourdomain.com` → Vercel
7. Set `ENV=production` and verify HTTPS, CORS, and rate limiting work correctly
8. Make GitHub repo public

---

## 6. Agent Briefs

Each agent reads this section for its role, then reads the relevant milestone(s) above.

### Foundation Agent

You are setting up the `shrt` project skeleton. Read [[#2. Project Structure]] for the exact folder layout, [[#3. Conventions]] for the Makefile and env var conventions, and [[#1. Design System]] to initialize the Next.js frontend correctly. Your goal is an empty but correctly structured project that builds and runs. Do not implement any business logic.

**Tech stack:** Go 1.22+, `github.com/go-chi/chi/v5`, `github.com/golang-migrate/migrate/v4`, Next.js (latest stable, App Router), TypeScript, Tailwind CSS, shadcn/ui (New York + Zinc), Geist font, Docker Compose.

**Deliver:** Passing GitHub Actions CI on both backend and frontend; `make docker-up && make migrate-up && make dev` runs without errors; `npm run dev` starts without errors.

---

### Backend Agent

You are building the Go API and redirect server for `shrt`. Read [[#3. Conventions]] for package structure rules, [[#4. API Contract]] for the exact endpoint shapes you must implement, and your assigned milestone(s) for the ordered task list. Do not deviate from the API contract — the frontend is built against it.

**Tech stack:** Go 1.22+, `github.com/go-chi/chi/v5` (router), `github.com/redis/go-redis/v9` (Redis), `github.com/jackc/pgx/v5` (Postgres driver), `sqlc-dev/sqlc` (query generation), `github.com/golang-migrate/migrate/v4` (migrations), `github.com/golang-jwt/jwt/v5` (JWT RS256), `golang.org/x/crypto/bcrypt` (password hashing).

Pin the latest stable release of each module when first imported in M1 (as of this writing: chi v5.3.0, pgx v5.10.0, go-redis v9.21.0, golang-jwt v5.3.1, golang-migrate v4.19.1, go-playground/validator v10.30.3). Same majors as listed above — take the newest patch/minor.

**Rules:**
- No service layer — all business logic lives in `store/`; handlers only parse and respond
- `server/` = HTTP concerns only; `store/` = everything below (DB, Redis, slug gen, Safe Browsing, auth)
- Never call `os.Getenv` outside of `internal/config/`
- Every handler has a corresponding integration test
- Return errors using the error envelope from [[#4. API Contract]]

---

### Frontend Agent

You are building the Next.js frontend for `shrt`. Read [[#1. Design System]] carefully — design decisions are locked and must be followed exactly. Read [[#4. API Contract]] to understand the backend you're integrating with. Read your assigned milestone for the ordered task list.

**Tech stack:** Next.js (latest stable, App Router), TypeScript, Tailwind CSS, shadcn/ui (New York + Zinc), Geist Sans + Geist Mono, Lucide React, Sonner (toasts), next-themes (dark mode), Playwright (E2E, added in M5).

**Rules:**
- Server Components by default; `"use client"` only when necessary
- All API calls through `lib/api.ts` — never fetch the backend directly from a component
- All shared types in `types/` — no `any`
- shadcn components in `components/ui/`; app components in `components/app/`
- Access tokens in memory only; refresh tokens in httpOnly cookies via Next.js API route

---

### QA Agent

You are doing the final quality and testing pass on `shrt`. Both the backend and frontend are complete. Read [[#3.4 Testing]] for the testing conventions. Your milestone is [[#M5 — QA & Polish]].

**Your job:** Fix lint warnings (do not rewrite logic — fix warnings only), write Playwright E2E tests for the critical user paths, manually verify the security checklist, and ensure the README and CONTRIBUTING docs are complete enough for a stranger to self-host the project.

---

## 7. Git Workflow

```
main          ←── PRs from dev only (CI must pass)
  └── dev     ←── PRs from feature branches (CI must pass)
        ├── feat/m0-foundation
        ├── feat/m1-redirect
        ├── feat/m2-link-api
        ├── feat/m3-auth
        ├── feat/m4-frontend
        └── chore/m5-qa
```

Each milestone is a single PR from its feature branch into `dev`. When a milestone is complete and the PR is merged, the next milestone's agent branches off the updated `dev`.

Commit message format (Conventional Commits):
```
feat(redirect): add Redis cache with TTL fallback
fix(auth): correct bcrypt cost factor to 12
chore(ci): add golangci-lint to backend workflow
test(links): add integration test for expired link 410 response
docs(readme): add local dev setup instructions
```

---

*Author: [[Ahmed Sukhera]] | Repository: https://github.com/sukhera/shrt | Last updated: 2026-06-27*
