package queries

import (
	"context"
	"generic-shop-sample/db"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductSummary struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Price   int64     `json:"price"`
	PubDate time.Time `json:"pub_date"`
}

type Product struct {
	IsAvailable bool              `json:"is_available"`
	IsActive    bool              `json:"is_active"`
	Description *string           `json:"description"`
	Details     map[string]string `json:"details"`
	ProductSummary
}

type OwnedProduct struct {
	UserID int32 `json:"user_id"`
	Product
}

type ProductRepository struct {
	session db.Session
}

type ProductStore interface {
	Create(ctx context.Context, product *OwnedProduct) error
	List(ctx context.Context, pagination, page int) ([]ProductSummary, error)
	FullList(ctx context.Context, id int32, pagination, page int) ([]ProductSummary, error)
	Get(ctx context.Context, id string) (*OwnedProduct, error)
	Update(ctx context.Context, product *OwnedProduct) error
	Delete(ctx context.Context, id string, userID int32) error
	SetAvailable(ctx context.Context, id string, isActive bool) error
	SetActive(ctx context.Context, id string, isActive bool) error
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
		"Price":       product.Price,
		"Details":     product.Details,
		"IsAvailable": product.IsAvailable,
		"IsActive":    product.IsActive,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) List(ctx context.Context, pagination, page int) ([]ProductSummary, error) {
	const q = `SELECT id, name, price, pub_date FROM products
		WHERE is_active = true AND is_available = true
		ORDER BY pub_date DESC
		LIMIT $1
		OFFSET $2`
	return list[ProductSummary](ctx, pr.session, q, pagination, (page-1)*pagination)
}

func (pr *ProductRepository) FullList(ctx context.Context, id int32, pagination, page int) ([]ProductSummary, error) {
	const baseQuery = `SELECT id, name, price, pub_date FROM products`
	const limitOffset = ` LIMIT @Pagination OFFSET @Offset`
	args := pgx.NamedArgs{
		"Pagination": pagination,
		"Offset":     (page - 1) * pagination,
	}

	if id == 0 {
		q := baseQuery + ` ORDER BY pub_date DESC, is_active` + limitOffset
		return list[ProductSummary](ctx, pr.session, q, args)
	}
	args["ID"] = id
	q := baseQuery + ` WHERE user_id = @ID ORDER BY pub_date DESC` + limitOffset
	return list[ProductSummary](ctx, pr.session, q, args)
}

func (pr *ProductRepository) Get(ctx context.Context, id string) (*OwnedProduct, error) {
	const q = `SELECT id, user_id, name, description, details, price, pub_date, is_available, is_active FROM products
		WHERE id = $1::UUID
		LIMIT 1`
	return get[OwnedProduct](ctx, pr.session, q, id)
}

func (pr *ProductRepository) Update(ctx context.Context, product *OwnedProduct) error {
	const q = `UPDATE products
		SET name = @Name, description = @Description , details = @Details::JSONB, is_available = @IsAvailable
		WHERE id = @ID and user_id = @UserID`
	args := pgx.NamedArgs{
		"Name":        product.Name,
		"Description": product.Description,
		"Details":     product.Details,
		"IsAvailable": product.IsAvailable,
		"ID":          product.ID,
		"UserID":      product.UserID,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) Delete(ctx context.Context, id string, userID int32) error {
	const q = `DELETE FROM products WHERE id = $1::UUID and user_id = $2`
	return execOne(ctx, pr.session, q, id, userID)
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
	Create(ctx context.Context, tag string) error
	List(ctx context.Context) ([]Category, error)
	Delete(ctx context.Context, product_id int32) error
}

func NewCategoryStore(session db.Session) CategoryStore {
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
	session db.Session
}

type PCStore interface {
	SetTags(ctx context.Context, product_id string, tags []string) error
	List(ctx context.Context, product_id string) ([]string, error)
}

func NewPCStore(session db.Session) PCStore {
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
