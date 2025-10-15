package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func OrderRouter(router *gin.RouterGroup) {
	oh := orderHandler{
		os: queries.NewOrderStore(db.NewSession()),
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", oh.create)
	router.GET("/customer", oh.customerList)
	router.GET("/vendor", oh.vendorList)
	router.PUT("/verify/:id", oh.verifyUserInfo)
	router.GET("/:id", oh.get)
}

type orderHandler struct {
	os queries.OrderStore
}

func (uh *orderHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	output, err := uh.os.Create(c.Request.Context(), claims.ID)
	if err != nil {
		BadRequest(c, "")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order_id": output})
}

func (uh *orderHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	page := GetPage(c)
	output, err := uh.os.CustomerList(c.Request.Context(), claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (uh *orderHandler) vendorList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	page := GetPage(c)
	output, err := uh.os.VendorList(c.Request.Context(), claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (uh *orderHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	output, err := uh.os.Get(c.Request.Context(), id, claims.ID)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (uh *orderHandler) verifyUserInfo(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := uh.os.VerifyUserInfo(c.Request.Context(), id, claims.ID, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func OrderItemsRouter(router *gin.RouterGroup) {
	oih := orderItemsHandler{queries.NewOrderItemsStore(db.NewSession())}

	router.Use(md.AuthMiddleware())
	router.POST("/", oih.create)
	router.DELETE("/", oih.delete)
	router.GET("/customer/:id", oih.customerList)
	router.GET("/vendor/:id", oih.vendorList)
	router.GET("/full/:id", oih.fullList)
	router.PUT("/set-items-total/:total", oih.setItemsTotal)
	router.PUT("/set-packed", oih.setPacked)
}

type ItemsTotal struct {
	Total int32 `url:"total" binding:"required"`
}

type orderItemsHandler struct {
	ois queries.OrderItemsStore
}

func (oih *orderItemsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.OrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if err := oih.ois.Create(c.Request.Context(), claims.ID, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (oih *orderItemsHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	output, err := oih.ois.CustomerList(c.Request.Context(), id, claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
	}
	c.JSON(http.StatusOK, output)
}

func (oih *orderItemsHandler) vendorList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	output, err := oih.ois.VendorList(c.Request.Context(), id, claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (oih *orderItemsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	id := c.Param("id")
	page := GetPage(c)
	output, err := oih.ois.FullList(c.Request.Context(), id, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (oih *orderItemsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.BaseOrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if err := oih.ois.Delete(c.Request.Context(), claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func (oih *orderItemsHandler) setItemsTotal(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var url ItemsTotal
	if err := c.ShouldBindUri(&url); err != nil {
		BadRequest(c, "")
		return
	}
	var input queries.BaseOrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if err := oih.ois.SetItemsTotal(c.Request.Context(), claims.ID, url.Total, &input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (oih *orderItemsHandler) setPacked(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input queries.OrderItemPackStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if err := oih.ois.SetPacked(c.Request.Context(), claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
