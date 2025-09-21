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

func DBManagerSetup(ctx context.Context) db.DBManager {
	config := app.NewAppConfig()
	engine, err := db.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Panicln("error setup db: ", err)
	}
	return engine
}
