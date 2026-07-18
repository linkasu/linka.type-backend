package httpmiddleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/requestid"
)

func TestRecoverySanitizesPanicAndCorrelatesLog(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := RequestID(Recovery(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private phrase from request")
	})))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(requestid.Header, "request-123")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	var payload httpapi.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ErrorCode != "internal_error" || payload.RequestID != "request-123" {
		t.Fatalf("unexpected response: %+v", payload)
	}
	if strings.Contains(logs.String(), "private phrase") {
		t.Fatalf("panic content leaked to logs: %s", logs.String())
	}
	if !strings.Contains(logs.String(), `"request_id":"request-123"`) ||
		!strings.Contains(logs.String(), `"error_code":"internal_error"`) {
		t.Fatalf("log is not correlated: %s", logs.String())
	}
}

func TestRecoveryDoesNotOverwriteCommittedResponse(t *testing.T) {
	handler := Recovery(slog.New(slog.NewTextHandler(io.Discard, nil)))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		panic("after commit")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
