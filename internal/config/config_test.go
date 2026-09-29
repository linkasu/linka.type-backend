package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func TestLoadTTSConfig(t *testing.T) {
	t.Setenv("TTS_SERVICE_TOKEN", "service-token")
	t.Setenv("TTS_MAX_AUDIO_BYTES", "1234")
	t.Setenv("TTS_TIMEOUT", "45s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.TTS.ServiceToken != "service-token" {
		t.Fatalf("ServiceToken = %q", cfg.TTS.ServiceToken)
	}
	if cfg.TTS.MaxAudioBytes != 1234 {
		t.Fatalf("MaxAudioBytes = %d", cfg.TTS.MaxAudioBytes)
	}
	if cfg.TTS.Timeout != 45*time.Second {
		t.Fatalf("Timeout = %s", cfg.TTS.Timeout)
	}
}

func TestLoadTTSConfigDefaults(t *testing.T) {
	t.Setenv("TTS_SERVICE_TOKEN", "")
	t.Setenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64", "")
	t.Setenv("TTS_SERVICE_JWT_KEY_ID", "")
	t.Setenv("TTS_SERVICE_JWT_TTL", "")
	t.Setenv("TTS_MAX_AUDIO_BYTES", "")
	t.Setenv("TTS_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.TTS.ServiceToken != "" {
		t.Fatalf("ServiceToken = %q", cfg.TTS.ServiceToken)
	}
	if cfg.TTS.MaxAudioBytes != 50*1024*1024 {
		t.Fatalf("MaxAudioBytes = %d", cfg.TTS.MaxAudioBytes)
	}
	if cfg.TTS.Timeout != 120*time.Second {
		t.Fatalf("Timeout = %s", cfg.TTS.Timeout)
	}
	if cfg.TTS.ServiceJWTTTL != 5*time.Minute {
		t.Fatalf("ServiceJWTTTL = %s", cfg.TTS.ServiceJWTTTL)
	}
}

func TestLoadTTSServiceJWTConfig(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64", base64.StdEncoding.EncodeToString(privateKey))
	t.Setenv("TTS_SERVICE_JWT_KEY_ID", "key-1")
	t.Setenv("TTS_SERVICE_JWT_TTL", "4m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(cfg.TTS.ServiceJWTPrivateKey) != string(privateKey) || cfg.TTS.ServiceJWTKeyID != "key-1" || cfg.TTS.ServiceJWTTTL != 4*time.Minute {
		t.Fatal("TTS service JWT config was not loaded")
	}
}

func TestLoadRejectsInvalidTTSServiceJWTConfig(t *testing.T) {
	t.Setenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64", base64.StdEncoding.EncodeToString(make([]byte, ed25519.PrivateKeySize)))
	t.Setenv("TTS_SERVICE_JWT_KEY_ID", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a TTS service JWT key without key ID")
	}

	t.Setenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64", "not-base64")
	t.Setenv("TTS_SERVICE_JWT_KEY_ID", "key-1")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted malformed TTS service JWT key")
	}

	t.Setenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64", "")
	t.Setenv("TTS_SERVICE_JWT_KEY_ID", "key-1")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a TTS service JWT key ID without key")
	}

	t.Setenv("TTS_SERVICE_JWT_PRIVATE_KEY_BASE64", base64.StdEncoding.EncodeToString(make([]byte, ed25519.PrivateKeySize)))
	t.Setenv("TTS_SERVICE_JWT_KEY_ID", "key-1")
	t.Setenv("TTS_SERVICE_JWT_TTL", "6m")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a TTS service JWT TTL above five minutes")
	}
}

func TestTTSControlDisabledDoesNotRequireDependencies(t *testing.T) {
	t.Setenv("TTS_CONTROL_PLANE_ENABLED", "false")
	t.Setenv("TTS_CONTROL_POSTGRES_DSN", "")
	t.Setenv("TTS_CONTROL_REDIS_ADDR", "")
	if _, err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestTTSControlEnabledRequiresDependencies(t *testing.T) {
	t.Setenv("TTS_CONTROL_PLANE_ENABLED", "true")
	t.Setenv("TTS_CONTROL_POSTGRES_DSN", "postgres://example")
	t.Setenv("TTS_CONTROL_REDIS_ADDR", "redis:6379")
	t.Setenv("TTS_CONTROL_TOKEN_SIGNING_KEY", "12345678901234567890123456789012")
	t.Setenv("TTS_CONTROL_IP_HASH_KEY", "12345678901234567890123456789012")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.TTSControl.Enabled || cfg.TTSControl.AnonymousDailyChunks != 30 || cfg.TTSControl.AuthenticatedMaxChunks != 21 {
		t.Fatalf("TTS control config = %#v", cfg.TTSControl)
	}
}

func TestTTSControlRejectsInvalidConfiguredLimit(t *testing.T) {
	t.Setenv("TTS_CONTROL_PLANE_ENABLED", "true")
	t.Setenv("TTS_CONTROL_ANONYMOUS_DAILY_CHUNKS", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted invalid TTS control limit")
	}
}
