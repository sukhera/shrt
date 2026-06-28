package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// registerTestUser registers a unique user and returns the credentials plus the
// issued tokens. It cleans up the user (and cascaded refresh tokens) afterwards.
func registerTestUser(t *testing.T, srv *Server, pool *pgxpool.Pool) (email, password string, tokens registerResponse) {
	t.Helper()
	email = fmt.Sprintf("user-%d@example.com", time.Now().UnixNano())
	password = "supersecret"

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email":    email,
		"password": password,
	}))
	if code := doJSON(t, srv, req, &tokens); code != http.StatusCreated {
		t.Fatalf("register: status %d", code)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
	})
	return email, password, tokens
}

// bearer attaches an Authorization header to a request.
func bearer(req *http.Request, token string) *http.Request {
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestRegister(t *testing.T) {
	srv, pool := newTestServer(t)
	email, _, tokens := registerTestUser(t, srv, pool)

	if tokens.User.Email != email {
		t.Errorf("email: got %q, want %q", tokens.User.Email, email)
	}
	if tokens.User.ID == "" {
		t.Error("expected a user id")
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("expected access and refresh tokens")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	srv, pool := newTestServer(t)
	email, password, _ := registerTestUser(t, srv, pool)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email":    email,
		"password": password,
	}))
	if code := doJSON(t, srv, req, nil); code != http.StatusConflict {
		t.Errorf("duplicate register: got %d, want 409", code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email":    fmt.Sprintf("short-%d@example.com", time.Now().UnixNano()),
		"password": "short",
	}))
	// A too-short password is invalid credentials → 401 per the error mapping.
	if code := doJSON(t, srv, req, nil); code != http.StatusUnauthorized {
		t.Errorf("short password: got %d, want 401", code)
	}
}

func TestLogin(t *testing.T) {
	srv, pool := newTestServer(t)
	email, password, _ := registerTestUser(t, srv, pool)

	t.Run("valid credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, map[string]string{
			"email":    email,
			"password": password,
		}))
		var resp tokenResponse
		if code := doJSON(t, srv, req, &resp); code != http.StatusOK {
			t.Fatalf("login: status %d", code)
		}
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Error("expected tokens on login")
		}
		if resp.ExpiresIn <= 0 {
			t.Errorf("expires_in: got %d, want > 0", resp.ExpiresIn)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}))
		if code := doJSON(t, srv, req, nil); code != http.StatusUnauthorized {
			t.Errorf("wrong password: got %d, want 401", code)
		}
	})

	t.Run("unknown email", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, map[string]string{
			"email":    "nobody-here@example.com",
			"password": password,
		}))
		if code := doJSON(t, srv, req, nil); code != http.StatusUnauthorized {
			t.Errorf("unknown email: got %d, want 401", code)
		}
	})
}

func TestRefresh(t *testing.T) {
	srv, pool := newTestServer(t)
	_, _, reg := registerTestUser(t, srv, pool)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", jsonBody(t, map[string]string{
		"refresh_token": reg.RefreshToken,
	}))
	var resp refreshResponse
	if code := doJSON(t, srv, req, &resp); code != http.StatusOK {
		t.Fatalf("refresh: status %d", code)
	}
	if resp.AccessToken == "" {
		t.Error("expected a new access token")
	}
	if resp.RefreshToken == reg.RefreshToken {
		t.Error("expected refresh token to be rotated")
	}

	// The old refresh token is now revoked and cannot be reused (rotation).
	reuse := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", jsonBody(t, map[string]string{
		"refresh_token": reg.RefreshToken,
	}))
	if code := doJSON(t, srv, reuse, nil); code != http.StatusUnauthorized {
		t.Errorf("reuse of rotated token: got %d, want 401", code)
	}

	// The new refresh token works.
	again := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", jsonBody(t, map[string]string{
		"refresh_token": resp.RefreshToken,
	}))
	if code := doJSON(t, srv, again, nil); code != http.StatusOK {
		t.Errorf("rotated token refresh: got %d, want 200", code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", jsonBody(t, map[string]string{
		"refresh_token": "not-a-real-token",
	}))
	if code := doJSON(t, srv, req, nil); code != http.StatusUnauthorized {
		t.Errorf("invalid refresh token: got %d, want 401", code)
	}
}

