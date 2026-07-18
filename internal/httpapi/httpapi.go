package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/linkasu/linka.type-backend/internal/requestid"
)

const fallbackErrorCode = "internal_error"

// ErrorResponse is the only error payload exposed by HTTP and realtime APIs.
type ErrorResponse struct {
	ErrorCode string     `json:"error_code"`
	RequestID string     `json:"request_id"`
	Error     ErrorAlias `json:"error"`
}

// ErrorAlias preserves the existing error.code contract without exposing messages.
type ErrorAlias struct {
	Code string `json:"code"`
}

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

// NewError returns a sanitized error payload suitable for HTTP or WebSocket use.
func NewError(code, requestID string) ErrorResponse {
	code = sanitizeErrorCode(code)
	return ErrorResponse{
		ErrorCode: code,
		RequestID: requestid.Normalize(requestID),
		Error:     ErrorAlias{Code: code},
	}
}

// WriteError writes a sanitized JSON error. Legacy messages are intentionally ignored.
func WriteError(w http.ResponseWriter, status int, code string, _ ...string) {
	rid := requestid.Normalize(w.Header().Get(requestid.Header))
	w.Header().Set(requestid.Header, rid)
	WriteJSON(w, status, NewError(code, rid))
}

func sanitizeErrorCode(code string) string {
	if code == "" || len(code) > 64 {
		return fallbackErrorCode
	}
	for _, ch := range code {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}
		return fallbackErrorCode
	}
	return code
}
