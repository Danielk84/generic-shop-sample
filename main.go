package main

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/internal"
	"log"
	"os/signal"
	"syscall"

	"generic-shop-sample/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := internal.NewConfig()
	engine, err := db.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Panicf("invalid DATABASE_URL env variable, %s\n", err)
		return
	}
	defer engine.Close()

	cmd.Execute(ctx)
}
