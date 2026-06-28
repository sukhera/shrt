# Contributing to shrt

Thanks for your interest in improving shrt! This guide covers how to run the
project locally, the branch and PR workflow, and the code style we follow.

## Running locally

See the [Local development](README.md#local-development) section of the README
for full setup. In short:

```bash
cp .env.example .env
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem
make tools        # installs air, sqlc, migrate, golangci-lint
make docker-up    # Postgres + Redis
make migrate-up
make dev          # backend :8080 + frontend :3000
```

## Project layout

- `backend/` — Go API + redirect server (single binary). Business logic lives in
  `store/`; HTTP concerns in `server/`. No service layer.
- `frontend/` — Next.js App Router. API calls go through `lib/api.ts`; shared
  types in `types/api.ts`.
- See `IMPLEMENTATION-PLAN.md` for the full architecture and `backend/CLAUDE.md`
  / `frontend/CLAUDE.md` for coding standards.

## Branch & PR workflow

- `main` — always deployable; no direct pushes.
- `dev` — integration branch; feature PRs merge here first.
- Branch names: `feat/<short-description>`, `fix/<short-description>`,
  `chore/<short-description>`.
- Open one PR per logical change into `dev`. Keep PRs focused and reviewable.

Before opening a PR, run the checks below and make sure they pass.

## Required checks

```bash
# Backend
cd backend
go vet ./...
golangci-lint run ./...
go test -race ./...
go build ./cmd/shrt

# Frontend
cd frontend
npm run type-check       # tsc --noEmit
npm run lint             # eslint .
npm run build
npm run e2e              # Playwright (needs the full stack running)
```

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(redirect): add Redis cache with TTL fallback
fix(auth): correct bcrypt cost factor
chore(ci): add golangci-lint to backend workflow
test(links): integration test for expired link 410 response
docs(readme): add deployment guide
```

Scope is the area touched (`auth`, `links`, `redirect`, `frontend`, etc.).

## Code style

**Go**
- Wrap errors with context: `fmt.Errorf("doing thing: %w", err)`; match with
  `errors.Is` / `errors.As`, never string comparison.
- Sentinel errors live in `store/errors.go` and map to HTTP codes in
  `server/response.go`.
- Never call `os.Getenv` outside `internal/config/`.
- Never hand-edit sqlc-generated files; change `db/queries/*.sql` and run
  `make sqlc`.
- Every handler has a corresponding integration test.

**TypeScript / React**
- Server Components by default; add `"use client"` only when needed.
- All backend calls go through `lib/api.ts` — never `fetch` directly in a
  component or hook.
- All shared types in `types/api.ts`. No `any`.
- Use semantic Tailwind tokens (`bg-background`, `text-primary`) so dark mode
  works — never raw colours like `bg-white`.
- shadcn components in `components/ui/` are generated; don't hand-edit them.

## Reporting issues

Open a GitHub issue with steps to reproduce, expected vs. actual behaviour, and
your environment (OS, Go/Node versions). For security issues, please disclose
privately rather than in a public issue.
