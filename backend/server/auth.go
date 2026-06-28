package server

import (
	"net/http"

	"github.com/sukhera/shrt/backend/store"
)

// userResponse is the public representation of a user in auth responses — it
// never includes the password hash.
type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func toUserResponse(u store.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email}
}

// credentialsRequest is the shared body for register and login.
type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// registerResponse is the POST /auth/register body (API contract § 4.2).
type registerResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// handleRegister creates an account and returns the user plus an initial token
// pair.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	res, err := s.store.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, registerResponse{
		User:         toUserResponse(res.User),
		AccessToken:  res.Tokens.AccessToken,
		RefreshToken: res.Tokens.RefreshToken,
	})
}

// tokenResponse is the POST /auth/login body (API contract § 4.2).
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// handleLogin verifies credentials and returns a fresh token pair.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	res, err := s.store.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  res.Tokens.AccessToken,
		RefreshToken: res.Tokens.RefreshToken,
		ExpiresIn:    res.Tokens.ExpiresIn,
	})
}

// refreshRequest is the POST /auth/refresh body.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// refreshResponse is the POST /auth/refresh body. The contract returns a new
// access token; we also rotate and return a new refresh token so the client can
// replace the one it just spent.
type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// handleRefresh exchanges a valid refresh token for a new token pair, rotating
// the refresh token.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	if req.RefreshToken == "" {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "A refresh token is required.")
		return
	}

	tokens, err := s.store.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondStoreError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, refreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// handleLogout revokes the supplied refresh token. It always returns 204:
// revoking an unknown or already-revoked token is a no-op, since the goal — that
// the token can no longer be used — is satisfied either way.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.RefreshToken != "" {
		if err := s.store.Logout(r.Context(), req.RefreshToken); err != nil {
			respondStoreError(w, err)
			return
		}
	}
	respondJSON(w, http.StatusNoContent, nil)
}
