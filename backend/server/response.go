package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
