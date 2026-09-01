package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"

	"github.com/jackc/pgx/v5"
)

type CategoryTag struct {
	Tag string `json:"tag" binding:"required,min=1"`
}

type Category struct {
	ID int32 `json:"id"`
	CategoryTag
}

type categoryRepository struct {
	session database.Session
	log     logger.Logger
}

type CategoryStore interface {
	Create(ctx context.Context, tag string) error
	List(ctx context.Context) ([]Category, error)
	Delete(ctx context.Context, id int32) error
}

func NewCategoryStore(session database.Session, log logger.Logger) CategoryStore {
	return &categoryRepository{session, log}
}

func (c *categoryRepository) Create(ctx context.Context, tag string) (err error) {
	const q = `INSERT INTO product_s.categories(tag) VALUES ($1)`
	if err = execOne(ctx, c.session, q, tag); err != nil {
		c.log.Debug("CategoryRepository.Create", "error", err)
	}
	return
}

func (c *categoryRepository) List(ctx context.Context) (items []Category, err error) {
	const q = `SELECT id, tag FROM product_s.categories`
	items, err = list[Category](ctx, c.session, q)
	if err != nil {
		c.log.Debug("CategoryRepository.List", "error", err)
	}
	return
}

func (c *categoryRepository) Delete(ctx context.Context, id int32) (err error) {
	const q = `DELETE FROM product_s.categories WHERE id = $1`
	if err = execOne(ctx, c.session, q, id); err != nil {
		c.log.Debug("CategoryRepository.Delete", "error", err)
	}
	return
}

type pcRepository struct {
	session database.Session
	log     logger.Logger
}

type PCStore interface {
	SetTags(ctx context.Context, id string, tags []string) error
	List(ctx context.Context, id string) ([]string, error)
}

func NewPCStore(session database.Session, log logger.Logger) PCStore {
	return &pcRepository{session, log}
}

func (p *pcRepository) SetTags(ctx context.Context, id string, tags []string) (err error) {
	const q = `DELETE FROM product_s.products_categories
		WHERE product_id = $1::UUID`

	tagsLen := len(tags)
	if tagsLen == 0 {
		p.log.Debug("pcRepository.SetTags",
			"error", "return nil: empty tag list")
		return
	}
	err = pgx.BeginFunc(ctx, p.session, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, q, id); err != nil {
			p.log.Debug("pcRepository.SetTags", "error", err)
			return err
		}
		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{"product_s", "products_categories"},
			[]string{"product_id", "tag"},
			pgx.CopyFromSlice(tagsLen, func(i int) ([]any, error) {
				return []any{id, tags[i]}, nil
			}),
		)
		if err != nil {
			p.log.Debug("pcRepository.SetTags", "error", err)
		}
		return err
	})
	if err != nil {
		p.log.Debug("PCRepository.SetTags", "error", err)
	}
	return
}

func (p *pcRepository) List(ctx context.Context, id string) (items []string, err error) {
	const q = `SELECT tag FROM product_s.products_categories WHERE product_id = $1::UUID`
	items, err = list[string](ctx, p.session, q, id)
	if err != nil {
		p.log.Debug("PCRepository.List", "error", err)
	}
	return
}
