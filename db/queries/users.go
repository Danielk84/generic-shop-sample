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

const insertQuery = `INSERT INTO users (username, password_hash, permission_type, is_active)
VALUES (@Username, @PasswordHash, @PermissionType, @IsActive)`

func (um *UserManager) Create(ctx context.Context, user *User) error {
	args := pgx.NamedArgs{
		"Username":       user.Username,
		"PasswordHash":   user.PasswordHash,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	cTag, err := um.session.Exec(ctx, insertQuery, args)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return ErrNoRowInserted
	}
	return nil
}

const addOrUpdateEmailQuery = `UPDATE ONLY users SET email = @Email WHERE id = @ID`

func (um *UserManager) SetEmail(ctx context.Context, user *User) error {
	args := pgx.NamedArgs{
		"Email": user.Email,
		"ID":    user.ID,
	}
	cTag, err := um.session.Exec(ctx, addOrUpdateEmailQuery, args)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return ErrNoRowFoundToUpdate
	}
	return nil
}

const readUserQuery = `SELECT id, username, email::VARCHAR, password_hash::VARCHAR, permission_type, is_active FROM users
WHERE username = $1
LIMIT 1`

func (um *UserManager) Read(ctx context.Context, username string) (*User, error) {
	rows, err := um.session.Query(ctx, readUserQuery, username)
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, ErrNoRowFound
	}
	return &users[0], nil
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
	cTag, err := um.session.Exec(ctx, updateUserQuery, args)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return ErrNoRowFoundToDelete
	}
	return nil
}

const deleteUserQuery = `DELETE FROM users WHERE id = $1`

func (um *UserManager) Delete(ctx context.Context, id int32) error {
	cTag, err := um.session.Exec(ctx, deleteUserQuery, id)
	if err != nil {
		return err
	}
	if cTag.RowsAffected() != 1 {
		return ErrNoRowFoundToDelete
	}
	return nil
}
