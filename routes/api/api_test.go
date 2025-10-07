package api_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal/auth"
	tu "generic-shop-sample/internal/testutils"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	engine := tu.DBManagerSetup(ctx)
	session := db.NewSession()
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

	us := queries.NewUserStore(session)
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
	engine.Close()

	tu.CeckErrList(errs)

	os.Exit(exitVal)
}
