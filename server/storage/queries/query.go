package queries

import (
	"context"
	"errors"
	"generic-shop-sample/storage/database"
	"reflect"

	"github.com/jackc/pgx/v5"
)

var (
	ErrFullCapacity = errors.New("error full capacity")
	ErrNotFound     = errors.New("error item not found")
)

func getOffsetFromPageNum(pagination, page int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pagination
}

func execOne(ctx context.Context, session database.Session, query string, args ...any) error {
	cTag, err := session.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

func list[T any](ctx context.Context, session database.Session, query string, args ...any) ([]T, error) {
	rows, err := session.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	items, err := pgx.CollectRows(rows, getRowToFunc[T]())
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}

	return items, nil
}

func get[T any](ctx context.Context, session database.Session, query string, args ...any) (item T, err error) {
	rows, err := session.Query(ctx, query, args...)
	if err != nil {
		return
	}
	item, err = pgx.CollectOneRow(rows, getRowToFunc[T]())
	return
}

func getRowToFunc[T any]() pgx.RowToFunc[T] {
	var t T
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Struct {
		return pgx.RowToStructByName[T]
	} else {
		return pgx.RowTo[T]
	}
}
