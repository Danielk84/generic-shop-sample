package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func execOne(ctx context.Context, session *pgxpool.Pool, query string, args ...any) error {
	cTag, err := session.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

func list[T any](ctx context.Context, session *pgxpool.Pool, query string, args ...any) (*[]T, error) {
	rows, err := session.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return &items, nil
}

func get[T any](ctx context.Context, session *pgxpool.Pool, query string, args ...any) (*T, error) {
	items, err := list[T](ctx, session, query, args...)
	if err != nil {
		return nil, err
	}
	if len(*items) != 1 {
		return nil, pgx.ErrTooManyRows
	}
	return &(*items)[0], nil
}
