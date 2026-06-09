package queries

import (
	"context"
	"generic-shop-sample/db/database"

	"github.com/jackc/pgx/v5"
)

type CategoryTag struct {
	Tag string `json:"tag" binding:"required"`
}

type Category struct {
	ID int32 `json:"id"`
	CategoryTag
}

type CategoryRepository struct {
	session database.Session
}

type CategoryStore interface {
	Create(ctx context.Context, tag string) error
	List(ctx context.Context) ([]Category, error)
	Delete(ctx context.Context, product_id int32) error
}

func NewCategoryStore(session database.Session) CategoryStore {
	return &CategoryRepository{session}
}

func (cr *CategoryRepository) Create(ctx context.Context, tag string) error {
	const q = `INSERT INTO categories(tag) VALUES ($1)`
	return execOne(ctx, cr.session, q, tag)
}

func (cr *CategoryRepository) List(ctx context.Context) ([]Category, error) {
	const q = `SELECT id, tag FROM categories`
	return list[Category](ctx, cr.session, q)
}

func (cr *CategoryRepository) Delete(ctx context.Context, id int32) error {
	const q = `DELETE FROM categories WHERE id = $1`
	return execOne(ctx, cr.session, q, id)
}

type PCRepository struct {
	session database.Session
}

type PCStore interface {
	SetTags(ctx context.Context, product_id string, tags []string) error
	List(ctx context.Context, product_id string) ([]string, error)
}

func NewPCStore(session database.Session) PCStore {
	return &PCRepository{session}
}

func (pcr *PCRepository) SetTags(ctx context.Context, product_id string, tags []string) error {
	const q = `DELETE FROM products_categories WHERE product_id = $1::UUID`

	tagsLen := len(tags)
	if tagsLen == 0 {
		return nil
	}
	return pgx.BeginFunc(ctx, pcr.session, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, q, product_id); err != nil {
			return err
		}
		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{"products_categories"},
			[]string{"product_id", "tag"},
			pgx.CopyFromSlice(tagsLen, func(i int) ([]any, error) {
				return []any{product_id, tags[i]}, nil
			}),
		)
		return err
	})
}

func (pcr *PCRepository) List(ctx context.Context, id string) ([]string, error) {
	const q = `SELECT tag FROM products_categories WHERE product_id = $1::UUID`
	return list[string](ctx, pcr.session, q, id)
}
