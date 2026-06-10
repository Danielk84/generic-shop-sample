package api

import (
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(router *gin.RouterGroup) {
	ph := productsHandler{
		store:           queries.NewProductStore(database.GetSession()),
		cache:           cache.GetCache(cache.ProductsCache),
		baseCacheKey:    "products",
		cacheExpiration: 1 * time.Hour,
	}

	router.GET("/", ph.list)
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{ph.create}},
		{http.MethodPut, "/", []gin.HandlerFunc{ph.update}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{ph.delete}},
		{http.MethodGet, "/full", []gin.HandlerFunc{ph.fullList}},
		{http.MethodGet, "/overview/:id", []gin.HandlerFunc{ph.get}},
		{http.MethodPut, "/incr/:id", []gin.HandlerFunc{ph.incrBy}},
		{http.MethodPut, "/decr/:id", []gin.HandlerFunc{ph.decrBy}},
		{http.MethodPut, "set-available/:id", []gin.HandlerFunc{ph.setAvailable}},
		{http.MethodPut, "set-active/:id", []gin.HandlerFunc{ph.setActive}},
	})
	router.GET("/:id", ph.get)
}

type AvailableQuantity struct {
	Num int32 `json:"num" binding:"required,gt=0"`
}

type productsHandler struct {
	store           queries.ProductStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
}

func (ph *productsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input queries.CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	if err := ph.store.Create(c.Request.Context(), claims.ID, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (ph *productsHandler) list(c *gin.Context) {
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:list:%d", ph.baseCacheKey, page)
	var output []queries.ProductSummaryResponse
	if err := GetJSONCache(ctx, ph.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = ph.store.List(ctx, defaultPagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ph.cache, cacheKey, ph.cacheExpiration, output); err != nil {
			LogCacheErr("SetCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (ph *productsHandler) fullList(c *gin.Context) {
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
	cacheKey := fmt.Sprintf("%s:full:%d:%d", ph.baseCacheKey, id, page)
	var output []queries.ProductStatusResponse
	if err := GetJSONCache(ctx, ph.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = ph.store.FullList(ctx, id, defaultPagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ph.cache, cacheKey, ph.cacheExpiration, output); err != nil {
			LogCacheErr("SetCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (ph *productsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	id := c.Param("id")

	cacheKey := fmt.Sprintf("%s:%s", ph.baseCacheKey, id)
	var output *queries.OwnedProductResponse
	if err := GetJSONCache(ctx, ph.cache, cacheKey, output); err != nil {
		LogCacheErr("HGetALl", ph.baseCacheKey, err)

		output, err = ph.store.Get(ctx, id)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ph.cache, cacheKey, ph.cacheExpiration, *output); err != nil {
			LogCacheErr("SetHCacheEx", ph.baseCacheKey, err)
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

func (ph *productsHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := ph.store.Update(ctx, claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", ph.baseCacheKey, input.ID)
	if _, err := ph.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	Accepted(c, "")
}

func (ph *productsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	if err := ph.store.Delete(ctx, id, claims.ID); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", ph.baseCacheKey, id)
	if _, err := ph.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	c.Status(http.StatusNoContent)
}

func (ph *productsHandler) incrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input AvailableQuantity
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := ph.store.IncrBy(c.Request.Context(), id, claims.ID, input.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (ph *productsHandler) decrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input AvailableQuantity
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := ph.store.DecrBy(c.Request.Context(), id, claims.ID, input.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (ph *productsHandler) setAvailable(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := ph.store.SetAvailable(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (ph *productsHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := ph.store.SetActive(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func ProductImagesRouter(router *gin.RouterGroup) {
	config := internal.GetConfig()
	session := database.GetSession()
	pih := productImagesHandler{
		productStore:    queries.NewProductStore(session),
		imagesStore:     queries.NewProductImagesStore(session),
		cache:           cache.GetCache(cache.ProductsCache),
		baseCacheKey:    "images",
		cacheExpiration: 1 * time.Hour,
		uploadPath:      config.Opt.UploadPath,
	}

	router.GET("/:productID", pih.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/:productID", []gin.HandlerFunc{pih.create}},
		{http.MethodDelete, "/:productID/:id", []gin.HandlerFunc{pih.delete}},
	})
}

type productImagesHandler struct {
	productStore    queries.ProductStore
	imagesStore     queries.ProductImagesStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	uploadPath      string
}

func (pih *productImagesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	output, err := pih.productStore.Get(ctx, productID)
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
		BadRequest(c, "")
		return
	}
	resultPath, err := UploadFile(file, claims, "product-images", "")
	if err != nil {
		BadRequest(c, "")
		return
	}
	if err := pih.imagesStore.Create(ctx, productID, resultPath); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", pih.baseCacheKey, productID)
	if _, err := pih.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}
	Created(c, "")
}

func (pih *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("productID")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", pih.baseCacheKey, productID)
	var output []queries.ProductImageResponse
	if err := GetJSONCache(ctx, pih.cache, cacheKey, &cacheKey); err != nil {
		LogCacheErr("HGetAll", pih.baseCacheKey, err)

		output, err = pih.imagesStore.List(ctx, productID)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, pih.cache, cacheKey, pih.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (pih *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	output, err := pih.productStore.Get(ctx, productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if output.UserID != claims.ID {
		Forbidden(c, "")
		return
	}
	id := c.Param("id")
	imgPath, err := pih.imagesStore.Delete(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := os.Remove(fmt.Sprintf("%s/%s", pih.uploadPath, imgPath)); err != nil {
		if err != os.ErrNotExist {
			Forbidden(c, "")
			return
		}
		slog.Info("error on removing img", "img_path", imgPath, "error", err)
	}
	cacheKey := fmt.Sprintf("%s:%s", pih.baseCacheKey, productID)
	if _, err := pih.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	c.Status(http.StatusNoContent)
}
