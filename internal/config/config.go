package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config aggregates configuration used by services.
type Config struct {
	Env          string
	HTTP         HTTPConfig
	Firebase     FirebaseConfig
	YDB          YDBConfig
	Feature      FeatureConfig
	TTS          TTSConfig
	TTSControl   TTSControlConfig
	Sync         SyncConfig
	Predictor    PredictorConfig
	Dialog       DialogHelperConfig
	DialogWorker DialogWorkerConfig
	JWT          JWTConfig
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
	APIKey          string
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
	ProxyEnabled         bool
	BaseURL              string
	ServiceToken         string
	ServiceJWTPrivateKey ed25519.PrivateKey
	ServiceJWTKeyID      string
	ServiceJWTTTL        time.Duration
	MaxAudioBytes        int64
	Timeout              time.Duration
}

// TTSControlConfig controls the isolated TTS control-plane foundation.
type TTSControlConfig struct {
	Enabled                      bool
	PostgresDSN                  string
	RedisAddr                    string
	RedisUsername                string
	RedisPassword                string
	RedisDB                      int
	TokenSigningKey              string
	PreviousTokenSigningKey      string
	IPHashKey                    string
	AnonymousDailyChunks         int
	AuthenticatedDailyChunks     int
	AnonymousMaxChunks           int
	AuthenticatedMaxChunks       int
	AnonymousGlobalDailyBudget   int
	AnonymousGlobalMonthlyBudget int
	TrustedProxyHops             int
	ChunkChars                   int
}

// SyncConfig controls sync-worker behavior.
type SyncConfig struct {
	PollInterval    time.Duration
	StreamEnabled   bool
	StreamPath      string
	StreamReconnect time.Duration
}

// PredictorConfig controls Yandex Predictor API integration.
type PredictorConfig struct {
	APIKey string
}

// DialogHelperConfig controls dialog-helper integration.
type DialogHelperConfig struct {
	BaseURL       string
	APIKey        string
	MaxAudioBytes int64
	Timeout       time.Duration
}

// DialogWorkerConfig controls dialog suggestion worker behavior.
type DialogWorkerConfig struct {
	PollInterval time.Duration
	FolderID     string
	ModelURI     string
}

