package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CategoriesRouter(router *gin.RouterGroup) {
	cs := categoriesHandler{queries.NewCategoryStore(db.NewSession())}

	router.GET("/", cs.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{cs.create}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{cs.delete}},
	})
}

type categoriesHandler struct {
	cs queries.CategoryStore
}

func (ch *categoriesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var json queries.CategoryTag
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}

	if err := ch.cs.Create(c.Request.Context(), json.Tag); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (ch *categoriesHandler) list(c *gin.Context) {
	categories, err := ch.cs.List(c.Request.Context())
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (ch *categoriesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		BadRequest(c, "")
		return
	}

	if err := ch.cs.Delete(c.Request.Context(), int32(id)); err != nil {
		NotFound(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func PCRouter(router *gin.RouterGroup) {
	pch := pcHandler{queries.NewPCStore(db.NewSession())}

	router.GET("/:id", pch.list)
	router.POST("/set-tags/:id", md.AuthMiddleware(), pch.setTags)
}

type PC struct {
	Tags []string `json:"tags" binding:"required"`
}

type pcHandler struct {
	pcs queries.PCStore
}

func (pch *pcHandler) setTags(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	var json PC
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")

	ps := queries.NewProductStore(db.NewSession())
	product, err := ps.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if product.UserID != claims.ID && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}

	pcs := queries.NewPCStore(db.NewSession())
	if err := pcs.SetTags(c.Request.Context(), product.ID, json.Tags); err != nil {
		slog.Debug(err.Error())
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (pch *pcHandler) list(c *gin.Context) {
	id := c.Param("id")
	items, err := pch.pcs.List(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, items)
}
