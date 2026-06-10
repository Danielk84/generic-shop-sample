package background

import (
	"context"
	"generic-shop-sample/db/cache"
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
)

type BackgroundTask interface {
	start(ctx context.Context)
}

func StartTasks(ctx context.Context, session database.Session) {
	config := internal.NewConfig()
	tasks := []BackgroundTask{
		&expiredOrdersCleaner{queries.NewOrderStore(session)},
		&emailBroker{
			cache:    cache.GetCache(cache.UsersCache),
			from:     config.FromEmail,
			host:     config.SMTPHost,
			port:     config.SMTPPort,
			password: config.SMTPPassword,
		},
	}
	for _, t := range tasks {
		go t.start(ctx)
	}
}
