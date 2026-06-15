package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"time"

	"github.com/jackc/pgx/v5"
)

type OrderUserInfoRequest struct {
	Address string `json:"address" binding:"required"`
	ZipCode string `json:"zip_code" binding:"required"`
}

type OrderSummaryResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	StartedAt   time.Time `json:"started_at"`
	IsPaid      bool      `json:"is_paid"`
	IsDelivered bool      `json:"is_delivered"`
}

type OrderResponse struct {
	OrderSummaryResponse
	ItemsTotal  string `json:"items_total"`
	TotalBill   int64  `json:"total_bill"`
	Address     string `json:"address"`
	ZipCode     string `json:"zip_code"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type PaymentStatus struct {
	PaymentSummary string `json:"payment_summary"`
	IsPaid         bool   `json:"is_paid"`
}

type OrderRepository struct {
	session database.Session
	log     logger.Logger
}

type OrderStore interface {
	Create(ctx context.Context, userID int32) (string, error)
	CustomerList(ctx context.Context, userID int32, pagination, page int) ([]OrderSummaryResponse, error)
	VendorList(ctx context.Context, vendorID int32, pagination, page int) ([]OrderSummaryResponse, error)
	FullList(ctx context.Context, pagination, page int) ([]OrderSummaryResponse, error)
	Get(ctx context.Context, id string, userID int32) (OrderResponse, error)
	SetUserInfo(ctx context.Context, id string, userID int32, info *OrderUserInfoRequest) error
	VerifyUserInfo(ctx context.Context, id string, userID int32, isVerified bool) error
	SetPaymentStatus(ctx context.Context, id string, userID int32, status *PaymentStatus) error
	DeleteExpiredOrders(ctx context.Context) error
}

func NewOrderStore(session database.Session, log logger.Logger) OrderStore {
	return &OrderRepository{session, log}
}

func (o *OrderRepository) Create(ctx context.Context, userID int32) (item string, err error) {
	const q = `WITH create_new_order AS (
			INSERT INTO orders(user_id)
				SELECT $1 WHERE NOT EXISTS(SELECT 1 FROM orders WHERE user_id = $1 AND is_paid = FALSE)
		)
		SELECT id FROM orders WHERE user_id = $1 AND is_paid = FALSE`
	if err = o.session.QueryRow(ctx, q, userID).Scan(&item); err != nil {
		o.log.Warn("OrderRepository.Create", "error", err)
	}
	return
}

func (o *OrderRepository) CustomerList(ctx context.Context, userID int32, pagination, page int) (items []OrderSummaryResponse, err error) {
	const q = `SELECT id, user_id, started_at, is_paid, is_delivered FROM orders
		WHERE user_id = @UserID
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

func (o *OrderRepository) VendorList(ctx context.Context, vendorID int32, pagination, page int) (items []OrderSummaryResponse, err error) {
	const q = `SELECT id, user_id, started_at, is_paid, is_delivered FROM orders
		WHERE id in (
			SELECT DISTINCT o.order_id FROM order_items AS o JOIN products AS p ON o.product_id = p.id AND p.user_id = @VendorID
		) AND (is_paid = TRUE OR payment_summary IS NOT NULL)
		ORDER BY started_at DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"VendorID": vendorID,
		"Limit":    pagination,
		"Offset":   getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OrderSummaryResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderRepository.VendorList", "error", err)
	}
	return
}

func (o *OrderRepository) FullList(ctx context.Context, pagination, page int) (items []OrderSummaryResponse, err error) {
	const q = `SELECT id, user_id, started_at, is_paid, is_delivered FROM orders
		ORDER BY started_at DESC
		LIMIT $1
		OFFSET $2`
	items, err = list[OrderSummaryResponse](ctx, o.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		o.log.Debug("OrderRepository.FullList", "error", err)
	}
	return
}

func (o *OrderRepository) Get(ctx context.Context, id string, userID int32) (item OrderResponse, err error) {
	const q = `SELECT id, user_id, started_at, items_total, total_bill, is_paid, address, zip_code, is_confirmed, is_delivered FROM orders
		WHERE id = $1::UUID AND user_id = $2 AND is_confirmed = TRUE`
	item, err = get[OrderResponse](ctx, o.session, q, id, userID)
	if err != nil {
		o.log.Debug("OrderRepository.Get", "error", err)
	}
	return
}

func (o *OrderRepository) SetUserInfo(ctx context.Context, id string, userID int32, info *OrderUserInfoRequest) (err error) {
	const q = `UPDATE orders SET address = @Address, zip_code = @ZipCode, is_confirmed = FALSE
		WHERE id = @ID::UUID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Address": info.Address,
		"ZipCode": info.ZipCode,
		"ID":      id,
		"UserID":  userID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderRepository.SetUserInfo", "error", err)
	}
	return
}

