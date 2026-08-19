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

type EmailAddrRequest struct {
	Email string `json:"email" binding:"required,email,min=10,max=256"`
}

type PhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,iran_phone_number"`
}

type PermissionRequest struct {
	PermissionType PermissionType `json:"permission_type" binding:"number,gte=0,lt=4"`
}

type ValidUserRequest struct {
	EmailAddrRequest
	PermissionRequest
	ID string
}

type UserPermissionRequest struct {
	PermissionRequest
	IsActive bool `json:"is_active" binding:"boolean"`
}

type CreateUserRequest struct {
	EmailAddrRequest
	UserPermissionRequest
}

type UserInfoRequest struct {
	FirstName    string `json:"first_name" binding:"required,min=1,max=50"`
	LastName     string `json:"last_name" binding:"required,min=1,max=60"`
	NationalCode string `json:"national_code" binding:"required,length=10"`
}

type UserResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PermissionType int32  `json:"permission_type"`
	IsActive       bool   `json:"is_active"`
	IsVerified     string `json:"is_verified"`
}

type UserInfoResponse struct {
	UserResponse
	Email          string `json:"email"`
	IsVEmail       bool   `json:"is_v_email"`
	PhoneNumber    string `json:"phone_number"`
	IsVPhoneNumber bool   `json:"is_v_phone_number"`
	NationalCode   string `json:"national_code"`
}

type userRepository struct {
	session database.Session
	log     logger.Logger
}

type UserStore interface {
	IsUserExists(ctx context.Context, email string) bool
	IsValidUser(ctx context.Context, user ValidUserRequest) bool

	Create(ctx context.Context, user CreateUserRequest) error
	List(ctx context.Context, pagination, page int) ([]UserResponse, error)
	Get(ctx context.Context, email string) (UserInfoResponse, error)
	Delete(ctx context.Context, id, email string) error

	UpdatePermission(ctx context.Context, id string, user UserPermissionRequest) error

	SetInfo(ctx context.Context, id string, info UserInfoRequest) error
	VerifyUser(ctx context.Context, id string, isVerified bool) error

	SetEmail(ctx context.Context, id string, email EmailAddrRequest) error
	VerifyEmail(ctx context.Context, id string, isVerified bool) error

	SetPhoneNumber(ctx context.Context, id string, phoneNumber PhoneNumberRequest) error
	VerifyPhoneNumber(ctx context.Context, id string, isVerified bool) error
}

func NewUserStore(session database.Session, log logger.Logger) UserStore {
	return &userRepository{session, log}
}

func (u *userRepository) IsUserExists(ctx context.Context, email string) bool {
	const q = `SELECT EXISTS(
		SELECT 1
		FROM user_s.users
		WHERE email = $1)`
	var isExists bool
	if err := u.session.QueryRow(ctx, q, email).Scan(&isExists); err != nil || !isExists {
		u.log.Debug("UserRepository.IsUserExists", "error", err)
		return false
	}
	return true
}

func (u *userRepository) IsValidUser(ctx context.Context, user ValidUserRequest) bool {
	const q = `SELECT EXISTS(
		SELECT 1
		FROM user_s.users
		WHERE
			id = @ID::UUID AND
			email = @Email AND
			is_v_email = TRUE AND
			permission_type = @PermissionType AND
			is_active = TRUE)`
	args := pgx.NamedArgs{
		"ID":             user.ID,
		"Email":          user.Email,
		"PermissionType": user.PermissionType,
	}
	var isExists bool
	if err := u.session.QueryRow(ctx, q, args).Scan(&isExists); err != nil || !isExists {
		if err != nil {
			u.log.Debug("UserRepository.IsValidUser", "error", err)
		}
		return false
	}
	return true
}

func (u *userRepository) Create(ctx context.Context, user CreateUserRequest) (err error) {
	const q = `INSERT INTO user_s.users (email, permission_type, is_active)
			VALUES (@Email, @PermissionType, @IsActive)
			RETURNING id`
	args := pgx.NamedArgs{
		"Email":          user.Email,
		"PermissionType": user.PermissionType,
		"IsActive":       user.IsActive,
	}
	if err = execOne(ctx, u.session, q, args); err != nil {
		u.log.Debug("UserRepository.Create", "error", err)
	} else {
		u.log.Info("UserRepository.Create", "email", user.Email)
	}
	return
}

