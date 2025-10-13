package queries

import (
	"context"
	"generic-shop-sample/db"

	"github.com/jackc/pgx/v5"
)

type OrderUserInfoRequest struct {
	Address string `json:"address" binding:"required"`
	ZipCode bool   `json:"zip_code" binding:"required"`
}

type OrderSummaryResponse struct {
	ID         string `json:"id"`
	StartedAt  string `json:"update_at"`
	IsPaid     bool   `json:"is_paid"`
	IsDeliverd bool   `json:"is_deliverd"`
}

type OrderResponse struct {
	OrderSummaryResponse
	ItemsTotal string `json:"items_total"`
	TotalBill  int64  `json:"total_bill"`
	Address    string `json:"address"`
	ZipCode    string `json:"zip_code"`
}

type OrderRepository struct {
	session db.Session
}

type OrderStore interface {
	Create(ctx context.Context, userID int32) (string, error)
	CustomerList(ctx context.Context, userID int32, pagination, page int) ([]OrderSummaryResponse, error)
	VendorList(ctx context.Context, vendorID int32, pagination, page int) ([]OrderSummaryResponse, error)
	Get(ctx context.Context, id string, userID int32) (*OrderResponse, error)
	SetUserInfo(ctx context.Context, id string, userID int32, info *OrderUserInfoRequest) error
	VerifyUserInfo(ctx context.Context, id string, userID int32, isVerified bool) error
	DeleteExpiredOrders(ctx context.Context) error
}

func NewOrderStore(session db.Session) OrderStore {
	return &OrderRepository{session}
}

func (or *OrderRepository) Create(ctx context.Context, userID int32) (string, error) {
	const q = `INSERT INTO orders(user_id) VALUES ($1) RETURNING order_id`
	id := ""
	err := or.session.QueryRow(ctx, q, userID).Scan(&id)
	return id, err
}

func (or *OrderRepository) CustomerList(ctx context.Context, userID int32, pagination, page int) ([]OrderSummaryResponse, error) {
	const q = `SELECT id, started_at, is_paid, is_delivered FROM orders
		WHERE user_id = @UserID
		ORDER BY started_at DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"UserID": userID,
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	return list[OrderSummaryResponse](ctx, or.session, q, args)
}

func (or *OrderRepository) VendorList(ctx context.Context, vendorID int32, pagination, page int) ([]OrderSummaryResponse, error) {
	const q = `SELECT id, started_id, is_paid, is_delivered FROM orders
		WHERE id in (
			SELECT DISTINCT o.order_id FROM order_items AS o JOIN products AS p ON o.product_id = p.id AND p.user_id = @VendoreID
		) AND (is_paid = TRUE OR payment_summary IS NOT NULL)
		ORDER BY started_at DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"VendorID": vendorID,
		"Limit":    pagination,
		"Offset":   getOffsetFromPageNum(pagination, page),
	}
	return list[OrderSummaryResponse](ctx, or.session, q, args)
}

func (or *OrderRepository) Get(ctx context.Context, id string, userID int32) (*OrderResponse, error) {
	const q = `SELECT id, started_at, items_total, total_bill, is_paid, address, zip_code, is_delivered FROM orders
		WHERE id = $1::UUID AND user_id = $2 AND is_confirmed = TRUE`
	return get[OrderResponse](ctx, or.session, q, id, userID)
}

func (or *OrderRepository) SetUserInfo(ctx context.Context, id string, userID int32, info *OrderUserInfoRequest) error {
	const q = `UPDATE orders SET address = @Address, zip_code = @ZipCode, is_confirmed = FALSE
		WHERE id = @ID::UUID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Address": info.Address,
		"ZipCode": info.ZipCode,
		"ID":      id,
		"UserID":  userID,
	}
	return execOne(ctx, or.session, q, args)
}

func (or *OrderRepository) VerifyUserInfo(ctx context.Context, id string, userID int32, isVerified bool) error {
	const q = `UPDATE orders SET is_confirmed = @IsVerfied WHERE id = @ID::UUID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"IsVerified": isVerified,
		"ID":         id,
		"UserID":     userID,
	}
	return execOne(ctx, or.session, q, args)
}

