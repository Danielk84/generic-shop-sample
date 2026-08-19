package api

import (
	"context"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/file_storage"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func UsersRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := usersHandler{
		userStore:  queries.NewUserStore(deps.DB.GetSession(), log),
		shopStore:  queries.NewShopStore(deps.DB.GetSession(), log),
		fileStore:  file_storage.NewFileStore(deps.Ctx, deps.Config.FileStore, deps.FileStore, "user-profile"),
		cache:      deps.Cache.GetCache(cache.UsersCache),
		log:        log,
		pagination: deps.Config.Pagination,
	}

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
		{http.MethodGet, "/", []gin.HandlerFunc{h.list}},
		{http.MethodGet, "/:id", []gin.HandlerFunc{h.get}},
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
	userStore  queries.UserStore
	shopStore  queries.ShopStore
	fileStore  file_storage.FileStore
	cache      cache.CacheClient
	log        logger.Logger
	pagination int
}

func (h *usersHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	output, err := h.userStore.List(c.Request.Context(), h.pagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *usersHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	id := c.Param("id")
	output, err := h.userStore.Get(c.Request.Context(), id)
	if err != nil {
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
	id := c.Param("id")
	var input queries.UserPermissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("usersHandler.updateUserPermission", "err", err)
		BadRequest(c, "invalid permission_id or is_active")
		return
	}

	if err := h.userStore.UpdatePermission(c.Request.Context(), id, input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("login:%s", claims.ID)
	if _, err := h.cache.Del(c.Request.Context(), cacheKey).Result(); err != nil {
		LogCacheErr("Del", "usersHandler.updateUserPermission", err)
	}
	Accepted(c, "")
}

func (h *usersHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	if HasPermissions(nil, claims.PermissionType, queries.Admin, queries.Vendor) {
		imgPath, err := h.shopStore.GetImgPath(ctx, claims.ID)
		if err != nil {
			NotFound(c, "")
			return
		}
		if imgPath != "" {
			if err := h.fileStore.Delete(ctx, imgPath); err != nil {
				Unprocessable(c, "")
				return
			}
		}
	}
	if err := h.userStore.Delete(ctx, claims.ID, claims.Email); err != nil {
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
	if err := h.userStore.SetEmail(ctx, claims.ID, input); err != nil {
		BadRequest(c, "email already exists")
		return
	}

	randKey := RandVerifyNum()
	cacheKey := fmt.Sprintf("verify:email:%s", input.Email)
	if _, err := h.cache.Set(ctx, cacheKey, randKey, 2*time.Minute).Result(); err != nil {
		LogCacheErr("Set", "usersHandler.setEmail", err)
		Unprocessable(c, "Failed to set verifier key")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := background.SendMail(ctx, h.cache, background.MailMessage{
			To:      []string{input.Email},
			Subject: "pass-key",
			Msg:     strconv.Itoa(randKey),
		}); err != nil {
			LogCacheErr("SendMail", "usersHandler.setEmail", err)
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
	cacheKey := fmt.Sprintf("verify:email:%s", claims.Email)
	key, err := h.cache.GetDel(ctx, cacheKey).Result()
	if err != nil {
		LogCacheErr("GetDel", "usersHandler.verifyEmail", err)
		NotFound(c, "Verfier key not found")
		return
	}
	if key != strconv.Itoa(input.Key) {
		Forbidden(c, "")
		return
	}
	if err := h.userStore.VerifyEmail(ctx, claims.ID, true); err != nil {
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
	if err := h.userStore.SetPhoneNumber(c.Request.Context(), claims.ID, json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func ShopRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := ShopHandler{
		store:      queries.NewShopStore(deps.DB.GetSession(), log),
		log:        log,
		fileStore:  file_storage.NewFileStore(deps.Ctx, deps.Config.FileStore, deps.FileStore, "user-profile"),
		pagination: deps.Config.Pagination,
	}

	router.Use(md.AuthMiddleware(deps, log))
	router.POST("/", h.upsert)
	router.GET("/:id", h.get)
	router.GET("/", h.list)
	router.PUT("/:id", h.setPhoneNumber)
	router.POST("/upload", h.uploadProfileImg)
	router.DELETE("/", h.deleteImgPath)
}

type ShopHandler struct {
	store      queries.ShopStore
	log        logger.Logger
	fileStore  file_storage.FileStore
	pagination int
}

func (h *ShopHandler) upsert(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input queries.UpsertShopRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("ShopHandler.upsert", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.Upsert(c.Request.Context(), claims.ID, input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *ShopHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	ctx := c.Request.Context()
	id := c.Param("id")
	output, err := h.store.Get(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *ShopHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	ctx := c.Request.Context()
	output, err := h.store.List(ctx, h.pagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *ShopHandler) setPhoneNumber(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	var input queries.ShopPhoneNumberRequest
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := h.store.SetPhoneNumber(ctx, claims.ID, input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *ShopHandler) uploadProfileImg(c *gin.Context) {
	claims := md.GetUserClaims(c)
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "")
		return
	}

	var fileKey string
	ctx := c.Request.Context()
	if fpath, err := h.store.GetImgPath(ctx, claims.ID); err == nil {
		fileKey = fpath
	} else {
		fileKey = GenFileKey(claims, claims.ID)
	}
	resultPath, err := h.fileStore.Upload(ctx, file, fileKey)
	if err != nil {
		h.log.Error("failed to upload file", "error", err)
		BadRequest(c, "failed to process file")
		return
	}
	if resultPath != fileKey {
		if err := h.store.SetImgPath(ctx, claims.ID, resultPath); err != nil {
			h.log.Error("failed to set img path", "error", err)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to save file"})
			return
		}
	}
	Accepted(c, "")
}

func (h *ShopHandler) deleteImgPath(c *gin.Context) {
	claims := md.GetUserClaims(c)
	imgPath, err := h.store.DeleteImgPath(c.Request.Context(), claims.ID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := h.fileStore.Delete(c.Request.Context(), imgPath); err != nil {
		Forbidden(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}
