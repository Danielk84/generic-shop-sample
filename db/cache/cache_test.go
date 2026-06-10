package cache_test

import (
	"generic-shop-sample/db/cache"
	"generic-shop-sample/internal"
	"testing"
)

func TestCacheManager(t *testing.T) {
	config := internal.GetConfig()
	dbs := []int{cache.PublicCache, cache.UsersCache}
	cm, err := cache.New(t.Context(), config.Opt.CacheURL, dbs)
	if err != nil {
		t.Errorf("failed to create new cache manager, %s", err)
	}
	defer cm.Close()

	for _, i := range dbs {
		cache := cache.GetCache(i)
		if _, err := cache.Ping(t.Context()).Result(); err != nil {
			t.Errorf(`failed to ping cache db "%d", %s`, i, err)
		}
	}
}
