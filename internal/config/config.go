package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config aggregates configuration used by services.
type Config struct {
	Env      string
	HTTP     HTTPConfig
	Firebase FirebaseConfig
	YDB      YDBConfig
	Feature  FeatureConfig
	TTS      TTSConfig
	Sync     SyncConfig
}

// HTTPConfig controls HTTP server behavior.
type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// FirebaseConfig holds Firebase admin and RTDB settings.
type FirebaseConfig struct {
	ProjectID       string
	DatabaseURL     string
	CredentialsFile string
	CredentialsJSON string
}

// YDBConfig holds YDB connection settings.
type YDBConfig struct {
	Endpoint string
	Database string
	Token    string
}

// FeatureConfig controls rollout behavior.
type FeatureConfig struct {
	ReadSource    string
	CohortPercent int
}

// TTSConfig controls the optional proxy.
type TTSConfig struct {
	ProxyEnabled bool
	BaseURL      string
}

// SyncConfig controls sync-worker behavior.
type SyncConfig struct {
	PollInterval    time.Duration
	StreamEnabled   bool
	StreamPath      string
	StreamReconnect time.Duration
}

// Load reads config from environment variables.
func Load() (Config, error) {
	var cfg Config

	cfg.Env = getenv("ENV", "dev")
	cfg.HTTP = HTTPConfig{
		Addr:            httpAddr(),
		ReadTimeout:     getenvDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getenvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getenvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getenvDuration("HTTP_SHUTDOWN_TIMEOUT", 20*time.Second),
	}

	cfg.Firebase = FirebaseConfig{
		ProjectID:       getenv("FIREBASE_PROJECT_ID", ""),
		DatabaseURL:     getenv("FIREBASE_DATABASE_URL", ""),
		CredentialsFile: getenv("FIREBASE_CREDENTIALS_FILE", ""),
		CredentialsJSON: getenv("FIREBASE_CREDENTIALS_JSON", ""),
	}

	cfg.YDB = YDBConfig{
		Endpoint: getenv("YDB_ENDPOINT", ""),
		Database: getenv("YDB_DATABASE", ""),
		Token:    getenv("YDB_TOKEN", ""),
	}

	cfg.Feature = FeatureConfig{
		ReadSource:    getenv("FEATURE_READ_SOURCE", "firebase_only"),
		CohortPercent: getenvInt("FEATURE_COHORT_PERCENT", 0),
	}

	cfg.TTS = TTSConfig{
		ProxyEnabled: getenvBool("TTS_PROXY_ENABLED", false),
		BaseURL:      getenv("TTS_BASE_URL", "https://tts.linka.su"),
	}

	cfg.Sync = SyncConfig{
		PollInterval:    getenvDuration("SYNC_POLL_INTERVAL", 5*time.Second),
		StreamEnabled:   getenvBool("SYNC_STREAM_ENABLED", false),
		StreamPath:      getenv("SYNC_STREAM_PATH", "users"),
		StreamReconnect: getenvDuration("SYNC_STREAM_RECONNECT", 5*time.Second),
	}

	if cfg.Feature.CohortPercent < 0 || cfg.Feature.CohortPercent > 100 {
		return cfg, fmt.Errorf("FEATURE_COHORT_PERCENT must be between 0 and 100")
	}

	return cfg, nil
}

func httpAddr() string {
	if addr := getenv("HTTP_ADDR", ""); addr != "" {
		return addr
	}
	if port := getenv("PORT", ""); port != "" {
		return ":" + port
	}
	return ":8080"
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return parsed
}
