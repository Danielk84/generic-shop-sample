package queries

import (
	"context"
	"generic-shop-sample/db/database"
	"generic-shop-sample/internal"
	"time"

	"github.com/jackc/pgx/v5"
)

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required,min=4,max=256"`
	Price       int64  `json:"price" binding:"required,number,gte=0"`
	Description string `json:"description" binding:"required"`
	Details     string `json:"details" binding:"json"`
	IsAvailable bool   `json:"is_available" binding:"boolean"`
}

type UpdateProductRequest struct {
	ID string `json:"id" binding:"uuid"`
	CreateProductRequest
}

type ProductSummaryResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Price   int64     `json:"price"`
	PubDate time.Time `json:"pub_date"`
}

type ProductStatusResponse struct {
	ProductSummaryResponse
	AvailableQuantity int32 `json:"available_quantity"`
	IsAvailable       bool  `json:"is_available"`
	IsActive          bool  `json:"is_active"`
}

type OwnedProductResponse struct {
	ProductStatusResponse
	UserID      int32  `json:"user_id"`
	Description string `json:"description"`
	Details     string `json:"details"`
}

type ProductRepository struct {
	session database.Session
}

type ProductStore interface {
	Create(ctx context.Context, userID int32, Product *CreateProductRequest) error
	List(ctx context.Context, pagination, page int) ([]ProductSummaryResponse, error)
	FullList(ctx context.Context, userID int32, pagination, page int) ([]ProductStatusResponse, error)
	Get(ctx context.Context, id string) (*OwnedProductResponse, error)
	Update(ctx context.Context, userID int32, product *UpdateProductRequest) error
	Delete(ctx context.Context, id string, userID int32) error
	IncrBy(ctx context.Context, id string, userID, num int32) error
	DecrBy(ctx context.Context, id string, userID, num int32) error
	SetAvailable(ctx context.Context, id string, isActive bool) error
	SetActive(ctx context.Context, id string, isActive bool) error
}

func NewProductStore(session database.Session) ProductStore {
	return &ProductRepository{session: session}
}

