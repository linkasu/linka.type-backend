package firebase

import (
	"context"
	"fmt"
	"os"

	"github.com/linkasu/linka.type-backend/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	scopeDatabase = "https://www.googleapis.com/auth/firebase.database"
	scopeEmail    = "https://www.googleapis.com/auth/userinfo.email"
)

// TokenSource returns an OAuth2 token source for Firebase RTDB REST streaming.
func TokenSource(ctx context.Context, cfg config.FirebaseConfig) (oauth2.TokenSource, error) {
	if token := os.Getenv("FIREBASE_ACCESS_TOKEN"); token != "" {
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}), nil
	}

	scopes := []string{scopeDatabase, scopeEmail}

	if cfg.CredentialsJSON != "" {
		creds, err := google.CredentialsFromJSON(ctx, []byte(cfg.CredentialsJSON), scopes...)
		if err != nil {
			return nil, fmt.Errorf("firebase credentials json: %w", err)
		}
		return creds.TokenSource, nil
	}

	if cfg.CredentialsFile != "" {
		data, err := os.ReadFile(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("firebase credentials file: %w", err)
		}
		creds, err := google.CredentialsFromJSON(ctx, data, scopes...)
		if err != nil {
			return nil, fmt.Errorf("firebase credentials file parse: %w", err)
		}
		return creds.TokenSource, nil
	}

	creds, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return nil, fmt.Errorf("firebase default credentials: %w", err)
	}
	return creds.TokenSource, nil
}
