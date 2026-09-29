package rediscache

import (
	"context"
	"fmt"
	"time"

	"github.com/linkasu/linka.type-backend/internal/ttscontrol"
	"github.com/redis/go-redis/v9"
)

type Cache struct{ client redis.UniversalClient }

func New(client redis.UniversalClient) *Cache { return &Cache{client: client} }

func (c *Cache) SetDaily(ctx context.Context, quota ttscontrol.DailyQuota) error {
	key := fmt.Sprintf("tts:quota:daily:%s:%s", quota.Day.UTC().Format("2006-01-02"), quota.Key)
	expiresAt := quota.Day.UTC().Add(48 * time.Hour)
	return c.client.Set(ctx, key, quota.Used, time.Until(expiresAt)).Err()
}
