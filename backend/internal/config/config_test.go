package config

import (
	"testing"
	"time"
)

func TestLoad_RequiredVarsPresent(t *testing.T) {
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port: got %q, want default 8080", cfg.Port)
	}
	if cfg.DefaultRedirectCode != 302 {
		t.Errorf("DefaultRedirectCode: got %d, want default 302", cfg.DefaultRedirectCode)
	}
	if cfg.JWTAccessTTL != time.Hour {
		t.Errorf("JWTAccessTTL: got %v, want default 1h", cfg.JWTAccessTTL)
	}
}

func TestLoad_PanicsOnMissingRequired(t *testing.T) {
	// BASE_URL intentionally unset.
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")

	defer func() {
		if recover() == nil {
			t.Fatal("expected Load to panic when BASE_URL is missing")
		}
	}()

	Load()
}

func TestLoad_PanicsOnMalformedInt(t *testing.T) {
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("SLUG_LENGTH", "notanumber")

	defer func() {
		if recover() == nil {
			t.Fatal("expected Load to panic on malformed SLUG_LENGTH")
		}
	}()

	Load()
}
