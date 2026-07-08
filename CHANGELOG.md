# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — 2026-07-08

Initial release. URL shortener with authenticated link management, deployed at
[shrt.sukhera.dev](https://shrt.sukhera.dev).

### Added

- **Redirect service** with Redis cache-aside for fast lookups (M1)
- **Link CRUD API** with nanoid slug generation, Google Safe Browsing check, and
  rate limiting (M2)
- **JWT authentication** — register, login, token refresh, logout (M3)
- **Next.js frontend** — home (anonymous shorten), auth pages, link dashboard,
  dark mode (M4)
- **Playwright E2E tests** for critical paths (M5)
- **Combined Dockerfile** for Coolify deployment with Caddy reverse proxy
- API reference, OpenAPI spec, and architecture documentation
- README, CONTRIBUTING guide, MIT license

### Fixed

- Backend URL default for auth routes and theme-toggle hydration (M4)
- API routing: `/api/auth` → Next.js, `/api/v1` → Go backend
- Dockerfile public folder reference and `NEXT_PUBLIC_API_URL` default

[Unreleased]: https://github.com/sukhera/shrt/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/sukhera/shrt/releases/tag/v0.1.0
