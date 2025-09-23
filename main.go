package main

import (
	"context"
	"generic-shop-sample/app"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := app.NewConfig()
	app := app.NewApp(ctx, config)

	app.Run()
}
