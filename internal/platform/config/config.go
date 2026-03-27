package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	HTTP          HTTPConfig
	Database      DatabaseConfig
	Worker        WorkerConfig
	Storage       StorageConfig
	Observability ObservabilityConfig
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type WorkerConfig struct {
	PollInterval time.Duration
}

type StorageConfig struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type ObservabilityConfig struct {
	ServiceName string
	LogLevel    string
}

type envLookup func(string) (string, bool)

func Load() (Config, error) {
	_ = godotenv.Load()

	return loadFromEnv(os.LookupEnv)
}

func loadFromEnv(lookup envLookup) (Config, error) {
	cfg := Config{
		AppEnv: envOrDefault(lookup, "APP_ENV", "development"),
		HTTP: HTTPConfig{
			Addr:         envOrDefault(lookup, "HTTP_ADDR", ":8080"),
			ReadTimeout:  durationOrDefault(lookup, "HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: durationOrDefault(lookup, "HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  durationOrDefault(lookup, "HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:               envOrDefault(lookup, "DATABASE_URL", ""),
			MaxConns:          int32OrDefault(lookup, "DB_MAX_CONNS", 10),
			MinConns:          int32OrDefault(lookup, "DB_MIN_CONNS", 1),
			MaxConnLifetime:   durationOrDefault(lookup, "DB_MAX_CONN_LIFETIME", 30*time.Minute),
			MaxConnIdleTime:   durationOrDefault(lookup, "DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
			HealthCheckPeriod: durationOrDefault(lookup, "DB_HEALTH_CHECK_PERIOD", 30*time.Second),
		},
		Worker: WorkerConfig{
			PollInterval: durationOrDefault(lookup, "WORKER_POLL_INTERVAL", 5*time.Second),
		},
		Storage: StorageConfig{
			Endpoint:  envOrDefault(lookup, "STORAGE_ENDPOINT", "localhost:9000"),
			Bucket:    envOrDefault(lookup, "STORAGE_BUCKET", "coralhub-local"),
			Region:    envOrDefault(lookup, "STORAGE_REGION", "us-east-1"),
			AccessKey: envOrDefault(lookup, "STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey: envOrDefault(lookup, "STORAGE_SECRET_KEY", "minioadmin"),
			UseSSL:    boolOrDefault(lookup, "STORAGE_USE_SSL", false),
		},
		Observability: ObservabilityConfig{
			ServiceName: envOrDefault(lookup, "OTEL_SERVICE_NAME", "coralhub-backend"),
			LogLevel:    strings.ToUpper(envOrDefault(lookup, "LOG_LEVEL", "INFO")),
		},
	}

	if strings.TrimSpace(cfg.Database.URL) == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(lookup envLookup, key string, fallback string) string {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func durationOrDefault(lookup envLookup, key string, fallback time.Duration) time.Duration {
	value := envOrDefault(lookup, key, "")
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("invalid duration for %s: %v", key, err))
	}

	return parsed
}

func int32OrDefault(lookup envLookup, key string, fallback int32) int32 {
	value := envOrDefault(lookup, key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("invalid int value for %s: %v", key, err))
	}

	return int32(parsed)
}

func boolOrDefault(lookup envLookup, key string, fallback bool) bool {
	value := envOrDefault(lookup, key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf("invalid bool value for %s: %v", key, err))
	}

	return parsed
}
