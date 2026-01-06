package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/linkasu/linka.type-backend/internal/auth"
	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/coreapi"
	"github.com/linkasu/linka.type-backend/internal/firebase"
	"github.com/linkasu/linka.type-backend/internal/jwt"
	"github.com/linkasu/linka.type-backend/internal/logging"
	"github.com/linkasu/linka.type-backend/internal/service"
	"github.com/linkasu/linka.type-backend/internal/store/legacy"
	"github.com/linkasu/linka.type-backend/internal/store/ydbstore"
	"github.com/linkasu/linka.type-backend/internal/ydb"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	logger := logging.New("core-api", cfg.Env)

	// Validate Firebase credentials are present
	if cfg.Firebase.CredentialsJSON == "" && cfg.Firebase.CredentialsFile == "" {
		logger.Error("firebase credentials are required: set FIREBASE_CREDENTIALS_JSON, FIREBASE_CREDENTIALS_B64, or FIREBASE_CREDENTIALS_FILE")
		os.Exit(1)
	}
	if cfg.Firebase.ProjectID == "" {
		logger.Error("FIREBASE_PROJECT_ID is required")
		os.Exit(1)
	}

	fbClients, err := firebase.NewClients(ctx, cfg.Firebase)
	if err != nil {
		logger.Error("failed to init firebase", "error", err)
		os.Exit(1)
	}
	fbVerifier := auth.NewFirebaseVerifier(fbClients.Auth)

	var jwtManager *jwt.Manager
	var verifier auth.Verifier = fbVerifier

	if cfg.JWT.Secret != "" {
		jwtManager = jwt.NewManager(jwt.Config{
			Secret:               cfg.JWT.Secret,
			AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
			RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
		})
		jwtVerifier := auth.NewJWTVerifier(jwtManager)
		verifier = auth.NewCompositeVerifier(jwtVerifier, fbVerifier)
		logger.Info("jwt auth enabled", "access_duration", cfg.JWT.AccessTokenDuration, "refresh_duration", cfg.JWT.RefreshTokenDuration)
	} else {
		logger.Warn("jwt auth disabled, using firebase only (set JWT_SECRET to enable)")
	}

	ydbClient, err := ydb.New(ctx, cfg.YDB)
	if err != nil {
		logger.Error("failed to init ydb", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = ydbClient.Close(ctx)
	}()

	var legacyWriter *legacy.Writer
	var legacyReader *legacy.Reader
	if fbClients.DB != nil {
		legacyWriter, err = legacy.New(fbClients.DB)
		if err != nil {
			logger.Error("failed to init legacy writer", "error", err)
			os.Exit(1)
		}
		legacyReader, err = legacy.NewReader(fbClients.DB)
		if err != nil {
			logger.Error("failed to init legacy reader", "error", err)
			os.Exit(1)
		}
	}

	svc := &service.Service{
		Store:        ydbstore.New(ydbClient),
		LegacyWriter: legacyWriter,
		LegacyReader: legacyReader,
		Feature:      cfg.Feature,
	}

	handler := coreapi.New(svc, verifier, fbClients.Auth, jwtManager, cfg)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("core-api listening", "addr", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCh
	logger.Info("shutdown requested")

	ctxTimeout, cancel := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		logger.Error("shutdown failed", "error", err)
	}
	logger.Info("shutdown complete")
}
