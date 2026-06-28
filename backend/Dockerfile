# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer — only re-runs when go.mod changes)
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/shrt \
    ./cmd/shrt

# ── Runtime stage ─────────────────────────────────────────────────────────────
# distroless: no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/bin/shrt /shrt

EXPOSE 8080

# Run as non-root
USER nonroot:nonroot

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/shrt", "-health"] || exit 1

ENTRYPOINT ["/shrt"]
