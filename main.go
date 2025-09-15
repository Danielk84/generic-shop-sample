package main

import "generic-shop-sample/app"

func main() {
	config := app.NewAppConfig()
	app := app.NewApp(config)

	app.Run()
}
