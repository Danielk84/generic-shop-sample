package background

import (
	"context"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
)

const (
	cacheCleanerChannel = "cache-cleaner-channel"
	cacheScanSize       = 1000
)

type CacheCleanerMessage struct {
	CacheDB int
	Keys    []string `json:"keys"`
}

type CacheCleaner struct {
	Cache  cache.CacheManager
	Client cache.CacheClient
	Log    logger.Logger
}

func (c *CacheCleaner) Start(ctx context.Context) {
	sub := c.Client.Subscribe(ctx, cacheCleanerChannel)
	defer func() { _ = sub.Close() }()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				c.Log.Warn("cache cleaner subscriber is closed")
				return
			}

			var input CacheCleanerMessage
			if err := json.Unmarshal([]byte(msg.Payload), &input); err != nil {
				c.Log.Error("failed decode EmailMessage", "error", err)
				continue
			}

			client := c.Cache.GetCache(input.CacheDB)
			pipe := client.Pipeline()
			for _, prefix := range input.Keys {
				var cursor uint64 = 0
			Loop:
				for {
					scan := client.Scan(
						ctx,
						cursor,
						fmt.Sprintf("%s*", prefix),
						cacheScanSize)
					keys, newCursor, err := scan.Result()
					if err != nil {
						c.Log.Warn("failed to scan keys on cache cleaner", "error", err)
						break Loop
					}
					if len(keys) > 0 {
						pipe.Del(ctx, keys...)
					}
					if newCursor == 0 {
						break Loop
					}
					cursor = newCursor
				}
			}
			if _, err := pipe.Exec(ctx); err != nil {
				c.Log.Error("failed to clean caches",
					"error", err)
			}
		}
	}
}

var _ = (*CacheCleaner)(nil)

func SendCacheCleanr(ctx context.Context, cache cache.CacheClient, cc CacheCleanerMessage) error {
	msg, err := json.Marshal(cc)
	if err != nil {
		return fmt.Errorf("failed to encode CacheCleanerMessage, %s", err)
	}
	if err := cache.Publish(ctx, cacheCleanerChannel, msg).Err(); err != nil {
		return err
	}
	return nil
}
