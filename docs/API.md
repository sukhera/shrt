# shrt API Reference

The shrt backend exposes a JSON API for links and authentication, plus a
top-level redirect endpoint. All API routes are versioned under `/api/v1`. The
redirect endpoint (`GET /{slug}`) and the health check live at the root.

- **Base URL (local):** `http://localhost:8080`
- **Content type:** `application/json` for all request and response bodies
- **Timestamps:** ISO 8601 / RFC 3339, UTC (e.g. `2026-06-28T10:00:00Z`)
- **IDs:** UUID v4

A machine-readable [OpenAPI 3.1 spec](../openapi.yaml) is also available.

## Table of contents

- [Authentication](#authentication)
- [Auth endpoints](#auth-endpoints)
  - [Register](#post-apiv1authregister)
  - [Login](#post-apiv1authlogin)
  - [Refresh](#post-apiv1authrefresh)
  - [Logout](#post-apiv1authlogout)
- [Link endpoints](#link-endpoints)
  - [Create link](#post-apiv1links)
  - [List links](#get-apiv1links)
  - [Get link](#get-apiv1linksslug)
  - [Update link](#patch-apiv1linksslug)
  - [Delete link](#delete-apiv1linksslug)
- [Redirect](#redirect)
- [Health](#health)
- [Errors](#errors)
- [Rate limiting](#rate-limiting)

---

## Authentication

shrt uses JWT access tokens signed with RS256.

- **Access token** — short-lived (default 1h). Sent on each request as a
  `Authorization: Bearer <token>` header. Held in memory by the web client.
- **Refresh token** — long-lived (default 30d), opaque, single-use (rotated on
  each refresh). Only its SHA-256 hash is stored server-side.

The access token's `sub` claim is the user ID; a `role` claim carries the user's
role. Tokens that are missing, malformed, or expired yield `401 UNAUTHORIZED`.

> **Web client note:** the Next.js frontend never exposes the refresh token to
> JavaScript. Its route handlers under `/api/auth/*` and `/api/refresh` store the
> refresh token in an httpOnly cookie and proxy to this API. The endpoints below
> describe the **backend** API directly.

| Endpoint | Auth required |
|----------|---------------|
| `POST /api/v1/auth/register` | No |
| `POST /api/v1/auth/login` | No |
| `POST /api/v1/auth/refresh` | No (uses the refresh token in the body) |
| `POST /api/v1/auth/logout` | Yes (Bearer access token) |
| `POST /api/v1/links` | Optional (associates the link with a user if present) |
| `GET /api/v1/links` | Yes |
| `GET /api/v1/links/{slug}` | Yes |
| `PATCH /api/v1/links/{slug}` | Yes |
| `DELETE /api/v1/links/{slug}` | Yes |
| `GET /{slug}` | No (redirect) |
| `GET /health` | No |

---

## Auth endpoints

### `POST /api/v1/auth/register`

Create an account. Returns the user and an initial token pair. The password must
be at least 8 characters.

**Request**

```json
{
  "email": "user@example.com",
  "password": "a-strong-password"
}
```

**Response — `201 Created`**

```json
{
  "user": { "id": "018f4a2b-...", "email": "user@example.com" },
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "9f8c1d2e..."
}
```

**Errors**

| Status | Code | When |
|--------|------|------|
| 409 | `EMAIL_TAKEN` | Email already registered |
| 401 | `INVALID_CREDENTIALS` | Password too short / email malformed |

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"a-strong-password"}'
```

---

### `POST /api/v1/auth/login`

Verify credentials and return a fresh token pair.

**Request**

```json
{ "email": "user@example.com", "password": "a-strong-password" }
```

**Response — `200 OK`**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "9f8c1d2e...",
  "expires_in": 3600
}
```

`expires_in` is the access-token lifetime in seconds.

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `INVALID_CREDENTIALS` | Unknown email or wrong password (intentionally indistinguishable) |

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"a-strong-password"}'
```

---

### `POST /api/v1/auth/refresh`

Exchange a valid refresh token for a new token pair. The presented refresh token
is **rotated**: it is revoked and a new one is issued, so each refresh token can
be used only once.

**Request**

```json
{ "refresh_token": "9f8c1d2e..." }
```

**Response — `200 OK`**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "1a2b3c4d...",
  "expires_in": 3600
}
```

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Refresh token missing, unknown, expired, or already used/revoked |

---

### `POST /api/v1/auth/logout`

Revoke a refresh token. Requires a valid access token. Revoking an unknown or
already-revoked token is a no-op — the response is always `204`.

**Request** — `Authorization: Bearer <access_token>`

```json
{ "refresh_token": "1a2b3c4d..." }
```

**Response — `204 No Content`**

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid access token |

---

## Link endpoints

### The Link object

```json
{
  "id": "018f4a2b-...",
  "slug": "abc1234",
  "short_url": "https://shrt.example.com/abc1234",
  "original_url": "https://www.example.com/long/path",
  "is_custom": false,
  "expires_at": null,
  "created_at": "2026-06-28T10:00:00Z",
  "updated_at": "2026-06-28T10:00:00Z"
}
```

`short_url` is built from the server's `BASE_URL`. `expires_at` is `null` for
links that never expire.

---

### `POST /api/v1/links`

Shorten a URL. Authentication is **optional**: if a valid Bearer token is
present, the link is owned by that user and appears in their dashboard;
otherwise an ownerless (anonymous) link is created. This route is rate limited
(see [Rate limiting](#rate-limiting)).

**Request** — `alias` and `expires_at` are optional

```json
{
  "url": "https://www.example.com/long/path",
  "alias": "my-link",
  "expires_at": "2026-12-31T23:59:59Z"
}
```

- `url` — required; must be an `http`/`https` URL with a host.
- `alias` — optional custom slug; if omitted a random 7-char base62 slug is
  generated.
- `expires_at` — optional expiry timestamp.

**Response — `201 Created`** — a [Link object](#the-link-object).

**Errors**

| Status | Code | When |
|--------|------|------|
| 422 | `INVALID_URL` | Missing/malformed URL, unsupported scheme, or invalid JSON |
| 422 | `UNSAFE_URL` | Flagged by Google Safe Browsing (when enabled) |
| 409 | `ALIAS_TAKEN` | Requested alias is already in use |
| 429 | `RATE_LIMITED` | Rate limit exceeded |

```bash
# Anonymous
curl -X POST http://localhost:8080/api/v1/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.example.com/long/path"}'

# Authenticated (owned link)
curl -X POST http://localhost:8080/api/v1/links \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.example.com","alias":"my-link"}'
```

---

### `GET /api/v1/links`

List the authenticated user's links, paginated.

**Query parameters**

| Param | Type | Default | Notes |
|-------|------|---------|-------|
| `page` | int | `1` | 1-based page number |
| `limit` | int | `20` | Page size, capped at `100` |
| `sort` | string | `created_at` | `created_at` or `expires_at` |
| `order` | string | `desc` | `asc` or `desc` |
| `q` | string | — | Search term (matches slug / original URL) |

**Response — `200 OK`**

```json
{
  "data": [ /* Link objects */ ],
  "pagination": { "page": 1, "limit": 20, "total": 47 }
}
```

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid access token |

```bash
curl "http://localhost:8080/api/v1/links?page=1&limit=20&sort=created_at&order=desc" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### `GET /api/v1/links/{slug}`

Get a single link owned by the authenticated user.

**Response — `200 OK`** — a [Link object](#the-link-object).

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid access token |
| 404 | `LINK_NOT_FOUND` | No such link, or it is not owned by the caller |

> Links owned by another user return `404` (not `403`) — the API does not reveal
> that a slug exists to a non-owner.

---

### `PATCH /api/v1/links/{slug}`

Update a link. Send any subset of fields; omitted fields are unchanged. To clear
an expiry, send `"expires_at": null` explicitly.

**Request**

```json
{
  "url": "https://new-destination.example.com",
  "alias": "new-alias",
  "expires_at": null
}
```

**Response — `200 OK`** — the updated [Link object](#the-link-object).

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid access token |
| 404 | `LINK_NOT_FOUND` | No such link, or not owned by the caller |
| 409 | `ALIAS_TAKEN` | New alias already in use |
| 422 | `INVALID_URL` / `UNSAFE_URL` | New URL invalid or unsafe |

---

### `DELETE /api/v1/links/{slug}`

Soft-delete a link. The slug stops redirecting immediately and the cache is
invalidated. A deleted slug can later be recycled as a new custom alias.

**Response — `204 No Content`**

**Errors**

| Status | Code | When |
|--------|------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid access token |
| 404 | `LINK_NOT_FOUND` | No such link, or not owned by the caller |

---

## Redirect

### `GET /{slug}`

Resolve a slug to its destination. This is the public hot path, served from a
Redis cache in front of Postgres. Not under `/api/v1`.

| Status | Meaning |
|--------|---------|
| `301 Moved Permanently` | Valid link with no expiry, when `DEFAULT_REDIRECT_CODE=301` |
| `302 Found` | Valid link (default), or any link that has an expiry |
| `404 Not Found` | Unknown or deleted slug |
| `410 Gone` | The link has expired |

The destination is returned in the `Location` header. Links with an expiry always
use `302` so they are not cached by browsers past their lifetime.

```bash
curl -i http://localhost:8080/abc1234
# HTTP/1.1 302 Found
# Location: https://www.example.com/long/path
```

---

## Health

### `GET /health`

Liveness/readiness check. Verifies Postgres and Redis connectivity.

- `200 OK` → `{ "status": "ok" }`
- `503 Service Unavailable` → `{ "error": { "code": "UNHEALTHY", ... } }`

---

## Errors

All API errors share one envelope:

```json
{
  "error": {
    "code": "ALIAS_TAKEN",
    "message": "That alias is already in use.",
    "status": 409
  }
}
```

- `code` — stable machine-readable identifier (match on this, not the message).
- `message` — human-readable description; wording may change.
- `status` — mirrors the HTTP status code.

### Error codes

| Code | Status | Meaning |
|------|--------|---------|
| `INVALID_URL` | 422 | Malformed URL, unsupported scheme, or invalid request body |
| `UNSAFE_URL` | 422 | URL flagged by Google Safe Browsing |
| `ALIAS_TAKEN` | 409 | Requested custom alias already exists |
| `EMAIL_TAKEN` | 409 | Email already registered |
| `INVALID_CREDENTIALS` | 401 | Bad email/password, or password too short on register |
| `LINK_NOT_FOUND` | 404 | Slug does not exist, is deleted, or is not owned by the caller |
| `UNAUTHORIZED` | 401 | Missing, invalid, or expired token |
| `FORBIDDEN` | 403 | Authenticated but not permitted |
| `RATE_LIMITED` | 429 | Too many requests |
| `UNHEALTHY` | 503 | A dependency (DB/Redis) is unavailable |
| `INTERNAL` | 500 | Unexpected server error |

---

## Rate limiting

Link creation (`POST /api/v1/links`) is rate limited using a fixed one-hour
window, keyed per identity:

| Caller | Key | Default limit |
|--------|-----|---------------|
| Authenticated | user ID | `RATE_LIMIT_USER` (200/hour) |
| Anonymous | client IP | `RATE_LIMIT_ANON` (10/hour) |

When the limit is exceeded the API responds `429 RATE_LIMITED`. If Redis is
unavailable the limiter fails open (the request is allowed and a warning is
logged), so an outage never blocks shortening.
