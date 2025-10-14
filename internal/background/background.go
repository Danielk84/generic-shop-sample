package background

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
)

type BackgroundTask interface {
	start(ctx context.Context)
}

func StartTasks(ctx context.Context, session db.Session) {
	tasks := []BackgroundTask{
		&expiredOrdersCleaner{queries.NewOrderStore(session)},
	}
	for _, t := range tasks {
		t.start(ctx)
	}
}
