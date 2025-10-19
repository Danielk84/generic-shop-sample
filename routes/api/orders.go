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
		store: queries.NewOrderStore(db.NewSession()),
	}

	router.Use(md.AuthMiddleware())
	router.POST("/", oh.create)
	router.GET("/customer", oh.customerList)
	router.GET("/vendor", oh.vendorList)
	router.PUT("/set-user-info/:id", oh.setUserInfo)
	router.PUT("/verify/:id", oh.verifyUserInfo)
	router.GET("/:id", oh.get)
}

type orderHandler struct {
	store queries.OrderStore
}

// @Summary		Create a new order
// @Description	Creates a new order for the authenticated user
// @Tags			orders
// @Accept			json
// @Produce		json
// @Success		201	{object}	queries.OrderResponse
// @Failure		400	{object}	map[string]string	"Bad Request"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Security		CookieAuth
// @Router			/orders [post]
func (uh *orderHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	output, err := uh.store.Create(c.Request.Context(), claims.ID)
	if err != nil {
		BadRequest(c, "")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order_id": output})
}

// @Summary		Get orders for the authenticated customer
// @Description	List all orders created by the authenticated customer
// @Tags			orders
// @Produce		json
// @Param			page	query		int	false	"Page number"	default(1)
// @Success		200		{object}	queries.OrderSummaryResponse
// @Failure		404		{object}	map[string]string	"Not Found"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Security		CookieAuth
// @Router			/orders/customer [get]
func (uh *orderHandler) customerList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	page := GetPage(c)
	output, err := uh.store.CustomerList(c.Request.Context(), claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Get orders for the authenticated vendor
// @Description	List all orders assigned to the authenticated vendor
// @Tags			orders
// @Produce		json
// @Param			page	query		int	false	"Page number"	default(1)
// @Success		200		{object}	queries.OrderSummaryResponse
// @Failure		404		{object}	map[string]string	"Not Found"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Security		CookieAuth
// @Router			/orders/vendor [get]
func (uh *orderHandler) vendorList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	page := GetPage(c)
	output, err := uh.store.VendorList(c.Request.Context(), claims.ID, defaultPagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Get details of a specific order
// @Description	Get order details for the authenticated user
// @Tags			orders
// @Produce		json
// @Param			id	path		string	true	"Order ID"
// @Success		200	{object}	queries.OrderResponse
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/orders/{id} [get]
func (uh *orderHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	output, err := uh.store.Get(c.Request.Context(), id, claims.ID)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

// @Summary	set user info for specific order
// @Tags		orders
// @Accept		json
// @Produce	json
// @Param		id		path		string							true	"Order ID"
// @Param		input	body		queries.OrderUserInfoRequest	true	"User Info"
// @Success	200		{object}	queries.OrderResponse
// @Failure	403		{object}	map[string]string	"Forbidden"
// @Failure	404		{object}	map[string]string	"Not Found"
// @Security	CookieAuth
// @Router		/orders/set-user-info/{id} [put]
func (uh *orderHandler) setUserInfo(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.OrderUserInfoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := uh.store.SetUserInfo(c.Request.Context(), id, claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

// @Summary		Verify user's information for an order
// @Description	Admin or authorized vendor can verify user information for an order
// @Tags			orders
// @Produce		json
// @Param			id		path		string				true	"Order ID"
// @Param			input	body		SetFlag				true	"Verification status"
// @Success		202		{object}	map[string]string	"Accepted"
// @Failure		400		{object}	map[string]string	"Bad Request"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/orders/verify/{id} [put]
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
	if err := uh.store.VerifyUserInfo(c.Request.Context(), id, claims.ID, input.Accepted); err != nil {
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

// @Summary		Add an item to an order
// @Description	Adds an item to the authenticated user's order
// @Tags			order-items
// @Accept			json
// @Produce		json
// @Param			input	body		queries.OrderItemRequest	true	"Order item details"
// @Success		201		{object}	map[string]string			"Created"
// @Failure		400		{object}	map[string]string			"Bad Request"
// @Failure		403		{object}	map[string]string			"Forbidden"
// @Security		CookieAuth
// @Router			/orders/items [post]
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

// @Summary		List order items for a customer
// @Description	Returns items for the given order of the authenticated customer
// @Tags			order-items
// @Produce		json
// @Param			id		path		string	true	"Order ID"
// @Param			page	query		int		false	"Page number"	default(1)
// @Success		200		{array}		queries.OrderItemRequest
// @Failure		400		{object}	map[string]string	"Bad Request"
// @Failure		404		{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/orders/items/customer/{id} [get]
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

// @Summary		List order items for a vendor
// @Description	Returns items for the given order accessible by vendor
// @Tags			order-items
// @Produce		json
// @Param			id		path		string	true	"Order ID"
// @Param			page	query		int		false	"Page number"	default(1)
// @Success		200		{array}		queries.OrderItemRequest
// @Failure		404		{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/orders/items/vendor/{id} [get]
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

// @Summary		Get all items of an order (admin only)
// @Description	Admin can retrieve all items for a specific order
// @Tags			order-items
// @Produce		json
// @Param			id		path		string	true	"Order ID"
// @Param			page	query		int		false	"Page number"	default(1)
// @Success		200		{array}		queries.OrderItemRequest
// @Failure		404		{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/orders/items/full/{id} [get]
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

// @Summary		Remove an item from an order
// @Description	Deletes an order item for the authenticated user
// @Tags			order-items
// @Accept			json
// @Produce		json
// @Param			input	body	queries.BaseOrderItemRequest	true	"Order item to delete"
// @Success		204		"No Content"
// @Failure		400		{object}	map[string]string	"Bad Request"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/orders/items [delete]
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

// @Summary		Set total quantity of items in an order
// @Description	Updates total items for an order
// @Tags			order-items
// @Accept			json
// @Produce		json
// @Param			total	path		int								true	"Total number of items"
// @Param			input	body		queries.BaseOrderItemRequest	true	"Order details"
// @Success		202		{object}	map[string]string				"Accepted"
// @Failure		400		{object}	map[string]string				"Bad Request"
// @Failure		403		{object}	map[string]string				"Forbidden"
// @Failure		404		{object}	map[string]string				"Not Found"
// @Security		CookieAuth
// @Router			/orders/items/set-items-total/{total} [put]
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

// @Summary		Mark order item as packed
// @Description	Admin or vendor can mark an order item as packed
// @Tags			order-items
// @Accept			json
// @Produce		json
// @Param			input	body		queries.OrderItemPackStatusRequest	true	"Packed status"
// @Success		202		{object}	map[string]string					"Accepted"
// @Failure		400		{object}	map[string]string					"Bad Request"
// @Failure		404		{object}	map[string]string					"Not Found"
// @Security		CookieAuth
// @Router			/orders/items/set-packed [put]
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
