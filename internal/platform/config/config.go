package config

import (
	"errors"
	"fmt"
	"net/url"
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
	Firebase      FirebaseConfig
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
	Host              string
	Port              uint16
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type WorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int32
	MaxAttempts  int32
	RetryBackoff time.Duration
	LeaseTimeout time.Duration
}

type FirebaseConfig struct {
	Enabled         bool
	CredentialsFile string
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
			Host:              envOrDefault(lookup, "DB_HOST", ""),
			Port:              uint16OrDefault(lookup, "DB_PORT", 5432),
			User:              envOrDefault(lookup, "DB_USER", ""),
			Password:          envOrDefault(lookup, "DB_PASSWORD", ""),
			Name:              envOrDefault(lookup, "DB_NAME", ""),
			SSLMode:           envOrDefault(lookup, "DB_SSL_MODE", "disable"),
			MaxConns:          int32OrDefault(lookup, "DB_MAX_CONNS", 10),
			MinConns:          int32OrDefault(lookup, "DB_MIN_CONNS", 1),
			MaxConnLifetime:   durationOrDefault(lookup, "DB_MAX_CONN_LIFETIME", 30*time.Minute),
			MaxConnIdleTime:   durationOrDefault(lookup, "DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
			HealthCheckPeriod: durationOrDefault(lookup, "DB_HEALTH_CHECK_PERIOD", 30*time.Second),
		},
		Firebase: FirebaseConfig{
			Enabled:         boolOrDefault(lookup, "FIREBASE_ENABLED", false),
			CredentialsFile: envOrDefault(lookup, "FIREBASE_CREDENTIALS_FILE", ""),
		},
		Worker: WorkerConfig{
			PollInterval: durationOrDefault(lookup, "WORKER_POLL_INTERVAL", 5*time.Second),
			BatchSize:    int32OrDefault(lookup, "WORKER_BATCH_SIZE", 10),
			MaxAttempts:  int32OrDefault(lookup, "WORKER_MAX_ATTEMPTS", 3),
			RetryBackoff: durationOrDefault(lookup, "WORKER_RETRY_BACKOFF", time.Minute),
			LeaseTimeout: durationOrDefault(lookup, "WORKER_LEASE_TIMEOUT", 30*time.Second),
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

	if strings.TrimSpace(cfg.Database.Host) == "" {
		return Config{}, errors.New("DB_HOST is required")
	}

	if strings.TrimSpace(cfg.Database.User) == "" {
		return Config{}, errors.New("DB_USER is required")
	}

	if strings.TrimSpace(cfg.Database.Password) == "" {
		return Config{}, errors.New("DB_PASSWORD is required")
	}

	if strings.TrimSpace(cfg.Database.Name) == "" {
		return Config{}, errors.New("DB_NAME is required")
	}

	if cfg.Firebase.Enabled && strings.TrimSpace(cfg.Firebase.CredentialsFile) == "" {
		return Config{}, errors.New("FIREBASE_CREDENTIALS_FILE is required when FIREBASE_ENABLED=true")
	}

	return cfg, nil
}

func (c DatabaseConfig) ConnectionString() string {
	query := url.Values{}
	query.Set("sslmode", c.SSLMode)

	connectionURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.Name,
		RawQuery: query.Encode(),
	}

	return connectionURL.String()
}

func (c StorageConfig) EndpointURL() string {
	endpoint := strings.TrimSpace(c.Endpoint)
	if endpoint == "" {
		return ""
	}

	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}

	scheme := "http"
	if c.UseSSL {
		scheme = "https"
	}

	return scheme + "://" + endpoint
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

func uint16OrDefault(lookup envLookup, key string, fallback uint16) uint16 {
	value := envOrDefault(lookup, key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		panic(fmt.Sprintf("invalid uint16 value for %s: %v", key, err))
	}

	return uint16(parsed)
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
