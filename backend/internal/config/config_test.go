package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_allDefaults(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	if cfg.Port != "8000" {
		t.Errorf("expected default port 8000, got %s", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected default env development, got %s", cfg.AppEnv)
	}
	if cfg.JWTAccessTTL != 15*time.Minute {
		t.Errorf("expected 15m access TTL, got %v", cfg.JWTAccessTTL)
	}
	if cfg.JWTRefreshTTL != 168*time.Hour {
		t.Errorf("expected 168h refresh TTL, got %v", cfg.JWTRefreshTTL)
	}
	if cfg.CrossmintEnv != "staging" {
		t.Errorf("expected staging crossmint env, got %s", cfg.CrossmintEnv)
	}
}

func TestLoad_fromEnv(t *testing.T) {
	os.Clearenv()
	t.Setenv("PORT", "9999")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_ACCESS_TTL", "5m")
	t.Setenv("DATABASE_URL", "postgres://localhost/mydb")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Errorf("expected port 9999, got %s", cfg.Port)
	}
	if cfg.JWTSecret != "super-secret" {
		t.Errorf("expected secret, got %s", cfg.JWTSecret)
	}
	if cfg.JWTAccessTTL != 5*time.Minute {
		t.Errorf("expected 5m, got %v", cfg.JWTAccessTTL)
	}
	if cfg.DatabaseURL != "postgres://localhost/mydb" {
		t.Errorf("expected db url, got %s", cfg.DatabaseURL)
	}
}
