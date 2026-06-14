package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"generic-shop-sample/app"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"

	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{
		Use:     "cmd",
		Short:   "app manager cli",
		Example: `cmd -c="path/to/file" [commands]`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			config := internal.NewConfig(configFile)
			db, err := database.New(ctx, config.Opt.DatabaseURL)
			if err != nil {
				return fmt.Errorf("invalid DATABASE_URL opt variable, %s", err)
			}
			defer db.Close()
			fmt.Println(configFile)
			return nil
		},
	}
	setSubCommands(rootCmd)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setSubCommands(rootCmd *cobra.Command) {
	newAdminCmd := &cobra.Command{
		Use:     "new-admin",
		Short:   "add new admin user",
		Example: `cmd -c="path/to/file" new-admin --username="username" --password="password"`,
		RunE:    newAdmin,
	}
	newAdminCmd.PersistentFlags().StringP("username", "u", "", "admin username")
	newAdminCmd.PersistentFlags().StringP("password", "p", "", "admin password")

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run server",
		RunE:  run,
	}

	rootCmd.AddCommand(newAdminCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "", "absolute path to yaml config file")
}

func newAdmin(cmd *cobra.Command, args []string) error {
	var err error
	username, err := cmd.PersistentFlags().GetString("username")
	if err != nil {
		return err
	}
	password, err := cmd.PersistentFlags().GetString("password")
	if err != nil {
		return err
	}
	us := queries.NewUserStore(database.GetSession(), logger.GetLogger())
	user := queries.CreateUserRequest{
		LoginRequest: queries.LoginRequest{
			Username: username,
			Password: password,
		},
		UserPermissionRequest: queries.UserPermissionRequest{
			PermissionType: queries.Admin,
			IsActive:       true,
		},
	}
	validate := internal.GetValidator()
	if err = validate.Struct(user); err != nil {
		if username == "" || password == "" {
			return fmt.Errorf("username and password required")
		}
		return fmt.Errorf("failed to validate input, %s", err)
	}
	if user.Password, err = auth.PasswordHash(password); err != nil {
		return fmt.Errorf("failed to hash password, %s", err)
	}
	if err = us.Create(cmd.Context(), &user); err != nil {
		return fmt.Errorf("failed to create admin, %s", err)

	}
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	config := internal.GetConfig()
	ctx := cmd.Context()

	cacheDBs := []int{cache.PublicCache, cache.UsersCache, cache.ProductsCache, cache.PaymentCache}
	cache, err := cache.New(ctx, config.Opt.CacheURL, cacheDBs)
	if err != nil {
		return err
	}
	defer cache.Close()

	server := app.NewApp(ctx, config)
	server.Run()
	return nil
}