// JWTConfig controls JWT token generation and validation.
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	CookieDomain         string
	CookieSecure         bool
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

	credentialsJSON, err := firebaseCredentialsJSON()
	if err != nil {
		return cfg, err
	}
	cfg.Firebase = FirebaseConfig{
		ProjectID:       getenv("FIREBASE_PROJECT_ID", ""),
		DatabaseURL:     getenv("FIREBASE_DATABASE_URL", ""),
		CredentialsFile: getenv("FIREBASE_CREDENTIALS_FILE", ""),
		CredentialsJSON: credentialsJSON,
		APIKey:          getenv("FIREBASE_API_KEY", ""),
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

	ttsServiceJWTConfigured := os.Getenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64") != "" || os.Getenv("TTS_SERVICE_JWT_KEY_ID") != ""
	ttsServiceJWTPrivateKey, err := ttsServiceJWTPrivateKey()
	if err != nil {
		return cfg, err
	}
	ttsServiceJWTTTL, err := ttsServiceJWTTTL(ttsServiceJWTConfigured)
	if err != nil {
		return cfg, err
	}
	cfg.TTS = TTSConfig{
		ProxyEnabled:         getenvBool("TTS_PROXY_ENABLED", false),
		BaseURL:              getenv("TTS_BASE_URL", "https://tts.linka.su"),
		ServiceToken:         getenv("TTS_SERVICE_TOKEN", ""),
		ServiceJWTPrivateKey: ttsServiceJWTPrivateKey,
		ServiceJWTKeyID:      getenv("TTS_SERVICE_JWT_KEY_ID", ""),
		ServiceJWTTTL:        ttsServiceJWTTTL,
		MaxAudioBytes:        int64(getenvInt("TTS_MAX_AUDIO_BYTES", 50*1024*1024)),
		Timeout:              getenvDuration("TTS_TIMEOUT", 120*time.Second),
	}
	if err := validateTTSServiceJWT(cfg.TTS); err != nil {
		return cfg, err
	}
	ttsControlEnabled := getenvBool("TTS_CONTROL_PLANE_ENABLED", false)
	ttsRedisDB, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_REDIS_DB", 0)
	if err != nil {
		return cfg, err
	}
	ttsAnonymousDailyChunks, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_ANONYMOUS_DAILY_CHUNKS", 30)
	if err != nil {
		return cfg, err
	}
	ttsAuthenticatedDailyChunks, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_AUTH_DAILY_CHUNKS", 200)
	if err != nil {
		return cfg, err
	}
	ttsAnonymousMaxChunks, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_ANONYMOUS_MAX_CHUNKS", 5)
	if err != nil {
		return cfg, err
	}
	ttsAuthenticatedMaxChunks, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_AUTH_MAX_CHUNKS", 21)
	if err != nil {
		return cfg, err
	}
	ttsAnonymousGlobalDailyBudget, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_ANONYMOUS_GLOBAL_DAILY_BUDGET", 0)
	if err != nil {
		return cfg, err
	}
	ttsAnonymousGlobalMonthlyBudget, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_ANONYMOUS_GLOBAL_MONTHLY_BUDGET", 0)
	if err != nil {
		return cfg, err
	}
	ttsTrustedProxyHops, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_TRUSTED_PROXY_HOPS", 1)
	if err != nil {
		return cfg, err
	}
	ttsChunkChars, err := getenvTTSControlInt(ttsControlEnabled, "TTS_CONTROL_CHUNK_CHARS", 240)
	if err != nil {
		return cfg, err
	}
	cfg.TTSControl = TTSControlConfig{
		Enabled:                      ttsControlEnabled,
		PostgresDSN:                  getenv("TTS_CONTROL_POSTGRES_DSN", ""),
		RedisAddr:                    getenv("TTS_CONTROL_REDIS_ADDR", ""),
		RedisUsername:                getenv("TTS_CONTROL_REDIS_USERNAME", ""),
		RedisPassword:                getenv("TTS_CONTROL_REDIS_PASSWORD", ""),
		RedisDB:                      ttsRedisDB,
		TokenSigningKey:              getenv("TTS_CONTROL_TOKEN_SIGNING_KEY", ""),
		PreviousTokenSigningKey:      getenv("TTS_CONTROL_TOKEN_PREVIOUS_SIGNING_KEY", ""),
		IPHashKey:                    getenv("TTS_CONTROL_IP_HASH_KEY", ""),
		AnonymousDailyChunks:         ttsAnonymousDailyChunks,
		AuthenticatedDailyChunks:     ttsAuthenticatedDailyChunks,
		AnonymousMaxChunks:           ttsAnonymousMaxChunks,
		AuthenticatedMaxChunks:       ttsAuthenticatedMaxChunks,
		AnonymousGlobalDailyBudget:   ttsAnonymousGlobalDailyBudget,
		AnonymousGlobalMonthlyBudget: ttsAnonymousGlobalMonthlyBudget,
		TrustedProxyHops:             ttsTrustedProxyHops,
		ChunkChars:                   ttsChunkChars,
	}

	cfg.Sync = SyncConfig{
		PollInterval:    getenvDuration("SYNC_POLL_INTERVAL", 5*time.Second),
		StreamEnabled:   getenvBool("SYNC_STREAM_ENABLED", false),
		StreamPath:      getenv("SYNC_STREAM_PATH", "users"),
		StreamReconnect: getenvDuration("SYNC_STREAM_RECONNECT", 5*time.Second),
	}

	predictorKey := getenv("YANDEX_PREDICTOR_API_KEY", "")
	if predictorKey == "" {
		predictorKey = getenv("PREDICTOR_API_KEY", "")
	}
	cfg.Predictor = PredictorConfig{
		APIKey: predictorKey,
	}

	cfg.Dialog = DialogHelperConfig{
		BaseURL:       getenv("DIALOG_HELPER_URL", ""),
		APIKey:        getenv("DIALOG_HELPER_API_KEY", ""),
		MaxAudioBytes: int64(getenvInt("DIALOG_HELPER_MAX_AUDIO_BYTES", 8*1024*1024)),
		Timeout:       getenvDuration("DIALOG_HELPER_TIMEOUT", 20*time.Second),
	}

	cfg.DialogWorker = DialogWorkerConfig{
		PollInterval: getenvDuration("DIALOG_WORKER_INTERVAL", 15*time.Second),
		FolderID:     getenv("YC_FOLDER_ID", ""),
		ModelURI:     getenv("YC_GPT_MODEL_URI", ""),
	}

	cfg.JWT = JWTConfig{
		Secret:               getenv("JWT_SECRET", ""),
		AccessTokenDuration:  getenvDuration("JWT_ACCESS_TOKEN_DURATION", time.Hour),
		RefreshTokenDuration: getenvDuration("JWT_REFRESH_TOKEN_DURATION", 90*24*time.Hour),
		CookieDomain:         getenv("JWT_COOKIE_DOMAIN", ""),
		CookieSecure:         getenvBool("JWT_COOKIE_SECURE", cfg.Env != "dev"),
	}

	if cfg.Feature.CohortPercent < 0 || cfg.Feature.CohortPercent > 100 {
		return cfg, fmt.Errorf("FEATURE_COHORT_PERCENT must be between 0 and 100")
	}
	if err := validateTTSControl(cfg.TTSControl); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func validateTTSControl(cfg TTSControlConfig) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.PostgresDSN == "" || cfg.RedisAddr == "" {
		return fmt.Errorf("TTS_CONTROL_POSTGRES_DSN and TTS_CONTROL_REDIS_ADDR are required when TTS_CONTROL_PLANE_ENABLED=true")
	}
	if len(cfg.TokenSigningKey) < 32 || (cfg.PreviousTokenSigningKey != "" && len(cfg.PreviousTokenSigningKey) < 32) || len(cfg.IPHashKey) < 32 {
		return fmt.Errorf("TTS_CONTROL_TOKEN_SIGNING_KEY and TTS_CONTROL_IP_HASH_KEY must each be at least 32 bytes when TTS_CONTROL_PLANE_ENABLED=true")
	}
	if cfg.RedisDB < 0 || cfg.TrustedProxyHops < 0 || cfg.AnonymousDailyChunks <= 0 || cfg.AuthenticatedDailyChunks <= 0 || cfg.AnonymousMaxChunks <= 0 || cfg.AuthenticatedMaxChunks <= 0 || cfg.ChunkChars <= 0 {
		return fmt.Errorf("TTS control limits must be positive and TTS_CONTROL_REDIS_DB must not be negative")
	}
	if cfg.AnonymousGlobalDailyBudget < 0 || cfg.AnonymousGlobalMonthlyBudget < 0 {
		return fmt.Errorf("TTS control anonymous global budgets must not be negative")
	}
	return nil
}

