package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sukhera/shrt/backend/store"
)

// linkResponse is the API representation of a link (IMPLEMENTATION-PLAN.md § 4.1).
type linkResponse struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	IsCustom    bool       `json:"is_custom"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClickCount  *int64     `json:"click_count,omitempty"`
}

// toLinkResponse adapts a store.LinkDetail into the API shape, building the full
// short URL from the configured base URL.
func (s *Server) toLinkResponse(d *store.LinkDetail) linkResponse {
	return linkResponse{
		ID:          d.ID,
		Slug:        d.Slug,
		ShortURL:    strings.TrimRight(s.cfg.BaseURL, "/") + "/" + d.Slug,
		OriginalURL: d.OriginalURL,
		IsCustom:    d.IsCustom,
		ExpiresAt:   d.ExpiresAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// createLinkRequest is the POST /links body.
type createLinkRequest struct {
	URL       string     `json:"url"`
	Alias     string     `json:"alias"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// handleCreateLink shortens a URL. Auth is optional: an authenticated request
// associates the link with the user; an anonymous one creates an ownerless link.
func (s *Server) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	var req createLinkRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.URL == "" {
		respondError(w, http.StatusUnprocessableEntity, "INVALID_URL", "A url is required.")
		return
	}

	in := store.CreateLinkInput{
		URL:       req.URL,
		Alias:     req.Alias,
		ExpiresAt: req.ExpiresAt,
	}
	if userID, ok := userIDFromContext(r.Context()); ok {
		in.UserID = &userID
	}

	link, err := s.store.CreateLink(r.Context(), in)
	if err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, s.toLinkResponse(link))
}

// listLinksResponse is the GET /links body.
type listLinksResponse struct {
	Data       []linkResponse `json:"data"`
	Pagination pagination     `json:"pagination"`
}

type pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// handleListLinks returns the authenticated user's paginated links.
func (s *Server) handleListLinks(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}

	in := parseListParams(r, userID)
	links, total, err := s.store.ListByUser(r.Context(), in)
	if err != nil {
		respondStoreError(w, err)
		return
	}

	// Fetch click counts in one query (no N+1).
	clickCounts, _ := s.store.GetClickCountsByUser(r.Context(), userID)

	data := make([]linkResponse, len(links))
	for i := range links {
		lr := s.toLinkResponse(&links[i])
		if clickCounts != nil {
			if c, ok := clickCounts[links[i].ID]; ok {
				lr.ClickCount = &c
			} else {
				zero := int64(0)
				lr.ClickCount = &zero
			}
		}
		data[i] = lr
	}
	respondJSON(w, http.StatusOK, listLinksResponse{
		Data:       data,
		Pagination: pagination{Page: in.Page, Limit: in.Limit, Total: total},
	})
}

// parseListParams reads and clamps the list query parameters, applying defaults
// from the API contract (page 1, limit 20 capped at 100, newest first).
func parseListParams(r *http.Request, userID string) store.ListLinksInput {
	q := r.URL.Query()

	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	limit := atoiDefault(q.Get("limit"), 20)
	switch {
	case limit < 1:
		limit = 20
	case limit > 100:
		limit = 100
	}

	sort := q.Get("sort")
	if sort != "expires_at" {
		sort = "created_at"
	}
	order := q.Get("order")
	if order != "asc" {
		order = "desc"
	}

	return store.ListLinksInput{
		UserID: userID,
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Order:  order,
		Search: q.Get("q"),
	}
}

// handleGetLink returns a single link owned by the authenticated user.
func (s *Server) handleGetLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}
	slug := chi.URLParam(r, "slug")

	link, err := s.store.GetDetailBySlug(r.Context(), userID, slug)
	if err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, s.toLinkResponse(link))
}

// updateLinkRequest is the PATCH /links/:slug body. Pointers distinguish an
// omitted field from a provided one; expiresAtSet captures an explicit null.
type updateLinkRequest struct {
	URL       *string    `json:"url"`
	Alias     *string    `json:"alias"`
	ExpiresAt *time.Time `json:"expires_at"`

	expiresAtSet bool
}

// UnmarshalJSON records whether expires_at was present in the body so the handler
// can tell "clear the expiry" (null) from "leave it unchanged" (absent).
func (u *updateLinkRequest) UnmarshalJSON(data []byte) error {
	type alias updateLinkRequest
	if err := json.Unmarshal(data, (*alias)(u)); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	_, u.expiresAtSet = raw["expires_at"]
	return nil
}

// handleUpdateLink applies a partial update to a user's link.
func (s *Server) handleUpdateLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}
	slug := chi.URLParam(r, "slug")

	var req updateLinkRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	in := store.UpdateLinkInput{
		URL:         req.URL,
		Alias:       req.Alias,
		ExpiresAt:   req.ExpiresAt,
		ClearExpiry: req.expiresAtSet && req.ExpiresAt == nil,
	}

	link, err := s.store.UpdateLink(r.Context(), userID, slug, in)
	if err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, s.toLinkResponse(link))
}

// handleDeleteLink soft-deletes a user's link.
func (s *Server) handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required.")
		return
	}
	slug := chi.URLParam(r, "slug")

	if err := s.store.DeleteLink(r.Context(), userID, slug); err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

// decodeJSON reads a JSON body into v, rejecting unknown fields and oversized
// bodies. On failure it writes a 422 and returns the error so the caller stops.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "INVALID_URL", "Request body is not valid JSON.")
		return err
	}
	return nil
}

// atoiDefault parses s as an int, returning def when s is empty or malformed.
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
