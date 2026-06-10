package main

import (
	"context"
	"generic-shop-sample/db/database"
	"generic-shop-sample/internal"
	"log"
	"os/signal"
	"syscall"

	"generic-shop-sample/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	config := internal.GetConfig()
	db, err := database.New(ctx, config.Opt.DatabaseURL)
	if err != nil {
		log.Panicf("invalid DATABASE_URL env variable, %s\n", err)
		return
	}
	defer db.Close()

	cmd.Execute(ctx)
}