func ttsServiceJWTPrivateKey() (ed25519.PrivateKey, error) {
	encoded := os.Getenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64")
	if encoded == "" {
		return nil, nil
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode TTS_SERVICE_JWT_PRIVATE_KEY_BASE64: %w", err)
	}
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64 must decode to %d bytes", ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(key), nil
}

func ttsServiceJWTTTL(configured bool) (time.Duration, error) {
	const maxTTL = 5 * time.Minute
	if !configured {
		return maxTTL, nil
	}
	value := os.Getenv("TTS_SERVICE_JWT_TTL")
	if value == "" {
		return maxTTL, nil
	}
	ttl, err := time.ParseDuration(value)
	if err != nil || ttl <= 0 || ttl > maxTTL {
		return 0, fmt.Errorf("TTS_SERVICE_JWT_TTL must be positive and no more than %s", maxTTL)
	}
	return ttl, nil
}

func validateTTSServiceJWT(cfg TTSConfig) error {
	if (len(cfg.ServiceJWTPrivateKey) == 0) != (cfg.ServiceJWTKeyID == "") {
		return fmt.Errorf("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64 and TTS_SERVICE_JWT_KEY_ID must be configured together")
	}
	return nil
}

func getenvStrictInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return parsed, nil
}

func getenvTTSControlInt(enabled bool, key string, fallback int) (int, error) {
	if !enabled {
		return fallback, nil
	}
	return getenvStrictInt(key, fallback)
}

func firebaseCredentialsJSON() (string, error) {
	if json := getenv("FIREBASE_CREDENTIALS_JSON", ""); json != "" {
		return json, nil
	}
	if b64 := getenv("FIREBASE_CREDENTIALS_B64", ""); b64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return "", fmt.Errorf("decode FIREBASE_CREDENTIALS_B64: %w", err)
		}
		return string(decoded), nil
	}
	return "", nil
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
