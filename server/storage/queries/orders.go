package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"time"

	"github.com/jackc/pgx/v5"
)

type OrderUserInfo struct {
	Address string `json:"address" binding:"required"`
	ZipCode string `json:"zip_code" binding:"required"`
}

type OrderID struct {
	ID     string `json:"id" binding:"required,uuid"`
	UserID string `json:"user_id" binding:"required,uuid"`
}

type OrderSummaryResponse struct {
	OrderID
	StartedAt   time.Time `json:"started_at"`
	IsPaid      bool      `json:"is_paid"`
	IsDelivered bool      `json:"is_delivered"`
}

type OrderResponse struct {
	OrderSummaryResponse
	OrderUserInfo
	ItemsTotal     int32             `json:"items_total"`
	TotalBill      int64             `json:"total_bill"`
	IsVerified     bool              `json:"is_verified"`
	IsConfirmed    bool              `json:"is_confirmed"`
	PaymentSummary []ProductProperty `json:"payment_summary"`
}

type PaymentStatus struct {
	PaymentSummary ProductProperty `json:"payment_summary" binding:"required,min=0,dive,keys,required,min=1,endkeys,required,min=1"`
	IsPaid         bool            `json:"is_paid" binding:"required"`
}

type orderRepository struct {
	session database.Session
	log     logger.Logger
}

type OrderStore interface {
	Create(ctx context.Context, userID string) (string, error)
	CustomerList(ctx context.Context, userID string, pagination, page int) ([]OrderSummaryResponse, error)
	FullList(ctx context.Context, pagination, page int) ([]OrderSummaryResponse, error)
	NotConfirmedList(ctx context.Context, pagination, page int) ([]OrderSummaryResponse, error)
	Get(ctx context.Context, id OrderID) (OrderResponse, error)
	SetUserInfo(ctx context.Context, id OrderID, info OrderUserInfo) error
	VerifyUserInfo(ctx context.Context, id OrderID, isVerified bool) error
	SetPaymentStatus(ctx context.Context, id OrderID, status PaymentStatus) error
	SetConfirmed(ctx context.Context, id OrderID, IsConfirmed bool) error
	DeleteExpiredOrders(ctx context.Context) error
}

func NewOrderStore(session database.Session, log logger.Logger) OrderStore {
	return &orderRepository{session, log}
}

func (o *orderRepository) Create(ctx context.Context, userID string) (item string, err error) {
	// this query check, if there is not ended order (not paid), return existed id
	// instead of creating new one.
	const q = `WITH matched AS (
			SELECT id FROM order_s.orders 
			WHERE user_id = $1::UUID AND is_paid = FALSE
			LIMIT 1
		),
		inserted AS (
			INSERT INTO order_s.orders (user_id)
			SELECT $1::UUID
			WHERE NOT EXISTS (SELECT 1 FROM matched)
			RETURNING id
		)
		SELECT id FROM matched
		UNION ALL
		SELECT id FROM inserted`
	if err = o.session.QueryRow(ctx, q, userID).Scan(&item); err != nil {
		o.log.Warn("OrderRepository.Create", "error", err)
	}
	return
}

func (o *orderRepository) CustomerList(ctx context.Context, userID string, pagination, page int) (items []OrderSummaryResponse, err error) {
	const q = `SELECT id, user_id, started_at, is_paid, is_delivered
		FROM order_s.orders
		WHERE user_id = @UserID::UUID
		ORDER BY started_at DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"UserID": userID,
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OrderSummaryResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderRepository.CustomerList", "error", err)
	}
	return
}

func (o *orderRepository) NotConfirmedList(ctx context.Context, pagination, page int) (items []OrderSummaryResponse, err error) {
	const q = `SELECT id, user_id, started_at, is_paid, is_delivered
		FROM order_s.orders
		WHERE is_confirmed = FALSE AND is_paid = TRUE
		ORDER BY started_at DESC
		LIMIT $1
		OFFSET $2`
	items, err = list[OrderSummaryResponse](ctx, o.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		o.log.Debug("OrderRepository.NotConfirmedList", "error", err)
	}
	return
}

func (o *orderRepository) FullList(ctx context.Context, pagination, page int) (items []OrderSummaryResponse, err error) {
	const q = `SELECT id, user_id, started_at, is_paid, is_delivered
		FROM order_s.orders
		ORDER BY started_at DESC, is_confirmed
		LIMIT $1
		OFFSET $2`
	items, err = list[OrderSummaryResponse](ctx, o.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		o.log.Debug("OrderRepository.FullList", "error", err)
	}
	return
}

func (o *orderRepository) Get(ctx context.Context, id OrderID) (item OrderResponse, err error) {
	const q = `SELECT
			id, user_id, started_at,
			items_total, total_bill, is_paid,
			COALESCE(address, '') AS address, COALESCE(zip_code, '') AS zip_code, is_confirmed,
			is_verified, is_delivered, payment_summary
		FROM order_s.orders
		WHERE id = $1::UUID AND user_id = $2::UUID`
	item, err = get[OrderResponse](ctx, o.session, q, id.ID, id.UserID)
	if err != nil {
		o.log.Debug("OrderRepository.Get", "error", err)
	}
	return
}

func (o *orderRepository) SetUserInfo(ctx context.Context, id OrderID, info OrderUserInfo) (err error) {
	const q = `UPDATE order_s.orders
		SET address = @Address, zip_code = @ZipCode, is_verified = FALSE
		WHERE id = @ID::UUID AND user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"Address": info.Address,
		"ZipCode": info.ZipCode,
		"ID":      id.ID,
		"UserID":  id.UserID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderRepository.SetUserInfo", "error", err)
	}
	return
}

