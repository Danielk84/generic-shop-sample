package api

import (
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

func (h *vendorOrderHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	ctx := c.Request.Context()
	page := GetPage(c)
	output, err := h.store.List(ctx, claims.ID, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "vendorOrder",
		pagination: h.pagination,
		getMaxPage: h.store.MaxPage(claims.ID),
	})
	c.JSON(http.StatusOK, output)
}

func (h *vendorOrderHandler) setIsDelivered(c *gin.Context) {
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
	if err := h.store.SetIsDelivered(ctx, input); err != nil {
		NotFound(c, "")
		return
	}

	Accepted(c, "")
}
