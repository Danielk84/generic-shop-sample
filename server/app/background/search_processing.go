package background

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	"time"
)

type SearchIndexProcess struct {
	Store queries.SearchStore
	Log   logger.Logger
}

func (s *SearchIndexProcess) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Store.DeleteAll(ctx); err != nil {
				s.Log.Error("failed to process search index", "error", err)
			}
		}
	}
}

var _ BackgroundTask = (*SearchIndexProcess)(nil)
