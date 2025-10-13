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
	UpdateAt   string `json:"update_at"`
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
	List(ctx context.Context, userID int32, pagination, page int) ([]OrderSummaryResponse, error)
	Get(ctx context.Context, id string, userID int32) (*OrderResponse, error)
	SetUserInfo(ctx context.Context, id string, userID int32, info *OrderUserInfoRequest) error
	VerifyUserInfo(ctx context.Context, id string, userID int32, isVerified bool) error
	DeleteExpiredOrders(ctx context.Context) error
}

func NewOrderStore(session db.Session) OrderStore {
	return &OrderRepository{session}
}

func (or *OrderRepository) Create(ctx context.Context, userID int32) (string, error) {
	const q = `INSERT INTO orders(user_id) VALUES ($1) RETURNING id`
	id := ""
	err := or.session.QueryRow(ctx, q, userID).Scan(&id)
	return id, err
}

func (or *OrderRepository) List(ctx context.Context, userID int32, pagination, page int) ([]OrderSummaryResponse, error) {
	const q = `SELECT id, updated_at, is_paid, is_deliverd FROM orders
		WHERE user_id = @UserID
		LIMIT @Limit
		OFFSET @Offset`
	args := pgx.NamedArgs{
		"UserID": userID,
		"Limit":  pagination,
		"Offset": getOffsetFromPageNum(pagination, page),
	}
	return list[OrderSummaryResponse](ctx, or.session, q, args)
}

func (or *OrderRepository) Get(ctx context.Context, id string, userID int32) (*OrderResponse, error) {
	const q = `SELECT id, updated_at, items_total, total_bill, is_paid, address, zip_code, is_deliverd FROM orders
		WHERE id = $1::UUID AND user_id = $2 AND is_confirmed = TRUE`
	return get[OrderResponse](ctx, or.session, q, id, userID)
}

func (or *OrderRepository) SetUserInfo(ctx context.Context, id string, userID int32, info *OrderUserInfoRequest) error {
	const q = `UPDATE orders SET address = @Address, zip_code = @ZipCode, is_confirmed = FALSE
		WHERE id = @ID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"Address": info.Address,
		"ZipCode": info.ZipCode,
		"ID":      id,
		"UserID":  userID,
	}
	return execOne(ctx, or.session, q, args)
}

func (or *OrderRepository) VerifyUserInfo(ctx context.Context, id string, userID int32, isVerified bool) error {
	const q = `UPDATE orders SET is_confirmed = @IsVerfied WHERE id = @ID AND user_id = @UserID`
	args := pgx.NamedArgs{
		"IsVerified": isVerified,
		"ID":         id,
		"UserID":     userID,
	}
	return execOne(ctx, or.session, q, args)
}

func (or *OrderRepository) DeleteExpiredOrders(ctx context.Context) error {
	const q = `DELETE FROM orders WHERE started_at > NOW() AND is_paid = FALSE AND payment_summary = NULL`
	_, err := or.session.Exec(ctx, q)
	return err
}