func (u *userRepository) List(ctx context.Context, pagination, page int) (items []UserResponse, err error) {
	const q = `SELECT
			id, (first_name || ' ' || last_name) as name
			permission_type, is_active, is_verified
		FROM user_s.users
		ORDER BY is_active DESC, is_verified
		LIMIT $1
		OFFSET $2`
	items, err = list[UserResponse](ctx, u.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		u.log.Debug("UserRepository.List", "error", err)
	}
	return
}

func (u *userRepository) Get(ctx context.Context, id string) (item UserInfoResponse, err error) {
	const q = `SELECT
			id, (first_name || ' ' || last_name) as name,
			permission_type, is_active, is_verified,
			COALESCE(email, ''), is_v_email,
			COALESCE(phone_number, ''), is_v_phone_number,
			national_code
		FROM user_s.users
		WHERE id = $1::UUID
		LIMIT 1`
	item, err = get[UserInfoResponse](ctx, u.session, q, id)
	if err != nil {
		u.log.Debug("UserRepository.Get", "error", err)
	}
	return
}

func (u *userRepository) Delete(ctx context.Context, id, email string) (err error) {
	const q = `WITH remove_shop AS (
			DELETE FROM user_s.shop WHERE user_id = $1::UUID
		), delete_related_comments AS (
			UPDATE user_s.comments SET name = NULL WHERE user_id = $1::UUID
		)
		UPDATE user_s.users
        SET email = NULL, phone_number = NULL, is_active = FALSE
        WHERE id = $1::UUID AND email = $2`
	if err = execOne(ctx, u.session, q, id, email); err != nil {
		u.log.Warn("UserRepository.Delete", "error", err)
	}
	return
}

func (u *userRepository) UpdatePermission(ctx context.Context, id string, user UserPermissionRequest) (err error) {
	const q = `UPDATE user_s.users
		SET permission_type = @PermissionType, is_active = @IsActive
		WHERE id = @ID::UUID`
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

func (u *userRepository) SetInfo(ctx context.Context, id string, info UserInfoRequest) (err error) {
	const q = `UPDATE user_s.users
		SET
			first_name = @FirstName,
			last_name = @LastName,
			national_code = @NationalCode,
			is_verified = FALSE
		WHERE id = @ID::UUID`
	args := pgx.NamedArgs{
		"FirstName":    info.FirstName,
		"LastName":     info.LastName,
		"NationalCode": info.NationalCode,
		"ID":           id,
	}
	if err = execOne(ctx, u.session, q, args); err != nil {
		u.log.Error("UserRepository.SetInfo", "error", err)
	}
	return
}

func (u *userRepository) VerifyUser(ctx context.Context, id string, isVerified bool) (err error) {
	const q = `UPDATE user_s.users
		SET is_verified = $2
		WHERE
			id = $1::UUID AND
			first_name != '' AND
			last_name != '' AND
			national_code != ''`
	return
}

func (u *userRepository) SetEmail(ctx context.Context, id string, email EmailAddrRequest) (err error) {
	const q = `UPDATE user_s.users
		SET email = NULLIF($1, ''), is_v_email = FALSE
		WHERE id = $2::UUID`
	if err = execOne(ctx, u.session, q, email.Email, id); err != nil {
		u.log.Debug("UserRepository.SetEmail", "error", err)
	}
	return
}

func (u *userRepository) VerifyEmail(ctx context.Context, id string, isVerified bool) (err error) {
	const q = `UPDATE user_s.users
		SET is_v_email = $1
		WHERE id = $2::UUID AND email IS NOT NULL`
	if err = execOne(ctx, u.session, q, isVerified, id); err != nil {
		u.log.Warn("UserRepository.VerifyEmail", "error", err)
	}
	return
}

func (u *userRepository) SetPhoneNumber(ctx context.Context, id string, phoneNumber PhoneNumberRequest) (err error) {
	const q = `UPDATE user_s.users
		SET phone_number = NULLIF($1, ''), is_v_phone_number = FALSE
		WHERE id = $2::UUID`
	if err = execOne(ctx, u.session, q, phoneNumber.PhoneNumber, id); err != nil {
		u.log.Debug("UserRepository.SetPhoneNumber", "error", err)
	}
	return
}

func (u *userRepository) VerifyPhoneNumber(ctx context.Context, id string, isVerified bool) (err error) {
	const q = `UPDATE user_s.users
		SET is_v_phone_number = $1
		WHERE id = $2 AND phone_number IS NOT NULL`
	if err = execOne(ctx, u.session, q, isVerified, id); err != nil {
		u.log.Debug("UserRepository.VerifyPhoneNumber", "error", err)
	}
	return
}

type ShopPhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,max=15"`
}

