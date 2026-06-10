package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PublicCache int = iota
	UsersCache
	ProductsCache
	PaymentCache
)

type CacheClient = *redis.Client

type CacheManager interface {
	Close()
}

type Cache struct {
	clients map[int]CacheClient
	mu      sync.RWMutex
}

var defaultCache *Cache

func NewCache(ctx context.Context, addr string, dbs []int) (*Cache, error) {
	if addr == "" {
		return nil, fmt.Errorf("empty cache storage address")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// base on redis min, max database number
	for _, db := range dbs {
		if db < 0 || db > 15 {
			return nil, fmt.Errorf(`invalid database number "%d", must be 0 <= db <= 15`, db)
		}
	}

	option, err := redis.ParseURL(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL, %s", err)
	}

	clients := make(map[int]CacheClient)
	for _, db := range dbs {
		opt := *option
		opt.DB = db
		rdb := redis.NewClient(&opt)

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			return nil, fmt.Errorf(`failed to ping redis db="%d", %s`, db, err)
		}
		clients[db] = rdb
	}
	return &Cache{clients: clients}, nil
}

func (ce *Cache) Close() {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	for _, client := range ce.clients {
		_ = client.Close()
	}
}

func New(ctx context.Context, addr string, dbs []int) (CacheManager, error) {
	var err error
	if defaultCache == nil {
		defaultCache, err = NewCache(ctx, addr, dbs)
	}
	return defaultCache, err
}

func GetCache(db int) CacheClient {
	if defaultCache == nil {
		panic("not initiated cache engine")
	}
	defaultCache.mu.RLock()
	defer defaultCache.mu.RUnlock()

	return defaultCache.clients[db]
}
