package ydb

import (
	"context"
	"fmt"

	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

// Client wraps YDB driver and table client.
type Client struct {
	driver   *ydb.Driver
	database string
}

// New creates a YDB client.
func New(ctx context.Context, cfg config.YDBConfig) (*Client, error) {
	if cfg.Endpoint == "" || cfg.Database == "" {
		return nil, fmt.Errorf("YDB_ENDPOINT and YDB_DATABASE are required")
	}

	opts := []ydb.Option{
		ydb.WithDatabase(cfg.Database),
	}
	if cfg.Token != "" {
		opts = append(opts, ydb.WithAccessTokenCredentials(cfg.Token))
	}

	driver, err := ydb.Open(ctx, cfg.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("open ydb: %w", err)
	}

	return &Client{driver: driver, database: cfg.Database}, nil
}

// Table returns a table client.
func (c *Client) Table() table.Client {
	return c.driver.Table()
}

// Database returns the configured database path.
func (c *Client) Database() string {
	return c.database
}

// Close shuts down the driver.
func (c *Client) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}
