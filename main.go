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
	defer engine.Close()

	cacheDBs := []int{db.PublicCache, db.UsersCache, db.ProductsCache, db.PaymentCache}
	cache, err := db.NewCacheManager(ctx, config.CacheURL, cacheDBs)
	if err != nil {
		log.Panicln(err)
	}
	defer cache.Close()

	app := app.NewApp(ctx, config)
	app.Run()
}
