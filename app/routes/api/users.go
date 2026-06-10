package api

import (
	"context"
	"fmt"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UsersRouter(router *gin.RouterGroup) {
	uh := usersHandler{
		store: queries.NewUserStore(database.GetSession()),
		cache: cache.GetCache(cache.UsersCache),
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
	cache cache.CacheClient
}

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

func (uh *usersHandler) get(c *gin.Context) {
	username := c.Param("username")
	output, err := uh.store.GetDetails(c.Request.Context(), username)
	if err != nil || !HasPermissions(nil, output.PermissionType, queries.Admin, queries.Vendor) {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

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

func (uh *usersHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if err := uh.store.Delete(c.Request.Context(), claims.ID, claims.Username); err != nil {
		NotFound(c, "")
		return
	}

	c.Status(http.StatusNoContent)
}

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
	config := internal.GetConfig()
	uph := userProfileHandler{
		store:      queries.NewUserProfileStore(database.GetSession()),
		uploadPath: config.Opt.UploadPath,
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
