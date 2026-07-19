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
	GetSession() Session
}

type DB struct {
	session Session
}

func New(ctx context.Context, addr string) (DBManager, error) {
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

func (d *DB) Close() {
	if d.session != nil {
		d.session.Close()
	}
}

func (d *DB) GetSession() Session {
	if d.session != nil {
		return d.session
	}
	panic("undefined db session")
}
