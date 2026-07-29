package database_test

import (
	"context"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/tests/internal/testutils"
	"testing"
	"time"
)

func TestDBEngine(t *testing.T) {
	config := testutils.ConfigTestSetup()
	db, err := database.New(t.Context(), config.DatabaseURL)
	if err != nil {
		t.Errorf("incorrect database connection: %s", err)
	}
	defer db.Close()

	session := db.GetSession()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	var n int32
	if err := session.QueryRow(ctx, "select $1::int", 1).Scan(&n); err != nil {
		t.Errorf("error on running QueryRow in session: %s", err)
	}
	if n != 1 {
		t.Error("not expexcted result from session.QueryRow")
	}
}
