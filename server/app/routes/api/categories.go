package api

import (
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CategoriesRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := categoriesHandler{
		store:        queries.NewCategoryStore(database.GetSession(), log),
		cache:        cache.GetCache(cache.ProductsCache),
		baseCacheKey: "categories",
		log:          log,
	}

	router.GET("/", h.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(log)}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{h.create}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{h.delete}},
	})
}

type categoriesHandler struct {
	store        queries.CategoryStore
	cache        cache.CacheClient
	baseCacheKey string
	log          logger.Logger
}

func (h *categoriesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var input queries.CategoryTag
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("categoriesHandler.create", "error", err)
		BadRequest(c, "Invalid tag")
		return
	}

	ctx := c.Request.Context()
	if err := h.store.Create(ctx, input.Tag); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := h.cache.Del(ctx, h.baseCacheKey).Result(); err != nil {
		LogCacheErr("Del", h.baseCacheKey, err)
	}
	Created(c, "")
}

func (h *categoriesHandler) list(c *gin.Context) {
	ctx := c.Request.Context()
	var output []queries.Category
	if err := GetJSONCache(ctx, h.cache, h.baseCacheKey, &output); err != nil {
		LogCacheErr("HGetAll", h.baseCacheKey, err)

		output, err = h.store.List(ctx)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err = SetJSONCacheEx(ctx, h.cache, h.baseCacheKey, 24*time.Hour, output); err != nil {
			LogCacheErr("SetHCacheEx", h.baseCacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *categoriesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	if err := h.store.Delete(ctx, int32(id)); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := h.cache.Del(ctx, h.baseCacheKey).Result(); err != nil {
		LogCacheErr("Del", h.baseCacheKey, err)
	}
	c.Status(http.StatusNoContent)
}

func PCRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	session := database.GetSession()
	h := pcHandler{
		pcStore:      queries.NewPCStore(session, log),
		productStore: queries.NewProductStore(session, log),
		cache:        cache.GetCache(cache.ProductsCache),
		baseCacheKey: "pc",
		log:          log,
	}

	router.GET("/:id", h.list)
	router.POST("/set-tags/:id", md.AuthMiddleware(log), h.setTags)
}

type PC struct {
	Tags []string `json:"tags" binding:"required"`
}

type pcHandler struct {
	pcStore      queries.PCStore
	productStore queries.ProductStore
	cache        cache.CacheClient
	baseCacheKey string
	log          logger.Logger
}

func (h *pcHandler) setTags(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	var input PC
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("pcHandler.setTags", "error", err)
		BadRequest(c, "Invalid tags")
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	product, err := h.productStore.Get(ctx, id)
	if err != nil {
		h.log.Debug("pcHandler.setTags", "error", err)
		NotFound(c, "Related product not found")
		return
	}
	if product.UserID != claims.ID && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}

	if err := h.pcStore.SetTags(ctx, product.ID, input.Tags); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := h.cache.Del(ctx, fmt.Sprintf("%s:%s", h.baseCacheKey, product.ID)).Result(); err != nil {
		LogCacheErr("Del", h.baseCacheKey, err)
	}
	Accepted(c, "")
}

func (h *pcHandler) list(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)
	var output []string
	if err := GetJSONCache(ctx, h.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = h.pcStore.List(ctx, id)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, h.cache, cacheKey, 12*time.Hour, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}
