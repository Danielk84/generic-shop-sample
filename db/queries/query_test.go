package queries_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	tu "generic-shop-sample/internal/testutils"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type DatabaseFixtureFn func(context.Context, db.Session) error

var setupFuncs = []DatabaseFixtureFn{
	createTemporaryUsers,
}

var teardownFuncs = []DatabaseFixtureFn{
	truncateTables,
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	engine := tu.DBManagerSetup(ctx)
	session := db.NewSession()

	errs := []error{}
	for _, fn := range setupFuncs {
		if err := fn(ctx, session); err != nil {
			errs = append(errs, err)
		}
	}

	var exitVal int
	if len(errs) == 0 {
		exitVal = m.Run()
	}

	for _, fn := range teardownFuncs {
		if err := fn(ctx, session); err != nil {
			errs = append(errs, err)
		}
	}

	engine.Close()
	cancel()

	tu.CeckErrList(errs)

	os.Exit(exitVal)
}

func createTemporaryUsers(ctx context.Context, session db.Session) error {
	batch := &pgx.Batch{}
	const q = `INSERT INTO users (username, password, permission_type, is_active)
		VALUES (@Username, @Password, @PermissionType, @IsActive)`
	password := "secure password"
	var users = []queries.CreateUserRequest{
		{queries.LoginRequest{"adminUser", password}, queries.UserPermissionRequest{queries.Admin, true}},
		{queries.LoginRequest{"vendorUser", password}, queries.UserPermissionRequest{queries.Vendor, true}},
		{queries.LoginRequest{"customerUser", password}, queries.UserPermissionRequest{queries.Customer, true}},
		{queries.LoginRequest{"blockUser", password}, queries.UserPermissionRequest{queries.BlockUser, false}},
	}

	for _, user := range users {
		args := pgx.NamedArgs{
			"Username":       user.Username,
			"Password":       user.Password,
			"PermissionType": user.PermissionType,
			"IsActive":       user.IsActive,
		}
		batch.Queue(q, args)
	}

	sb := session.SendBatch(ctx, batch)
	return sb.Close()
}

func truncateTables(ctx context.Context, session db.Session) error {
	const q = `TRUNCATE users, products, categories RESTART IDENTITY CASCADE`
	_, err := session.Exec(ctx, q)
	return err
}
