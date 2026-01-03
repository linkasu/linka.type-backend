package ydb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	defaultMetadataURL   = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
	metadataHeader       = "Metadata-Flavor"
	metadataHeaderValue  = "Google"
	metadataTokenSkew    = 30 * time.Second
	metadataRequestLimit = 5 * time.Second
)

type metadataCredentials struct {
	url    string
	client *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

type metadataToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func newMetadataCredentials() *metadataCredentials {
	if os.Getenv("YDB_METADATA_DISABLED") != "" {
		return nil
	}
	url := os.Getenv("YDB_METADATA_URL")
	if url == "" {
		url = defaultMetadataURL
	}
	return &metadataCredentials{
		url:    url,
		client: &http.Client{Timeout: metadataRequestLimit},
	}
}

func (c *metadataCredentials) Token(ctx context.Context) (string, error) {
	if c == nil {
		return "", errors.New("metadata credentials disabled")
	}

	now := time.Now()
	c.mu.Lock()
	if c.token != "" && now.Add(metadataTokenSkew).Before(c.expiresAt) {
		token := c.token
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return "", fmt.Errorf("create metadata request: %w", err)
	}
	req.Header.Set(metadataHeader, metadataHeaderValue)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch metadata token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metadata token request failed: %s", resp.Status)
	}

	var payload metadataToken
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode metadata token: %w", err)
	}
	if payload.AccessToken == "" {
		return "", errors.New("metadata token response missing access_token")
	}

	expiresAt := now.Add(time.Duration(payload.ExpiresIn) * time.Second)
	c.mu.Lock()
	c.token = payload.AccessToken
	c.expiresAt = expiresAt
	c.mu.Unlock()

	return payload.AccessToken, nil
}
