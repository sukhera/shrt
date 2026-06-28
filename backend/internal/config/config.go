// Package config loads environment variables into a typed Config struct.
//
// This is the only package permitted to read environment variables via
// os.Getenv (or os.LookupEnv). Everything else receives a *Config.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration, loaded once at startup.
type Config struct {
	Port        string
	Env         string
	BaseURL     string
	FrontendURL string

	DatabaseURL string
	RedisURL    string

	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration

	SafeBrowsingAPIKey string

	RateLimitAnon int
	RateLimitUser int

	SlugLength          int
	DefaultRedirectCode int
}

// Load reads configuration from the process environment. It panics if a
// required variable is missing or malformed — the app must not start in a
// half-configured state.
func Load() *Config {
	cfg := &Config{
		Port:        getDefault("PORT", "8080"),
		Env:         getDefault("ENV", "development"),
		BaseURL:     mustGet("BASE_URL"),
		FrontendURL: getDefault("FRONTEND_URL", "http://localhost:3000"),

		DatabaseURL: mustGet("DATABASE_URL"),
		RedisURL:    mustGet("REDIS_URL"),

		JWTPrivateKeyPath: getDefault("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
		JWTPublicKeyPath:  getDefault("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
		JWTAccessTTL:      getDuration("JWT_ACCESS_TTL", time.Hour),
		JWTRefreshTTL:     getDuration("JWT_REFRESH_TTL", 720*time.Hour),

		SafeBrowsingAPIKey: os.Getenv("SAFE_BROWSING_API_KEY"), // optional — empty skips the check

		RateLimitAnon: getInt("RATE_LIMIT_ANON", 10),
		RateLimitUser: getInt("RATE_LIMIT_USER", 200),

		SlugLength:          getInt("SLUG_LENGTH", 7),
		DefaultRedirectCode: getInt("DEFAULT_REDIRECT_CODE", 302),
	}

	return cfg
}

// mustGet returns the value of key or panics if it is unset/empty.
func mustGet(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	return v
}

// getDefault returns the value of key, or def if it is unset/empty.
func getDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// getInt parses key as an integer, or panics if it is set but malformed.
func getInt(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("config: %s must be an integer, got %q", key, v))
	}
	return n
}

// getDuration parses key as a Go duration (e.g. "1h", "720h"), or panics if
// it is set but malformed.
func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Sprintf("config: %s must be a duration (e.g. 1h), got %q", key, v))
	}
	return d
}
