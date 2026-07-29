package app_test

import (
	"context"
	"generic-shop-sample/app"
	tu "generic-shop-sample/tests/internal/testutils"
	"testing"
	"time"
)

func TestApp(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	config := tu.ConfigTestSetup()

	a := app.NewApp(ctx, config.App)

	if a == nil {
		t.Fatal("app is nil")
	}

	done := make(chan struct{})
	go func() {
		a.Run()
		close(done)
	}()
	time.Sleep(200 * time.Millisecond)

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("server shutdown timeout")
	}
	a.Close()
}
