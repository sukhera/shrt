package server

import (
	"net"
	"net/http"
	"time"
)

// rateLimitWindow is the fixed window for link-creation rate limiting (per hour,
// matching the RATE_LIMIT_* config which is expressed in shortens per hour).
const rateLimitWindow = time.Hour

// cors applies a minimal CORS policy: it allows the configured frontend origin
// and the standard methods/headers the API uses. Preflight OPTIONS requests are
// answered immediately. Auth and rate-limiting middleware arrive in later
// milestones.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.cfg.FrontendURL {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimitCreate limits link-creation requests. Authenticated users are limited
// per user ID at the higher RATE_LIMIT_USER threshold; anonymous requests are
// limited per client IP at RATE_LIMIT_ANON. The window is one hour. Redis
// failures fail open inside the store, so this never blocks on an outage.
func (s *Server) rateLimitCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, limit := s.rateLimitIdentity(r)
		key := "shorten:" + identity

		if !s.store.RateLimit(r.Context(), key, limit, rateLimitWindow) {
			respondError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please slow down.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimitIdentity returns the rate-limit bucket key and the applicable limit
// for a request: the user ID and user limit when authenticated, otherwise the
// client IP and anonymous limit.
func (s *Server) rateLimitIdentity(r *http.Request) (string, int) {
	if userID, ok := userIDFromContext(r.Context()); ok {
		return "user:" + userID, s.cfg.RateLimitUser
	}
	return "ip:" + clientIP(r), s.cfg.RateLimitAnon
}

// clientIP returns the request's remote IP. RealIP middleware has already
// normalised RemoteAddr from X-Forwarded-For/X-Real-IP where trusted, so we
// strip any port and use what remains.
func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
