package queries

import (
	"context"
	"generic-shop-sample/db"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductSummary struct {
	ID      string
	Name    string
	Price   int64
	PubDate time.Time
}

type Product struct {
	IsAvailable bool
	IsActive    bool
	Description *string
	Details     PropertyMap
	ProductSummary
}

type OwnedProduct struct {
	UserID int32
	Product
}

type ProductRepository struct {
	session db.Session
}

type ProductStore interface {
	Create(context.Context, *OwnedProduct) error
	List(context.Context, int, int) (*[]ProductSummary, error)
	Get(context.Context, string) (*Product, error)
	Update(context.Context, *Product) error
	Delete(context.Context, string) error
	SetAvailable(context.Context, string, bool) error
	SetActive(context.Context, string, bool) error
}

func NewProductStore(session db.Session) ProductStore {
	return &ProductRepository{session: session}
}

func (pr *ProductRepository) Create(ctx context.Context, product *OwnedProduct) error {
	const q = `INSERT INTO products(user_id, name, description, details, price, is_available, is_active)
		VALUES (@UserID, @Name, @Description, @Details, @Price, @IsAvailable, @IsActive)`
	args := pgx.NamedArgs{
		"UserID":      product.UserID,
		"Name":        product.Name,
		"Description": product.Description,
		"Details":     product.Details,
		"IsAvailable": product.IsAvailable,
		"IsActive":    product.IsActive,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) List(ctx context.Context, pagination, page int) (*[]ProductSummary, error) {
	const q = `SELECT id, name, price, pub_time FROM products
		WHERE is_active = true AND is_available = true
		ORDER BY pub_time DESC
		LIMIT $1
		OFFSET $2`
	return list[ProductSummary](ctx, pr.session, q, pagination, (page-1)*pagination)
}

func (pr *ProductRepository) Get(ctx context.Context, id string) (*Product, error) {
	const q = `SELECT id, name, description, details, price, is_available, is_active FROM products
		WHERE id = $1::UUID
		LIMIT 1`
	return get[Product](ctx, pr.session, q, id)
}

func (pr *ProductRepository) Update(ctx context.Context, product *Product) error {
	const q = `UPDATE products
		SET name = @Name, description = @Description , details = @Details::JSONB, is_available = @IsAvailable, is_active = @IsActive
		WHERE id = @ID`
	args := pgx.NamedArgs{
		"Name":        product.Name,
		"Description": product.Description,
		"Details":     product.Details,
		"IsAvailable": product.IsAvailable,
		"IsActive":    product.IsActive,
		"ID":          product.ID,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM products WHERE id = $1::UUID`
	return execOne(ctx, pr.session, q, id)
}

func (pr *ProductRepository) SetAvailable(ctx context.Context, id string, isAvailable bool) error {
	const q = `UPDATE products SET is_available = $1 WHERE id = $2`
	return execOne(ctx, pr.session, q, isAvailable, id)
}

func (pr *ProductRepository) SetActive(ctx context.Context, id string, isActive bool) error {
	const q = `UPDATE products SET is_active = $1 WHERE id = $2`
	return execOne(ctx, pr.session, q, isActive, id)
}

type Category struct {
	ID  int32
	Tag string
}

type CategoryRepository struct {
	session db.Session
}

type CategoryStore interface {
	Create(context.Context, *Category) error
	List(context.Context) (*[]Category, error)
	Delete(context.Context, int32) error
}

func NewCategoryStore(session db.Session) CategoryStore {
	return &CategoryRepository{session}
}

func (cr *CategoryRepository) Create(ctx context.Context, category *Category) error {
	const q = `INSERT INTO categories(tag) VALUES ($1)`
	return execOne(ctx, cr.session, q, category.Tag)
}

func (cr *CategoryRepository) List(ctx context.Context) (*[]Category, error) {
	const q = `SELECT id, tag FROM categories`
	return list[Category](ctx, cr.session, q)
}

func (cr *CategoryRepository) Delete(ctx context.Context, id int32) error {
	const q = `DELETE FROM categories WHERE id = $1`
	return execOne(ctx, cr.session, q, id)
}

type PCRepository struct {
	session db.Session
}

type PCStore interface {
	SetTags(context.Context, string, []string) error
	List(context.Context, string) (*[]string, error)
}

func NewPCStore(session db.Session) PCStore {
	return &PCRepository{session}
}

func (pcr *PCRepository) SetTags(ctx context.Context, product_id string, tags []string) error {
	const q = `DELETE FROM products_categories WHERE product_id = $1`

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
			[]string{"product_id", "category_id"},
			pgx.CopyFromSlice(tagsLen, func(i int) ([]any, error) {
				return []any{product_id, tags[i]}, nil
			}),
		)
		return err
	})
}

func (pcr *PCRepository) List(ctx context.Context, id string) (*[]string, error) {
	const q = `SELECT category_id FROM products_categories WHERE product_id = $1`
	return list[string](ctx, pcr.session, q, id)
}
