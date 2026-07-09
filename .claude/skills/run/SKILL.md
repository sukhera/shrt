---
name: run
description: How to launch and drive the shrt app locally to see a change working or capture screenshots — the sanctioned run path for this repo, including the freeze-safe way to do anything that launches a browser on a low-RAM machine. Use whenever asked to run/start/screenshot shrt, verify a change in the real app, or capture README images. Read this BEFORE running next dev + a local browser — that combination can freeze an 8GB Mac.
---

# Running shrt locally

## Full stack (dev)

```bash
make docker-up      # Postgres + Redis
make migrate-up     # apply migrations
make dev            # API (air) + Next.js (next dev) together
```

Backend: http://localhost:8080  ·  Frontend: http://localhost:3000
Requires `.env` (copy from `.env.example`) and JWT keys in `backend/keys/` — the backend panics at startup without them.

## Screenshots / local Chromium / E2E — the freeze-safe path (IMPORTANT)

`next dev` uses **Turbopack**, which stays resident live-compiling and churns allocations. On a memory-constrained machine (e.g. 8 GB) that churn competes with Chromium's launch spike for the OS's memory reclaim, and the box can thrash into swap and freeze — even though it looks like there's free RAM. **Use a production build for anything that drives a real browser locally:**

```bash
# backend as a prebuilt binary (no resident Go compiler)
cd backend && set -a && . ../.env && set +a && go build -o bin/shrt ./cmd/shrt && ./bin/shrt &

# frontend as a PRODUCTION build, not next dev
cd frontend && npm run build && npm run start &   # next start, port 3000

# then screenshot / Playwright against http://localhost:3000
```

This is exactly how the original tree captured its README screenshots, and it's reproducibly clean where `next dev` + Chromium froze.

### Guardrails on this 8GB Mac

- Check headroom with **`memory_pressure`** (free % — the real gauge), NOT `Pages free` (near-zero by design on macOS). Want a healthy green free %.
- Arm a swap-watch that kills node/next/chromium if `sysctl vm.swapusage` used crosses ~1 GB, BEFORE launching the browser. Watch for a *fast* swap climb — that's the freeze signature, not a static low free-page count.
- Safer still: point E2E at a deployed target via `E2E_FRONTEND_URL` / `E2E_BACKEND_URL`, or let the user drive the browser while Claude reads the log.
- The combination to AVOID: `next dev` (Turbopack) + local Chromium.

## Verify gate before a PR

```bash
cd backend  && go vet ./... && golangci-lint run ./... && go test -race ./... && go build ./cmd/shrt
cd frontend && npx tsc --noEmit && npm run lint && npm run build
```
