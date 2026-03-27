package config

import (
	"testing"
	"time"
)

func TestLoadFromEnvUsesDefaults(t *testing.T) {
	cfg, err := loadFromEnv(func(key string) (string, bool) {
		values := map[string]string{
			"DATABASE_URL": "postgres://coralhub:coralhub@localhost:5432/coralhub?sslmode=disable",
		}
		value, ok := values[key]
		return value, ok
	})
	if err != nil {
		t.Fatalf("loadFromEnv() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}

	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":8080")
	}

	if cfg.Worker.PollInterval != 5*time.Second {
		t.Fatalf("Worker.PollInterval = %v, want %v", cfg.Worker.PollInterval, 5*time.Second)
	}
}

func TestLoadFromEnvRequiresDatabaseURL(t *testing.T) {
	_, err := loadFromEnv(func(key string) (string, bool) {
		return "", false
	})
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}
