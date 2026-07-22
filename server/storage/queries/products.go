package queries

import (
	"context"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductProperty = map[string]string

type CreateProductRequest struct {
	Name         string          `json:"name" binding:"required,min=4,max=256"`
	Description  string          `json:"description" binding:"required"`
	CommonDetail ProductProperty `json:"common_detail" binding:"required,min=0,dive,keys,required,min=1,endkeys,required,min=1"`
}

type ProductIDRequest struct {
	ID string `json:"id" binding:"required,uuid"`
}

type UpdateProductRequest struct {
	ProductIDRequest
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

type ProductVendor struct {
	UserID   string `json:"user_id" binding:"required,uuid"`
	Quantity int32  `json:"quantity" binding:"required,min=0"`
}

type ProductPropertyRequest struct {
	Property ProductProperty `json:"property" binding:"required,min=0,dive,keys,required,min=1,endkeys,required,min=1"`
}

type ProductVariantDetail struct {
	ProductPropertyRequest
	Price   int64           `json:"price" binding:"required,min=0"`
	Vendors []ProductVendor `json:"vendors" binding:"required,min=0,dive"`
}

type UpdateProductVendor struct {
	ProductIDRequest
	ProductVendor
	ProductPropertyRequest
}

type ProductResponse struct {
	ProductStatusResponse
	Description   string                 `json:"description"`
	CommonDetail  ProductProperty        `json:"common_detail"`
	VariantDetail []ProductVariantDetail `json:"variant_detail"`
}

type productRepository struct {
	session database.Session
	log     logger.Logger
}

type ProductStore interface {
	Create(ctx context.Context, product CreateProductRequest) error
	List(ctx context.Context, pagination, page int) ([]ProductSummaryResponse, error)
	AdminList(ctx context.Context, pagination, page int) ([]ProductStatusResponse, error)
	Get(ctx context.Context, id string) (ProductResponse, error)
	Update(ctx context.Context, product UpdateProductRequest) error
	Delete(ctx context.Context, id string) error
	SetVariantDetail(ctx context.Context, id string, variantDetail []ProductVariantDetail) error
	SetVendor(ctx context.Context, vendor UpdateProductVendor) error
	SetActive(ctx context.Context, id string, isActive bool) error
	GetVendors(ctx context.Context, id string, property ProductProperty) ([]string, error)
	GetQuantity(ctx context.Context, id string, property ProductProperty, userID string) (int32, error)
}

func NewProductStore(session database.Session, log logger.Logger) ProductStore {
	return &productRepository{session, log}
}

func (p *productRepository) Create(ctx context.Context, product CreateProductRequest) (err error) {
	const q = `INSERT INTO product_s.products(name, description, details, price, is_available, variant_detail)
		VALUES (@Name, @Description, '@CommonDetail'::JSONB, @Price, @IsAvailable, '[]'::JSONB)`
	args := pgx.NamedArgs{
		"Name":         product.Name,
		"Description":  product.Description,
		"CommonDetail": product.CommonDetail,
	}
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Debug("ProductRepository.Create", "error", err)
	}
	return
}

func (p *productRepository) List(ctx context.Context, pagination, page int) (items []ProductSummaryResponse, err error) {
	const q = `SELECT id, name, price, pub_date
		FROM product_s.products
		WHERE is_active = true
		ORDER BY
			pub_date DESC,
			available_quantity DESC,
			price,
			is_available DESC
		LIMIT $1
		OFFSET $2`

	items, err = list[ProductSummaryResponse](ctx, p.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		p.log.Debug("ProductRepository.List", "error", err)
		return nil, err
	}
	return
}

func (p *productRepository) AdminList(ctx context.Context, pagination, page int) (items []ProductStatusResponse, err error) {
	const q = `SELECT id, name, price, pub_date, available_quantity, is_available, is_active
		FROM product_s.products
		ORDER BY pub_date DESC, is_active
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	items, err = list[ProductStatusResponse](ctx, p.session, q, args)
	if err != nil {
		p.log.Debug("ProductRepository.AdminList", "error", err)
	}
	return
}

func (p *productRepository) Get(ctx context.Context, id string) (item ProductResponse, err error) {
	const q = `SELECT id, name, description,
			common_detail::TEXT, variant_detail::TEXT,
			price, pub_date, available_quantity, is_available, is_active
		FROM product_s.products
		WHERE id = '$1'::UUID
		LIMIT 1`
	item, err = get[ProductResponse](ctx, p.session, q, id)
	if err != nil {
		p.log.Debug("ProductRepository.Get", "error", err)
	}
	return
}

func (p *productRepository) Update(ctx context.Context, product UpdateProductRequest) (err error) {
	const q = `UPDATE product_s.products
		SET
			name = @Name,
			price = @Price,
			description = @Description,
			common_detail = NULLIF(@CommonDetail, '')::JSONB,
			is_available = @IsAvailable
		WHERE id = @ID`
	args := pgx.NamedArgs{
		"Name":        product.Name,
		"Description": product.Description,
		"Details":     product.CommonDetail,
		"ID":          product.ID,
	}
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Warn("ProductRepository.Update")
	}
	return
}

func (p *productRepository) Delete(ctx context.Context, id string) (err error) {
	const q = `DELETE FROM product_s.products WHERE id = $1::UUID`
	if err = execOne(ctx, p.session, q, id); err != nil {
		p.log.Error("ProductRepository.Delete", "error", err)
	}
	return
}

func (p *productRepository) SetVariantDetail(ctx context.Context, id string, variantDetail []ProductVariantDetail) (err error) {
	const q = `UPDATE product_s.products
		SET variant_detail = '@VariantDetail'::JSONB
		WHERE id = '@ID'::UUID`
	args := pgx.NamedArgs{
		"ID":            id,
		"VariantDetail": variantDetail,
	}
	if err = execOne(ctx, p.session, q, args); err != nil {
		p.log.Error("ProductRepository.SetVariantDetail", "error", err)
	}
	return
}

func (p *productRepository) SetVendor(ctx context.Context, vendor UpdateProductVendor) (err error) {
	var product ProductResponse
	if product, err = p.Get(ctx, vendor.ID); err != nil {
		p.log.Debug("ProductRepository.SetQuantity", "error", err)
		return
	}

	propertyFound := false
OuterLoop:
	for i, item := range product.VariantDetail {
		if reflect.DeepEqual(item.Property, vendor.Property) {
			propertyFound = true
			vendorFound := false
		InnerLoop:
			for j, v := range item.Vendors {
				if v.UserID == vendor.UserID {
					vendorFound = true
					product.VariantDetail[i].Vendors[j].Quantity = vendor.Quantity
					break InnerLoop
				}
			}
			if !vendorFound {
				product.VariantDetail[i].Vendors = append(product.VariantDetail[i].Vendors, vendor.ProductVendor)
			}
			break OuterLoop
		}
	}
	if !propertyFound {
		err = ErrNotFound
		p.log.Debug("ProductRepository.SetVendor", "error", err)
		return
	}

	if err = p.SetVariantDetail(ctx, vendor.ID, product.VariantDetail); err != nil {
		p.log.Debug("ProductRepository.SetVendor", "error", err)
	}
	return
}

func (p *productRepository) SetActive(ctx context.Context, id string, isActive bool) (err error) {
	const q = `UPDATE product_s.products SET is_active = $1 WHERE id = '$2'::UUID`
	if err = execOne(ctx, p.session, q, isActive, id); err != nil {
		p.log.Error("ProductRepository.SetActive", "error", err)
	}
	return
}

func (p *productRepository) GetVendors(ctx context.Context, id string, property ProductProperty) (items []string, err error) {
	const q = `SELECT unnest
		FROM unnest(product_s.get_vendors('@ProductID'::UUID, '@Property'::JSONB))`
	args := pgx.NamedArgs{
		"ProductID": id,
		"Property":  property,
	}
	if items, err = list[string](ctx, p.session, q, args); err != nil {
		p.log.Debug("ProductRepository.GetVendors", "error", err)
		return
	}
	return
}

func (p *productRepository) GetQuantity(ctx context.Context, id string, property ProductProperty, userID string) (item int32, err error) {
	const q = `SELECT product_s.get_quantity('@ProductID'::UUID, '@Property'::JSONB, @UserID)`
	args := pgx.NamedArgs{
		"ProductID": id,
		"Property":  property,
		"UserID":    userID,
	}
	if item, err = get[int32](ctx, p.session, q, args); err != nil {
		p.log.Debug("ProductRepository.GetQuantity", "error", err)
	}
	return
}

type ProductImageResponse struct {
	ID      string `json:"id"`
	ImgPath string `json:"img_path"`
}

type productImagesRepository struct {
	session database.Session
	log     logger.Logger
	config  config.ProductImageConfig
}

type ProductImagesStore interface {
	Create(ctx context.Context, productID, imgPath string) error
	List(ctx context.Context, productID string) ([]ProductImageResponse, error)
	Delete(ctx context.Context, id string) (string, error)
	BulkDelete(ctx context.Context, productID string) error
}

func NewProductImagesStore(session database.Session, logger logger.Logger, config config.ProductImageConfig) ProductImagesStore {
	return &productImagesRepository{session, logger, config}
}

func (p *productImagesRepository) Create(ctx context.Context, productID, imgPath string) (err error) {
	const productImagesCountQuery = `SELECT count(*)
		FROM product_s.product_images
		WHERE product_id = '$1'::UUID`
	count, err := get[int](ctx, p.session, productImagesCountQuery, productID)
	if err != nil {
		p.log.Debug("ProductImagesRepository.Create", "error", err)
		return
	}
	if count >= p.config.MaxProductImagesAmount {
		return ErrFullCapacity
	}

	const createProductImageQuery = `INSERT INTO product_s.product_images(product_id, img_path)
		VALUES ($1::UUID, $2)`
	if err = execOne(ctx, p.session, createProductImageQuery, productID, imgPath); err != nil {
		p.log.Debug("ProductImagesRepository.Create", "error", err)
	}
	return
}

func (p *productImagesRepository) List(ctx context.Context, productID string) (items []ProductImageResponse, err error) {
	const q = `SELECT id, img_path
		FROM product_s.product_images
		WHERE product_id = '$1'::UUID`
	items, err = list[ProductImageResponse](ctx, p.session, q, productID)
	if err != nil {
		p.log.Debug("ProductImagesRepository.List", "error", err)
	}
	return
}

func (p *productImagesRepository) Delete(ctx context.Context, id string) (imgPath string, err error) {
	const q = `DELETE FROM product_s.product_images
		WHERE id = '$1'::UUID
		RETURNING img_path`
	if err = p.session.QueryRow(ctx, q, id).Scan(&imgPath); err != nil {
		p.log.Debug("productImagesRepository.Delete", "error", err)
	}
	return
}

func (p *productImagesRepository) BulkDelete(ctx context.Context, productID string) (err error) {
	const q = `DELETE FROM product_s.product_images
		WHERE product_id = '$1'::UUID`
	if err = execOne(ctx, p.session, q, productID); err != nil {
		p.log.Debug("productImagesRepository.BulkDelete", "error", err)
	}
	return
}
