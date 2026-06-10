package cmd

import (
	"generic-shop-sample/app"
	"generic-shop-sample/db/cache"
	"generic-shop-sample/internal"
	"log"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "serve",
	Short: "run server",
	Run:   serve,
}

func serve(cmd *cobra.Command, args []string) {
	config := internal.NewConfig()
	ctx := cmd.Context()

	cacheDBs := []int{cache.PublicCache, cache.UsersCache, cache.ProductsCache, cache.PaymentCache}
	cache, err := cache.New(ctx, config.CacheURL, cacheDBs)
	if err != nil {
		log.Panicln(err)
	}
	defer cache.Close()

	app := app.NewApp(ctx, config)
	app.Run()
}

func init() {
	rootCmd.AddCommand(runCmd)
}
