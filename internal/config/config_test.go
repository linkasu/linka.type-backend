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
