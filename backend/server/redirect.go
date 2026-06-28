package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sukhera/shrt/backend/store"
)

// handleRedirect resolves GET /{slug} to its destination:
//
//   - unknown / deleted slug → 404
//   - expired link           → 410
//   - link with an expiry     → 302 (temporary; not browser-cached)
//   - link without an expiry  → 301 (permanent; browser-cacheable)
//
// When DEFAULT_REDIRECT_CODE is set to 302, all valid redirects use 302
// regardless of expiry (preserves future analytics tracking).
func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	link, err := s.store.GetBySlug(r.Context(), slug)
	switch {
	case errors.Is(err, store.ErrNotFound):
		respondError(w, http.StatusNotFound, "LINK_NOT_FOUND", "That short link does not exist.")
		return
	case errors.Is(err, store.ErrExpired):
		respondError(w, http.StatusGone, "LINK_EXPIRED", "This short link has expired.")
		return
	case err != nil:
		respondError(w, http.StatusInternalServerError, "INTERNAL", "Something went wrong. Please try again.")
		return
	}

	http.Redirect(w, r, link.OriginalURL, s.redirectStatus(link))
}

// redirectStatus picks the HTTP status for a valid redirect. A link with an
// expiry always uses 302 (it must not be cached past its lifetime); a permanent
// link uses the configured default (301 or 302).
func (s *Server) redirectStatus(link *store.Link) int {
	if link.ExpiresAt != nil {
		return http.StatusFound // 302
	}
	if s.cfg.DefaultRedirectCode == http.StatusMovedPermanently {
		return http.StatusMovedPermanently // 301
	}
	return http.StatusFound // 302
}
