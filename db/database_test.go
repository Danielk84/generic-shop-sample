package db_test

import (
	"context"
	"generic-shop-sample/app"
	"generic-shop-sample/db"
	"testing"
	"time"
)

func TestDBEngine(t *testing.T) {
	config := app.NewAppConfig()
	dbEngine, err := db.SetupDBEngine(t.Context(), config.DatabaseURL)
	if err != nil {
		t.Errorf("incorrect database connection: %s", err)
	}
	defer dbEngine.Close()

	session := db.Session()
	timeoutCtx, timeoutCancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer timeoutCancel()

	var n int32
	if err := session.QueryRow(timeoutCtx, "select $1::int", 1).Scan(&n); err != nil {
		t.Errorf("error on running QueryRow in session: %s", err)
	}
	if n != 1 {
		t.Error("not expexcted result from session.QueryRow")
	}
}
