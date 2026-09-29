package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol/migrate"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config load failed")
		os.Exit(1)
	}
	if cfg.TTSControl.PostgresDSN == "" {
		fmt.Fprintln(os.Stderr, "TTS_CONTROL_POSTGRES_DSN is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.TTSControl.PostgresDSN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "postgres init failed")
		os.Exit(1)
	}
	defer pool.Close()
	if err := migrate.Run(ctx, pool); err != nil {
		fmt.Fprintln(os.Stderr, "TTS migration failed")
		os.Exit(1)
	}
	fmt.Println("TTS schema applied")
}
