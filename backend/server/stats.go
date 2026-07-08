package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// handleGetLinkStats returns 30-day click data for a single link.
//
//	GET /api/v1/links/{slug}/stats
//
// The response includes the total click count and a daily breakdown (with
// zero-days filled in) suitable for rendering a sparkline.
func (s *Server) handleGetLinkStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}
	slug := chi.URLParam(r, "slug")

	// Verify ownership via the existing detail lookup.
	link, err := s.store.GetDetailBySlug(r.Context(), userID, slug)
	if err != nil {
		respondStoreError(w, err)
		return
	}

	stats, err := s.store.GetLinkStats(r.Context(), link.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL", "Couldn't load stats.")
		return
	}
	respondJSON(w, http.StatusOK, stats)
}
