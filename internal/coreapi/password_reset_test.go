package coreapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linkasu/linka.type-backend/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type resetResponseBody struct {
	*bytes.Reader
	closed bool
}

func (b *resetResponseBody) Close() error {
	b.closed = true
	return nil
}

func TestAuthResetPasswordDoesNotEnumerateAccounts(t *testing.T) {
	const responseDelay = 25 * time.Millisecond
	type result struct {
		status      int
		contentType string
		body        string
		duration    time.Duration
	}

	run := func(upstreamStatus int, upstreamBody string) result {
		t.Helper()
		body := &resetResponseBody{Reader: bytes.NewReader([]byte(upstreamBody))}
		api := &API{
			config: config.Config{Firebase: config.FirebaseConfig{APIKey: "test-key"}},
			httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: upstreamStatus,
					Body:       body,
					Header:     make(http.Header),
				}, nil
			})},
			passwordResetDelay: responseDelay,
		}
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/reset", bytes.NewBufferString(`{"email":"person@example.com"}`))

		started := time.Now()
		api.authResetPassword(recorder, req)
		duration := time.Since(started)

		if !body.closed {
			t.Fatal("upstream response body was not closed")
		}
		remaining, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("read upstream remainder: %v", err)
		}
		if len(remaining) != 0 {
			t.Fatalf("upstream response body was not drained: %q", remaining)
		}
		return result{
			status:      recorder.Code,
			contentType: recorder.Header().Get("Content-Type"),
			body:        recorder.Body.String(),
			duration:    duration,
		}
	}

	existing := run(http.StatusOK, `{"email":"person@example.com"}`)
	missing := run(http.StatusBadRequest, `{"error":{"message":"EMAIL_NOT_FOUND"}}`)

	if existing.status != http.StatusAccepted || missing.status != existing.status {
		t.Fatalf("statuses differ: existing=%d missing=%d", existing.status, missing.status)
	}
	if existing.contentType != missing.contentType || existing.body != missing.body {
		t.Fatalf("responses differ: existing=%q missing=%q", existing.body, missing.body)
	}
	if existing.body != "{\"status\":\"accepted\"}\n" {
		t.Fatalf("body = %q, want generic accepted response", existing.body)
	}
	for name, duration := range map[string]time.Duration{"existing": existing.duration, "missing": missing.duration} {
		if duration < responseDelay-2*time.Millisecond || duration > responseDelay+100*time.Millisecond {
			t.Fatalf("%s response duration %s is outside fixed timing class", name, duration)
		}
	}
}

func TestPasswordResetCORSAllowsPWAOrigin(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/reset", nil)
	req.Header.Set("Origin", "https://type.linka.su")

	corsMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("preflight request reached password reset handler")
	})).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "https://type.linka.su" {
		t.Fatalf("Access-Control-Allow-Origin = %q", origin)
	}
}
