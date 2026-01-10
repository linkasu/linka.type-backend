package auth

import (
	"context"
)

// CompositeVerifier tries multiple verifiers in order.
type CompositeVerifier struct {
	verifiers []Verifier
}

// NewCompositeVerifier creates a verifier that tries multiple verifiers.
func NewCompositeVerifier(verifiers ...Verifier) *CompositeVerifier {
	return &CompositeVerifier{verifiers: verifiers}
}

// Verify tries each verifier in order until one succeeds.
func (v *CompositeVerifier) Verify(ctx context.Context, token string) (User, error) {
	for _, verifier := range v.verifiers {
		user, err := verifier.Verify(ctx, token)
		if err == nil {
			return user, nil
		}
	}
	return User{}, ErrUnauthorized
}


