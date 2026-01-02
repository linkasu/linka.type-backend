package auth

import (
	"context"
	"errors"
)

// User represents the authenticated user.
type User struct {
	UID   string
	Email string
}

// Verifier validates a bearer token and returns a User.
type Verifier interface {
	Verify(ctx context.Context, token string) (User, error)
}

var ErrUnauthorized = errors.New("unauthorized")
