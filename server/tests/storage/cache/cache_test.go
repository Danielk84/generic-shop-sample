package cache_test

import (
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/tests/internal/testutils"
	"testing"
)

func TestCacheManager(t *testing.T) {
	config := testutils.ConfigTestSetup()
	dbs := []int{cache.PublicCache, cache.UsersCache}
	cm, err := cache.New(t.Context(), config.CacheURL, dbs)
	if err != nil {
		t.Errorf("failed to create new cache manager, %s", err)
	}
	defer cm.Close()

	for _, i := range dbs {
		c := cm.GetCache(i)
		if _, err := c.Ping(t.Context()).Result(); err != nil {
			t.Errorf(`failed to ping cache db "%d", %s`, i, err)
		}
	}
}
