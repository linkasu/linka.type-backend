package requestid

import (
	"context"
	"strings"

	"github.com/linkasu/linka.type-backend/internal/id"
)

// Header is the HTTP header used for request IDs.
const Header = "X-Request-Id"

const maxLength = 128

type ctxKey struct{}

// New returns a new request ID.
func New() string {
	return id.New()
}

// Normalize accepts a bounded caller-provided request ID or creates a new one.
func Normalize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxLength {
		return New()
	}
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' {
			continue
		}
		return New()
	}
	return value
}

// WithContext stores the request ID in context.
func WithContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, requestID)
}

// FromContext reads the request ID from context.
func FromContext(ctx context.Context) string {
	if val, ok := ctx.Value(ctxKey{}).(string); ok {
		return val
	}
	return ""
}
