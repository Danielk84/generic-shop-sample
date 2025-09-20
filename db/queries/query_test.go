package queries_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	tu "generic-shop-sample/internal/testutils"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemporaryDataManagerFn func(context.Context, *pgxpool.Pool) error

var setupFuncs = []TemporaryDataManagerFn{
	createTemporaryUsers,
}

var teardownFuncs = []TemporaryDataManagerFn{
	truncateTables,
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	dbEngine := tu.DBEngineSetup(ctx)
	session := db.Session()

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

	dbEngine.Close()
	cancel()

	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	os.Exit(exitVal)
}

const createTemporaryQuery = `INSERT INTO users (username, permission_type, is_active)
VALUES (@Username, @PermissionType, @IsActive)`

var testUsersData = []queries.User{
	{Username: "adminUser", PermissionType: queries.Admin, IsActive: true},
	{Username: "vendorUser", PermissionType: queries.Vendor, IsActive: true},
	{Username: "customerUser", PermissionType: queries.Customer, IsActive: true},
	{Username: "blockUser", PermissionType: queries.BlockUser, IsActive: false},
}

func createTemporaryUsers(ctx context.Context, session *pgxpool.Pool) error {
	batch := &pgx.Batch{}

	for _, user := range testUsersData {
		args := pgx.NamedArgs{
			"Username":       user.Username,
			"PermissionType": user.PermissionType,
			"IsActive":       user.IsActive,
		}
		batch.Queue(createTemporaryQuery, args)
	}

	sb := session.SendBatch(ctx, batch)
	return sb.Close()
}

const truncateUsersQuery = `TRUNCATE users RESTART IDENTITY CASCADE`

func truncateTables(ctx context.Context, session *pgxpool.Pool) error {
	_, err := session.Exec(ctx, truncateUsersQuery)
	return err
}
