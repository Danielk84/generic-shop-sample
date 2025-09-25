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

type User struct {
	ID             int32          `json:"id"`
	Username       string         `json:"username"`
	Email          *string        `json:"email"`
	PasswordHash   *string        `json:"-"`
	PermissionType PermissionType `json:"permission-type"`
	IsActive       bool           `json:"is_active"`
}

type UserRepository struct {
	session db.Session
}

type UserStore interface {
	IsUsernameExists(ctx context.Context, username string) bool
	Create(ctx context.Context, user *User) error
	Get(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int32) error
	SetEmail(ctx context.Context, user *User) error
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

func (ur *UserRepository) Create(ctx context.Context, user *User) error {
	const q = `INSERT INTO users (username, password_hash, permission_type, is_active)
		VALUES (@Username, @PasswordHash, @PermissionType, @IsActive)`
	args := pgx.NamedArgs{
		"Username":       user.Username,
		"PasswordHash":   user.PasswordHash,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	return execOne(ctx, ur.session, q, args)
}

func (ur *UserRepository) Get(ctx context.Context, username string) (*User, error) {
	const readUserQuery = `SELECT id, username, email, password_hash, permission_type, is_active FROM users
	WHERE username = $1
	LIMIT 1`
	return get[User](ctx, ur.session, readUserQuery, username)
}

func (ur *UserRepository) Update(ctx context.Context, user *User) error {
	const q = `UPDATE users
		SET permission_type = @PermissionType, is_active = @IsActive
		WHERE id = @ID`
	args := pgx.NamedArgs{
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
		"ID":             user.ID,
	}
	return execOne(ctx, ur.session, q, args)
}

func (ur *UserRepository) Delete(ctx context.Context, id int32) error {
	const q = `DELETE FROM users WHERE id = $1`
	return execOne(ctx, ur.session, q, id)
}

func (ur *UserRepository) SetEmail(ctx context.Context, user *User) error {
	const q = `UPDATE ONLY users SET email = @Email WHERE id = @ID`
	args := pgx.NamedArgs{
		"Email": user.Email,
		"ID":    user.ID,
	}
	return execOne(ctx, ur.session, q, args)
}
