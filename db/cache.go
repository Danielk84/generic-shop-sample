package db

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
)

type CacheClient = *redis.Client

type CacheManager interface {
	Close()
}

type CacheEngine struct {
	clients map[int]CacheClient
	mu      sync.RWMutex
}

var (
	DefaultCacheEngine *CacheEngine
	onceCacheClient    sync.Once
)

func NewCacheEngine(ctx context.Context, addr string, dbs []int) (*CacheEngine, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, db := range dbs {
		if db < 0 || db > 15 {
			return nil, fmt.Errorf(`invalid db number "%d", db must be 0 <= db <= 15`, db)
		}
	}

	option, err := redis.ParseURL(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL, %s\n", err)
	}

	clients := make(map[int]CacheClient)
	for _, db := range dbs {
		opt := *option
		opt.DB = db
		rdb := redis.NewClient(&opt)

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			return nil, fmt.Errorf(`failed to ping redis db="%d", %s\n`, db, err)
		}
		clients[db] = rdb
	}
	return &CacheEngine{clients: clients}, nil
}

func (ce *CacheEngine) Close() {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	for _, client := range ce.clients {
		_ = client.Close()
	}
}

func NewCacheManager(ctx context.Context, addr string, dbs []int) (CacheManager, error) {
	var err error
	onceCacheClient.Do(func() {
		DefaultCacheEngine, err = NewCacheEngine(ctx, addr, dbs)
	})
	return DefaultCacheEngine, err
}

func NewCache(db int) CacheClient {
	if DefaultCacheEngine == nil {
		panic("not initiated cache engine")
	}
	DefaultCacheEngine.mu.RLock()
	defer DefaultCacheEngine.mu.RUnlock()

	return DefaultCacheEngine.clients[db]
}
