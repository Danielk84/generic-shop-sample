package api

import (
	"context"
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/background"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UsersRouter(router *gin.RouterGroup) {
	uh := usersHandler{
		store: queries.NewUserStore(db.NewSession()),
		cache: db.NewCache(db.UsersCache),
	}

	router.GET("/:username", uh.get)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodGet, "/", []gin.HandlerFunc{uh.list}},
		{http.MethodPost, "/", []gin.HandlerFunc{uh.createUserByAdmin}},
		{http.MethodDelete, "/", []gin.HandlerFunc{uh.delete}},
		{http.MethodPut, "/set-email", []gin.HandlerFunc{uh.setEmail}},
		{http.MethodPut, "/set-phone-number", []gin.HandlerFunc{uh.setPhoneNumber}},
		{http.MethodPost, "/verify-email", []gin.HandlerFunc{uh.verifyEmail}},
		{http.MethodPut, "/:id", []gin.HandlerFunc{uh.updateUserPermission}},
	})
}

type VerfierKey struct {
	Key int `json:"num" binding:"required"`
}

type usersHandler struct {
	store queries.UserStore
	cache db.CacheClient
}

// @Summary		Create a new user
// @Description	Admin-only endpoint to create a user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			user	body		queries.CreateUserRequest	true	"New user data"
// @Success		201		{object}	map[string]string			"Created"
// @Failure		400		{object}	map[string]string			"Bad Request"
// @Failure		409		{object}	map[string]string			"Conflict: username exists"
// @Security		CookieAuth
// @Router			/users/ [post]c
func (uh *usersHandler) createUserByAdmin(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input queries.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "invalid user data")
		return
	}

	ctx := c.Request.Context()
	if uh.store.IsUsernameExists(ctx, input.Username) {
		c.JSON(http.StatusConflict, gin.H{"msg": "username already exists"})
		return
	}
	var err error
	input.Password, err = auth.PasswordHash(input.Password)
	if err != nil {
		BadRequest(c, "invalid password string")
		return
	}
	if err = uh.store.Create(ctx, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

// @Summary		List users
// @Description	List all users (admin only)
// @Tags			users
// @Produce		json
// @Success		200	{array}		queries.UserDetailsResponse
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/users/ [get]
func (uh *usersHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	output, err := uh.store.List(c.Request.Context(), defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Get user details
// @Description	Retrieve a user's information by username (admin/vendor accessible)
// @Tags			users
// @Produce		json
// @Param			username	path		string	true	"Username"
// @Success		200			{object}	queries.UserDetailsResponse
// @Failure		404			{object}	map[string]string	"Not Found"
// @Router			/users/{username} [get]
func (uh *usersHandler) get(c *gin.Context) {
	username := c.Param("username")
	output, err := uh.store.GetDetails(c.Request.Context(), username)
	if err != nil || !HasPermissions(nil, output.PermissionType, queries.Admin, queries.Vendor) {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Update user permission
// @Description	Admin-only endpoint to update user permission and activation status
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id			path		int								true	"User ID"
// @Param			permission	body		queries.UserPermissionRequest	true	"Permission update"
// @Success		202			{object}	map[string]string				"Accepted"
// @Failure		400			{object}	map[string]string				"Bad Request"
// @Failure		404			{object}	map[string]string				"Not Found"
// @Security		CookieAuth
// @Router			/users/{id} [put]
func (uh *usersHandler) updateUserPermission(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid id")
		return
	}
	var input queries.UserPermissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "invalid permission_id or is_active")
		return
	}

	if err := uh.store.UpdatePermission(c.Request.Context(), int32(id), &input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("login:%d", claims.ID)
	if _, err := uh.cache.Del(c.Request.Context(), cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}
	Accepted(c, "")
}

// @Summary		Delete current user
// @Description	Deletes the authenticated user's account
// @Tags			users
// @Produce		json
// @Success		204	{object}	map[string]string	"No Content"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/users/ [delete]
func (uh *usersHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if err := uh.store.Delete(c.Request.Context(), claims.ID, claims.Username); err != nil {
		NotFound(c, "")
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary		Set user email
// @Description	Update email and send verification key (auth required)
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			email	body		queries.EmailAddrRequest	true	"Email address"
// @Success		202		{object}	map[string]string			"Accepted"
// @Failure		400		{object}	map[string]string			"Bad Request"
// @Failure		422		{object}	map[string]string			"Unprocessable"
// @Security		CookieAuth
// @Router			/users/set-email [put]
func (uh *usersHandler) setEmail(c *gin.Context) {
	var input queries.EmailAddrRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "invalid email address")
		return
	}
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	if err := uh.store.SetEmail(ctx, claims.ID, &input); err != nil {
		BadRequest(c, "email already exists")
		return
	}

	randKey := RandVerifyNum()
	cacheKey := fmt.Sprintf("verify:email:%s", claims.Username)
	if _, err := uh.cache.SetEx(ctx, cacheKey, randKey, 2*time.Minute).Result(); err != nil {
		LogCacheErr("SetEx", cacheKey, err)
		Unprocessable(c, "Failed to set verifier key")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := background.SendMail(ctx, uh.cache, &background.MailMessage{
			To:  input.Email,
			Msg: []byte(strconv.Itoa(randKey)),
		}); err != nil {
			LogCacheErr("SendMail", "send mail", err)
		}
	}()
	Accepted(c, "")
}

// @Summary		Verify email
// @Description	Verify email with the numeric verification key
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			key	body		VerfierKey			true	"Verification Key"
// @Success		202	{object}	map[string]string	"Accepted"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Verifier key not found"
// @Security		CookieAuth
// @Router			/users/verify-email [post]
func (uh *usersHandler) verifyEmail(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input VerfierKey
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	key, err := uh.cache.GetDel(ctx, fmt.Sprintf("verify:email:%s", claims.Username)).Result()
	if err != nil {
		NotFound(c, "Verfier key not found")
		return
	}
	if key != strconv.Itoa(input.Key) {
		Forbidden(c, "")
		return
	}
	if err := uh.store.VerifyEmail(ctx, claims.ID, true); err != nil {
		slog.Error("unxpected error to UserStore.VerifyEmail", "error", err)
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

// @Summary		Set user phone number
// @Description	Update authenticated user's phone number
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			phone	body		queries.PhoneNumberRequest	true	"Phone number"
// @Success		202		{object}	map[string]string			"Accepted"
// @Failure		400		{object}	map[string]string			"Bad Request"
// @Failure		404		{object}	map[string]string			"Not Found"
// @Security		CookieAuth
// @Router			/users/set-phone-number [put]
func (uh *usersHandler) setPhoneNumber(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.PhoneNumberRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	if err := uh.store.SetPhoneNumber(c.Request.Context(), claims.ID, &json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func UserProfileRouter(router *gin.RouterGroup) {
	config := internal.NewConfig()
	uph := userProfileHandler{
		store:      queries.NewUserProfileStore(db.NewSession()),
		uploadPath: config.UploadPath,
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", uph.upsert)
	router.POST("/upload", uph.uploadProfileImg)
	router.DELETE("/", uph.deleteImgPath)
}

type userProfileHandler struct {
	store      queries.UserProfileStore
	uploadPath string
}

// @Summary		Create or update user profile
// @Description	Inserts a new profile or updates existing profile for the authenticated user
// @Tags			user-profile
// @Accept			json
// @Produce		json
// @Param			profile	body		queries.UserProfileRequest	true	"User profile data"
// @Success		202		{object}	map[string]string			"Accepted"
// @Failure		400		{object}	map[string]string			"Bad Request"
// @Failure		404		{object}	map[string]string			"Not Found"
// @Security		CookieAuth
// @Router			/users/profile/ [post]
func (uph *userProfileHandler) upsert(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.UserProfileRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	if err := uph.store.Upsert(c.Request.Context(), claims.ID, &json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

// @Summary		Upload user profile image
// @Description	Uploads a profile image for the authenticated user. Replaces the old image if exists.
// @Tags			user-profile
// @Accept			multipart/form-data
// @Produce		json
// @Param			file	formData	file				true	"Profile image file"
// @Success		202		{object}	map[string]string	"Accepted"
// @Failure		400		{object}	map[string]string	"Bad Request"
// @Failure		422		{object}	map[string]string	"Unprocessable"
// @Security		CookieAuth
// @Router			/users/profile/upload [post]
func (uph *userProfileHandler) uploadProfileImg(c *gin.Context) {
	claims := md.GetUserClaims(c)
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "")
		return
	}
	dst := ""
	ctx := c.Request.Context()
	if fpath, err := uph.store.GetImgPath(ctx, claims.ID); err == nil {
		dst = fpath
	}
	resultPath, err := UploadFile(file, claims, "user-profile", dst)
	if err != nil {
		slog.Error("failed to upload file", "error", err)
		BadRequest(c, "failed to process file")
		return
	}
	if resultPath != dst {
		if err := uph.store.SetImgPath(ctx, claims.ID, resultPath); err != nil {
			slog.Error("failed to set img path", "error", err)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to save file"})
			return
		}
	}
	Accepted(c, "")
}

// @Summary		Delete user profile image
// @Description	Deletes the profile image for the authenticated user
// @Tags			user-profile
// @Produce		json
// @Success		204	{object}	map[string]string	"No Content"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/users/profile/ [delete]
func (upr *userProfileHandler) deleteImgPath(c *gin.Context) {
	claims := md.GetUserClaims(c)
	imgPath, err := upr.store.DeleteImgPath(c.Request.Context(), claims.ID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := os.Remove(fmt.Sprintf("%s/%s", upr.uploadPath, imgPath)); err != nil {
		slog.Info(`failed to remove file "%s", %s`, imgPath, err)
	}
	c.Status(http.StatusNoContent)
}