func (o *orderRepository) VerifyUserInfo(ctx context.Context, id OrderID, isVerified bool) (err error) {
	const q = `UPDATE order_s.orders
		SET is_verified = @IsVerified
		WHERE id = @ID::UUID AND user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"IsVerified": isVerified,
		"ID":         id.ID,
		"UserID":     id.UserID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderRepository.VerifyUserInfo", "error", err)
	}
	return
}

func (o *orderRepository) SetPaymentStatus(ctx context.Context, id OrderID, status PaymentStatus) (err error) {
	const q = `UPDATE order_s.orders
		SET payment_summary = jsonb_insert(payment_summary, '{-1}', @Summary::JSONB), is_paid = @IsPaid
		WHERE id = @ID::UUID AND user_id = @UserID::UUID`
	args := &pgx.NamedArgs{
		"Summary": status.PaymentSummary,
		"IsPaid":  status.IsPaid,
		"ID":      id.ID,
		"UserID":  id.UserID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Error("OrderRepository.SetPaymentStatus", "error", err)
	}
	return
}

func (o *orderRepository) SetConfirmed(ctx context.Context, id OrderID, IsConfirmed bool) (err error) {
	const q = `UPDATE order_s.orders
		SET is_confirmed = @IsConfirmed
		WHERE id = @ID::UUID AND user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"IsConfirmed": IsConfirmed,
		"ID":          id.ID,
		"UserID":      id.UserID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderRepository.SetConfirmed", "error", err)
	}
	return
}

func (o *orderRepository) DeleteExpiredOrders(ctx context.Context) (err error) {
	const q = `DELETE FROM order_s.orders
		WHERE is_paid = FALSE
		AND started_at <= NOW()
		AND is_confirmed = FALSE`
	if _, err = o.session.Exec(ctx, q); err != nil {
		o.log.Warn("OrderRepository.DeleteExpiredOrders", "error", err)
	}
	return
}

type OrderItem struct {
	OrderID   string `json:"order_id" binding:"required,uuid"`
	ProductID string `json:"product_id" binding:"required,uuid"`
}

type OrderItemID struct {
	OrderItem
	UserID string `json:"user_id" binding:"required,uuid"`
}

type OrderItemRequest struct {
	OrderItem
	Property ProductProperty `json:"property" binding:"required,min=0,dive,keys,required,min=1,endkeys,required,min=1"`
	Price    int64           `json:"price"  binding:"required,min=0"`
}

type OwnedOrderItemResponse struct {
	OrderItem
	ItemsTotal       int32           `json:"items_total"`
	ProcessedItems   int32           `json:"processed_items"`
	Price            int64           `json:"price"`
	ConfirmedVendors []ProductVendor `json:"confirmed_vendors"`
	Name             string          `json:"name"`
}

type OrderItemResponse struct {
	OwnedOrderItemResponse
	Property ProductProperty
}

type orderItemsRepository struct {
	session database.Session
	log     logger.Logger
}

type OrderItemsStore interface {
	Create(ctx context.Context, userID string, item OrderItemRequest) error
	CustomerList(ctx context.Context, id OrderID, pagination, page int) ([]OwnedOrderItemResponse, error)
	AdminList(ctx context.Context, orderID string, pagination, page int) ([]OwnedOrderItemResponse, error)
	FullList(ctx context.Context, orderID string, pagination, page int) ([]OrderItemResponse, error)
	Delete(ctx context.Context, id OrderItemID) error
	SetItemsTotal(ctx context.Context, id OrderItemID, itemsTotal int32) error
	SetConfirmedVendors(ctx context.Context, id OrderItemID, confirmedVendors []ProductVendor) error
}

