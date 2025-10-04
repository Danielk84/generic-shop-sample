package db_test

import (
	"generic-shop-sample/db"
	"generic-shop-sample/internal"
	"testing"
)

func TestCacheManager(t *testing.T) {
	config := internal.NewConfig()
	dbs := []int{db.PublicCache, db.UsersCache}
	cm, err := db.NewCacheManager(t.Context(), config.CacheURL, dbs)
	if err != nil {
		t.Errorf("failed to create new cache manager, %s", err)
	}
	defer cm.Close()

	for _, i := range dbs {
		cache := db.NewCache(i)
		if _, err := cache.Ping(t.Context()).Result(); err != nil {
			t.Errorf(`failed to ping cache db "%d", %s`, i, err)
		}
	}
}
