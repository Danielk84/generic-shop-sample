package queries

import (
	"context"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
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
	log     logger.Logger
}

type ProductStore interface {
	Create(ctx context.Context, userID int32, Product *CreateProductRequest) error
	List(ctx context.Context, pagination, page int) ([]ProductSummaryResponse, error)
	FullList(ctx context.Context, userID int32, pagination, page int) ([]ProductStatusResponse, error)
	Get(ctx context.Context, id string) (OwnedProductResponse, error)
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

func (p *ProductRepository) Create(ctx context.Context, userID int32, product *CreateProductRequest) (err error) {
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
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Debug("ProductRepository.Create", "error", err)
	}
	return
}

func (p *ProductRepository) List(ctx context.Context, pagination, page int) (items []ProductSummaryResponse, err error) {
	const q = `SELECT id, name, price, pub_date FROM products
		WHERE is_active = true AND is_available = true
		ORDER BY pub_date DESC, available_quantity DESC, price
		LIMIT $1
		OFFSET $2`
	items, err = list[ProductSummaryResponse](ctx, p.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		p.log.Debug("ProductRepository.List", "error", err)
		return nil, err
	}
	return
}

func (p *ProductRepository) FullList(ctx context.Context, userID int32, pagination, page int) (items []ProductStatusResponse, err error) {
	const baseQuery = `SELECT id, name, price, pub_date, available_quantity, is_available, is_active FROM products`
	const limitOffset = ` LIMIT @Limit OFFSET @Offset`
	args := pgx.NamedArgs{
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}

	var q string
	if userID == 0 {
		q = baseQuery + ` ORDER BY pub_date DESC, is_active` + limitOffset
	} else {
		args["UserID"] = userID
		q = baseQuery + ` WHERE user_id = @UserID ORDER BY pub_date DESC` + limitOffset
	}
	items, err = list[ProductStatusResponse](ctx, p.session, q, args)
	if err != nil {
		p.log.Debug("ProductRepository.FullList", "error", err)
	}
	return
}

func (p *ProductRepository) Get(ctx context.Context, id string) (item OwnedProductResponse, err error) {
	const q = `SELECT id, user_id, name, description, COALESCE(details::TEXT, '{}') AS details,
			price, pub_date, available_quantity, is_available, is_active
		FROM products
		WHERE id = $1::UUID
		LIMIT 1`
	item, err = get[OwnedProductResponse](ctx, p.session, q, id)
	if err != nil {
		p.log.Debug("ProductRepository.Get", "error", err)
	}
	return
}

func (p *ProductRepository) Update(ctx context.Context, userID int32, product *UpdateProductRequest) (err error) {
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
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Warn("ProductRepository.Update")
	}
	return
}

func (p *ProductRepository) Delete(ctx context.Context, id string, userID int32) (err error) {
	const q = `DELETE FROM products WHERE id = $1::UUID and user_id = $2`
	if err = execOne(ctx, p.session, q, id, userID); err != nil {
		p.log.Error("ProductRepository.Delete", "error", err)
	}
	return
}

func (p *ProductRepository) IncrBy(ctx context.Context, id string, userID, num int32) (err error) {
	const q = `UPDATE products SET available_quantity = available_quantity + @Num WHERE id = @ID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Num":    num,
		"ID":     id,
		"UserID": userID,
	}
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Error("ProductRepository.IncrBy", "error", err)
	}
	return
}

func (p *ProductRepository) DecrBy(ctx context.Context, id string, userID, num int32) (err error) {
	const q = `UPDATE products SET available_quantity = available_quantity - @Num WHERE id = @ID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Num":    num,
		"ID":     id,
		"UserID": userID,
	}
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Error("ProductRepository.DecrBy", "error", err)
	}
	return
}

func (p *ProductRepository) SetAvailable(ctx context.Context, id string, isAvailable bool) (err error) {
	const q = `UPDATE products SET is_available = $1 WHERE id = $2`
	if err = execOne(ctx, p.session, q, isAvailable, id); err != nil {
		p.log.Error("ProductRepository.SetAvailable", "error", err)
	}
	return
}

func (p *ProductRepository) SetActive(ctx context.Context, id string, isActive bool) (err error) {
	const q = `UPDATE products SET is_active = $1 WHERE id = $2`
	if err = execOne(ctx, p.session, q, isActive, id); err != nil {
		p.log.Error("ProductRepository.SetActive", "error", err)
	}
	return
}

type ProductImageResponse struct {
	ID      string `json:"id"`
	ImgPath string `json:"img_path"`
}

type ProductImagesRepository struct {
	session database.Session
	log     logger.Logger
}

type ProductImagesStore interface {
	Create(ctx context.Context, productID, imgPath string) error
	List(ctx context.Context, productID string) ([]ProductImageResponse, error)
	Delete(ctx context.Context, id string) (string, error)
}

func NewProductImagesStore(session database.Session, logger logger.Logger) ProductImagesStore {
	return &ProductImagesRepository{session, logger}
}

func (p *ProductImagesRepository) Create(ctx context.Context, productID, imgPath string) (err error) {
	const productImagesCountQuery = `SELECT count(*) FROM product_images WHERE product_id = $1::UUID`
	count, err := get[int](ctx, p.session, productImagesCountQuery, productID)
	if err != nil {
		p.log.Debug("ProductImagesRepository.Create", "error", err)
		return
	}
	config := internal.GetConfig()
	if count >= config.Opt.MaxProductImagesAmount {
		return ErrFullCapacity
	}

	const createProductImageQuery = `INSERT INTO product_images(product_id, img_path) VALUES ($1::UUID, $2)`
	if err = execOne(ctx, p.session, createProductImageQuery, productID, imgPath); err != nil {
		p.log.Debug("ProductImagesRepository.Create", "error", err)
	}
	return
}

func (p *ProductImagesRepository) List(ctx context.Context, productID string) (items []ProductImageResponse, err error) {
	const q = `SELECT id, img_path FROM product_images WHERE product_id = $1::UUID`
	items, err = list[ProductImageResponse](ctx, p.session, q, productID)
	if err != nil {
		p.log.Debug("ProductImagesRepository.List", "error", err)
	}
	return
}

func (p *ProductImagesRepository) Delete(ctx context.Context, id string) (imgPath string, err error) {
	const q = `DELETE FROM product_images WHERE id = $1::UUID RETURNING img_path`
	if err = p.session.QueryRow(ctx, q, id).Scan(&imgPath); err != nil {
		p.log.Debug("ProductImagesRepository.Delete", "error", err)
	}
	return
}
