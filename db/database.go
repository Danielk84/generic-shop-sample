package db

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IDBEngine interface {
	Close()
}

type DBEngine struct {
	dbPool *pgxpool.Pool
}

var (
	DefaultDBEngine *DBEngine
	once            sync.Once
)

func newDBEngine(ctx context.Context, addr string) (*DBEngine, error) {
	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(connCtx, addr)
	if err != nil {
		return nil, err
	}

	if err := dbPool.Ping(connCtx); err != nil {
		dbPool.Close()
		return nil, err
	}
	return &DBEngine{dbPool: dbPool}, nil
}

func (db *DBEngine) Close() {
	if db.dbPool != nil {
		db.dbPool.Close()
	}
}

func SetupDBEngine(ctx context.Context, addr string) (IDBEngine, error) {
	var err error
	once.Do(func() {
		var engine *DBEngine
		engine, err = newDBEngine(ctx, addr)
		if err == nil {
			DefaultDBEngine = engine
		}
	})
	return DefaultDBEngine, nil
}

func Session() *pgxpool.Pool {
	return DefaultDBEngine.dbPool
}
