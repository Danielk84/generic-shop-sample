package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"generic-shop-sample/app"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/file_storage"
	"generic-shop-sample/storage/queries"

	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	m := manager{}
	rootCmd := &cobra.Command{
		Use:                "cmd",
		Short:              "app manager cli",
		Example:            `cmd -c="path/to/file" [commands]`,
		PersistentPreRunE:  m.persistentPreRunE,
		PersistentPostRunE: m.persistentPostRunE,
	}

	newAdminCmd := &cobra.Command{
		Use:     "new-admin",
		Short:   "add new admin user",
		Example: `cmd -c="path/to/file" new-admin --email="person@example.com"`,
		RunE:    m.newAdmin,
	}
	newAdminCmd.PersistentFlags().StringP("email", "e", "", "admin email")

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "run server",
		RunE:  m.run,
	}

	rootCmd.AddCommand(newAdminCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "", "absolute path to yaml config file")

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type manager struct {
	config config.Config
	db     database.DBManager
	cache  cache.CacheManager
}

func (m *manager) persistentPreRunE(cmd *cobra.Command, args []string) error {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	m.config = config.NewConfig(configFile)

	internal.SetCustomValidators()

	ctx := cmd.Context()
	m.db, err = database.New(ctx, m.config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL opt variable, %s", err)
	}
	return nil
}

func (m *manager) persistentPostRunE(cmd *cobra.Command, args []string) error {
	if m.db != nil {
		m.db.Close()
	}
	if m.cache != nil {
		m.cache.Close()
	}
	return nil
}

func (m *manager) newAdmin(cmd *cobra.Command, args []string) error {
	var err error
	email, err := cmd.PersistentFlags().GetString("email")
	if err != nil {
		return err
	}
	store := queries.NewUserStore(m.db.GetSession(), logger.GetLogger())
	user := queries.CreateUserRequest{
		EmailAddrRequest: queries.EmailAddrRequest{
			Email: email,
		},
		UserPermissionRequest: queries.UserPermissionRequest{
			PermissionRequest: queries.PermissionRequest{
				PermissionType: queries.Admin,
			},
			IsActive: true,
		},
	}
	validate := internal.GetValidator()
	if err = validate.Struct(user); err != nil {
		if email == "" {
			return fmt.Errorf("email required")
		}
		return fmt.Errorf("failed to validate input, %s", err)
	}
	if err = store.Create(cmd.Context(), user); err != nil {
		return fmt.Errorf("failed to create admin, %s", err)

	}
	return nil
}

func (m *manager) run(cmd *cobra.Command, args []string) (err error) {
	ctx := cmd.Context()

	cacheDBs := []int{
		cache.PublicCache,
		cache.UsersCache,
		cache.ProductsCache,
		cache.OrdersCache,
		cache.PaymentCache}
	m.cache, err = cache.New(ctx, m.config.CacheURL, cacheDBs)
	if err != nil {
		return
	}

	fileStore, err := file_storage.NewFileStoreClient(ctx, m.config.FileStore.AwsS3)
	if err != nil {
		return err
	}
	deps := &app.ServiceDeps{
		Ctx:       ctx,
		DB:        m.db,
		Cache:     m.cache,
		Config:    m.config,
		FileStore: fileStore,
	}
	sv := newServer(deps)
	sv.run()
	return
}
