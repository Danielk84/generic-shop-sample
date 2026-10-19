package api

import (
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

func OrderRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := orderHandler{
		store:           queries.NewOrderStore(deps.DB.GetSession(), log),
		cache:           deps.Cache.GetCache(cache.OrdersCache),
		baseCacheKey:    "orders",
		cacheExpiration: 1 * time.Hour,
		log:             log,
		pagination:      deps.Config.Pagination,
	}

	router.Use(md.AuthMiddleware(deps, log))
	router.POST("/", h.create)
	router.GET("/customer", h.customerList)
	router.PUT("/set-user-info/:id", h.setUserInfo)
	router.PUT("/verify/:id", h.verifyUserInfo)
	router.GET("/:id", h.get)
}

type orderHandler struct {
	store           queries.OrderStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	log             logger.Logger
	pagination      int
}

func (h *orderHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	output, err := h.store.Create(c.Request.Context(), claims.ID)
	if err != nil {
		BadRequest(c, "")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order_id": output})
}

func (h *orderHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	ctx := c.Request.Context()
	page := GetPage(c)
	output, err := h.store.CustomerList(ctx, claims.ID, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "orders-customerList",
		pagination: h.pagination,
		getMaxPage: h.store.MaxCustomerListPage(claims.ID),
	})
	c.JSON(http.StatusOK, output)
}

func (h *orderHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	ctx := c.Request.Context()
	id := c.Param("id")
	output, err := h.store.Get(ctx, queries.OrderID{
		ID:     id,
		UserID: claims.ID,
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *orderHandler) setUserInfo(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.OrderUserInfo
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderHandler.setUserInfo", "error", err)
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	err := h.store.SetUserInfo(c.Request.Context(), queries.OrderID{
		ID:     id,
		UserID: claims.ID,
	}, input)
	if err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *orderHandler) verifyUserInfo(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderHandler.VerifyUserInfo")
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	err := h.store.VerifyUserInfo(c.Request.Context(), queries.OrderID{
		ID:     id,
		UserID: claims.ID,
	}, input.Accepted)
	if err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func OrderItemsRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	session := deps.DB.GetSession()
	h := orderItemsHandler{
		orderItemsStore: queries.NewOrderItemsStore(session, log),
		log:             log,
		baseCacheKey:    "order-items",
		cacheExpiration: 1 * time.Hour,
		cache:           deps.Cache.GetCache(cache.OrdersCache),
		pagination:      deps.Config.Pagination,
	}

	router.Use(md.AuthMiddleware(deps, log))
	router.POST("/", h.create)
	router.DELETE("/", h.delete)
	router.GET("/daily-sales", h.dailySales)
	router.GET("/customer/:id", h.customerList)
	router.GET("/admin/:id", h.adminList)
	router.PUT("/set-items-total/:total", h.setItemsTotal)
}

type ItemsTotal struct {
	Total int32 `uri:"total" binding:"required"`
}

type orderItemsHandler struct {
	orderItemsStore queries.OrderItemsStore
	log             logger.Logger
	baseCacheKey    string
	cacheExpiration time.Duration
	cache           cache.CacheClient
	pagination      int
}

func (h *orderItemsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.OrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderItemsHandler.create", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.orderItemsStore.Create(c.Request.Context(), claims.ID, input); err != nil {
		NotFound(c, "")
		return
	}
	Created(c, "")
}

func (h *orderItemsHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	ctx := c.Request.Context()
	output, err := h.orderItemsStore.CustomerList(ctx, queries.OrderID{
		ID:     id,
		UserID: claims.ID,
	}, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	count, err := h.orderItemsStore.MaxCustomerListPage(queries.OrderID{
		ID:     id,
		UserID: claims.ID,
	})(ctx, h.pagination)
	if err == nil {
		c.Header("X-Max-Page", strconv.Itoa(count))
	}

	c.JSON(http.StatusOK, output)
}

func (h *orderItemsHandler) adminList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	page := GetPage(c)
	output, err := h.orderItemsStore.AdminList(ctx, id, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "orderItems-adminList",
		pagination: h.pagination,
		getMaxPage: h.orderItemsStore.MaxAdminListPage(id),
	})

	c.JSON(http.StatusOK, output)
}

func (h *orderItemsHandler) dailySales(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	page := GetPage(c)
	ctx := c.Request.Context()
	output, err := h.orderItemsStore.DailySales(ctx, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}

	c.JSON(http.StatusOK, output)
}

func (h *orderItemsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.OrderItem
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderItemsHandler.delete", "error", err)
		BadRequest(c, "")
		return
	}
	err := h.orderItemsStore.Delete(c.Request.Context(), queries.OrderItemID{
		OrderItem: input,
		UserID:    claims.ID,
	})
	if err != nil {
		NotFound(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *orderItemsHandler) setItemsTotal(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var url ItemsTotal
	if err := c.ShouldBindUri(&url); err != nil {
		h.log.Debug("orderItemsHandler.setItemsTotal", "step", "bindUri", "error", err)
		BadRequest(c, "")
		return
	}
	var input queries.OrderItem
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderItemsHandler.setItemsTotal", "step", "bindJSON", "error", err)
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	err := h.orderItemsStore.SetItemsTotal(ctx, queries.OrderItemID{
		OrderItem: input,
		UserID:    claims.ID,
	}, url.Total)
	if err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
