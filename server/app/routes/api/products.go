package api

import (
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	config := internal.GetConfig()
	h := productsHandler{
		store:           queries.NewProductStore(database.GetSession(), log),
		cache:           cache.GetCache(cache.ProductsCache),
		baseCacheKey:    "products",
		cacheExpiration: 1 * time.Hour,
		log:             log,
		pagination:      config.Opt.Pagination,
	}

	router.GET("/", h.list)
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(log)}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{h.create}},
		{http.MethodPut, "/", []gin.HandlerFunc{h.update}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{h.delete}},
		{http.MethodGet, "/full", []gin.HandlerFunc{h.fullList}},
		{http.MethodGet, "/overview/:id", []gin.HandlerFunc{h.get}},
		{http.MethodPut, "/incr/:id", []gin.HandlerFunc{h.incrBy}},
		{http.MethodPut, "/decr/:id", []gin.HandlerFunc{h.decrBy}},
		{http.MethodPut, "set-available/:id", []gin.HandlerFunc{h.setAvailable}},
		{http.MethodPut, "set-active/:id", []gin.HandlerFunc{h.setActive}},
	})
	router.GET("/:id", h.get)
}

type AvailableQuantity struct {
	Num int32 `json:"num" binding:"required,gt=0"`
}

type productsHandler struct {
	store           queries.ProductStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	log             logger.Logger
	pagination      int
}

func (h *productsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input queries.CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.create", "error", err)
		BadRequest(c, "")
		return
	}

	if err := h.store.Create(c.Request.Context(), claims.ID, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (h *productsHandler) list(c *gin.Context) {
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:list:%d", h.baseCacheKey, page)
	var output []queries.ProductSummaryResponse
	if err := GetJSONCache(ctx, h.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = h.store.List(ctx, h.pagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, h.cache, cacheKey, h.cacheExpiration, output); err != nil {
			LogCacheErr("SetCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := claims.ID
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		id = 0
	}
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:full:%d:%d", h.baseCacheKey, id, page)
	var output []queries.ProductStatusResponse
	if err := GetJSONCache(ctx, h.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = h.store.FullList(ctx, id, h.pagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, h.cache, cacheKey, h.cacheExpiration, output); err != nil {
			LogCacheErr("SetCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	id := c.Param("id")

	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)
	var output queries.OwnedProductResponse
	if err := GetJSONCache(ctx, h.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetALl", h.baseCacheKey, err)

		output, err = h.store.Get(ctx, id)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, h.cache, cacheKey, h.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", h.baseCacheKey, err)
		}
	}

	if !output.IsActive {
		if claims == nil {
			Unauthorized(c, "")
			return
		}
		if claims.ID != output.UserID && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
			Forbidden(c, "")
			return
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.update", "error", err)
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := h.store.Update(ctx, claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, input.ID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	Accepted(c, "")
}

func (h *productsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	if err := h.store.Delete(ctx, id, claims.ID); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	c.Status(http.StatusNoContent)
}

func (h *productsHandler) incrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input AvailableQuantity
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.incrBy", "error", err)
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := h.store.IncrBy(c.Request.Context(), id, claims.ID, input.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *productsHandler) decrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input AvailableQuantity
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.decrBy", "error", err)
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := h.store.DecrBy(c.Request.Context(), id, claims.ID, input.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *productsHandler) setAvailable(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.setAvailable", "error", err)
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := h.store.SetAvailable(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *productsHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.setActive", "error", err)
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := h.store.SetActive(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func ProductImagesRouter(router *gin.RouterGroup) {
	config := internal.GetConfig()
	session := database.GetSession()
	log := logger.GetLogger()
	h := productImagesHandler{
		productStore:    queries.NewProductStore(session, log),
		imagesStore:     queries.NewProductImagesStore(session, log),
		cache:           cache.GetCache(cache.ProductsCache),
		baseCacheKey:    "images",
		cacheExpiration: 1 * time.Hour,
		uploadPath:      config.Opt.UploadPath,
		fileUploader:    GetFileUploader(config, log),
		log:             log,
	}

	router.GET("/:productID", h.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(log)}, []RouteSpec{
		{http.MethodPost, "/:productID", []gin.HandlerFunc{h.create}},
		{http.MethodDelete, "/:productID/:id", []gin.HandlerFunc{h.delete}},
	})
}

type productImagesHandler struct {
	productStore    queries.ProductStore
	imagesStore     queries.ProductImagesStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	uploadPath      string
	fileUploader    FileUploaderFunc
	log             logger.Logger
}

func (h *productImagesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	output, err := h.productStore.Get(ctx, productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if output.UserID != claims.ID {
		Forbidden(c, "")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.log.Debug("productImagesHandler.create:c.FormFile", "error", err)
		BadRequest(c, "")
		return
	}
	resultPath, err := h.fileUploader(file, claims, "product-images", "")
	if err != nil {
		h.log.Debug("productImagesHandler.create:h.fileUploader", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.imagesStore.Create(ctx, productID, resultPath); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, productID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}
	Created(c, "")
}

func (h *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("productID")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, productID)
	var output []queries.ProductImageResponse
	if err := GetJSONCache(ctx, h.cache, cacheKey, &cacheKey); err != nil {
		LogCacheErr("HGetAll", h.baseCacheKey, err)

		output, err = h.imagesStore.List(ctx, productID)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, h.cache, cacheKey, h.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	output, err := h.productStore.Get(ctx, productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if output.UserID != claims.ID {
		Forbidden(c, "")
		return
	}
	id := c.Param("id")
	imgPath, err := h.imagesStore.Delete(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := os.Remove(fmt.Sprintf("%s/%s", h.uploadPath, imgPath)); err != nil {
		if err != os.ErrNotExist {
			Forbidden(c, "")
			return
		}
		h.log.Error("productImagesHandler.delete ", "img_path", imgPath, "error", err)
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, productID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	c.Status(http.StatusNoContent)
}
