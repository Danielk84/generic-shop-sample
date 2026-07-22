package api

import (
	"fmt"
	"generic-shop-sample/app"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func VendorOrderRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := vendorOrderHandler{
		log:             log,
		store:           queries.NewVendorOrderStore(deps.DB.GetSession(), log),
		cache:           deps.Cache.GetCache(cache.UsersCache),
		pagination:      deps.Config.Pagination,
		cacheExpiration: 1 * time.Hour,
		baseCacheKey:    "vendor-order-router",
	}
	router.Use(md.AuthMiddleware(deps, log))
	router.GET("/", h.list)
	router.PUT("/set-delivered", h.setIsDelivered)
}

type vendorOrderHandler struct {
	log             logger.Logger
	store           queries.VendorOrderStore
	cache           cache.CacheClient
	pagination      int
	cacheExpiration time.Duration
	baseCacheKey    string
}

func (v *vendorOrderHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:%s", v.baseCacheKey, claims.ID)
	var output []queries.VendorOrder
	if err := GetJSONCache(ctx, v.cache, cacheKey, &output); err != nil {
		LogCacheErr("GetJSONCache", "vendorOrderHandler.list", err)
		output, err = v.store.List(ctx, claims.ID, v.pagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, v.cache, cacheKey, v.cacheExpiration, output); err != nil {
			LogCacheErr("SetJSONCacheEx", "vendorOrderHandler.list", err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (v *vendorOrderHandler) setIsDelivered(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	var input queries.VendorOrderDelivere
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if claims.ID != input.UserID {
		Forbidden(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := v.store.SetIsDelivered(ctx, input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", v.baseCacheKey, claims.ID)
	if err := v.cache.Del(ctx, cacheKey).Err(); err != nil {
		LogCacheErr("Del", "vendorOrderHandler.setIsDelivered", err)
	}
	Accepted(c, "")
}
