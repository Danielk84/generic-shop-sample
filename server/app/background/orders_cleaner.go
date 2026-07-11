package background

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	"time"
)

type expiredOrdersCleaner struct {
	store queries.OrderStore
	log   logger.Logger
}

func (b *expiredOrdersCleaner) start(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.store.DeleteExpiredOrders(ctx); err != nil {
				logger.GetLogger().Error("failed to delete expired orders", "error", err)
			}
		}
	}
}
