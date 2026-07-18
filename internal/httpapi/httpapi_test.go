package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/linkasu/linka.type-backend/internal/requestid"
)

func TestWriteErrorSanitizesResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set(requestid.Header, "request-123")

	WriteError(recorder, http.StatusInternalServerError, "store_failed", "private phrase from storage")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "private phrase") || strings.Contains(recorder.Body.String(), "message") {
		t.Fatalf("response exposes error detail: %s", recorder.Body.String())
	}

	var payload ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ErrorCode != "store_failed" || payload.Error.Code != payload.ErrorCode {
		t.Fatalf("unexpected error code payload: %+v", payload)
	}
	if payload.RequestID != "request-123" {
		t.Fatalf("request_id = %q, want request-123", payload.RequestID)
	}
}

func TestNewErrorRejectsUnboundedCodeAndRequestID(t *testing.T) {
	payload := NewError("Unsafe Error", "private/request/id")
	if payload.ErrorCode != fallbackErrorCode {
		t.Fatalf("error_code = %q, want %q", payload.ErrorCode, fallbackErrorCode)
	}
	if payload.RequestID == "private/request/id" || payload.RequestID == "" {
		t.Fatalf("request_id was not sanitized: %q", payload.RequestID)
	}
}
