package queries

import (
	"context"
	"fmt"
	"generic-shop-sample/storage/database"
	"reflect"

	"github.com/jackc/pgx/v5"
)

var (
	ErrFullCapacity = fmt.Errorf("error full capacity")
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

	var items []T
	var t T
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Struct {
		items, err = pgx.CollectRows(rows, pgx.RowToStructByName[T])
	} else {
		items, err = pgx.CollectRows(rows, pgx.RowTo[T])
	}
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}

	return items, nil
}

func get[T any](ctx context.Context, session database.Session, query string, args ...any) (*T, error) {
	items, err := list[T](ctx, session, query, args...)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, pgx.ErrTooManyRows
	}
	return &items[0], nil
}
