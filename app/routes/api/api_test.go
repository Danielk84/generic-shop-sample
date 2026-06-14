package api_test

import (
	"context"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	tu "generic-shop-sample/internal/testutils"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := tu.DBManagerSetup(ctx)
	session := database.GetSession()
	cache := tu.CacheSetup(ctx)

	password := "securePassword"
	var users = []queries.CreateUserRequest{
		{
			LoginRequest:          queries.LoginRequest{Username: "admin_user", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Admin, IsActive: true},
		},
		{
			LoginRequest:          queries.LoginRequest{Username: "admin_user2", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Admin, IsActive: true},
		},
		{
			LoginRequest:          queries.LoginRequest{Username: "vendor_user", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Vendor, IsActive: true},
		},
		{
			LoginRequest:          queries.LoginRequest{Username: "customer_user", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Customer, IsActive: true},
		},
		{
			LoginRequest:          queries.LoginRequest{Username: "block_user", Password: password},
			UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.BlockUser, IsActive: false},
		},
	}

	log := logger.GetLogger()
	us := queries.NewUserStore(session, log)
	errs := []error{}
	var err error
	for _, user := range users {
		if user.Password, err = auth.PasswordHash(user.Password); err != nil {
			errs = append(errs, err)
		} else if err = us.Create(ctx, &user); err != nil {
			errs = append(errs, err)
		}
	}

	var exitVal int
	if len(errs) == 0 {
		exitVal = m.Run()
	}

	if _, err := session.Exec(ctx,
		"TRUNCATE users, products, categories RESTART IDENTITY CASCADE",
	); err != nil {
		errs = append(errs, err)
	}

	cache.Close()
	db.Close()

	tu.CeckErrList("api", errs)

	os.Exit(exitVal)
}
