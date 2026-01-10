package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/dialoghelper"
	"github.com/linkasu/linka.type-backend/internal/dialogworker"
	"github.com/linkasu/linka.type-backend/internal/logging"
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

	logger := logging.New("dialog-worker", cfg.Env)

	ydbClient, err := ydb.New(ctx, cfg.YDB)
	if err != nil {
		logger.Error("failed to init ydb", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = ydbClient.Close(ctx)
	}()

	dialogClient := dialoghelper.New(cfg.Dialog.BaseURL, cfg.Dialog.APIKey, cfg.Dialog.Timeout)
	worker := dialogworker.New(ydbstore.New(ydbClient), dialogClient, logger)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("dialog-worker started", "interval", cfg.DialogWorker.PollInterval.String())
		if err := worker.Run(ctx, cfg.DialogWorker.PollInterval); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-shutdownCh:
		logger.Info("shutdown requested")
	case err := <-errCh:
		logger.Error("dialog-worker failed", "error", err)
	}
}
