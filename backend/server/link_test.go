package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// authed wraps a request so handlers see it as coming from the given user. It
// stands in for the M3 auth middleware, which will populate the same context
// value from a validated bearer token.
func authed(req *http.Request, userID string) *http.Request {
	return req.WithContext(withUserID(req.Context(), userID))
}

// doJSON issues req against the server and decodes the JSON response into out
// (when non-nil), returning the status code.
func doJSON(t *testing.T, srv *Server, req *http.Request, out any) int {
	t.Helper()
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if out != nil && rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("decode response (%d): %v\nbody: %s", rec.Code, err, rec.Body.String())
		}
	}
	return rec.Code
}

// newUserID returns a fresh UUID for use as a test owner. There is no users-table
// FK on links in the test schema, so a synthetic id is sufficient.
func newUserID(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), "SELECT gen_random_uuid()").Scan(&id); err != nil {
		t.Fatalf("gen user id: %v", err)
	}
	return id
}

// jsonBody marshals v into a request body reader.
func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return bytes.NewReader(b)
}

// cleanupSlug removes a link by slug after the test.
func cleanupSlug(t *testing.T, pool *pgxpool.Pool, slug string) {
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM links WHERE slug = $1", slug)
	})
}

func TestCreateLink_Anonymous(t *testing.T) {
	srv, pool := newTestServer(t)

	body := map[string]any{"url": "https://example.com/a/long/path"}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, body))

	var resp linkResponse
	code := doJSON(t, srv, req, &resp)
	if code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201", code)
	}
	cleanupSlug(t, pool, resp.Slug)

	if resp.Slug == "" {
		t.Error("expected a generated slug")
	}
	if resp.IsCustom {
		t.Error("generated slug should not be custom")
	}
	if resp.ShortURL != srv.cfg.BaseURL+"/"+resp.Slug {
		t.Errorf("short_url: got %q", resp.ShortURL)
	}
	if resp.OriginalURL != body["url"] {
		t.Errorf("original_url: got %q", resp.OriginalURL)
	}
}

func TestCreateLink_CustomAlias(t *testing.T) {
	srv, pool := newTestServer(t)
	alias := fmt.Sprintf("my-link-%d", time.Now().UnixNano())
	cleanupSlug(t, pool, alias)

	body := map[string]any{"url": "https://example.com", "alias": alias}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, body))

	var resp linkResponse
	if code := doJSON(t, srv, req, &resp); code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201", code)
	}
	if resp.Slug != alias {
		t.Errorf("slug: got %q, want %q", resp.Slug, alias)
	}
	if !resp.IsCustom {
		t.Error("expected is_custom true for a custom alias")
	}

	// A second request with the same alias must conflict.
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, body))
	if code := doJSON(t, srv, req2, nil); code != http.StatusConflict {
		t.Errorf("duplicate alias: got %d, want 409", code)
	}
}

func TestCreateLink_InvalidURL(t *testing.T) {
	srv, _ := newTestServer(t)

	tests := []struct {
		name string
		url  string
	}{
		{"missing url", ""},
		{"bad scheme", "ftp://example.com"},
		{"no host", "https://"},
		{"garbage", "not a url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]any{"url": tt.url}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, body))
			if code := doJSON(t, srv, req, nil); code != http.StatusUnprocessableEntity {
				t.Errorf("status: got %d, want 422", code)
			}
		})
	}
}

func TestListLinks_PaginationAndOwnership(t *testing.T) {
	srv, pool := newTestServer(t)
	user := newUserID(t, pool)
	other := newUserID(t, pool)

	// Create three links for the user and one for someone else.
	for i := 0; i < 3; i++ {
		body := map[string]any{"url": fmt.Sprintf("https://example.com/%d", i)}
		req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, body)), user)
		var resp linkResponse
		if code := doJSON(t, srv, req, &resp); code != http.StatusCreated {
			t.Fatalf("seed create %d: status %d", i, code)
		}
		cleanupSlug(t, pool, resp.Slug)
	}
	otherReq := authed(httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, map[string]any{"url": "https://other.example.com"})), other)
	var otherResp linkResponse
	doJSON(t, srv, otherReq, &otherResp)
	cleanupSlug(t, pool, otherResp.Slug)

	// List for the user — should see exactly their three links.
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/links?limit=2&page=1", nil), user)
	var page listLinksResponse
	if code := doJSON(t, srv, req, &page); code != http.StatusOK {
		t.Fatalf("list: status %d", code)
	}
	if page.Pagination.Total != 3 {
		t.Errorf("total: got %d, want 3", page.Pagination.Total)
	}
	if len(page.Data) != 2 {
		t.Errorf("page size: got %d, want 2 (limit)", len(page.Data))
	}

	// Unauthenticated list is rejected.
	anon := httptest.NewRequest(http.MethodGet, "/api/v1/links", nil)
	if code := doJSON(t, srv, anon, nil); code != http.StatusUnauthorized {
		t.Errorf("anon list: got %d, want 401", code)
	}
}

