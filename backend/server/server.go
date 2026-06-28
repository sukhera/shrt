// Package server holds all HTTP concerns for shrt: routing, middleware, and
// handlers. Handlers parse requests, call into store, and write responses —
// they contain no business logic.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sukhera/shrt/backend/internal/config"
	"github.com/sukhera/shrt/backend/store"
)

// Server wires configuration and the store to an HTTP server.
type Server struct {
	cfg   *config.Config
	store *store.Store
	http  *http.Server
}

// New builds a Server with all routes and middleware registered and HTTP
// timeouts configured.
func New(cfg *config.Config, st *store.Store) *Server {
	s := &Server{cfg: cfg, store: st}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	// RealIP trusts X-Forwarded-For / X-Real-IP, which are spoofable when the app
	// is exposed directly. shrt is intended to run behind a trusted reverse proxy
	// that sets these headers; the deployment guide documents that requirement.
	r.Use(middleware.RealIP) //nolint:staticcheck // trusted-proxy deployment; see deployment notes
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.cors)

	r.Get("/health", s.handleHealth)

	// API routes. Auth middleware is added in M3; for now POST /links is
	// anonymous-capable and the rest read a user ID that is not yet populated.
	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/links", func(links chi.Router) {
			links.With(s.rateLimitCreate).Post("/", s.handleCreateLink)
			links.Get("/", s.handleListLinks)
			links.Get("/{slug}", s.handleGetLink)
			links.Patch("/{slug}", s.handleUpdateLink)
			links.Delete("/{slug}", s.handleDeleteLink)
		})
	})

	// Redirect is registered last: its catch-all /{slug} must not shadow the
	// API or health routes above.
	r.Get("/{slug}", s.handleRedirect)

	s.http = &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return s
}

// Handler exposes the configured router for tests.
func (s *Server) Handler() http.Handler {
	return s.http.Handler
}

// Start runs the HTTP server until ListenAndServe returns.
func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// handleHealth reports service health, including DB and Redis connectivity.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, "UNHEALTHY", "dependency check failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
