package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/firebase"
	"github.com/linkasu/linka.type-backend/internal/logging"
	"github.com/linkasu/linka.type-backend/internal/store/legacy"
	"github.com/linkasu/linka.type-backend/internal/store/ydbstore"
	"github.com/linkasu/linka.type-backend/internal/syncworker"
	"github.com/linkasu/linka.type-backend/internal/ydb"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	logger := logging.New("sync-worker", cfg.Env)

	fbClients, err := firebase.NewClients(ctx, cfg.Firebase)
	if err != nil {
		logger.Error("failed to init firebase", "error", err)
		os.Exit(1)
	}
	if fbClients.DB == nil {
		logger.Error("firebase database url is required")
		os.Exit(1)
	}

	ydbClient, err := ydb.New(ctx, cfg.YDB)
	if err != nil {
		logger.Error("failed to init ydb", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = ydbClient.Close(ctx)
	}()

	legacyReader, err := legacy.NewReader(fbClients.DB)
	if err != nil {
		logger.Error("failed to init legacy reader", "error", err)
		os.Exit(1)
	}

	worker := syncworker.New(ydbClient, ydbstore.New(ydbClient), fbClients.DB, legacyReader)
	if cfg.Sync.StreamEnabled {
		var tokenSource oauth2.TokenSource
		tokenSource, err = firebase.TokenSource(ctx, cfg.Firebase)
		if err != nil {
			logger.Error("failed to init firebase token source", "error", err)
			os.Exit(1)
		}
		worker.EnableStream(cfg.Firebase.DatabaseURL, tokenSource, cfg.Sync.StreamPath, cfg.Sync.StreamReconnect)
		logger.Info("rtdb streaming enabled", "path", cfg.Sync.StreamPath)
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("sync-worker started", "interval", cfg.Sync.PollInterval.String())
		if err := worker.Run(ctx, cfg.Sync.PollInterval); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-shutdownCh:
		logger.Info("shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("sync-worker failed", "error", err)
		}
	}
}
