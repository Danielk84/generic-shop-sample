package testutils

import (
	"context"
	"generic-shop-sample/app"

	"github.com/gin-gonic/gin"
)

func RouterSetup(ctx context.Context) *gin.Engine {
	config := app.NewAppConfig()
	app := app.NewApp(ctx, config)

	return app.Router
}
