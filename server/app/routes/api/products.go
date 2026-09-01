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
	"time"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := productsHandler{
		productStore:       queries.NewProductStore(deps.DB.GetSession(), log),
		productImagesStore: queries.NewProductImagesStore(deps.DB.GetSession(), log, deps.Config.ProductImage),
		cache:              deps.Cache.GetCache(cache.ProductsCache),
		publicCache:        deps.Cache.GetCache(cache.PublicCache),
		fileStore:          file_storage.NewFileStore(deps.Ctx, deps.Config.FileStore, deps.FileStore, "product-images"),
		baseCacheKey:       "products",
		cacheExpiration:    1 * time.Hour,
		log:                log,
		pagination:         deps.Config.Pagination,
	}

	router.GET("/", h.list)
	router.GET("/popular", h.popular)
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{h.create}},
		{http.MethodPut, "/", []gin.HandlerFunc{h.update}},
		{http.MethodPut, "/set-vendor", []gin.HandlerFunc{h.setVendor}},
		{http.MethodPut, "/set-variant-detail/:id", []gin.HandlerFunc{h.setVariantDetail}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{h.delete}},
		{http.MethodGet, "/admin", []gin.HandlerFunc{h.adminList}},
		{http.MethodGet, "/overview/:id", []gin.HandlerFunc{h.get}},
		{http.MethodPut, "set-active/:id", []gin.HandlerFunc{h.setActive}},
	})
	router.GET("/:id", h.get)
}

type AvailableQuantity struct {
	Num int32 `json:"num" binding:"required,gt=0"`
}

type productsHandler struct {
	productStore       queries.ProductStore
	productImagesStore queries.ProductImagesStore
	cache              cache.CacheClient
	publicCache        cache.CacheClient
	fileStore          file_storage.FileStore
	baseCacheKey       string
	cacheExpiration    time.Duration
	log                logger.Logger
	pagination         int
}

func (h *productsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input queries.CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.create", "error", err)
		BadRequest(c, "")
		return
	}

	if err := h.productStore.Create(c.Request.Context(), input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (h *productsHandler) list(c *gin.Context) {
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:list:%d", h.baseCacheKey, page)

	output, err := JsonCache(JsonCacheInput[[]queries.ProductSummaryResponse]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) ([]queries.ProductSummaryResponse, error) {
			return h.productStore.List(sCtx, h.pagination, page)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "products-list",
		pagination: h.pagination,
		getMaxPage: h.productStore.MaxPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) popular(c *gin.Context) {
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:popular:%d", h.baseCacheKey, page)

	output, err := JsonCache(JsonCacheInput[[]queries.ProductSummaryResponse]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) ([]queries.ProductSummaryResponse, error) {
			return h.productStore.MostView(sCtx, h.pagination, page)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "products-popular",
		pagination: h.pagination,
		getMaxPage: h.productStore.MaxPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) adminList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	ctx := c.Request.Context()
	page := GetPage(c)

	output, err := h.productStore.AdminList(ctx, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "products-adminList",
		pagination: h.pagination,
		getMaxPage: h.productStore.MaxAdminListPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	id := c.Param("id")

	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)
	output, err := JsonCache(JsonCacheInput[queries.ProductResponse]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) (queries.ProductResponse, error) {
			return h.productStore.Get(sCtx, id)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	if !output.IsActive {
		if claims == nil {
			Unauthorized(c, "")
			return
		}
		if !HasPermissions(nil, claims.PermissionType, queries.Admin) {
			Forbidden(c, "")
			return
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *productsHandler) update(c *gin.Context) {
	var input queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.update", "error", err)
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := h.productStore.Update(ctx, input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, input.ID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", "productsHandler.update", err)
	}

	Accepted(c, "")
}

func (h *productsHandler) setVendor(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input queries.UpdateProductVendor
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.setVendor", "error", err)
		BadRequest(c, "")
		return
	}
	if claims.ID != input.UserID {
		Forbidden(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := h.productStore.SetVendor(ctx, input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, input.ID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", "productsHandler.setVendor", err)
	}

	Accepted(c, "")
}

func (h *productsHandler) setVariantDetail(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input struct {
		Items []queries.ProductVariantDetail `json:"items" binding:"required,min=1,dive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("productsHandler.setVariantDetail", "error", err)
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	ctx := c.Request.Context()
	if err := h.productStore.SetVariantDetail(ctx, id, input.Items); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", "productsHandler.setVariantDetail", err)
	}

	Accepted(c, "")
}

func (h *productsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	productImages, err := h.productImagesStore.List(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}
	imgs := make([]string, 0, len(productImages))
	for _, p := range productImages {
		imgs = append(imgs, p.ImgPath)
	}
	if err := h.fileStore.BulkDelete(ctx, imgs); err != nil {
		Unprocessable(c, "")
		return
	}
	if err := h.productStore.Delete(ctx, id); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", "productsHandler.delete", err)
	}

	c.Status(http.StatusNoContent)
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
	ctx := c.Request.Context()
	if err := h.productStore.SetActive(ctx, id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	err := background.SendCacheCleanr(ctx, h.publicCache, background.CacheCleanerMessage{
		CacheDB: cache.ProductsCache,
		Keys: []string{
			fmt.Sprintf("%s:%s", h.baseCacheKey, id),
			fmt.Sprintf("%s:list", h.baseCacheKey),
			fmt.Sprintf("%s:popular", h.baseCacheKey),
		},
	})
	if err != nil {
		h.log.Warn("productsHandler.setActive", "error", err)
	}
	Accepted(c, "")
}

func ProductImagesRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	session := deps.DB.GetSession()
	log := logger.GetLogger()
	h := productImagesHandler{
		productStore:    queries.NewProductStore(session, log),
		imagesStore:     queries.NewProductImagesStore(session, log, deps.Config.ProductImage),
		cache:           deps.Cache.GetCache(cache.ProductsCache),
		baseCacheKey:    "images",
		cacheExpiration: 1 * time.Hour,
		fileStore:       file_storage.NewFileStore(deps.Ctx, deps.Config.FileStore, deps.FileStore, "product-images"),
		log:             log,
	}

	router.GET("/:productID", h.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
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
	fileStore       file_storage.FileStore
	log             logger.Logger
}

func (h *productImagesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	file, err := c.FormFile("file")
	if err != nil {
		h.log.Debug("productImagesHandler.create:c.FormFile", "error", err)
		BadRequest(c, "")
		return
	}
	fileKey := GenFileKey(claims, productID)
	resultPath, err := h.fileStore.Upload(ctx, file, fileKey)
	if err != nil {
		h.log.Debug("productImagesHandler.create:h.fileStore.Upload", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.imagesStore.Create(ctx, productID, resultPath); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, productID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", "productImagesHandler.create", err)
	}
	Created(c, "")
}

func (h *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("productID")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, productID)

	output, err := JsonCache(JsonCacheInput[[]queries.ProductImageResponse]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) ([]queries.ProductImageResponse, error) {
			return h.imagesStore.List(sCtx, productID)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	id := c.Param("id")
	imgPath, err := h.imagesStore.Delete(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := h.fileStore.Delete(ctx, imgPath); err != nil {
		Forbidden(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, productID)
	if _, err := h.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", "productImagesHandler.delete", err)
	}

	c.Status(http.StatusNoContent)
}
