package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Session = *pgxpool.Pool

type DBManager interface {
	Close()
}

type DB struct {
	session Session
}

var defaultDB *DB

func NewDB(ctx context.Context, addr string) (*DB, error) {
	if addr == "" {
		return nil, fmt.Errorf("empty database address")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	session, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}

	if err := session.Ping(ctx); err != nil {
		session.Close()
		return nil, err
	}
	return &DB{session}, nil
}

func (db *DB) Close() {
	if db.session != nil {
		db.session.Close()
	}
}

func New(ctx context.Context, addr string) (DBManager, error) {
	var err error
	if defaultDB == nil {
		defaultDB, err = NewDB(ctx, addr)
	}
	return defaultDB, err
}

func GetSession() Session {
	if defaultDB == nil {
		panic("database not initiated")
	}
	return defaultDB.session
}
