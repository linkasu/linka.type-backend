package coreapi

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func TestTTSServiceJWTForEveryUpstreamRequest(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)
	var upstream []*http.Request
	api := &API{
		config: config.Config{TTS: config.TTSConfig{
			BaseURL:              "https://tts.example",
			ServiceToken:         "service-token",
			ServiceJWTPrivateKey: privateKey,
			ServiceJWTKeyID:      "key-1",
			ServiceJWTTTL:        4 * time.Minute,
			MaxAudioBytes:        1024,
		}},
		ttsNow: nowFunc(now),
		ttsHTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			upstream = append(upstream, req.Clone(req.Context()))
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString("audio"))}, nil
		})},
	}

	for _, test := range []struct {
		method string
		path   string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{http.MethodPost, "/v1/tts?voice=anna", api.proxyTTS},
		{http.MethodGet, "/v1/tts?voice=anna", api.proxyTTS},
		{http.MethodGet, "/v1/voices", api.proxyVoices},
	} {
		req := httptest.NewRequest(test.method, test.path, bytes.NewBufferString("hello"))
		req.Header.Set("Authorization", "Bearer user-token")
		req.Header.Set("Cookie", "session=secret")
		req.Header.Set("X-TTS-Service-JWT", "user-supplied")
		test.handle(httptest.NewRecorder(), req)
	}

	if len(upstream) != 3 {
		t.Fatalf("upstream requests = %d, want 3", len(upstream))
	}
	seen := make(map[string]bool)
	for _, req := range upstream {
		if got := req.Header.Get("X-TTS-Service-Token"); got != "service-token" {
			t.Fatalf("X-TTS-Service-Token = %q", got)
		}
		if req.Header.Get("Authorization") != "" || req.Header.Get("Cookie") != "" || req.Header.Get("X-TTS-Service-JWT") == "user-supplied" {
			t.Fatalf("user headers were forwarded: %#v", req.Header)
		}
		serviceJWT := req.Header.Get("X-TTS-Service-JWT")
		if serviceJWT == "" || seen[serviceJWT] {
			t.Fatal("each upstream request must have a distinct service JWT")
		}
		seen[serviceJWT] = true
		assertTTSServiceJWT(t, serviceJWT, publicKey, now, 4*time.Minute)
	}
}

func nowFunc(now time.Time) func() time.Time {
	return func() time.Time { return now }
}

func assertTTSServiceJWT(t *testing.T, raw string, publicKey ed25519.PublicKey, now time.Time, ttl time.Duration) {
	t.Helper()
	claims := &struct {
		Scope []string `json:"scope"`
		jwtgo.RegisteredClaims
	}{}
	token, err := jwtgo.NewParser(jwtgo.WithTimeFunc(nowFunc(now))).ParseWithClaims(raw, claims, func(token *jwtgo.Token) (any, error) {
		if token.Method.Alg() != jwtgo.SigningMethodEdDSA.Alg() {
			t.Fatalf("alg = %q", token.Method.Alg())
		}
		return publicKey, nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("service JWT invalid: %v", err)
	}
	if token.Header["kid"] != "key-1" || claims.Issuer != "linka-type-backend" || len(claims.Audience) != 1 || claims.Audience[0] != "tts-echo" || len(claims.Scope) != 1 || claims.Scope[0] != "synthesize" {
		t.Fatalf("service JWT header or claims = %#v %#v", token.Header, claims)
	}
	if !claims.IssuedAt.Time.Equal(now) || !claims.ExpiresAt.Time.Equal(now.Add(ttl)) {
		t.Fatalf("JWT lifetime = %v..%v", claims.IssuedAt, claims.ExpiresAt)
	}
	if _, err := uuid.Parse(claims.ID); err != nil {
		t.Fatalf("JWT jti = %q: %v", claims.ID, err)
	}
}
