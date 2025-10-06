package testutils

import (
	"context"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/db"
	"generic-shop-sample/internal"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func RouterSetup(ctx context.Context) *gin.Engine {
	config := internal.NewConfig()
	app := app.NewApp(ctx, config)

	return app.Router
}

func DBManagerSetup(ctx context.Context) db.DBManager {
	config := internal.NewConfig()
	engine, err := db.New(ctx, config.DatabaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to setup db, %s", err))
	}
	return engine
}

func CacheSetup(ctx context.Context) db.CacheManager {
	config := internal.NewConfig()
	cache, err := db.NewCacheManager(ctx, config.CacheURL, []int{db.PublicCache, db.UsersCache, db.ProductsCache})
	if err != nil {
		panic(fmt.Errorf("failed to setup cache, %s", err))
	}
	return cache
}

func CeckErrList(errs []error) {
	if len(errs) > 0 {
		for _, err := range errs {
			slog.Error("", "error", err)
		}
		os.Exit(1)
	}
}
