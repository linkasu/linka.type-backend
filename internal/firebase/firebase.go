package firebase

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/db"
	"github.com/linkasu/linka.type-backend/internal/config"
	"google.golang.org/api/option"
)

// Clients bundles Firebase Admin SDK clients.
type Clients struct {
	App  *firebase.App
	Auth *auth.Client
	DB   *db.Client
}

// NewClients initializes Firebase Admin clients.
func NewClients(ctx context.Context, cfg config.FirebaseConfig) (*Clients, error) {
	opts, err := credentialOptions(cfg)
	if err != nil {
		return nil, err
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID:   cfg.ProjectID,
		DatabaseURL: cfg.DatabaseURL,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase auth: %w", err)
	}

	var dbClient *db.Client
	if cfg.DatabaseURL != "" {
		dbClient, err = app.Database(ctx)
		if err != nil {
			return nil, fmt.Errorf("init firebase db: %w", err)
		}
	}

	return &Clients{
		App:  app,
		Auth: authClient,
		DB:   dbClient,
	}, nil
}

func credentialOptions(cfg config.FirebaseConfig) ([]option.ClientOption, error) {
	switch {
	case cfg.CredentialsJSON != "":
		return []option.ClientOption{option.WithCredentialsJSON([]byte(cfg.CredentialsJSON))}, nil
	case cfg.CredentialsFile != "":
		return []option.ClientOption{option.WithCredentialsFile(cfg.CredentialsFile)}, nil
	default:
		return nil, fmt.Errorf("firebase credentials are required: set FIREBASE_CREDENTIALS_JSON, FIREBASE_CREDENTIALS_B64, or FIREBASE_CREDENTIALS_FILE")
	}
}
