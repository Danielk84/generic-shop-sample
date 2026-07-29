package queries_test

import (
	"context"
	tu "generic-shop-sample/tests/internal/testutils"
	"os"
	"testing"
	"time"
)

var setupFuncs = []tu.TestFixtureFn{
	tu.CreateTemporaryUsers,
}

var teardownFuncs = []tu.TestFixtureFn{
	tu.TruncateTables,
	tu.FlushCache,
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := tu.ConfigTestSetup()
	deps := tu.ServiceDepsTestSetup(ctx, config)
	_ = tu.LogTestSetup(config)

	errs := []error{}
	for _, fn := range teardownFuncs {
		if err := fn(deps); err != nil {
			errs = append(errs, err)
		}
	}

	for _, fn := range setupFuncs {
		if err := fn(deps); err != nil {
			errs = append(errs, err)
		}
	}

	var exitVal int
	if len(errs) == 0 {
		exitVal = m.Run()
	}

	for _, fn := range teardownFuncs {
		if err := fn(deps); err != nil {
			errs = append(errs, err)
		}
	}

	tu.CloseServieDepsTestSetup(deps)
	tu.CeckErrList("query", errs)
	os.Exit(exitVal)
}
