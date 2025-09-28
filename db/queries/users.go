package queries

import (
	"context"
	"generic-shop-sample/db"

	"github.com/jackc/pgx/v5"
)

type PermissionType = int32

const (
	Admin PermissionType = iota
	Vendor
	Customer
	BlockUser
)

type EmailAddrRequest struct {
	Email string `json:"email" binding:"required,min=10,max=256,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=4,max=128,alphanum"`
	Password string `json:"password" binding:"required,min=8,max=32,ascii"`
}

type UserPermissionRequest struct {
	PermissionType PermissionType `json:"permission_type" binding:"required,number,gt=0,lte=4"`
	IsActive       bool           `json:"is_active" binding:"required,boolean"`
}

type CreateUserRequest struct {
	LoginRequest
	UserPermissionRequest
}

type UserResponse struct {
	ID             int32  `json:"id"`
	Username       string `json:"username"`
	Password       string `json:"-"`
	PermissionType int32  `json:"permission_type"`
	IsActive       bool   `json:"is_active"`
}

type UserRepository struct {
	session db.Session
}

type UserStore interface {
	IsUsernameExists(ctx context.Context, username string) bool
	Create(ctx context.Context, user *CreateUserRequest) error
	List(ctx context.Context, pagination, page int) ([]UserResponse, error)
	Get(ctx context.Context, username string) (*UserResponse, error)
	UpdatePermission(ctx context.Context, id int32, user *UserPermissionRequest) error
	Delete(ctx context.Context, id int32) error
	SetEmail(ctx context.Context, id int32, email *EmailAddrRequest) error
}

func NewUserStore(session db.Session) UserStore {
	return &UserRepository{session}
}

func (ur *UserRepository) IsUsernameExists(ctx context.Context, username string) bool {
	const q = "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	var isExists bool
	if err := ur.session.QueryRow(ctx, q, username).Scan(&isExists); err != nil || isExists {
		return true
	}
	return false
}

func (ur *UserRepository) Create(ctx context.Context, user *CreateUserRequest) error {
	const q = `INSERT INTO users (username, password, permission_type, is_active)
		VALUES (@Username, @Password, @PermissionType, @IsActive)`

	args := pgx.NamedArgs{
		"Username":       user.Username,
		"Password":       user.Password,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	return execOne(ctx, ur.session, q, args)
}

func (ur *UserRepository) List(ctx context.Context, pagination, page int) ([]UserResponse, error) {
	const q = `SELECT id, username, password, permission_type, is_active FROM users
	ORDER BY is_active DESC
	LIMIT $1
	OFFSET $2`
	return list[UserResponse](ctx, ur.session, q, pagination, (page-1)*pagination)
}

func (ur *UserRepository) Get(ctx context.Context, username string) (*UserResponse, error) {
	const q = `SELECT id, username, password, permission_type, is_active FROM users
	WHERE username = $1
	LIMIT 1`
	return get[UserResponse](ctx, ur.session, q, username)
}

func (ur *UserRepository) UpdatePermission(ctx context.Context, id int32, user *UserPermissionRequest) error {
	const q = `UPDATE users
		SET permission_type = @PermissionType, is_active = @IsActive
		WHERE id = @ID`
	args := pgx.NamedArgs{
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
		"ID":             id,
	}
	return execOne(ctx, ur.session, q, args)
}

func (ur *UserRepository) Delete(ctx context.Context, id int32) error {
	const q = `DELETE FROM users WHERE id = $1`
	return execOne(ctx, ur.session, q, id)
}

func (ur *UserRepository) SetEmail(ctx context.Context, id int32, email *EmailAddrRequest) error {
	const q = `UPDATE ONLY users SET email = @Email WHERE id = @ID`
	args := pgx.NamedArgs{
		"Email": email.Email,
		"ID":    id,
	}
	return execOne(ctx, ur.session, q, args)
}
