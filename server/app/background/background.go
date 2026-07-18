package background

import (
	"context"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/cache_query"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
)

type BackgroundTask interface {
	start(ctx context.Context)
}

func StartTasks(ctx context.Context, session database.Session) {
	config := internal.GetConfig()
	log := logger.GetLogger()
	tasks := []BackgroundTask{
		&expiredOrdersCleaner{
			store: queries.NewOrderStore(session, log),
			log:   log,
		},
		&emailBroker{
			cache:    cache.GetCache(cache.UsersCache),
			from:     config.Opt.FromEmail,
			host:     config.Opt.SMTPHost,
			port:     config.Opt.SMTPPort,
			password: config.Opt.SMTPPassword,
		},
		&ordersProcess{
			client:         cache.GetCache(cache.OrdersCache),
			log:            log,
			orderStore:     queries.NewOrderStore(session, log),
			orderItemStore: queries.NewOrderItemsStore(session, log),
			productStore:   queries.NewProductStore(session, log),
			vendorStore:    queries.NewVendorOrderStore(session, log),
			vendorCacheStore: cache_query.NewVendorCacheStore(
				cache.GetCache(cache.OrdersCache),
				log,
				queries.NewProductStore(session, log)),
			pagination: config.Opt.Pagination,
		},
	}
	for _, t := range tasks {
		go t.start(ctx)
	}
}
