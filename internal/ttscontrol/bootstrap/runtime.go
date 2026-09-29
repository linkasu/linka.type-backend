// Package bootstrap wires optional TTS control-plane infrastructure.
package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol/migrate"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol/pgstore"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol/rediscache"
	"github.com/redis/go-redis/v9"
)

// Runtime owns dependencies that exist only while the control-plane flag is enabled.
type Runtime struct {
	Service *ttscontrol.Service
	pool    *pgxpool.Pool
	redis   *redis.Client
}

func NewRuntime(ctx context.Context, cfg config.TTSControlConfig) (*Runtime, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("init TTS PostgreSQL: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping TTS PostgreSQL: %w", err)
	}
	if err := migrate.Check(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Username: cfg.RedisUsername, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		pool.Close()
		return nil, fmt.Errorf("ping TTS Redis: %w", err)
	}
	service, err := ttscontrol.NewService(pgstore.New(pool), rediscache.New(client), ttscontrol.ServiceConfig{
		SigningKey: cfg.TokenSigningKey, PreviousKey: cfg.PreviousTokenSigningKey, IPHashKey: cfg.IPHashKey,
		AnonymousDaily: cfg.AnonymousDailyChunks, AuthenticatedDaily: cfg.AuthenticatedDailyChunks,
		AnonymousMax: cfg.AnonymousMaxChunks, AuthenticatedMax: cfg.AuthenticatedMaxChunks,
	})
	if err != nil {
		_ = client.Close()
		pool.Close()
		return nil, err
	}
	return &Runtime{Service: service, pool: pool, redis: client}, nil
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	if r.redis != nil {
		_ = r.redis.Close()
	}
	if r.pool != nil {
		r.pool.Close()
	}
}