type UpsertShopRequest struct {
	Brand    string `json:"brand" binding:"required,min=2,max=100"`
	ShopAddr string `json:"shop_addr" binding:"required,max=1000"`
	ZipCode  string `json:"zip_code" binding:"required,max=10"`

	BusinessCode string `json:"business_code" binding:"required,max=100"`

	Bio string `json:"bio" binding:"required,max=650"`
}

type ShopInfoResponse struct {
	// user_s.users table
	UserResponse
	Email          string `json:"email"`
	IsVEmail       bool   `json:"is_v_email"`
	PhoneNumber    string `json:"phone_number"`
	IsVPhoneNumber bool   `json:"is_v_phone_number"`

	// user_s.shop table
	Brand    string `json:"brand"`
	ShopAddr string `json:"shop_addr"`
	ZipCode  string `json:"zip_code"`

	BusinessCode    string `json:"business_code"`
	ShopPhoneNumber string `json:"shop_phone_number"`

	ImgPath string `json:"img_path"`
	Bio     string `json:"bio"`
	IsShop  bool   `json:"is_shop"`
}

type ShopResponse struct {
	UserID      string `json:"user_id"`
	Brand       string `json:"brand"`
	PhoneNumber string `json:"phone_number"`
	IsShop      bool   `json:"is_shop"`
}

type shopRepository struct {
	session database.Session
	log     logger.Logger
}

type ShopStore interface {
	IsValidShop(ctx context.Context, userID string) bool
	Upsert(ctx context.Context, userID string, info UpsertShopRequest) error
	Get(ctx context.Context, userID string) (item ShopInfoResponse, err error)
	List(ctx context.Context, pagination, page int) ([]ShopResponse, error)
	SetPhoneNumber(ctx context.Context, userID string, phoneNumber ShopPhoneNumberRequest) error
	VerifyPhoneNumber(ctx context.Context, id string, isVerified bool) error
	GetImgPath(ctx context.Context, userID string) (string, error)
	SetImgPath(ctx context.Context, userID string, imgPath string) error
	DeleteImgPath(ctx context.Context, userID string) (string, error)
}

func NewShopStore(session database.Session, log logger.Logger) ShopStore {
	return &shopRepository{session, log}
}

func (s *shopRepository) IsValidShop(ctx context.Context, userID string) bool {
	const q = `SELECT EXISTS(
		SELECT 1
		FROM user_s.shop
		WHERE user_id = $1::UUID and is_shop = TRUE)`
	var isExists bool
	if err := s.session.QueryRow(ctx, q, userID).Scan(&isExists); err != nil || !isExists {
		if err != nil {
			s.log.Debug("ShopRepository.IsValidShop", "error", err.Error())
		}
		return false
	}
	return true
}

func (s *shopRepository) Upsert(ctx context.Context, userID string, info UpsertShopRequest) (err error) {
	const q = `INSERT INTO user_s.shop(user_id, brand, shop_addr, zip_code, business_code, bio)
        VALUES (
			@UserID::UUID, 
			NULLIF(@Brand, ''),
			@ShopCode,
			@ZipCode,
			@BusinessCode,
			@Bio)
		ON CONFLICT(user_id)
		DO UPDATE SET
			brand = NULLIF(@Brand, ''),
			shop_addr = @ShopCode,
			zip_code = @ZipCode,
			business_code = @BusinessCode,
			bio = @Bio,
			is_verified = FALSE`
	args := pgx.NamedArgs{
		"UserID":       userID,
		"Brand":        info.Brand,
		"ShopCode":     info.ShopAddr,
		"ZipCode":      info.ZipCode,
		"BusinessCode": info.BusinessCode,
		"Bio":          info.Bio,
	}
	if err = execOne(ctx, s.session, q, args); err != nil {
		s.log.Debug("ShopRepository.Upsert")
	}
	return
}

