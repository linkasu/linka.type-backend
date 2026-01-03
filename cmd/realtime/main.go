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
	"github.com/linkasu/linka.type-backend/internal/firebase"
	"github.com/linkasu/linka.type-backend/internal/logging"
	"github.com/linkasu/linka.type-backend/internal/realtime"
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
	logger := logging.New("realtime", cfg.Env)

	fbClients, err := firebase.NewClients(ctx, cfg.Firebase)
	if err != nil {
		logger.Error("failed to init firebase", "error", err)
		os.Exit(1)
	}
	verifier := auth.NewFirebaseVerifier(fbClients.Auth)

	ydbClient, err := ydb.New(ctx, cfg.YDB)
	if err != nil {
		logger.Error("failed to init ydb", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = ydbClient.Close(ctx)
	}()

	handler := realtime.New(ydbstore.New(ydbClient), verifier)

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
		logger.Info("realtime listening", "addr", cfg.HTTP.Addr)
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
