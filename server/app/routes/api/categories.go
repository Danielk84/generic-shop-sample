package api

import (
	"context"
	"fmt"
	"generic-shop-sample/app"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CategoriesRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := categoriesHandler{
		store:           queries.NewCategoryStore(deps.DB.GetSession(), log),
		cache:           deps.Cache.GetCache(cache.ProductsCache),
		baseCacheKey:    "categories",
		cacheExpiration: 1 * time.Hour,
		log:             log,
	}

	router.GET("/", h.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{h.create}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{h.delete}},
	})
}

type categoriesHandler struct {
	store           queries.CategoryStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	log             logger.Logger
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
		LogCacheErr("Del", "categoriesHandler.create", err)
	}
	Created(c, "")
}

func (h *categoriesHandler) list(c *gin.Context) {
	ctx := c.Request.Context()

	output, err := JsonCache(JsonCacheInput[[]queries.Category]{
		Ctx:        ctx,
		CacheKey:   h.baseCacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) ([]queries.Category, error) {
			return h.store.List(sCtx)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
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
		LogCacheErr("Del", "categoriesHandler.delete", err)
	}
	c.Status(http.StatusNoContent)
}

func PCRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	session := deps.DB.GetSession()
	h := pcHandler{
		pcStore:         queries.NewPCStore(session, log),
		productStore:    queries.NewProductStore(session, log),
		cache:           deps.Cache.GetCache(cache.ProductsCache),
		cacheExpiration: 1 * time.Hour,
		baseCacheKey:    "pc",
		log:             log,
	}

	router.GET("/:id", h.list)
	router.POST("/set-tags/:id", md.AuthMiddleware(deps, log), h.setTags)
}

type PC struct {
	Tags []string `json:"tags" binding:"required"`
}

type pcHandler struct {
	pcStore         queries.PCStore
	productStore    queries.ProductStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	log             logger.Logger
}

func (h *pcHandler) setTags(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
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

	if err := h.pcStore.SetTags(ctx, product.ID, input.Tags); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := h.cache.Del(ctx, fmt.Sprintf("%s:%s", h.baseCacheKey, product.ID)).Result(); err != nil {
		LogCacheErr("Del", "pcHandler.setTags", err)
	}
	Accepted(c, "")
}

func (h *pcHandler) list(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", h.baseCacheKey, id)

	output, err := JsonCache(JsonCacheInput[[]string]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) ([]string, error) {
			return h.pcStore.List(sCtx, id)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	c.JSON(http.StatusOK, output)
}
