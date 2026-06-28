-include .env
export

BACKEND_DIR := backend

.PHONY: dev dev-api dev-web build test lint sqlc migrate-up migrate-down docker-up docker-down tools

## Start both backend (air) and frontend (npm) — Ctrl+C stops both
dev:
	@echo "→ Starting API on :8080 and frontend on :3000 (Ctrl+C to stop both)"
	@trap 'kill 0' SIGINT; \
	(cd $(BACKEND_DIR) && air -c .air.toml) & \
	(cd frontend && npm run dev) & \
	wait

## Start backend only
dev-api:
	cd $(BACKEND_DIR) && air -c .air.toml

## Start frontend only
dev-web:
	cd frontend && npm run dev

## Build the Go binary
build:
	cd $(BACKEND_DIR) && go build -o bin/shrt ./cmd/shrt

## Run Go tests with race detector
test:
	cd $(BACKEND_DIR) && go test -race ./...

## Run golangci-lint
lint:
	cd $(BACKEND_DIR) && golangci-lint run ./...

## Regenerate sqlc code from db/queries/*.sql
sqlc:
	cd $(BACKEND_DIR) && sqlc generate

## Run all pending database migrations
migrate-up:
	migrate -path $(BACKEND_DIR)/db/migrations -database "$(DATABASE_URL)" up

## Roll back the last migration
migrate-down:
	migrate -path $(BACKEND_DIR)/db/migrations -database "$(DATABASE_URL)" down 1

## Start Postgres and Redis via Docker Compose
docker-up:
	docker compose up -d

## Stop Docker Compose services
docker-down:
	docker compose down

## Install Go dev tools (air, golangci-lint, sqlc, migrate)
tools:
	go install github.com/air-verse/air@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest
