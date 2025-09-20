package testutils

import (
	"context"
	"generic-shop-sample/app"
	"generic-shop-sample/db"
	"log"

	"github.com/gin-gonic/gin"
)

func RouterSetup(ctx context.Context) *gin.Engine {
	config := app.NewAppConfig()
	app := app.NewApp(ctx, config)

	return app.Router
}

func DBEngineSetup(ctx context.Context) db.IDBEngine {
	config := app.NewAppConfig()
	dbEngine, err := db.SetupDBEngine(ctx, config.DatabaseURL)
	if err != nil {
		log.Panicln("error setup db: ", err)
	}
	return dbEngine
}