func (o *OrderRepository) VerifyUserInfo(ctx context.Context, id string, userID int32, isVerified bool) (err error) {
	const q = `UPDATE orders SET is_confirmed = @IsVerified WHERE id = @ID::UUID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"IsVerified": isVerified,
		"ID":         id,
		"UserID":     userID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderRepository.VerifyUserInfo", "error", err)
	}
	return
}

func (o *OrderRepository) SetPaymentStatus(ctx context.Context, id string, userID int32, status *PaymentStatus) (err error) {
	const q = `UPDATE orders SET payment_summary = CONCAT(payment_summary::TEXT, ' ', @Summary::TEXT), is_paid = @IsPaid
		WHERE id = @ID::UUID AND user_id = @UserID`
	args := &pgx.NamedArgs{
		"Summary": status.PaymentSummary,
		"IsPaid":  status.IsPaid,
		"ID":      id,
		"UserID":  userID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Error("OrderRepository.SetPaymentStatus", "error", err)
	}
	return
}

func (o *OrderRepository) DeleteExpiredOrders(ctx context.Context) (err error) {
	const q = `DELETE FROM orders WHERE started_at > NOW() AND is_paid = FALSE AND payment_summary IS NULL`
	if _, err = o.session.Exec(ctx, q); err != nil {
		o.log.Warn("OrderRepository.DeleteExpiredOrders", "error", err)
	}
	return
}

type BaseOrderItemRequest struct {
	OrderID   string `json:"order_id" binding:"required,uuid"`
	ProductID string `json:"product_id" binding:"required,uuid"`
}

type OrderItemRequest struct {
	BaseOrderItemRequest
	Price int64 `json:"price" binding:"required"`
}

type OrderItemPackStatusRequest struct {
	BaseOrderItemRequest
	IsPacked bool `json:"is_packed" binding:"boolean"`
}

type OwnedOrderItemResponse struct {
	OrderID    string `json:"order_id"`
	ProductID  string `json:"product_id"`
	ItemsTotal string `json:"items_total"`
	Price      int64  `json:"price"`
	IsPacked   bool   `json:"is_packed"`
	Name       string `json:"name"`
}

type OrderItemsRepository struct {
	session database.Session
	log     logger.Logger
}

type OrderItemsStore interface {
	Create(ctx context.Context, userID int32, item *OrderItemRequest) error
	CustomerList(ctx context.Context, orderID string, userID int32, pagination, page int) ([]OwnedOrderItemResponse, error)
	VendorList(ctx context.Context, orderID string, userID int32, pagination, page int) ([]OwnedOrderItemResponse, error)
	FullList(ctx context.Context, orderID string, pagination, page int) ([]OwnedOrderItemResponse, error)
	Delete(ctx context.Context, userID int32, orderItem *BaseOrderItemRequest) error
	SetItemsTotal(ctx context.Context, userID int32, itemsTotal int32, orderItem *BaseOrderItemRequest) error
	SetPacked(ctx context.Context, vendorID int32, item *OrderItemPackStatusRequest) error
}

func NewOrderItemsStore(session database.Session, log logger.Logger) OrderItemsStore {
	return &OrderItemsRepository{session, log}
}

func (o *OrderItemsRepository) Create(ctx context.Context, userID int32, item *OrderItemRequest) (err error) {
	const q = `WITH is_owned_order AS (
		SELECT id FROM orders WHERE id = @OrderID::UUID AND user_id = @UserID LIMIT 1
	)
	INSERT INTO order_items(user_id, order_id, product_id, price)
		SELECT @UserID, id, @ProductID::UUID, @Price FROM is_owned_order`
	args := pgx.NamedArgs{
		"UserID":    userID,
		"OrderID":   item.OrderID,
		"ProductID": item.ProductID,
		"Price":     item.Price,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Warn("OrderItemRepository.Create", "error", err)
	}
	return
}

func (o *OrderItemsRepository) CustomerList(ctx context.Context,
	orderID string, userID int32, pagination, page int) (items []OwnedOrderItemResponse, err error) {
	const q = `SELECT o.order_id, o.product_id, o.items_total, o.price, o.is_packed, p.name
		FROM order_items AS o
			LEFT JOIN products AS p ON o.product_id = p.id
		WHERE o.order_id = @OrderID::UUID AND o.user_id = @UserID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"OrderID": orderID,
		"UserID":  userID,
		"Limit":   pagination,
		"Offset":  getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OwnedOrderItemResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderItemsRepository.CustomerList", "error", err)
	}
	return
}

