package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"

	"github.com/jackc/pgx/v5"
)

type VendorOrderDelivere struct {
	UserID      string          `json:"user_id"`
	OrderID     string          `json:"order_id"`
	ProductID   string          `json:"product_id"`
	Property    ProductProperty `json:"property"`
	IsDelivered bool            `json:"is_delivered"`
}

type VendorOrder struct {
	VendorOrderDelivere
	Quantity  int32 `json:"quantity"`
	TotalBill int64 `json:"total_bill"`
}

type VendorOrderRepository struct {
	session database.Session
	log     logger.Logger
}

type VendorOrderStore interface {
	Create(ctx context.Context, order VendorOrder) error
	List(ctx context.Context, userID string, pagination, page int) ([]VendorOrder, error)
	SetIsDelivered(ctx context.Context, order VendorOrderDelivere) error
}

func NewVendorOrderStore(session database.Session, log logger.Logger) VendorOrderStore {
	return &VendorOrderRepository{session, log}
}

func (v *VendorOrderRepository) Create(ctx context.Context, order VendorOrder) (err error) {
	const q = `INSERT INTO order_s.vendors_order(
			user_id, order_id, product_id,
			property, quantity, total_bill, is_delivered)
		VALUES (
			'@UserID'::UUID, '@OrderID'::UUID, '@ProductID'::UUID,
			'@Property'::JSONB, @Quantity, @TotalBill, @IsDelivered)`
	args := pgx.NamedArgs{
		"UserID":      order.UserID,
		"OrderID":     order.OrderID,
		"ProductId":   order.ProductID,
		"Property":    order.Property,
		"Quantity":    order.Quantity,
		"TotalBill":   order.TotalBill,
		"IsDelivered": order.IsDelivered,
	}
	if err = execOne(ctx, v.session, q, args); err != nil {
		v.log.Debug("VendorOrderRepository.Create", "error", err)
	}
	return
}

func (v *VendorOrderRepository) List(ctx context.Context, userID string, pagination, page int) (items []VendorOrder, err error) {
	const q = `SELECT
			user_id, order_id, product_id,
			property, quantity, total_bill, is_delivered)
		FROM order_s.vendors_order
		WHERE order_id = '@UserID'::UUID
		ORDER is_delivered DESC
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"UserID": userID,
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	if items, err = list[VendorOrder](ctx, v.session, q, args); err != nil {
		v.log.Debug("VendorOrderRepository.List", "error", err)
	}
	return
}

func (v *VendorOrderRepository) SetIsDelivered(ctx context.Context, order VendorOrderDelivere) (err error) {
	const q = `UPDATE order_s.vendors_order
		SET is_delivered = @IsDelivered
		WHERE
			user_id = '@UserID'::UUID AND
			order_id = '@OrderID'::UUID AND
			product_id = '@ProductID'::UUID AND
			property = '@Property'::JSONB`
	args := pgx.NamedArgs{
		"IsDelivered": order.IsDelivered,
		"UserID":      order.UserID,
		"OrderID":     order.OrderID,
		"ProductID":   order.ProductID,
		"Property":    order.Property,
	}
	if err = execOne(ctx, v.session, q, args); err != nil {
		v.log.Debug("VendorOrderRepository.SetIsDelivered", "error", err)
	}
	return
}
