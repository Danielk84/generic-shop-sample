package api

import (
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"

	"github.com/gin-gonic/gin"
)

func OrderRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := orderHandler{
		store: queries.NewOrderStore(database.GetSession(), log),
		log:   log,
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", h.create)
	router.GET("/customer", h.customerList)
	router.GET("/vendor", h.vendorList)
	router.PUT("/set-user-info/:id", h.setUserInfo)
	router.PUT("/verify/:id", h.verifyUserInfo)
	router.GET("/:id", h.get)
}

type orderHandler struct {
	store queries.OrderStore
	log   logger.Logger
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

	page := GetPage(c)
	output, err := h.store.CustomerList(c.Request.Context(), claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *orderHandler) vendorList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	page := GetPage(c)
	output, err := h.store.VendorList(c.Request.Context(), claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *orderHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	output, err := h.store.Get(c.Request.Context(), id, claims.ID)
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

	var input queries.OrderUserInfoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderHandler.setUserInfo", "error", err)
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := h.store.SetUserInfo(c.Request.Context(), id, claims.ID, &input); err != nil {
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
	if err := h.store.VerifyUserInfo(c.Request.Context(), id, claims.ID, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func OrderItemsRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := orderItemsHandler{
		store: queries.NewOrderItemsStore(database.GetSession(), log),
		log:   log,
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", h.create)
	router.DELETE("/", h.delete)
	router.GET("/customer/:id", h.customerList)
	router.GET("/vendor/:id", h.vendorList)
	router.GET("/full/:id", h.fullList)
	router.PUT("/set-items-total/:total", h.setItemsTotal)
	router.PUT("/set-packed", h.setPacked)
}

type ItemsTotal struct {
	Total int32 `uri:"total" binding:"required"`
}

type orderItemsHandler struct {
	store queries.OrderItemsStore
	log   logger.Logger
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
	if err := h.store.Create(c.Request.Context(), claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	Created(c, "")
}

func (h *orderItemsHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	output, err := h.store.CustomerList(c.Request.Context(), id, claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *orderItemsHandler) vendorList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	output, err := h.store.VendorList(c.Request.Context(), id, claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *orderItemsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	output, err := h.store.FullList(c.Request.Context(), id, defaultPagination, page)
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

	var input queries.BaseOrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderItemsHandler.delete", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.Delete(c.Request.Context(), claims.ID, &input); err != nil {
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
	var input queries.BaseOrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderItemsHandler.setItemsTotal", "step", "bindJSON", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.SetItemsTotal(c.Request.Context(), claims.ID, url.Total, &input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *orderItemsHandler) setPacked(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input queries.OrderItemPackStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("orderItemsHandler.setPacked", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.SetPacked(c.Request.Context(), claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
