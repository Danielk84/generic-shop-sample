package background

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	"time"
)

type ExpiredOrdersCleaner struct {
	Store queries.OrderStore
	Log   logger.Logger
}

func (b *ExpiredOrdersCleaner) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.Store.DeleteExpiredOrders(ctx); err != nil {
				logger.GetLogger().Error("failed to delete expired orders", "error", err)
			}
		}
	}
}

var _ BackgroundTask = (*ExpiredOrdersCleaner)(nil)
