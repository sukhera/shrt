# shrt Documentation

Reference documentation for the shrt URL shortener.

## Contents

| Document | What it covers |
|----------|----------------|
| [API.md](API.md) | REST API reference — every endpoint, request/response shapes, error codes, rate limits, curl examples |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design — components, redirect cache-aside, auth/token flow, data model (with diagrams) |
| [DEVELOPMENT.md](DEVELOPMENT.md) | Working in the codebase — env, make targets, migrations, sqlc, testing, common tasks |
| [`../openapi.yaml`](../openapi.yaml) | Machine-readable OpenAPI 3.1 spec (Swagger UI, client generation, contract testing) |

## Elsewhere in the repo

- [README.md](../README.md) — project overview and quick start
- [CONTRIBUTING.md](../CONTRIBUTING.md) — contribution workflow and code style
- [IMPLEMENTATION-PLAN.md](../IMPLEMENTATION-PLAN.md) — milestone plan and the original API contract
- [URL-Shortener-PRD.md](../URL-Shortener-PRD.md) — product requirements
- [backend/CLAUDE.md](../backend/CLAUDE.md) · [frontend/CLAUDE.md](../frontend/CLAUDE.md) — coding standards

## Viewing the OpenAPI spec

Render `openapi.yaml` interactively with any OpenAPI tool, e.g.:

```bash
npx @redocly/cli preview-docs openapi.yaml   # Redoc preview
# or import it into Swagger UI / Postman / Insomnia
```
