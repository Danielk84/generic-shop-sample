package middlewares

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRateLimiter(ctx context.Context, requestLimit int, timePeriod, cleanupPeriod time.Duration) *RateLimter {
	if timePeriod == 0 || cleanupPeriod == 0 {
		log.Panic("timePeriod or cleanupPeriod must be non-zero")
	}

	rl := &RateLimter{
		ctx:           ctx,
		ipRLs:         make(map[string]*ipRateLimiter),
		requestLimit:  requestLimit,
		timePeriod:    timePeriod,
		cleanupPeriod: cleanupPeriod,
	}

	go rl.removeExpired()

	return rl
}

type RateLimter struct {
	mu            sync.Mutex
	ctx           context.Context
	ipRLs         map[string]*ipRateLimiter
	requestLimit  int
	timePeriod    time.Duration
	cleanupPeriod time.Duration
}

func (rl *RateLimter) RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.ClientIP()
		ipRL := rl.createOrGetInfo(host)
		if ipRL.allow(rl.requestLimit, rl.timePeriod) {
			c.Next()
			return
		}

		retryAfter := strconv.Itoa(int(rl.timePeriod.Seconds()))
		c.Header("Retry-After", retryAfter)
		c.String(http.StatusTooManyRequests, "Too many requests, try later.")
	}
}

func (rl *RateLimter) createOrGetInfo(ip string) *ipRateLimiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if ipRL, ok := rl.ipRLs[ip]; ok {
		return ipRL
	}

	ipRL := &ipRateLimiter{startTime: time.Now(), maxRequest: rl.requestLimit}
	rl.ipRLs[ip] = ipRL

	return ipRL
}

func (rl *RateLimter) removeExpired() {
	ticker := time.NewTicker(rl.cleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()

			now := time.Now()
			for key, value := range rl.ipRLs {
				if now.Sub(value.startTime) >= rl.timePeriod {
					delete(rl.ipRLs, key)
				}
			}

			rl.mu.Unlock()
		}
	}
}

// GetLen return the number of stored IP entries, mainly for unit testing.
func (rl *RateLimter) GetLen() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.ipRLs)
}

type ipRateLimiter struct {
	mu         sync.Mutex
	startTime  time.Time
	maxRequest int
}

func (ipRL *ipRateLimiter) allow(rl int, tp time.Duration) bool {
	ipRL.mu.Lock()
	defer ipRL.mu.Unlock()

	if time.Since(ipRL.startTime) > tp {
		ipRL.startTime = time.Now()
		ipRL.maxRequest = rl
	}

	if ipRL.maxRequest > 0 {
		ipRL.maxRequest--
		return true
	}
	return false
}