func (pr *ProductRepository) Create(ctx context.Context, userID int32, product *CreateProductRequest) error {
	const q = `INSERT INTO products(user_id, name, description, details, price, is_available)
		VALUES (@UserID, @Name, @Description, NULLIF(@Details, '')::JSONB, @Price, @IsAvailable)`
	args := pgx.NamedArgs{
		"UserID":      userID,
		"Name":        product.Name,
		"Description": product.Description,
		"Details":     product.Details,
		"Price":       product.Price,
		"IsAvailable": product.IsAvailable,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) List(ctx context.Context, pagination, page int) ([]ProductSummaryResponse, error) {
	const q = `SELECT id, name, price, pub_date FROM products
		WHERE is_active = true AND is_available = true
		ORDER BY pub_date DESC, available_quantity DESC, price
		LIMIT $1
		OFFSET $2`
	return list[ProductSummaryResponse](ctx, pr.session, q, pagination, getOffsetFromPageNum(pagination, page))
}

func (pr *ProductRepository) FullList(ctx context.Context, userID int32, pagination, page int) ([]ProductStatusResponse, error) {
	const baseQuery = `SELECT id, name, price, pub_date, available_quantity, is_available, is_active FROM products`
	const limitOffset = ` LIMIT @Limit OFFSET @Offset`
	args := pgx.NamedArgs{
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}

	if userID == 0 {
		q := baseQuery + ` ORDER BY pub_date DESC, is_active` + limitOffset
		return list[ProductStatusResponse](ctx, pr.session, q, args)
	}
	args["UserID"] = userID
	q := baseQuery + ` WHERE user_id = @UserID ORDER BY pub_date DESC` + limitOffset
	return list[ProductStatusResponse](ctx, pr.session, q, args)
}

func (pr *ProductRepository) Get(ctx context.Context, id string) (*OwnedProductResponse, error) {
	const q = `SELECT id, user_id, name, description, COALESCE(details::TEXT, '{}') AS details, price, pub_date, available_quantity, is_available, is_active
		FROM products
		WHERE id = $1::UUID
		LIMIT 1`
	return get[OwnedProductResponse](ctx, pr.session, q, id)
}

func (pr *ProductRepository) Update(ctx context.Context, userID int32, product *UpdateProductRequest) error {
	const q = `UPDATE products
		SET name = @Name, price = @Price, description = @Description , details = NULLIF(@Details, '')::JSONB, is_available = @IsAvailable
		WHERE id = @ID and user_id = @UserID`
	args := pgx.NamedArgs{
		"Name":        product.Name,
		"Price":       product.Price,
		"Description": product.Description,
		"Details":     product.Details,
		"IsAvailable": product.IsAvailable,
		"ID":          product.ID,
		"UserID":      userID,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) Delete(ctx context.Context, id string, userID int32) error {
	const q = `DELETE FROM products WHERE id = $1::UUID and user_id = $2`
	return execOne(ctx, pr.session, q, id, userID)
}

func (pr *ProductRepository) IncrBy(ctx context.Context, id string, userID, num int32) error {
	const q = `UPDATE products SET available_quantity = available_quantity + @Num WHERE id = @ID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Num":    num,
		"ID":     id,
		"UserID": userID,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) DecrBy(ctx context.Context, id string, userID, num int32) error {
	const q = `UPDATE products SET available_quantity = available_quantity - @Num WHERE id = @ID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Num":    num,
		"ID":     id,
		"UserID": userID,
	}
	return execOne(ctx, pr.session, q, args)
}

func (pr *ProductRepository) SetAvailable(ctx context.Context, id string, isAvailable bool) error {
	const q = `UPDATE products SET is_available = $1 WHERE id = $2`
	return execOne(ctx, pr.session, q, isAvailable, id)
}

func (pr *ProductRepository) SetActive(ctx context.Context, id string, isActive bool) error {
	const q = `UPDATE products SET is_active = $1 WHERE id = $2`
	return execOne(ctx, pr.session, q, isActive, id)
}

type ProductImageResponse struct {
	ID      string `json:"id"`
	ImgPath string `json:"img_path"`
}

type ProductImagesRepository struct {
	session database.Session
}

type ProductImagesStore interface {
	Create(ctx context.Context, productID, imgPath string) error
	List(ctx context.Context, productID string) ([]ProductImageResponse, error)
	Delete(ctx context.Context, id string) (string, error)
}

func NewProductImagesStore(session database.Session) ProductImagesStore {
	return &ProductImagesRepository{session}
}

func (pir *ProductImagesRepository) Create(ctx context.Context, productID, imgPath string) error {
	const productImagesCountQuery = `SELECT count(*) FROM product_images WHERE product_id = $1::UUID`
	count, err := get[int](ctx, pir.session, productImagesCountQuery, productID)
	if err != nil {
		return err
	}
	config := internal.GetConfig()
	if *count >= config.Opt.MaxProductImagesAmount {
		return ErrFullCapacity
	}

	const createProductImageQuery = `INSERT INTO product_images(product_id, img_path) VALUES ($1::UUID, $2)`
	return execOne(ctx, pir.session, createProductImageQuery, productID, imgPath)
}

func (pir *ProductImagesRepository) List(ctx context.Context, productID string) ([]ProductImageResponse, error) {
	const q = `SELECT id, img_path FROM product_images WHERE product_id = $1::UUID`
	return list[ProductImageResponse](ctx, pir.session, q, productID)
}

func (pir *ProductImagesRepository) Delete(ctx context.Context, id string) (string, error) {
	const q = `DELETE FROM product_images WHERE id = $1::UUID RETURNING img_path`
	imgPath := ""
	err := pir.session.QueryRow(ctx, q, id).Scan(&imgPath)
	return imgPath, err
}
