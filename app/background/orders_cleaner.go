package background

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	"time"
)

type expiredOrdersCleaner struct {
	os queries.OrderStore
}

func (eoc *expiredOrdersCleaner) start(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := eoc.os.DeleteExpiredOrders(ctx); err != nil {
				logger.GetLogger().Error("failed to delete expired orders", "error", err)
			}
		}
	}
}
