package api

import (
	"context"
	"fmt"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UsersRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	config := internal.GetConfig()
	h := usersHandler{
		store:      queries.NewUserStore(database.GetSession(), log),
		cache:      cache.GetCache(cache.UsersCache),
		log:        log,
		pagination: config.Opt.Pagination,
	}

	router.GET("/:username", h.get)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodGet, "/", []gin.HandlerFunc{h.list}},
		{http.MethodPost, "/", []gin.HandlerFunc{h.createUserByAdmin}},
		{http.MethodDelete, "/", []gin.HandlerFunc{h.delete}},
		{http.MethodPut, "/set-email", []gin.HandlerFunc{h.setEmail}},
		{http.MethodPut, "/set-phone-number", []gin.HandlerFunc{h.setPhoneNumber}},
		{http.MethodPost, "/verify-email", []gin.HandlerFunc{h.verifyEmail}},
		{http.MethodPut, "/:id", []gin.HandlerFunc{h.updateUserPermission}},
	})
}

type VerfierKey struct {
	Key int `json:"num" binding:"required"`
}

type usersHandler struct {
	store      queries.UserStore
	cache      cache.CacheClient
	log        logger.Logger
	pagination int
}

func (h *usersHandler) createUserByAdmin(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input queries.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("usersHandler.createUserByAdmin", "error", err)
		BadRequest(c, "invalid user data")
		return
	}

	ctx := c.Request.Context()
	if h.store.IsUsernameExists(ctx, input.Username) {
		c.JSON(http.StatusConflict, gin.H{"msg": "username already exists"})
		return
	}
	var err error
	input.Password, err = auth.PasswordHash(input.Password)
	if err != nil {
		BadRequest(c, "invalid password string")
		return
	}
	if err = h.store.Create(ctx, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (h *usersHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	output, err := h.store.List(c.Request.Context(), h.pagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *usersHandler) get(c *gin.Context) {
	username := c.Param("username")
	output, err := h.store.GetDetails(c.Request.Context(), username)
	if err != nil || !HasPermissions(nil, output.PermissionType, queries.Admin, queries.Vendor) {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *usersHandler) updateUserPermission(c *gin.Context) {
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
		h.log.Debug("usersHandler.updateUserPermission", "err", err)
		BadRequest(c, "invalid permission_id or is_active")
		return
	}

	if err := h.store.UpdatePermission(c.Request.Context(), int32(id), &input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("login:%d", claims.ID)
	if _, err := h.cache.Del(c.Request.Context(), cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}
	Accepted(c, "")
}

func (h *usersHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if err := h.store.Delete(c.Request.Context(), claims.ID, claims.Username); err != nil {
		NotFound(c, "")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *usersHandler) setEmail(c *gin.Context) {
	var input queries.EmailAddrRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("usersHandler.setEmail", "error", err)
		BadRequest(c, "invalid email address")
		return
	}
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	if err := h.store.SetEmail(ctx, claims.ID, &input); err != nil {
		BadRequest(c, "email already exists")
		return
	}

	randKey := RandVerifyNum()
	cacheKey := fmt.Sprintf("verify:email:%s", claims.Username)
	if _, err := h.cache.Set(ctx, cacheKey, randKey, 2*time.Minute).Result(); err != nil {
		LogCacheErr("Set", cacheKey, err)
		Unprocessable(c, "Failed to set verifier key")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := background.SendMail(ctx, h.cache, &background.MailMessage{
			To:  input.Email,
			Msg: []byte(strconv.Itoa(randKey)),
		}); err != nil {
			LogCacheErr("SendMail", "send mail", err)
		}
	}()
	Accepted(c, "")
}

func (h *usersHandler) verifyEmail(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input VerfierKey
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("usersHandler.verifyEmail", "error", err)
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("verify:email:%s", claims.Username)
	key, err := h.cache.GetDel(ctx, cacheKey).Result()
	if err != nil {
		LogCacheErr("GetDel", cacheKey, err)
		NotFound(c, "Verfier key not found")
		return
	}
	if key != strconv.Itoa(input.Key) {
		Forbidden(c, "")
		return
	}
	if err := h.store.VerifyEmail(ctx, claims.ID, true); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *usersHandler) setPhoneNumber(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.PhoneNumberRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		h.log.Debug("usersHandler.setPhoneNumber", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.SetPhoneNumber(c.Request.Context(), claims.ID, &json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func UserProfileRouter(router *gin.RouterGroup) {
	config := internal.GetConfig()
	log := logger.GetLogger()
	h := userProfileHandler{
		store:        queries.NewUserProfileStore(database.GetSession(), log),
		uploadPath:   config.Opt.UploadPath,
		log:          log,
		fileUploader: GetFileUploader(config, log),
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", h.upsert)
	router.POST("/upload", h.uploadProfileImg)
	router.DELETE("/", h.deleteImgPath)
}

type userProfileHandler struct {
	store        queries.UserProfileStore
	uploadPath   string
	log          logger.Logger
	fileUploader FileUploaderFunc
}

func (h *userProfileHandler) upsert(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.UserProfileRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		h.log.Debug("userProfileHandler.upsert", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.Upsert(c.Request.Context(), claims.ID, &json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *userProfileHandler) uploadProfileImg(c *gin.Context) {
	claims := md.GetUserClaims(c)
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "")
		return
	}
	var dst string
	ctx := c.Request.Context()
	if fpath, err := h.store.GetImgPath(ctx, claims.ID); err == nil {
		dst = fpath
	}
	resultPath, err := h.fileUploader(file, claims, "user-profile", dst)
	if err != nil {
		h.log.Error("failed to upload file", "error", err)
		BadRequest(c, "failed to process file")
		return
	}
	if resultPath != dst {
		if err := h.store.SetImgPath(ctx, claims.ID, resultPath); err != nil {
			h.log.Error("failed to set img path", "error", err)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to save file"})
			return
		}
	}
	Accepted(c, "")
}

func (h *userProfileHandler) deleteImgPath(c *gin.Context) {
	claims := md.GetUserClaims(c)
	imgPath, err := h.store.DeleteImgPath(c.Request.Context(), claims.ID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := os.Remove(fmt.Sprintf("%s/%s", h.uploadPath, imgPath)); err != nil {
		h.log.Error("userProfileHandler.deleteImgPath", "imgPath", imgPath, "error", err)
	}
	c.Status(http.StatusNoContent)
}
