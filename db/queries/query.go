package queries

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"generic-shop-sample/db"

	"github.com/jackc/pgx/v5"
)

func execOne(ctx context.Context, session db.Session, query string, args ...any) error {
	cTag, err := session.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

func list[T any](ctx context.Context, session db.Session, query string, args ...any) (*[]T, error) {
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

func get[T any](ctx context.Context, session db.Session, query string, args ...any) (*T, error) {
	items, err := list[T](ctx, session, query, args...)
	if err != nil {
		return nil, err
	}
	if len(*items) != 1 {
		return nil, pgx.ErrTooManyRows
	}
	return &(*items)[0], nil
}

type PropertyMap map[string]any

func (p *PropertyMap) Value() (driver.Value, error) {
	return json.Marshal(*p)
}

func (p *PropertyMap) Scan(src any) error {
	value, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("type assertion, .([]byte) failed")
	}

	var i any
	if err := json.Unmarshal(value, &i); err != nil {
		return err
	}

	*p, ok = i.(PropertyMap)
	if !ok {
		return fmt.Errorf("type assertion, .([]PropertyMap) failed")
	}
	return nil
}
