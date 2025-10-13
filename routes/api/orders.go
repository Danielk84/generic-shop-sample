package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	md "generic-shop-sample/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func OrderRouter(router *gin.RouterGroup) {
	config := internal.NewConfig()
	oh := orderHandler{
		os:         queries.NewOrderStore(db.NewSession()),
		pagination: config.Pagination,
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", oh.create)
	router.GET("/customer", oh.customerList)
	router.GET("/vendor", oh.vendorList)
	router.PUT("/verify/:id", oh.verifyUserInfo)
	router.GET("/:id", oh.get)
}

type orderHandler struct {
	os         queries.OrderStore
	pagination int
}

func (uh *orderHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	order_id, err := uh.os.Create(c.Request.Context(), claims.ID)
	if err != nil {
		BadRequest(c, "")
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"order_id": order_id})
}

func (uh *orderHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	page := GetPage(c)
	orders, err := uh.os.CustomerList(c.Request.Context(), claims.ID, uh.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (uh *orderHandler) vendorList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	page := GetPage(c)
	orders, err := uh.os.VendorList(c.Request.Context(), claims.ID, uh.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (uh *orderHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	order, err := uh.os.Get(c.Request.Context(), id, claims.ID)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, order)
}

func (uh *orderHandler) verifyUserInfo(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := uh.os.VerifyUserInfo(c.Request.Context(), id, claims.ID, true); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