func (s *shopRepository) Get(ctx context.Context, userID string) (item ShopInfoResponse, err error) {
	const q = `SELECT
			u.id, (u.first_name || ' ' || u.last_name) as name
			u.permission_type, u.is_active, u.is_verified,
			COALESCE(u.email, ''), u.is_v_email,
			COALESCE(u.phone_numner, ''), u.is_v_phone_number,

			COALESCE(s.brand, ''), s.shop_addr, s.zip_code,
			s.business_code, COALESCE(s.phone_number, '') as shop_phone_number
			s.img_path, s.bio
		FROM user_s.users as u LEFT JOIN user_s.shop as s on u.id = s.user_id
		WHERE user_id = $1::UUID`
	if item, err = get[ShopInfoResponse](ctx, s.session, q, userID); err != nil {
		s.log.Debug("ShopRepository.Get", "error", err)
	}
	return
}

func (s *shopRepository) List(ctx context.Context, pagination, page int) (items []ShopResponse, err error) {
	const q = `SELECT user_id, COALESCE(brand, ''), COALESCE(phone_number, ''), is_shop
		FROM user_s.shop
		ORDER is_verified
		LIMIT $1
		OFFSET $2`
	items, err = list[ShopResponse](ctx, s.session, q, pagination, getOffsetFromPageNum(pagination, page))
	if err != nil {
		s.log.Debug("ShopRepository.List", "error", err)
	}
	return
}

func (s *shopRepository) SetPhoneNumber(ctx context.Context, userID string, phoneNumber ShopPhoneNumberRequest) (err error) {
	const q = `UPDATE user_s.shop
		SET phone_numner = NULLIF($1, ''), is_v_phone_number = FALSE
		WHERE user_id = $2::UUID`
	if err = execOne(ctx, s.session, q, phoneNumber.PhoneNumber, userID); err != nil {
		s.log.Debug("ShopRepository.SetPhoneNumber", "error", err)
	}
	return
}

func (u *shopRepository) VerifyPhoneNumber(ctx context.Context, id string, isVerified bool) (err error) {
	const q = `UPDATE user_s.shop
		SET is_v_phone_number = $1
		WHERE id = $2 AND phone_number IS NOT NULL`
	if err = execOne(ctx, u.session, q, isVerified, id); err != nil {
		u.log.Debug("ShopRepository.VerifyPhoneNumber", "error", err)
	}
	return
}

func (s *shopRepository) GetImgPath(ctx context.Context, userID string) (imgPath string, err error) {
	const q = `SELECT img_path
		FROM user_s.shop
		WHERE user_id = $1::UUID`
	if err = s.session.QueryRow(ctx, q, userID).Scan(&imgPath); err != nil {
		s.log.Debug("ShopRepository.GetImgPath", "error", err)
	}
	return
}

func (s *shopRepository) SetImgPath(ctx context.Context, userID string, imgPath string) (err error) {
	const q = `UPDATE user_s.shop
		SET img_path = $1
		WHERE user_id = $2::UUID`
	if err = execOne(ctx, s.session, q, imgPath, userID); err != nil {
		s.log.Debug("ShopRepository.SetImgPath", "error", err, "imgPath", imgPath)
	}
	return
}

func (s *shopRepository) DeleteImgPath(ctx context.Context, UserID string) (imgPath string, err error) {
	const q = `UPDATE user_s.shop
		SET img_path = ''
		WHERE user_id = $1::UUID
		RETURNING OLD.img_path`
	if err = s.session.QueryRow(ctx, q, UserID).Scan(&imgPath); err != nil {
		s.log.Error("ShopRepository.DeleteImgPath", "error", err)
	}
	return
}
