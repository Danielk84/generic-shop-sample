package db

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Session = *pgxpool.Pool

type DBManager interface {
	Close()
}

type DBEngine struct {
	session Session
}

var (
	DefaultDBEngine *DBEngine
	onceDBEngine    sync.Once
)

func NewDBEngine(ctx context.Context, addr string) (*DBEngine, error) {
	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	session, err := pgxpool.New(connCtx, addr)
	if err != nil {
		return nil, err
	}

	if err := session.Ping(connCtx); err != nil {
		session.Close()
		return nil, err
	}
	return &DBEngine{session}, nil
}

func (db *DBEngine) Close() {
	if db.session != nil {
		db.session.Close()
	}
}

func New(ctx context.Context, addr string) (DBManager, error) {
	var err error
	onceDBEngine.Do(func() {
		var engine *DBEngine
		engine, err = NewDBEngine(ctx, addr)
		if err == nil {
			DefaultDBEngine = engine
		}
	})
	return DefaultDBEngine, nil
}

func NewSession() Session {
	if DefaultDBEngine == nil {
		panic("not initiated db engine")
	}
	return DefaultDBEngine.session
}
