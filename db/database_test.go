package db_test

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/internal"
	"testing"
	"time"
)

func TestDBEngine(t *testing.T) {
	config := internal.NewConfig()
	engine, err := db.New(t.Context(), config.DatabaseURL)
	if err != nil {
		t.Errorf("incorrect database connection: %s", err)
	}
	defer engine.Close()

	session := db.NewSession()
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
