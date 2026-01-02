package requestid

import (
	"context"

	"github.com/linkasu/linka.type-backend/internal/id"
)

// Header is the HTTP header used for request IDs.
const Header = "X-Request-Id"

type ctxKey struct{}

// New returns a new request ID.
func New() string {
	return id.New()
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
