package background

import (
	"context"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
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
	}
	for _, t := range tasks {
		go t.start(ctx)
	}
}
