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
	session := deps.DB.GetSession()
	h := vendorOrderHandler{
		log:              log,
		vendorOrderStore: queries.NewVendorOrderStore(session, log),
		orderStore:       queries.NewOrderStore(session, log),
		orderItemsStore:  queries.NewOrderItemsStore(session, log),
		cache:            deps.Cache.GetCache(cache.UsersCache),
		pagination:       deps.Config.Pagination,
		cacheExpiration:  1 * time.Hour,
		baseCacheKey:     "vendor-order-router",
	}
	router.Use(md.AuthMiddleware(deps, log))
	router.GET("/", h.list)
	router.PUT("/set-delivered", h.setIsDelivered)
}

type vendorOrderHandler struct {
	log              logger.Logger
	vendorOrderStore queries.VendorOrderStore
	orderStore       queries.OrderStore
	orderItemsStore  queries.OrderItemsStore
	cache            cache.CacheClient
	pagination       int
	cacheExpiration  time.Duration
	baseCacheKey     string
}

func (h *vendorOrderHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	ctx := c.Request.Context()
	page := GetPage(c)
	output, err := h.vendorOrderStore.List(ctx, claims.ID, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       fmt.Sprintf("vendorOrder:%s", claims.ID),
		pagination: h.pagination,
		getMaxPage: h.vendorOrderStore.MaxPage(claims.ID),
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
		h.log.Debug("vendorOrderHandler.setIsDelivered", "error", err)
		BadRequest(c, "")
		return
	}
	if claims.ID != input.UserID {
		Forbidden(c, "")
		return
	}
	ctx := c.Request.Context()
	items, err := h.orderItemsStore.OrderList(ctx, input.OrderID)
	if err != nil || len(items) == 0 {
		h.log.Warn("vendorOrderHandler.setIsDelivered", "error", err)
		NotFound(c, "")
		return
	}
	allProcessed := true
	for _, i := range items {
		if i.ProcessedItems != i.ItemsTotal {
			allProcessed = false
			break
		}
	}
	if allProcessed {
		err := h.orderStore.SetDelivered(ctx, input.OrderID, input.IsDelivered)
		if err != nil {
			h.log.Warn("vendorOrderHandler.setIsDelivered", "error", err)
			NotFound(c, "")
			return
		}
	}
	if err := h.vendorOrderStore.SetIsDelivered(ctx, input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
