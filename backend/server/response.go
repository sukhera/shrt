package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sukhera/shrt/backend/store"
)

// errorBody is the error envelope returned for all API errors, per the API
// contract in IMPLEMENTATION-PLAN.md § 4.
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"error"`
}

// respondJSON writes v as a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response failed", "err", err)
	}
}

// respondError writes the standard error envelope.
func respondError(w http.ResponseWriter, status int, code, message string) {
	var body errorBody
	body.Error.Code = code
	body.Error.Message = message
	body.Error.Status = status
	respondJSON(w, status, body)
}

// respondStoreError maps a store sentinel error to the contract's error code and
// HTTP status. Unrecognised errors are logged and surface as a generic 500 so
// internal details never leak to clients.
func respondStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		respondError(w, http.StatusNotFound, "LINK_NOT_FOUND", "That short link does not exist.")
	case errors.Is(err, store.ErrAliasTaken):
		respondError(w, http.StatusConflict, "ALIAS_TAKEN", "That alias is already in use.")
	case errors.Is(err, store.ErrInvalidURL):
		respondError(w, http.StatusUnprocessableEntity, "INVALID_URL", "That URL is not valid.")
	case errors.Is(err, store.ErrUnsafeURL):
		respondError(w, http.StatusUnprocessableEntity, "UNSAFE_URL", "That URL was flagged as unsafe.")
	case errors.Is(err, store.ErrForbidden):
		respondError(w, http.StatusForbidden, "FORBIDDEN", "You do not have access to that link.")
	case errors.Is(err, store.ErrEmailTaken):
		respondError(w, http.StatusConflict, "EMAIL_TAKEN", "That email is already registered.")
	case errors.Is(err, store.ErrInvalidCredentials):
		respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect email or password.")
	case errors.Is(err, store.ErrInvalidToken):
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid or has expired.")
	default:
		slog.Error("unhandled store error", "err", err)
		respondError(w, http.StatusInternalServerError, "INTERNAL", "Something went wrong. Please try again.")
	}
}