func TestGetUpdateDeleteLink(t *testing.T) {
	srv, pool := newTestServer(t)
	user := newUserID(t, pool)

	// Create.
	createReq := authed(httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, map[string]any{"url": "https://example.com/original"})), user)
	var created linkResponse
	if code := doJSON(t, srv, createReq, &created); code != http.StatusCreated {
		t.Fatalf("create: status %d", code)
	}
	cleanupSlug(t, pool, created.Slug)

	// Get.
	getReq := authed(httptest.NewRequest(http.MethodGet, "/api/v1/links/"+created.Slug, nil), user)
	var got linkResponse
	if code := doJSON(t, srv, getReq, &got); code != http.StatusOK {
		t.Fatalf("get: status %d", code)
	}
	if got.OriginalURL != "https://example.com/original" {
		t.Errorf("get original_url: got %q", got.OriginalURL)
	}

	// Update destination.
	newURL := "https://example.com/updated"
	patchReq := authed(httptest.NewRequest(http.MethodPatch, "/api/v1/links/"+created.Slug, jsonBody(t, map[string]any{"url": newURL})), user)
	var updated linkResponse
	if code := doJSON(t, srv, patchReq, &updated); code != http.StatusOK {
		t.Fatalf("patch: status %d", code)
	}
	if updated.OriginalURL != newURL {
		t.Errorf("patch original_url: got %q, want %q", updated.OriginalURL, newURL)
	}

	// Another user cannot see or delete it.
	stranger := newUserID(t, pool)
	strangerGet := authed(httptest.NewRequest(http.MethodGet, "/api/v1/links/"+created.Slug, nil), stranger)
	if code := doJSON(t, srv, strangerGet, nil); code != http.StatusNotFound {
		t.Errorf("stranger get: got %d, want 404", code)
	}

	// Delete.
	delReq := authed(httptest.NewRequest(http.MethodDelete, "/api/v1/links/"+created.Slug, nil), user)
	if code := doJSON(t, srv, delReq, nil); code != http.StatusNoContent {
		t.Fatalf("delete: status %d", code)
	}

	// After delete it is gone.
	getAfter := authed(httptest.NewRequest(http.MethodGet, "/api/v1/links/"+created.Slug, nil), user)
	if code := doJSON(t, srv, getAfter, nil); code != http.StatusNotFound {
		t.Errorf("get after delete: got %d, want 404", code)
	}
}

func TestUpdateLink_ClearExpiry(t *testing.T) {
	srv, pool := newTestServer(t)
	user := newUserID(t, pool)

	expiry := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	createReq := authed(httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, map[string]any{
		"url":        "https://example.com",
		"expires_at": expiry,
	})), user)
	var created linkResponse
	if code := doJSON(t, srv, createReq, &created); code != http.StatusCreated {
		t.Fatalf("create: status %d", code)
	}
	cleanupSlug(t, pool, created.Slug)
	if created.ExpiresAt == nil {
		t.Fatal("expected expiry to be set on create")
	}

	// Clearing expiry: explicit null.
	patchReq := authed(httptest.NewRequest(http.MethodPatch, "/api/v1/links/"+created.Slug, bytes.NewReader([]byte(`{"expires_at":null}`))), user)
	var updated linkResponse
	if code := doJSON(t, srv, patchReq, &updated); code != http.StatusOK {
		t.Fatalf("patch: status %d", code)
	}
	if updated.ExpiresAt != nil {
		t.Errorf("expiry: got %v, want nil after clear", updated.ExpiresAt)
	}
}

func TestRateLimit_CreateReturns429(t *testing.T) {
	srv, pool := newTestServer(t)
	// Drive the anonymous limit down so the test is fast and deterministic.
	srv.cfg.RateLimitAnon = 3

	var lastCode int
	var slugs []string
	for i := 0; i < 5; i++ {
		body := map[string]any{"url": fmt.Sprintf("https://example.com/rl/%d", i)}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, body))
		req.RemoteAddr = "203.0.113.7:12345" // stable IP → same bucket
		var resp linkResponse
		lastCode = doJSON(t, srv, req, &resp)
		if resp.Slug != "" {
			slugs = append(slugs, resp.Slug)
		}
	}
	for _, s := range slugs {
		cleanupSlug(t, pool, s)
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("after exceeding limit: got %d, want 429", lastCode)
	}
}
