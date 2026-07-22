package queries

import (
	"context"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"

	"github.com/jackc/pgx/v5"
)

type PermissionType = int32

const (
	Admin PermissionType = iota
	Vendor
	Customer
	BlockUser
)

type ValidUserRequest struct {
	ID             string
	Username       string
	PermissionType PermissionType
}

type EmailAddrRequest struct {
	Email string `json:"email" binding:"required,email,min=10,max=256"`
}

type PhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,iran_phone_number"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required,password"`
}

type UserPermissionRequest struct {
	PermissionType PermissionType `json:"permission_type" binding:"number,gte=0,lt=4"`
	IsActive       bool           `json:"is_active" binding:"boolean"`
}

type CreateUserRequest struct {
	LoginRequest
	UserPermissionRequest
}

type UserResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Password       string `json:"-"`
	PermissionType int32  `json:"permission_type"`
	IsActive       bool   `json:"is_active"`
}

type UserDetailsResponse struct {
	// users table
	UserResponse
	Email          string `json:"email"`
	IsVEmail       bool   `json:"is_v_email"`
	PhoneNumber    string `json:"phone_number"`
	IsVPhoneNumber bool   `json:"is_v_phone_number"`

	// user_profile table
	ImgPath  string `json:"img_path"`
	Birthday string `json:"birthday"`
	Bio      string `json:"bio"`
}

type userRepository struct {
	session database.Session
	log     logger.Logger
}

type UserStore interface {
	IsUsernameExists(ctx context.Context, username string) bool
	IsValidUser(ctx context.Context, user ValidUserRequest) bool
	Create(ctx context.Context, user CreateUserRequest) error
	List(ctx context.Context, pagination, page int) ([]UserResponse, error)
	Get(ctx context.Context, username string) (UserResponse, error)
	GetDetails(ctx context.Context, username string) (UserDetailsResponse, error)
	UpdatePermission(ctx context.Context, id string, user UserPermissionRequest) error
	Delete(ctx context.Context, id string, username string) error
	SetEmail(ctx context.Context, id string, email EmailAddrRequest) error
	VerifyEmail(ctx context.Context, id string, isVerified bool) error
	SetPhoneNumber(ctx context.Context, id string, phoneNumber PhoneNumberRequest) error
	VerifyPhoneNumber(ctx context.Context, id string, isVerified bool) error
}

func NewUserStore(session database.Session, log logger.Logger) UserStore {
	return &userRepository{session, log}
}

func (u *userRepository) IsUsernameExists(ctx context.Context, username string) bool {
	const q = "SELECT EXISTS(SELECT 1 FROM user_s.users WHERE username = $1)"
	var isExists bool
	if err := u.session.QueryRow(ctx, q, username).Scan(&isExists); err != nil || isExists {
		if err != nil {
			u.log.Debug("UserRepository.IsUsernameExists", "error", err.Error())
		}
		return true
	}
	return false
}

func (u *userRepository) IsValidUser(ctx context.Context, user ValidUserRequest) bool {
	const q = `SELECT EXISTS(
		SELECT 1
		FROM user_s.users
		WHERE id = '@ID'::UUID AND username = @Username AND permission_type = @PermissionType AND is_active = true)`
	args := pgx.NamedArgs{
		"ID":             user.ID,
		"Username":       user.Username,
		"PermissionType": user.PermissionType,
	}
	var isExists bool
	if err := u.session.QueryRow(ctx, q, args).Scan(&isExists); err != nil || !isExists {
		if err != nil {
			u.log.Debug("UserRepository.IsValidUser", "error", err.Error())
		}
		return false
	}
	return true
}

func (u *userRepository) Create(ctx context.Context, user CreateUserRequest) (err error) {
	const q = `WITH create_user_cte AS (
		INSERT INTO user_s.users (username, password, permission_type, is_active)
			VALUES (@Username, @Password, @PermissionType, @IsActive)
			RETURNING id
		)
		INSERT INTO user_s.user_profile(user_id)
			SELECT c.id FROM create_user_cte c`
	args := pgx.NamedArgs{
		"Username":       user.Username,
		"Password":       user.Password,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	if err = execOne(ctx, u.session, q, args); err != nil {
		u.log.Debug("UserRepository.Create", "error", err)
	} else {
		u.log.Info("UserRepository.Create", "username", user.Username)
	}
	return
}

func (u *userRepository) List(ctx context.Context, pagination, page int) (items []UserResponse, err error) {
	const q = `SELECT
			id,
			COALESCE(username, '') AS username,
			'' as password,
			permission_type, is_active
		FROM user_s.users
		ORDER BY is_active DESC, username NULLS LAST
		LIMIT $1
		OFFSET $2`
	items, err = list[UserResponse](ctx, u.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		u.log.Debug("UserRepository.List", "error", err)
	}
	return
}

func (u *userRepository) Get(ctx context.Context, username string) (item UserResponse, err error) {
	const q = `SELECT id, username, password, permission_type, is_active
		FROM user_s.users
		WHERE username = $1
		LIMIT 1`
	item, err = get[UserResponse](ctx, u.session, q, username)
	if err != nil {
		u.log.Debug("UserRepository.Get", "error", err)
	}
	return
}

func (u *userRepository) GetDetails(ctx context.Context, username string) (item UserDetailsResponse, err error) {
	if username == "" || username == "deleted" {
		err = pgx.ErrNoRows
		return
	}
	const q = `SELECT
			u.id, u.username,
			COALESCE(u.email, '') as email, u.is_v_email, '' as password,
			    u.permission_type, u.is_active, COALESCE(u.phone_number, '') as phone_number,
			u.is_v_phone_number, COALESCE(up.img_path, '') as img_path,
			COALESCE(up.birthday::TEXT, '') as birthday, COALESCE(up.bio, '') as bio
		FROM user_s.users AS u LEFT JOIN user_s.user_profile AS up ON u.id = up.user_id
		WHERE username = $1
		LIMIT 1`
	item, err = get[UserDetailsResponse](ctx, u.session, q, username)
	if err != nil {
		u.log.Debug("UserRepository.GetDetails", "error", err)
	}
	return
}

func (u *userRepository) UpdatePermission(ctx context.Context, id string, user UserPermissionRequest) (err error) {
	const q = `UPDATE user_s.users
		SET permission_type = @PermissionType, is_active = @IsActive
		WHERE id = '@ID'::UUID`
	args := pgx.NamedArgs{
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
		"ID":             id,
	}
	if err = execOne(ctx, u.session, q, args); err != nil {
		u.log.Error("UserRepository.UpdatePermission", "error", err)
	}
	return
}

func (u *userRepository) Delete(ctx context.Context, id string, username string) (err error) {
	const q = `WITH remove_products AS (
			DELETE FROM product_s.products WHERE user_id = $1
		), remove_user_profile AS (
			DELETE FROM user_s.user_profile WHERE user_id = $1
		), delete_related_comments_username AS (
			UPDATE user_s.comments SET username = NULL WHERE username = $2
		)
		UPDATE user_s.users
        SET username = NULL, password = NULL, is_active = FALSE
        WHERE id = '$1'::UUID`
	if err = execOne(ctx, u.session, q, id, username); err != nil {
		u.log.Warn("UserRepository.Delete", "error", err)
	}
	return
}

func (u *userRepository) SetEmail(ctx context.Context, id string, email EmailAddrRequest) (err error) {
	const q = `WITH remove_not_used_email AS (
			UPDATE user_s.users
			SET
				email = NULL,
				is_v_email = FALSE
				WHERE email = $1 AND username IS NULL
		)
		UPDATE user_s.users
		SET email = $1 WHERE id = '$2'::UUID`
	if err = execOne(ctx, u.session, q, email.Email, id); err != nil {
		u.log.Debug("UserRepository.SetEmail", "error", err)
	}
	return
}

func (u *userRepository) VerifyEmail(ctx context.Context, id string, isVerified bool) (err error) {
	const q = `UPDATE user_s.users SET is_v_email = $1 WHERE id = '$2'::UUID`
	if err = execOne(ctx, u.session, q, isVerified, id); err != nil {
		u.log.Warn("UserRepository.VerifyEmail", "error", err)
	}
	return
}

func (u *userRepository) SetPhoneNumber(ctx context.Context, id string, phoneNumber PhoneNumberRequest) (err error) {
	const q = `WITH remove_not_used_phone_number AS (
			UPDATE user_s.users
			SET
				phone_number = NULL,
				is_v_phone_number = FALSE
				WHERE phone_number = $1 AND username IS NULL
		)
		UPDATE user_s.users SET phone_number = $1 WHERE id = '$2'::UUID`
	if err = execOne(ctx, u.session, q, phoneNumber.PhoneNumber, id); err != nil {
		u.log.Debug("UserRepository.SetPhoneNumber", "error", err)
	}
	return
}

func (u *userRepository) VerifyPhoneNumber(ctx context.Context, id string, isVerified bool) (err error) {
	const q = `UPDATE user_s.users
		SET is_v_phone_number = $1
		WHERE id = $2 AND username IS NOT NULL`
	if err = execOne(ctx, u.session, q, isVerified, id); err != nil {
		u.log.Debug("UserRepository.VerifyPhoneNumber", "error", err)
	}
	return
}

type UserProfileRequest struct {
	Birthday string `json:"birthday" binding:"required,date"`
	Bio      string `json:"bio" binding:"required,max=450"`
}

type userProfileRepository struct {
	session database.Session
	log     logger.Logger
}

type UserProfileStore interface {
	Upsert(ctx context.Context, userID string, userProfile UserProfileRequest) error
	GetImgPath(ctx context.Context, userID string) (string, error)
	SetImgPath(ctx context.Context, userID string, imgPath string) error
	DeleteImgPath(ctx context.Context, userID string) (string, error)
}

func NewUserProfileStore(session database.Session, log logger.Logger) UserProfileStore {
	return &userProfileRepository{session, log}
}

func (u *userProfileRepository) Upsert(ctx context.Context, userID string, userProfile UserProfileRequest) (err error) {
	const q = `INSERT INTO user_s.user_profile(user_id, birthday, bio)
        VALUES ('@UserID'::UUID, '@Birthday'::DATE, @Bio)
		ON CONFLICT(user_id)
		DO UPDATE SET birthday = @Birthday::DATE, bio = @Bio`
	args := pgx.NamedArgs{
		"UserID":   userID,
		"Birthday": userProfile.Birthday,
		"Bio":      userProfile.Bio,
	}
	if err = execOne(ctx, u.session, q, args); err != nil {
		u.log.Debug("UserProfileRepository.Upsert")
	}
	return
}

func (u *userProfileRepository) GetImgPath(ctx context.Context, userID string) (imgPath string, err error) {
	const q = `SELECT COALESCE(img_path, '') AS img_path
		FROM user_s.user_profile
		WHERE user_id = '$1'::UUID`
	if err = u.session.QueryRow(ctx, q, userID).Scan(&imgPath); err != nil {
		u.log.Debug("UserProfileRepository.GetImgPath", "error", err)
	}
	return
}

func (u *userProfileRepository) SetImgPath(ctx context.Context, userID string, imgPath string) (err error) {
	const q = `UPDATE user_s.user_profile
		SET img_path = NULLIF($1, '')
		WHERE user_id = '$2'::UUID`
	if err = execOne(ctx, u.session, q, imgPath, userID); err != nil {
		u.log.Debug("UserProfileRepository.SetImgPath", "error", err, "imgPath", imgPath)
	}
	return
}

func (u *userProfileRepository) DeleteImgPath(ctx context.Context, UserID string) (imgPath string, err error) {
	const q = `UPDATE user_s.user_profile
		SET img_path = null
		WHERE user_id = '$1'::UUID
		RETURNING OLD.img_path AS img_path`
	if err = u.session.QueryRow(ctx, q, UserID).Scan(&imgPath); err != nil {
		u.log.Error("UserProfileRepository.DeleteImgPath", "error", err)
	}
	return
}
