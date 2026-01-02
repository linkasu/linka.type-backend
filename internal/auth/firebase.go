package auth

import (
	"context"

	fbauth "firebase.google.com/go/v4/auth"
)

// FirebaseVerifier verifies Firebase ID tokens.
type FirebaseVerifier struct {
	client *fbauth.Client
}

// NewFirebaseVerifier creates a new verifier backed by Firebase Admin SDK.
func NewFirebaseVerifier(client *fbauth.Client) *FirebaseVerifier {
	return &FirebaseVerifier{client: client}
}

// Verify validates the Firebase ID token and returns the user.
func (v *FirebaseVerifier) Verify(ctx context.Context, token string) (User, error) {
	decoded, err := v.client.VerifyIDToken(ctx, token)
	if err != nil {
		return User{}, ErrUnauthorized
	}

	user := User{UID: decoded.UID}
	if email, ok := decoded.Claims["email"].(string); ok {
		user.Email = email
	}

	return user, nil
}