func NewOrderItemsStore(session database.Session, log logger.Logger) OrderItemsStore {
	return &orderItemsRepository{session, log}
}

func (o *orderItemsRepository) Create(ctx context.Context, userID string, item OrderItemRequest) (err error) {
	const q = `WITH is_owned_order AS (
		SELECT id
		FROM order_s.orders
		WHERE id = @OrderID::UUID AND user_id = @UserID::UUID AND is_paid = FALSE
		LIMIT 1
	)
	INSERT INTO order_s.order_items(user_id, order_id, product_id, price, property)
		SELECT @UserID::UUID, id, @ProductID::UUID, @Price, @Property::JSONB
		FROM is_owned_order`
	args := pgx.NamedArgs{
		"UserID":    userID,
		"OrderID":   item.OrderID,
		"ProductID": item.ProductID,
		"Price":     item.Price,
		"Property":  item.Property,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Warn("OrderItemRepository.Create", "error", err)
	}
	return
}

func (o *orderItemsRepository) CustomerList(ctx context.Context,
	id OrderID, pagination, page int) (items []OwnedOrderItemResponse, err error) {
	const q = `SELECT
			o.order_id, o.product_id, o.items_total,
			o.processed_items, o.price, p.name,
			o.confirmed_vendors
		FROM order_s.order_items AS o
			LEFT JOIN product_s.products AS p
			ON o.product_id = p.id
		WHERE o.order_id = @OrderID::UUID AND o.user_id = @UserID::UUID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"OrderID": id.ID,
		"UserID":  id.UserID,
		"Limit":   pagination,
		"Offset":  getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OwnedOrderItemResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderItemsRepository.CustomerList", "error", err)
	}
	return
}

func (o *orderItemsRepository) AdminList(ctx context.Context,
	orderID string, pagination, page int) (items []OwnedOrderItemResponse, err error) {
	const q = `SELECT
			order_id, product_id, items_total,
			processed_items, price, '' as name,
			confirmed_vendors
		FROM order_s.order_items
		WHERE order_id = @OrderID::UUID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"OrderID": orderID,
		"Limit":   pagination,
		"Offset":  getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OwnedOrderItemResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderItemsRepository.AdminList", "error", err)
	}
	return
}

func (o *orderItemsRepository) FullList(ctx context.Context,
	orderID string, pagination, page int) (items []OrderItemResponse, err error) {
	const q = `SELECT
			order_id, product_id, items_total,
			processed_items, price, '' as name,
			property, confirmed_vendors
		FROM order_s.order_items
		WHERE order_id = @OrderID::UUID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"OrderID": orderID,
		"Limit":   pagination,
		"Offset":  getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OrderItemResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderItemsRepository.FullList", "error", err)
	}
	return
}

func (o *orderItemsRepository) Delete(ctx context.Context, id OrderItemID) (err error) {
	const q = `DELETE FROM order_s.order_items
		WHERE order_id = @OrderID::UUID AND product_id = @ProductID::UUID AND user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"OrderID":   id.OrderID,
		"ProductID": id.ProductID,
		"UserID":    id.UserID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderItemsRepository.Delete", "error", err)
	}
	return
}

func (o *orderItemsRepository) SetItemsTotal(ctx context.Context, id OrderItemID, itemsTotal int32) (err error) {
	const q = `UPDATE order_s.order_items
		SET items_total = @ItemsTotal
		WHERE order_id = @OrderID::UUID AND product_id = @ProductID::UUID AND user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"OrderID":    id.OrderID,
		"ProductID":  id.ProductID,
		"UserID":     id.UserID,
		"ItemsTotal": itemsTotal,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Error("OrderItemsRepository.SetItemsTotal", "error", err)
	}
	return
}

func (o *orderItemsRepository) SetConfirmedVendors(ctx context.Context, id OrderItemID, confirmedVendors []ProductVendor) (err error) {
	const q = `UPDATE order_s.order_items
		SET confirmed_vendors = @ConfirmedVendors::JSONB
		WHERE order_id = @OrderID::UUID AND product_id = @ProductID::UUID AND user_id = @UserID::UUID`
	args := pgx.NamedArgs{
		"ConfirmedVendors": confirmedVendors,
		"OrderID":          id.OrderID,
		"ProductID":        id.ProductID,
		"UserID":           id.UserID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderItemsRepository.SetConfirmedVendors", "error", err)
	}
	return
}
