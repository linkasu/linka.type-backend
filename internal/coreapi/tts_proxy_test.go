package coreapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linkasu/linka.type-backend/internal/config"
)

func TestProxyTTSPostHardensUpstreamRequest(t *testing.T) {
	var upstream *http.Request
	api := &API{
		config: config.Config{TTS: config.TTSConfig{BaseURL: "https://tts.example", ServiceToken: "service-token", MaxAudioBytes: 1024}},
		ttsHTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			upstream = req.Clone(req.Context())
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"audio/mpeg"}}, Body: io.NopCloser(bytes.NewBufferString("audio"))}, nil
		})},
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/tts?voice=anna", bytes.NewBufferString("hello"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("X-Request-ID", "request-id")
	req.Header.Set("Authorization", "Bearer client-token")
	req.Header.Set("Cookie", "session=secret")
	req.Header.Set("X-Auth-Token", "secret")
	req.Header.Set("X-Forwarded-For", "203.0.113.1")

	api.proxyTTS(recorder, req)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "audio" {
		t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
	}
	if upstream.URL.String() != "https://tts.example/tts?voice=anna" {
		t.Fatalf("upstream URL = %q", upstream.URL)
	}
	for key, want := range map[string]string{
		"Content-Type":        "application/json",
		"Accept":              "audio/mpeg",
		"X-Request-ID":        "request-id",
		"X-TTS-Service-Token": "service-token",
	} {
		if got := upstream.Header.Get(key); got != want {
			t.Fatalf("upstream %s = %q, want %q", key, got, want)
		}
	}
	for _, key := range []string{"Authorization", "Cookie", "X-Auth-Token", "X-Forwarded-For"} {
		if got := upstream.Header.Get(key); got != "" {
			t.Fatalf("upstream unexpectedly includes %s: %q", key, got)
		}
	}
}

func TestProxyTTSGetHardensUpstreamRequest(t *testing.T) {
	var upstream *http.Request
	api := &API{
		config: config.Config{TTS: config.TTSConfig{BaseURL: "https://tts.example", ServiceToken: "service-token", MaxAudioBytes: 1024}},
		ttsHTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			upstream = req.Clone(req.Context())
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString("audio"))}, nil
		})},
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/tts?voice=anna", nil)
	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("X-Request-ID", "request-id")
	req.Header.Set("Authorization", "Bearer client-token")
	req.Header.Set("X-Forwarded-For", "203.0.113.1")

	api.proxyTTS(recorder, req)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "audio" {
		t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
	}
	if upstream.URL.String() != "https://tts.example/tts?voice=anna" {
		t.Fatalf("upstream URL = %q", upstream.URL)
	}
	if got := upstream.Header.Get("Accept"); got != "audio/mpeg" {
		t.Fatalf("upstream Accept = %q", got)
	}
	if got := upstream.Header.Get("X-Request-ID"); got != "request-id" {
		t.Fatalf("upstream X-Request-ID = %q", got)
	}
	if got := upstream.Header.Get("Authorization"); got != "" {
		t.Fatalf("upstream Authorization = %q", got)
	}
	if got := upstream.Header.Get("X-Forwarded-For"); got != "" {
		t.Fatalf("upstream X-Forwarded-For = %q", got)
	}
	if got := upstream.Header.Get("X-TTS-Service-Token"); got != "service-token" {
		t.Fatalf("upstream X-TTS-Service-Token = %q", got)
	}
}

func TestProxyTTSPostRejectsOversizedBodies(t *testing.T) {
	called := false
	api := &API{
		config: config.Config{TTS: config.TTSConfig{BaseURL: "https://tts.example"}},
		ttsHTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			called = true
			return nil, nil
		})},
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/tts", bytes.NewReader(make([]byte, ttsMaxRequestBytes+1)))

	api.proxyTTS(recorder, req)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if called {
		t.Fatal("oversized request reached upstream")
	}
}

func TestProxyTTSPostRejectsOversizedResponse(t *testing.T) {
	api := &API{
		config: config.Config{TTS: config.TTSConfig{BaseURL: "https://tts.example", MaxAudioBytes: 4}},
		ttsHTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString("audio"))}, nil
		})},
	}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/tts", bytes.NewBufferString("hello"))

	api.proxyTTS(recorder, req)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
}
