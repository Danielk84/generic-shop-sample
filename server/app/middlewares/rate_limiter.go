package middlewares

import (
	"context"
	"fmt"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func NewRateLimiter(ctx context.Context, rt RateLimiter) gin.HandlerFunc {
	ttl := time.Duration(rt.TTL) * time.Minute
	rt.retryAfter = strconv.Itoa(rt.TTL)

	return func(c *gin.Context) {
		cacheKey := rt.genKey(c.ClientIP())
		var output int
		err := rt.Cache.Get(ctx, cacheKey).Scan(&output)
		if err != nil && err != redis.Nil {
			rt.Log.Warn("NewRateLimiter:Get", "error", err)
			rt.limiterResponse(c)
			return

		}
		if output < rt.RequestLimit {
			args := redis.SetArgs{}
			// check if key already exists.
			if output == 0 {
				args.TTL = ttl
			} else {
				args.KeepTTL = true
			}
			err = rt.Cache.SetArgs(ctx, cacheKey, output+1, args).Err()
			if err != nil {
				rt.Log.Warn("NewRateLimiter:SetArgs", "error", err)
				rt.limiterResponse(c)
				return
			}
			c.Next()
		} else {
			rt.Log.Debug("NewRateLimiter",
				"cacheKey", cacheKey,
				"error", "limited ip")
			rt.limiterResponse(c)
		}
	}
}

type RateLimiter struct {
	Cache        cache.CacheClient
	Log          logger.Logger
	Scope        string
	RequestLimit int
	TTL          int
	retryAfter   string
}

func (r RateLimiter) genKey(ip string) string {
	return fmt.Sprintf("rate-limiter:%s:%s", r.Scope, ip)
}

func (r RateLimiter) limiterResponse(c *gin.Context) {
	c.Header("Retry-After", r.retryAfter)
	c.String(http.StatusTooManyRequests, "Too many requests, try later.")
}