func (or *OrderRepository) DeleteExpiredOrders(ctx context.Context) error {
	const q = `DELETE FROM orders WHERE started_at > NOW() AND is_paid = FALSE AND payment_summary IS NULL`
	_, err := or.session.Exec(ctx, q)
	return err
}

type BaseOrderItemsRequest struct {
	OrderID   string `json:"order_id" binding:"required,uuid"`
	ProductID string `json:"product_id" binding:"required,uuid"`
}

type OrderItemRequest struct {
	BaseOrderItemsRequest
	Price int64 `json:"price" binding:"required"`
}

type OrderItemPackStatusRequest struct {
	BaseOrderItemsRequest
	IsPacked bool `json:"is_packed" binding:"boolean"`
}

type OwnedOrderItemResponse struct {
	ID         string `json:"id"`
	ProductID  string `json:"product_id"`
	ItemsTotal string `json:"items_total"`
	Price      int64  `json:"price"`
	IsPacked   bool   `json:"is_packed"`
	Name       string `json:"name"`
}

type OrderItemsRepository struct {
	session db.Session
}

type OrderItemsStore interface {
	Create(ctx context.Context, userID, item *OrderItemRequest) error
	CustomerList(ctx context.Context, orderID string, userID int32) ([]OwnedOrderItemResponse, error)
	VendorList(ctx context.Context, orderID string, userID int32) ([]OwnedOrderItemResponse, error)
	Delete(ctx context.Context, orderID string, userID int32) error
	SetItemsTotal(ctx context.Context, orderID string, userID int32, itemsTotal int32) error
	SetPacked(ctx context.Context, vendorID int32, item *OrderItemPackStatusRequest) error
}

func NewOrderItemsStore(session db.Session) OrderItemsStore {
	return &OrderItemsRepository{session}
}

func (or *OrderItemsRepository) Create(ctx context.Context, userID, item *OrderItemRequest) error {
	const q = `WITH is_owned_order AS (
		SELECT id FROM orders WHERE id = @OrderID::UUID AND user_id = @UserID LIMIT 1
	)
	INSERT INTO order_items(order_id, product_id, price)
		SELECT @UserID, id, @ProductID::UUID, @Price FROM is_owned_order`
	args := pgx.NamedArgs{
		"UserID":    userID,
		"OrderID":   item.OrderID,
		"ProductID": item.ProductID,
		"Price":     item.Price,
	}
	return execOne(ctx, or.session, q, args)
}

func (or *OrderItemsRepository) CustomerList(ctx context.Context, orderID string, userID int32) ([]OwnedOrderItemResponse, error) {
	const q = `SELECT o.order_id, o.product_id, o.items_total, o.price, o.is_packed, p.name
		FROM order_items AS o
			JOIN products AS p ON o.product_id = p.id
		WHERE o.order_id = $1::UUID AND user_id = $2`
	return list[OwnedOrderItemResponse](ctx, or.session, q, orderID, userID)
}

func (or *OrderItemsRepository) VendorList(ctx context.Context, orderID string, vendorID int32) ([]OwnedOrderItemResponse, error) {
	const q = `SELECT o.order_id, o.product_id, o.items_total, o.price, o.is_packed, p.name 
		FROM order_items AS o
			JOIN products AS p ON o.product_id = p.id AND p.user_id = $2
		WHERE o.order_id = $1`
	return list[OwnedOrderItemResponse](ctx, or.session, q, orderID, vendorID)
}

func (or *OrderItemsRepository) Delete(ctx context.Context, orderID string, userID int32) error {
	const q = `DELETE FROM order_items WHERE order_id = $1::UUID AND user_id = $2`
	return execOne(ctx, or.session, q, orderID, userID)
}

func (or *OrderItemsRepository) SetItemsTotal(ctx context.Context, orderID string, userID int32, itemsTotal int32) error {
	const q = `UPDATE order_items SET items_total = @ItemsTotal
		WHERE order_id = @OrderID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"OrderID":    orderID,
		"UserID":     userID,
		"itemsTotal": itemsTotal,
	}
	return execOne(ctx, or.session, q, args)
}

func (or *OrderItemsRepository) SetPacked(ctx context.Context, vendorID int32, item *OrderItemPackStatusRequest) error {
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
	return execOne(ctx, or.session, q, args)
}