func (o *OrderItemsRepository) VendorList(ctx context.Context,
	orderID string, vendorID int32, pagination, page int) (items []OwnedOrderItemResponse, err error) {
	const q = `SELECT o.order_id, o.product_id, o.items_total, o.price, o.is_packed, p.name 
		FROM order_items AS o
			JOIN products AS p ON o.product_id = p.id AND p.user_id = @VendorID
		WHERE o.order_id = @OrderID::UUID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"OrderID":  orderID,
		"VendorID": vendorID,
		"Limit":    pagination,
		"Offset":   getOffsetFromPageNum(pagination, page),
	}
	items, err = list[OwnedOrderItemResponse](ctx, o.session, q, args)
	if err != nil {
		o.log.Debug("OrderItemsRepository.VendorList")
	}
	return
}

func (o *OrderItemsRepository) FullList(ctx context.Context,
	orderID string, pagination, page int) (items []OwnedOrderItemResponse, err error) {
	const q = `SELECT order_id, product_id, items_total, price, is_packed, '' as name FROM order_items
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
		o.log.Debug("OrderItemsRepository.FullList", "error", err)
	}
	return
}

func (o *OrderItemsRepository) Delete(ctx context.Context, userID int32, orderItem *BaseOrderItemRequest) (err error) {
	const q = `DELETE FROM order_items WHERE order_id = @OrderID::UUID AND product_id = @ProductID::UUID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"OrderID":   orderItem.OrderID,
		"ProductID": orderItem.ProductID,
		"UserID":    userID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Debug("OrderItemsRepository.Delete", "error", err)
	}
	return
}

func (o *OrderItemsRepository) SetItemsTotal(ctx context.Context,
	userID int32, itemsTotal int32, orderItem *BaseOrderItemRequest) (err error) {
	const q = `UPDATE order_items SET items_total = @ItemsTotal
		WHERE order_id = @OrderID::UUID AND product_id = @ProductID::UUID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"OrderID":    orderItem.OrderID,
		"ProductID":  orderItem.ProductID,
		"UserID":     userID,
		"ItemsTotal": itemsTotal,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Error("OrderItemsRepository.SetItemsTotal", "error", err)
	}
	return
}

func (o *OrderItemsRepository) SetPacked(ctx context.Context, vendorID int32, item *OrderItemPackStatusRequest) (err error) {
	const q = `UPDATE order_items SET is_packed = @IsPacked
		WHERE order_id = @OrderID AND product_id in (
			SELECT id FROM products WHERE id = @ProductID AND user_id = @VendorID  
		)`
	args := pgx.NamedArgs{
		"OrderID":   item.OrderID,
		"ProductID": item.ProductID,
		"IsPacked":  item.IsPacked,
		"VendorID":  vendorID,
	}
	if err = execOne(ctx, o.session, q, args); err != nil {
		o.log.Warn("OrderItemsRepository.SetPacked", "error", err)
	}
	return
}
