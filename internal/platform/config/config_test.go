package config

import (
	"testing"
	"time"
)

func TestLoadFromEnvUsesDefaults(t *testing.T) {
	cfg, err := loadFromEnv(func(key string) (string, bool) {
		values := map[string]string{
			"DB_HOST":     "localhost",
			"DB_PORT":     "5433",
			"DB_USER":     "coralhub",
			"DB_PASSWORD": "coralhub",
			"DB_NAME":     "coralhub",
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

	if cfg.HTTP.HandlerTimeout != 30*time.Second {
		t.Fatalf("HTTP.HandlerTimeout = %v, want %v", cfg.HTTP.HandlerTimeout, 30*time.Second)
	}

	if cfg.Firebase.Enabled {
		t.Fatal("Firebase.Enabled = true, want false by default")
	}

	if cfg.Firebase.CredentialsFile != "" {
		t.Fatalf("Firebase.CredentialsFile = %q, want empty string", cfg.Firebase.CredentialsFile)
	}

	if cfg.Worker.BatchSize != 10 {
		t.Fatalf("Worker.BatchSize = %d, want %d", cfg.Worker.BatchSize, 10)
	}

	if cfg.Worker.MaxAttempts != 3 {
		t.Fatalf("Worker.MaxAttempts = %d, want %d", cfg.Worker.MaxAttempts, 3)
	}

	if cfg.Worker.RetryBackoff != time.Minute {
		t.Fatalf("Worker.RetryBackoff = %v, want %v", cfg.Worker.RetryBackoff, time.Minute)
	}

	if cfg.Worker.LeaseTimeout != 30*time.Second {
		t.Fatalf("Worker.LeaseTimeout = %v, want %v", cfg.Worker.LeaseTimeout, 30*time.Second)
	}

	if cfg.Worker.NotificationRetention != 720*time.Hour {
		t.Fatalf("Worker.NotificationRetention = %v, want %v", cfg.Worker.NotificationRetention, 720*time.Hour)
	}

	if cfg.Database.ConnectionString() != "postgres://coralhub:coralhub@localhost:5433/coralhub?sslmode=disable" {
		t.Fatalf("Database.ConnectionString() = %q", cfg.Database.ConnectionString())
	}
}

func TestLoadFromEnvRequiresSplitDatabaseFields(t *testing.T) {
	_, err := loadFromEnv(func(key string) (string, bool) {
		return "", false
	})
	if err == nil {
		t.Fatal("expected error when split database config is missing")
	}
}

func TestLoadFromEnvRequiresFirebaseCredentialsWhenEnabled(t *testing.T) {
	_, err := loadFromEnv(func(key string) (string, bool) {
		values := map[string]string{
			"DB_HOST":          "localhost",
			"DB_PORT":          "5433",
			"DB_USER":          "coralhub",
			"DB_PASSWORD":      "coralhub",
			"DB_NAME":          "coralhub",
			"FIREBASE_ENABLED": "true",
		}
		value, ok := values[key]
		return value, ok
	})
	if err == nil {
		t.Fatal("expected error when firebase is enabled without credentials file")
	}
}
