package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionType int32

const (
	Admin PermissionType = iota
	Vendor
	Customer
	BlockUser
)

type User struct {
	ID             int32
	Username       string
	Email          *string
	PasswordHash   *string
	PermissionType PermissionType
	IsActive       bool
}

type IUserManager interface {
	IsUsernameExists(context.Context, string) bool
	Create(context.Context, *User) error
	SetEmail(context.Context, *User) error
	Read(context.Context, string) (*User, error)
	Update(context.Context, *User) error
	Delete(context.Context, int32) error
}

func NewUserManager(session *pgxpool.Pool) IUserManager {
	return &UserManager{session: session}
}

type UserManager struct {
	session *pgxpool.Pool
}

const isUsernameExistsQuery = "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

func (um *UserManager) IsUsernameExists(ctx context.Context, username string) bool {
	var isExists bool
	if err := um.session.QueryRow(ctx, isUsernameExistsQuery, username).Scan(&isExists); err != nil || isExists {
		return true
	}
	return false
}

const createUserQuery = `INSERT INTO users (username, password_hash, permission_type, is_active)
VALUES (@Username, @PasswordHash, @PermissionType, @IsActive)`

func (um *UserManager) Create(ctx context.Context, user *User) error {
	args := pgx.NamedArgs{
		"Username":       user.Username,
		"PasswordHash":   user.PasswordHash,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	return execQuery(ctx, um.session, createUserQuery, args)
}

const setEmailQuery = `UPDATE ONLY users SET email = @Email WHERE id = @ID`

func (um *UserManager) SetEmail(ctx context.Context, user *User) error {
	args := pgx.NamedArgs{
		"Email": user.Email,
		"ID":    user.ID,
	}
	return execQuery(ctx, um.session, setEmailQuery, args)
}

const readUserQuery = `SELECT id, username, email, password_hash, permission_type, is_active FROM users
WHERE username = $1
LIMIT 1`

func (um *UserManager) Read(ctx context.Context, username string) (*User, error) {
	return read[User](ctx, um.session, readUserQuery, username)
}

const updateUserQuery = `UPDATE users
SET password_hash = @PasswordHash, permission_type = @PermissionType, is_active = @IsActive
WHERE id = @ID`

func (um *UserManager) Update(ctx context.Context, user *User) error {
	args := pgx.NamedArgs{
		"PasswordHash":   user.PasswordHash,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
		"ID":             user.ID,
	}
	return execQuery(ctx, um.session, updateUserQuery, args)
}

const deleteUserQuery = `DELETE FROM users WHERE id = $1`

func (um *UserManager) Delete(ctx context.Context, id int32) error {
	return execQuery(ctx, um.session, deleteUserQuery, id)
}
