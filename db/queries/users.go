package queries

import (
	"context"
	"generic-shop-sample/db"
	"time"

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
	Email string `json:"email" binding:"required,email,min=10,max=256"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=128"`
	Password string `json:"password" binding:"required,password"`
}

type UserPermissionRequest struct {
	PermissionType PermissionType `json:"permission_type" binding:"number,gte=0,lt=4"`
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

type UserDetailsResponse struct {
	// users table
	UserResponse
	Email string `json:"email"`

	// user_profile table
	ImgPath     string `json:"img_path"`
	Birthday    string `json:"birthday"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
}

type UserRepository struct {
	session db.Session
}

type UserStore interface {
	IsUsernameExists(ctx context.Context, username string) bool
	Create(ctx context.Context, user *CreateUserRequest) error
	List(ctx context.Context, pagination, page int) ([]UserResponse, error)
	Get(ctx context.Context, username string) (*UserResponse, error)
	GetDetails(ctx context.Context, username string) (*UserDetailsResponse, error)
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
	const q = `WITH create_user_cte AS (
		INSERT INTO users (username, password, permission_type, is_active)
			VALUES (@Username, @Password, @PermissionType, @IsActive)
			RETURNING id
	)
	INSERT INTO user_profile(user_id) SELECT c.id FROM create_user_cte c`
	args := pgx.NamedArgs{
		"Username":       user.Username,
		"Password":       user.Password,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	return execOne(ctx, ur.session, q, args)
}

func (ur *UserRepository) List(ctx context.Context, pagination, page int) ([]UserResponse, error) {
	const q = `SELECT id, username, '' as password, permission_type, is_active FROM users
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

func (ur *UserRepository) GetDetails(ctx context.Context, username string) (*UserDetailsResponse, error) {
	const q = `SELECT u.id, u.username, COALESCE(u.email, '') as email, '' as password, u.permission_type, u.is_active,
		COALESCE(up.img_path, '') as img_path, COALESCE(up.birthday::TEXT, '') as birthday,
		COALESCE(up.phone_number, '') as phone_number, COALESCE(up.bio, '') as bio
	FROM users AS u LEFT JOIN user_profile AS up ON u.id = up.user_id
	WHERE username = $1
	LIMIT 1`
	return get[UserDetailsResponse](ctx, ur.session, q, username)
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
	const q = `UPDATE ONLY users SET email = $1 WHERE id = $2`
	return execOne(ctx, ur.session, q, email.Email, id)
}

type UserProfileRequest struct {
	Birthday time.Time `json:"age"`
	Bio      string    `json:"bio"`
}

type PhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type UserProfileRepository struct {
	session db.Session
}

type UserProfileStore interface {
	Upsert(ctx context.Context, userID int32, userProfile *UserProfileRequest) error
	GetImgPath(ctx context.Context, userID int32) (string, error)
	SetImgPath(ctx context.Context, userID int32, imgPath string) error
	SetPhoneNumber(ctx context.Context, userID int32, phoneNumber *PhoneNumberRequest) error
}

func NewUserProfileStore(session db.Session) UserProfileStore {
	return &UserProfileRepository{session}
}

func (upr *UserProfileRepository) Upsert(ctx context.Context, userID int32, userProfile *UserProfileRequest) error {
	const q = `INSERT INTO user_profile(user_id, birthday, bio) VALUES (@UserID, @Birthday, @Bio)
		ON CONFLICT(user_id)
		DO UPDATE SET birthday = @Birthday::DATE, bio = @Bio`
	args := pgx.NamedArgs{
		"UserID":   userID,
		"Birthday": userProfile.Birthday,
		"Bio":      userProfile.Bio,
	}
	return execOne(ctx, upr.session, q, args)
}

func (upr *UserProfileRepository) GetImgPath(ctx context.Context, userID int32) (string, error) {
	const q = `SELECT COALESCE(img_path, '') FROM user_profile WHERE user_id = $1`
	filepath, err := get[string](ctx, upr.session, q, userID)
	return *filepath, err
}

func (upr *UserProfileRepository) SetImgPath(ctx context.Context, userID int32, imgPath string) error {
	const q = `UPDATE user_profile SET img_path = NULLIF($1, '') WHERE user_id = $2`
	return execOne(ctx, upr.session, q, imgPath, userID)
}

func (upr *UserProfileRepository) SetPhoneNumber(ctx context.Context, userID int32, phoneNumber *PhoneNumberRequest) error {
	const q = `UPDATE user_profile SET phone_number = $1 WHERE user_id = $2`
	return execOne(ctx, upr.session, q, phoneNumber.PhoneNumber, userID)
}
