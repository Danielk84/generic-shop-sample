package main

import (
	"context"
	"generic-shop-sample/app"
	"generic-shop-sample/db"
	"generic-shop-sample/internal"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := internal.NewConfig()

	engine, err := db.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Panicln(err)
	}
	engine.Close()

	app := app.NewApp(ctx, config)

	app.Run()
}
