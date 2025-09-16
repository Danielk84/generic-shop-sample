package testutils

import (
	"context"
	a "generic-shop-sample/app"

	"github.com/gin-gonic/gin"
)

func RouterSetup(ctx context.Context) *gin.Engine {
	config := a.NewAppConfig()
	app := a.NewApp(ctx, config)

	return app.Router
}
