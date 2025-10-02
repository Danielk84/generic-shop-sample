package main

import (
	"context"
	"generic-shop-sample/cmd/cmd"
	"generic-shop-sample/db"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	engine, err := db.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Panicf("invalid DATABASE_URL env variable, %s\n", err)
		return
	}
	defer engine.Close()

	cmd.Execute(ctx)
}
