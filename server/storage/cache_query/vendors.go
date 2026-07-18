package cache_query

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"slices"
	"time"
)

const expiration time.Duration = time.Hour * 24 * 30 * 4 // 4 Month

type vendorCacheRepository struct {
	client cache.CacheClient
	log    logger.Logger
	store  queries.ProductStore
}

type VendorCacheStore interface {
	Turn(ctx context.Context, productID string, property queries.ProductProperty) (string, error)
}

func NewVendorCacheStore(client cache.CacheClient, log logger.Logger, store queries.ProductStore) VendorCacheStore {
	return &vendorCacheRepository{client, log, store}
}

func (v *vendorCacheRepository) Turn(ctx context.Context, productID string, property queries.ProductProperty) (userID string, err error) {
	vendors, err := v.store.GetVendors(ctx, productID, property)
	if err != nil {
		v.log.Debug("vendorCacheRepository.Turn", "error", err)
		return
	}
	if len(vendors) < 1 {
		err = queries.ErrNotFound
		v.log.Debug("vendorCacheRepository.Turn", "error", err)
		return
	}

	var buf []byte
	if buf, err = json.Marshal(property); err != nil {
		v.log.Debug("vendorCacheRepository.Turn", "error", err)
		return
	}
	key := fmt.Sprintf("%s:%x", productID, sha256.Sum256(buf))

	var values []string
	if err = v.client.LRange(ctx, key, 0, -1).ScanSlice(&values); err != nil {
		v.log.Debug("vendorCacheRepository.Turn", "error", err)
		return
	}

	for _, ven := range vendors {
		if !slices.Contains(values, ven) {
			if err = v.client.LPush(ctx, key, ven).Err(); err != nil {
				v.log.Warn("vendorCacheRepository.Turn", "error", err)
				return
			}
			userID = ven
			goto addExpire
		}
	}

	if err = v.client.Del(ctx, key).Err(); err != nil {
		v.log.Error("vendorCacheRepository.Turn", "error", err)
		return
	}
	userID = vendors[0]
	if err = v.client.LPush(ctx, key, userID).Err(); err != nil {
		v.log.Warn("vendorCacheRepository.Turn", "error", err)
		return
	}

addExpire:
	if err = v.client.Expire(ctx, key, expiration).Err(); err != nil {
		v.log.Warn("vendorCacheRepository.Turn", "error", err)
	}
	return
}
