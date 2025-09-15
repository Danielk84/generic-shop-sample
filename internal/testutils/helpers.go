package testutils

import (
	a "generic-shop-sample/app"

	"github.com/gin-gonic/gin"
)

func RouterSetup() *gin.Engine {
	config := a.NewAppConfig()
	app := a.NewApp(config)

	return app.Router
}
