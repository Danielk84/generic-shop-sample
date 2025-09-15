package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRateLimiter(requestLimit int, timePeriod int) *RateLimter {
	rl := &RateLimter{
		ipRLs:        make(map[string]*ipRateLimiter),
		requestLimit: requestLimit,
		timePeriod:   time.Duration(timePeriod) * time.Minute,
	}

	go rl.removeEndedTask()

	return rl
}

type RateLimter struct {
	mu           sync.Mutex
	ipRLs        map[string]*ipRateLimiter
	requestLimit int
	timePeriod   time.Duration
}

func (rl *RateLimter) RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.ClientIP()
		ipRL := rl.createOrGetInfo(host)
		if ipRL.allow(rl.requestLimit, rl.timePeriod) {
			c.Next()
			return
		}

		c.String(http.StatusTooManyRequests,
			fmt.Sprintf("Try %d minutes later", int64(rl.timePeriod/time.Minute)))
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

func (rl *RateLimter) removeEndedTask() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()

			newIpRLs := make(map[string]*ipRateLimiter)
			for key, value := range rl.ipRLs {
				if time.Since(value.startTime) > rl.timePeriod {
					continue
				}

				newIpRLs[key] = value
			}

			rl.ipRLs = newIpRLs

			rl.mu.Unlock()
		}
	}
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
