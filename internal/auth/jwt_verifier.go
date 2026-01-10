package auth

import (
	"context"

	"github.com/linkasu/linka.type-backend/internal/jwt"
)

// JWTVerifier verifies custom JWT tokens.
type JWTVerifier struct {
	manager *jwt.Manager
}

// NewJWTVerifier creates a new verifier backed by JWT manager.
func NewJWTVerifier(manager *jwt.Manager) *JWTVerifier {
	return &JWTVerifier{manager: manager}
}

// Verify validates the JWT access token and returns the user.
func (v *JWTVerifier) Verify(ctx context.Context, token string) (User, error) {
	claims, err := v.manager.ValidateAccessToken(token)
	if err != nil {
		return User{}, ErrUnauthorized
	}

	return User{
		UID:   claims.UID,
		Email: claims.Email,
	}, nil
}


