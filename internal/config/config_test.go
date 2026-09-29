package config

import (
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
