---
title: URL Shortener — Project Requirements Document
tags:
  - project
  - requirements
  - backend
  - api
  - open-source
status: draft
created: 2026-06-27
updated: 2026-06-27
version: "1.2"
owner: Ahmed Sukhera
---

# URL Shortener — Project Requirements Document

## Overview

An open-source URL shortening service built properly and made available for anyone to self-host. The author runs a public live instance, but the primary goal is a clean, well-structured codebase that developers can clone, understand, and deploy themselves.

This is not a SaaS product. There is no growth target, no paid tier, and no scaling ambition baked into v1. The measure of success is code quality, clarity, and ease of self-hosting.

**Repository:** GitHub (public, MIT license)
**Live instance:** hosted by the author at a custom domain

---

## Table of Contents

- [[#1. Goals & Non-Goals]]
- [[#2. User Personas]]
- [[#3. Functional Requirements]]
- [[#4. Non-Functional Requirements]]
- [[#5. Tech Stack]]
- [[#6. System Architecture]]
- [[#7. Data Models]]
- [[#8. API Specification]]
- [[#9. Frontend Requirements]]
- [[#10. Security Requirements]]
- [[#11. Error Handling]]
- [[#12. Deployment Options]]
- [[#13. Local Development]]
- [[#14. v1.1 — Deferred Features]]
- [[#15. Out of Scope]]
- [[#16. Milestones & Timeline]]
- [[#17. Open Questions]]

---

## 1. Goals & Non-Goals

### Goals

- Shorten any valid URL to a compact slug (e.g. `short.ly/abc123`)
- Redirect visitors from the short link to the original URL quickly
- Support optional custom aliases and link expiration for registered users
- Provide a dashboard where registered users can manage their links
- Be straightforward to self-host via Docker Compose
- Maintain a clean, readable codebase with good documentation

### Non-Goals

- Scaling to millions of users
- Marketing or user acquisition
- Monetisation or paid tiers
- High-availability infrastructure (no SLA commitments)
- Analytics in v1 — deferred to [[#14. v1.1 — Deferred Features]]
- OAuth / social login in v1

---

## 2. User Personas

### Anonymous User
Visits the site, pastes a URL, gets a short link. No account needed. Subject to rate limiting by IP.

### Registered User
Creates an account to own and manage their links — set custom aliases, expiry dates, edit or delete links via a dashboard.

### Self-Hoster
A developer who clones the repo and runs their own instance. First-class citizen: the project should be straightforward to configure and run locally or on any cloud provider.

### Admin
Has read/write access to all links. Can disable links that violate policy. Managed via a JWT role claim — no separate admin UI in v1.

---

## 3. Functional Requirements

### 3.1 URL Shortening

| ID | Requirement |
|----|-------------|
| F-01 | Accept a valid HTTP or HTTPS URL and return a shortened URL |
| F-02 | Validate that the input is a well-formed URL before processing |
| F-03 | Generate a unique alphanumeric slug (7 characters, base62) |
| F-04 | Detect duplicate long URLs for the same registered user and return the existing short link |
| F-05 | Support custom aliases for registered users (3–32 alphanumeric characters and hyphens) |
| F-06 | Check custom aliases for uniqueness before accepting |
| F-07 | Support an optional expiration date/time per link |
| F-08 | Expired links return HTTP 410 Gone |
| F-09 | Scan destination URLs against Google Safe Browsing before shortening |

### 3.2 Redirection

| ID | Requirement |
|----|-------------|
| F-10 | Visiting a valid short link redirects to the original URL |
| F-11 | Links without expiration use HTTP 301 (browser-cacheable) |
| F-12 | Links with expiration use HTTP 302 |
| F-13 | Invalid or deleted slugs return HTTP 404 with a branded error page |

### 3.3 Link Management

| ID | Requirement |
|----|-------------|
| F-14 | Registered users can view all their links in a paginated dashboard |
| F-15 | Registered users can edit the destination URL, alias, or expiry of their own links |
| F-16 | Registered users can delete their own links |
| F-17 | Deleting a link immediately stops it redirecting (cache invalidated) |
| F-18 | Dashboard supports sorting by created date and searching by slug or destination URL |
| F-19 | Default page size is 20; max is 100 |

### 3.4 Authentication

| ID | Requirement |
|----|-------------|
| F-20 | Users can register with an email and password |
| F-21 | Passwords hashed with bcrypt (cost factor ≥ 12) |
| F-22 | Login issues a signed JWT access token and a refresh token |
| F-23 | Access tokens expire after 1 hour; refresh tokens expire after 30 days |
| F-24 | Logout revokes the refresh token |

---

## 4. Non-Functional Requirements

These targets are appropriate for a self-hosted open-source project running on modest infrastructure — not a commercial SaaS.

### 4.1 Performance

| Metric | Target |
|--------|--------|
| Redirect latency (p50) | < 50ms |
| Redirect latency (p99) | < 200ms |
| API response time (p99) | < 500ms |

These are achievable on a single low-cost VM or free-tier cloud instance. Redis caching on the redirect path brings latency well within these targets.

### 4.2 Reliability

- The application should recover gracefully from transient DB or cache failures
- Redirects should fall back to a direct DB lookup if Redis is unavailable
- No uptime SLA — this is a self-hosted project

### 4.3 Rate Limiting

| Actor | Limit |
|-------|-------|
| Anonymous (by IP) | 10 shortens / hour |
| Registered user | 200 shortens / hour |

Rate-limited responses return HTTP 429 with a `Retry-After` header. Limits are configurable via environment variables so self-hosters can adjust to their needs.

### 4.4 Data Retention

- Soft-deleted links purged after 30 days
- No raw click data in v1

---

## 5. Tech Stack

These choices are final for v1.

### 5.1 Backend — Go

Go is the right choice for this project for two reasons. First, redirect latency is the most user-visible metric, and Go's native performance and goroutine model make hitting sub-50ms targets trivial even on minimal hardware. Second, Go compiles to a single static binary — no runtime to install, no dependency hell — which makes it significantly easier for self-hosters to build and run the project.

The API is structured as a standard Go project using the `chi` router. Database access uses `sqlc` (type-safe SQL from `.sql` files, no ORM magic). This keeps the code readable and the queries explicit.

### 5.2 Database — PostgreSQL

The data is relational — links belong to users, slugs must be globally unique, expiration is a first-class field. PostgreSQL enforces these constraints at the schema level. A partial index on `slug WHERE deleted_at IS NULL` keeps the redirect hot path fast. No exotic features required; any Postgres 15+ instance works.

### 5.3 Cache & Rate Limiting — Redis / Upstash

Redis handles two jobs: slug-to-URL caching on the redirect path, and rate limit counters. Both are simple key-value operations that Redis handles at sub-millisecond speed.

**Upstash** is the recommended option for the author's hosted instance — it's a serverless Redis API-compatible service with a generous free tier, no server to manage, and it works identically to standard Redis from the application's perspective.

Self-hosters running Docker Compose can use a standard Redis container instead.

The application code talks to Redis via the standard `go-redis` client and is indifferent to whether it's Upstash or a self-hosted Redis instance — just swap the connection string.

### 5.4 Frontend — Next.js on Vercel

Next.js with the App Router gives server-side rendering for the home page (useful for link previews and SEO) and a React SPA for the dashboard. TypeScript throughout. The project tracks the latest stable Next.js / React release.

The author deploys to Vercel's free tier — git-push deploy, automatic preview URLs per pull request, custom domain support included. Self-hosters can deploy the Next.js app anywhere that supports Node.js (Docker, Render, Railway, etc.) or export it as a static site if they remove SSR-dependent features.

### 5.5 Summary

| Layer | Choice | Notes |
|-------|--------|-------|
| API & Redirect | **Go** + `chi` router | Single binary, fast, easy to self-host |
| Database queries | **sqlc** | Type-safe SQL, no ORM |
| Database | **PostgreSQL 15+** | Relational integrity, partial indexes |
| Cache | **Redis** / **Upstash** | Slug cache + rate limiting; same client code for both |
| Frontend | **Next.js** (App Router, TypeScript) | SSR home page + SPA dashboard; track latest stable |
| Frontend hosting | **Vercel** (author's instance) | Free tier, git-push deploys |
| Auth | **JWT (RS256)** | Stateless; access + refresh token pattern |
| DB migrations | **golang-migrate** | Version-controlled SQL migrations |
| CI | **GitHub Actions** | Lint, test, build on every PR |
| License | **MIT** | Permissive, self-host-friendly |

---

## 6. System Architecture

```
┌─────────────┐     ┌──────────────────────┐
│   Browser   │────▶│     Go HTTP Server    │
└─────────────┘     │                       │
                    │  ┌─────────────────┐  │
                    │  │ Redirect Handler│  │──▶ Redis (slug cache)
                    │  └─────────────────┘  │         │
                    │  ┌─────────────────┐  │         │ cache miss
                    │  │   API Handler   │  │         ▼
                    │  │  (CRUD + auth)  │  │──▶ PostgreSQL
                    │  └─────────────────┘  │
                    └──────────────────────┘
                                ▲
                                │ API calls
                    ┌──────────────────────┐
                    │   Next.js Frontend   │
                    │   (Vercel)           │
                    └──────────────────────┘
```

The Go server runs as a single binary handling both redirect and API routes. This is intentional for v1 — one process to deploy, one set of logs to read, one binary to build. Splitting into two services adds operational complexity that isn't justified at this stage. If the redirect path ever needs to be scaled independently, it can be extracted later.

The frontend is a separate Next.js app hosted on Vercel. It communicates with the Go API over HTTPS.

---

## 7. Data Models

### 7.1 `users`

```sql
CREATE TABLE users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,             -- bcrypt hash
  role          TEXT NOT NULL DEFAULT 'user', -- 'user' | 'admin'
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 7.2 `links`

```sql
CREATE TABLE links (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug         TEXT NOT NULL,              -- uniqueness enforced by partial index below
  original_url TEXT NOT NULL,
  user_id      UUID REFERENCES users(id),  -- NULL = anonymous
  is_custom    BOOLEAN NOT NULL DEFAULT false,
  expires_at   TIMESTAMPTZ,
  deleted_at   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Redirect hot path: fast slug lookup, ignores deleted rows
CREATE UNIQUE INDEX idx_links_slug_active
  ON links(slug)
  WHERE deleted_at IS NULL;

-- Dashboard: list links by user
CREATE INDEX idx_links_user_id ON links(user_id);
```

### 7.3 `refresh_tokens`

```sql
CREATE TABLE refresh_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT UNIQUE NOT NULL,  -- SHA-256 of the raw token
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

> Click tracking tables are absent from v1. See [[#14. v1.1 — Deferred Features]].

---

## 8. API Specification

Base URL: `https://your-domain.com/api/v1`

Authenticated endpoints require `Authorization: Bearer <access_token>`.

### 8.1 Shorten a URL

```
POST /links
```

**Request:**
```json
{
  "url": "https://www.example.com/very/long/path",
  "alias": "my-alias",
  "expires_at": "2026-12-31T23:59:59Z"
}
```
`alias` and `expires_at` are optional. Works without authentication.

**Response `201 Created`:**
```json
{
  "id": "018f...",
  "slug": "my-alias",
  "short_url": "https://your-domain.com/my-alias",
  "original_url": "https://www.example.com/very/long/path",
  "expires_at": "2026-12-31T23:59:59Z",
  "created_at": "2026-06-27T10:00:00Z"
}
```

### 8.2 Get Link

```
GET /links/:slug
```

Returns the link object. Registered users only see their own links; admins see all.

### 8.3 Update a Link

```
PATCH /links/:slug
```

Accepts any subset of `{ url, alias, expires_at }`. Requires ownership or admin role.

### 8.4 Delete a Link

```
DELETE /links/:slug
```

Soft-deletes and immediately invalidates the cache entry. Returns `204 No Content`.

### 8.5 List My Links

```
GET /links?page=1&limit=20&sort=created_at&order=desc&q=searchterm
```

**Response `200 OK`:**
```json
{
  "data": [ /* link objects */ ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 143
  }
}
```

### 8.6 Auth Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/register` | Create account (`email`, `password`) |
| `POST` | `/auth/login` | Returns `access_token` + `refresh_token` |
| `POST` | `/auth/refresh` | Exchange refresh token for new access token |
| `POST` | `/auth/logout` | Revoke current refresh token |

### 8.7 Error Envelope

```json
{
  "error": {
    "code": "ALIAS_TAKEN",
    "message": "That alias is already in use.",
    "status": 409
  }
}
```

---

## 9. Frontend Requirements

### 9.1 Pages

| Page | Route | Auth Required |
|------|-------|--------------|
| Home / Shorten | `/` | No |
| Dashboard | `/dashboard` | Yes |
| Login | `/login` | No |
| Register | `/register` | No |
| Branded 404 | — | No |
| Branded 410 (expired) | — | No |

### 9.2 Home Page

- URL input with a "Shorten" button
- "Advanced" toggle for alias and expiration fields
- Result shown inline with a one-click copy button
- Anonymous users see the result and are invited (not forced) to register to manage it

### 9.3 Dashboard

- Table of the user's links: slug, short URL (with copy), destination (truncated), created date, expiration status
- Search by slug or destination URL
- Sort by created date or expiration date
- **Edit** (modal: change URL, alias, expiry) and **Delete** (with confirmation) actions
- Pagination controls
- Empty state with a prompt to create the first link

> No analytics charts or click counts in v1 — the dashboard is for link management only.

---

## 10. Security Requirements

| ID | Requirement |
|----|-------------|
| S-01 | All traffic served over HTTPS; HTTP redirects to HTTPS |
| S-02 | JWTs signed with RS256; private key stored as an environment secret |
| S-03 | All inputs validated and sanitized to prevent SQL injection and XSS |
| S-04 | Rate limiting enforced via Redis counters |
| S-05 | Passwords hashed with bcrypt (cost ≥ 12); never logged |
| S-06 | Destination URLs checked against Google Safe Browsing before a link is created |
| S-07 | Admin endpoints require a JWT with `role: admin` claim |
| S-08 | CORS restricted to known origins in production |
| S-09 | Secrets injected via environment variables; `.env.example` provided in repo, never `.env` |
| S-10 | Refresh tokens stored as SHA-256 hashes, never plaintext |

---

## 11. Error Handling

| Scenario | HTTP Status | Message |
|----------|-------------|---------|
| Invalid URL format | 422 | "Please enter a valid URL starting with http:// or https://" |
| URL flagged as unsafe | 422 | "This URL has been flagged as unsafe and cannot be shortened." |
| Alias already taken | 409 | "That alias is already in use. Please choose another." |
| Slug not found | 404 | Branded 404 page |
| Link expired | 410 | Branded expiration page |
| Unauthenticated | 401 | "Please log in to perform this action." |
| Forbidden | 403 | "You don't have permission to do that." |
| Rate limited | 429 | "Too many requests. Please try again in {n} seconds." |
| Server error | 500 | "Something went wrong. Please try again." |

---

## 12. Deployment Options

The application is designed to be self-hostable. Below are the documented options — the author's hosted instance uses the "Recommended for Author" column.

### 12.1 Go Backend

| Option | Cost | Notes |
|--------|------|-------|
| **Railway** | Free trial → ~$5/mo | Simplest: git-push deploy, managed env vars, built-in logs |
| **Fly.io** | Free tier (3 VMs) | Good free allowance; requires `fly.toml` config; Docker-based |
| **Render** | Free (sleeps after 15min idle) | Not suitable for the redirect path due to cold starts |
| **Self-hosted VPS** (Hetzner, DigitalOcean) | ~$4–6/mo | Most control; requires nginx + systemd setup; excellent for a project like this |
| **Docker Compose** | Infrastructure cost only | Recommended for self-hosters running on their own server |

> **Recommended for author's instance:** Railway or a $4/mo Hetzner VPS. Both are straightforward and cost-predictable.

### 12.2 PostgreSQL

| Option | Cost | Notes |
|--------|------|-------|
| **Neon** | Free (500MB) | Serverless Postgres; zero ops; branches for dev/prod |
| **Supabase** | Free (500MB) | Postgres + dashboard UI; good for visibility |
| **Railway** | Included if using Railway for Go | Simplest if already on Railway |
| **Self-hosted** | Infrastructure cost only | Standard Postgres in Docker Compose |

### 12.3 Redis / Cache

| Option | Cost | Notes |
|--------|------|-------|
| **Upstash** | Free (10k cmds/day) | Serverless Redis; recommended for author's instance; no server to manage |
| **Redis Cloud** | Free (30MB) | Alternative to Upstash |
| **Self-hosted Redis** | Infrastructure cost only | Use in Docker Compose; straightforward |

### 12.4 Frontend

| Option | Cost | Notes |
|--------|------|-------|
| **Vercel** | Free | Recommended; git-push deploy, preview URLs, custom domain |
| **Netlify** | Free | Alternative; similar feature set |
| **Self-hosted** | Infrastructure cost only | `next build` + `next start` in Docker |

### 12.5 Domain

| Option | Cost | Notes |
|--------|------|-------|
| **Cloudflare Registrar** | At-cost (~$10–12/yr for `.io`) | Cheapest; DNS managed in same dashboard; no markup |
| **Namecheap** | Similar pricing | Good alternative |

### 12.6 Estimated Cost (Author's Instance)

| Service | Monthly |
|---------|---------|
| Go backend (Railway or Hetzner VPS) | ~$4–5 |
| PostgreSQL (Neon free tier) | $0 |
| Redis (Upstash free tier) | $0 |
| Frontend (Vercel free) | $0 |
| Domain (amortised) | ~$1 |
| **Total** | **~$5–6/month** |

---

## 13. Local Development

The repo ships with a `docker-compose.yml` that starts PostgreSQL and Redis locally with a single command. The Go API and Next.js frontend run outside Docker during development for fast iteration.

```
docker-compose up -d        # start Postgres + Redis
make migrate                # run DB migrations
make dev                    # start Go API with hot reload (air)
cd frontend && npm run dev  # start Next.js dev server
```

A `.env.example` file documents every required environment variable. Developers copy it to `.env.local` and fill in their values. No secrets are ever committed to the repository.

Required environment variables:

```
# Database
DATABASE_URL=postgres://...

# Redis / Upstash
REDIS_URL=redis://...

# JWT
JWT_PRIVATE_KEY=...   # RS256 private key (PEM)
JWT_PUBLIC_KEY=...    # RS256 public key (PEM)

# Google Safe Browsing
SAFE_BROWSING_API_KEY=...

# App
BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
```

---

## 14. v1.1 — Deferred Features

### Analytics

Click analytics are the most natural v1.1 addition. The data model and architecture are deliberately left clean in v1 to make this easy to add:

- Record a click event per redirect (timestamp, IP hash, user agent, referrer)
- Geo enrichment via MaxMind GeoLite2
- Per-link stats: total clicks, unique visitors, clicks over time, top countries, top referrers
- Simple click counter on the dashboard as a first step (before full analytics)

### Other v1.1 Items

- OAuth 2.0 login (Google)
- API keys for programmatic access (distinct from session JWTs)
- QR code generation per link
- Click count displayed on dashboard (counter only, before full analytics)

---

## 15. Out of Scope

Not planned for any version:

- Paid tiers or billing
- Team/org accounts
- Browser extensions
- Native mobile apps
- Link password protection
- Bulk import/export
- Custom domain support per user (BYOD)

---

## 16. Milestones & Timeline

| Milestone | Deliverable | Target |
|-----------|-------------|--------|
| M1 — Foundation | Go project structure, DB schema, migrations, Docker Compose for local dev | Week 1 |
| M2 — Redirect | Slug generation, Redis cache, redirect endpoint (301/302/404/410) | Week 2 |
| M3 — API | Link CRUD, Safe Browsing check, rate limiting | Week 3–4 |
| M4 — Auth | Register, login, JWT + refresh token, logout | Week 5 |
| M5 — Frontend | Next.js setup, home page, login/register, dashboard | Week 6–8 |
| M6 — Polish | Error handling, `.env.example`, README, basic tests | Week 9 |
| M7 — Deploy | Author's instance live at custom domain, GitHub repo public | Week 10 |

---

## 17. Open Questions

- [ ] **301 vs 302 default:** 301 means browsers cache the redirect permanently and future click events (in v1.1) won't be tracked for returning visitors. Should we default to 302 now in anticipation of analytics?
- [ ] **Anonymous link ownership:** If an anonymous user creates a link and later registers, can they claim it?
- [ ] **Slug collision strategy:** Random retry on collision is simplest for v1 given low probability at 7 chars — confirm this is acceptable.
- [ ] **Safe Browsing API key:** Free tier limit is 10k lookups/day. Sufficient given this is a small self-hosted project — confirm.

---

*Author: [[Ahmed Sukhera]] | Last updated: 2026-06-27 | Version: 1.2 | License: MIT*