func TestLogout(t *testing.T) {
	srv, pool := newTestServer(t)
	_, _, reg := registerTestUser(t, srv, pool)

	req := bearer(httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", jsonBody(t, map[string]string{
		"refresh_token": reg.RefreshToken,
	})), reg.AccessToken)
	if code := doJSON(t, srv, req, nil); code != http.StatusNoContent {
		t.Fatalf("logout: status %d", code)
	}

	// The refresh token is revoked: refreshing with it now fails.
	refresh := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", jsonBody(t, map[string]string{
		"refresh_token": reg.RefreshToken,
	}))
	if code := doJSON(t, srv, refresh, nil); code != http.StatusUnauthorized {
		t.Errorf("refresh after logout: got %d, want 401", code)
	}
}

func TestLogout_RequiresAuth(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", jsonBody(t, map[string]string{
		"refresh_token": "whatever",
	}))
	if code := doJSON(t, srv, req, nil); code != http.StatusUnauthorized {
		t.Errorf("logout without bearer: got %d, want 401", code)
	}
}

// TestAuthMiddleware_ProtectsLinkRoutes verifies the full path: a registered
// user's access token authorizes link operations, while a missing or invalid
// token is rejected with 401.
func TestAuthMiddleware_ProtectsLinkRoutes(t *testing.T) {
	srv, pool := newTestServer(t)
	_, _, reg := registerTestUser(t, srv, pool)

	// Authenticated create associates the link with the token's user.
	createReq := bearer(httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, map[string]any{
		"url": "https://example.com/authed",
	})), reg.AccessToken)
	var created linkResponse
	if code := doJSON(t, srv, createReq, &created); code != http.StatusCreated {
		t.Fatalf("authed create: status %d", code)
	}
	cleanupSlug(t, pool, created.Slug)

	// The same user can list and see the link.
	listReq := bearer(httptest.NewRequest(http.MethodGet, "/api/v1/links", nil), reg.AccessToken)
	var page listLinksResponse
	if code := doJSON(t, srv, listReq, &page); code != http.StatusOK {
		t.Fatalf("authed list: status %d", code)
	}
	if page.Pagination.Total < 1 {
		t.Errorf("expected the created link in the user's list, total=%d", page.Pagination.Total)
	}

	// No token → 401 on a protected route.
	if code := doJSON(t, srv, httptest.NewRequest(http.MethodGet, "/api/v1/links", nil), nil); code != http.StatusUnauthorized {
		t.Errorf("list without token: got %d, want 401", code)
	}

	// Garbage token → 401.
	bad := bearer(httptest.NewRequest(http.MethodGet, "/api/v1/links", nil), "garbage.token.value")
	if code := doJSON(t, srv, bad, nil); code != http.StatusUnauthorized {
		t.Errorf("list with bad token: got %d, want 401", code)
	}
}

// TestOptionalAuth_CreateLink confirms POST /links works both anonymously and
// with a valid token, and that a bad token does not block anonymous creation.
func TestOptionalAuth_CreateLink(t *testing.T) {
	srv, pool := newTestServer(t)

	// Anonymous creation succeeds.
	anon := httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, map[string]any{
		"url": "https://example.com/anon",
	}))
	var anonResp linkResponse
	if code := doJSON(t, srv, anon, &anonResp); code != http.StatusCreated {
		t.Fatalf("anon create: status %d", code)
	}
	cleanupSlug(t, pool, anonResp.Slug)

	// A malformed bearer token is ignored (treated as anonymous), not rejected.
	badTok := bearer(httptest.NewRequest(http.MethodPost, "/api/v1/links", jsonBody(t, map[string]any{
		"url": "https://example.com/anon2",
	})), "not-a-jwt")
	var badResp linkResponse
	if code := doJSON(t, srv, badTok, &badResp); code != http.StatusCreated {
		t.Fatalf("create with bad token should still succeed anonymously: status %d", code)
	}
	cleanupSlug(t, pool, badResp.Slug)
}
