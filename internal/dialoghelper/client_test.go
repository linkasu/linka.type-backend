package dialoghelper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInferDoesNotIncludeExternalBodyInError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"detail":"private dialog text"}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-key", time.Second)
	_, err := client.Infer(context.Background(), InferPayload{
		Messages: []DialogMessage{{Role: "speaker", Content: "sensitive phrase"}},
	}, nil)
	if err == nil {
		t.Fatal("Infer() error = nil, want upstream status error")
	}
	if strings.Contains(err.Error(), "private dialog text") || strings.Contains(err.Error(), "sensitive phrase") {
		t.Fatalf("Infer() exposes content in error: %q", err)
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("Infer() error = %q, want bounded status", err)
	}
}
