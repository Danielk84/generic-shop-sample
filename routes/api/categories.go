package api

import (
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CategoriesRouter(router *gin.RouterGroup) {
	cs := categoriesHandler{
		store:        queries.NewCategoryStore(db.NewSession()),
		cache:        db.NewCache(db.ProductsCache),
		baseCacheKey: "categories",
	}

	router.GET("/", cs.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{cs.create}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{cs.delete}},
	})
}

type categoriesHandler struct {
	store        queries.CategoryStore
	cache        db.CacheClient
	baseCacheKey string
}

// @Summary		Create a category
// @Description	Creates a new product category (admin only)
// @Tags			categories
// @Accept			json
// @Produce		json
// @Param			input	body		queries.CategoryTag	true	"Category tag"
// @Success		201		{string}	string				"Created"
// @Failure		400		{string}	string				"Invalid tag"
// @Failure		403		{string}	string				"Forbidden"
// @Failure		404		{string}	string				"Not found"
// @Security		CookieAuth
// @Router			/categories [post]
func (ch *categoriesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var input queries.CategoryTag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid tag")
		return
	}

	ctx := c.Request.Context()
	if err := ch.store.Create(ctx, input.Tag); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := ch.cache.Del(ctx, ch.baseCacheKey).Result(); err != nil {
		LogCacheErr("Del", ch.baseCacheKey, err)
	}
	Created(c, "")
}

// @Summary		List categories
// @Description	Returns a list of all product categories
// @Tags			categories
// @Produce		json
// @Success		200	{array}		queries.Category
// @Failure		404	{string}	string	"Not found"
// @Router			/categories [get]
func (ch *categoriesHandler) list(c *gin.Context) {
	ctx := c.Request.Context()
	var output []queries.Category
	if err := GetJSONCache(ctx, ch.cache, ch.baseCacheKey, &output); err != nil {
		LogCacheErr("HGetAll", ch.baseCacheKey, err)

		output, err = ch.store.List(ctx)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err = SetJSONCacheEx(ctx, ch.cache, ch.baseCacheKey, 24*time.Hour, output); err != nil {
			LogCacheErr("SetHCacheEx", ch.baseCacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Delete a category
// @Description	Deletes a product category by ID (admin only)
// @Tags			categories
// @Param			id	path		int		true	"Category ID"
// @Success		204	{string}	string	"No Content"
// @Failure		400	{string}	string	"Bad Request"
// @Failure		403	{string}	string	"Forbidden"
// @Failure		404	{string}	string	"Not Found"
// @Security		CookieAuth
// @Router			/categories/{id} [delete]
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

	ctx := c.Request.Context()
	if err := ch.store.Delete(ctx, int32(id)); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := ch.cache.Del(ctx, ch.baseCacheKey).Result(); err != nil {
		LogCacheErr("Del", ch.baseCacheKey, err)
	}
	c.Status(http.StatusNoContent)
}

func PCRouter(router *gin.RouterGroup) {
	session := db.NewSession()
	pch := pcHandler{
		pcStore:      queries.NewPCStore(session),
		productStore: queries.NewProductStore(session),
		cache:        db.NewCache(db.ProductsCache),
		baseCacheKey: "pc",
	}

	router.GET("/:id", pch.list)
	router.POST("/set-tags/:id", md.AuthMiddleware(), pch.setTags)
}

type PC struct {
	Tags []string `json:"tags" binding:"required"`
}

type pcHandler struct {
	pcStore      queries.PCStore
	productStore queries.ProductStore
	cache        db.CacheClient
	baseCacheKey string
}

// @Summary		Set tags for a product
// @Description	Allows Admin or Vendor to set tags for a product
// @Tags			pc
// @Accept			json
// @Produce		json
// @Param			id		path		string				true	"Product ID"
// @Param			input	body		PC					true	"Tags input"
// @Success		202		{object}	map[string]string	"Accepted"
// @Failure		400		{object}	map[string]string	"Invalid tags"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Related product not found"
// @Security		CookieAuth
// @Router			/categories/pc/set-tags/{id} [post]
func (pch *pcHandler) setTags(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}
	var input PC
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid tags")
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	product, err := pch.productStore.Get(ctx, id)
	if err != nil {
		NotFound(c, "Related product not found")
		return
	}
	if product.UserID != claims.ID && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}

	if err := pch.pcStore.SetTags(ctx, product.ID, input.Tags); err != nil {
		NotFound(c, "")
		return
	}
	if _, err := pch.cache.Del(ctx, fmt.Sprintf("%s:%s", pch.baseCacheKey, product.ID)).Result(); err != nil {
		LogCacheErr("Del", pch.baseCacheKey, err)
	}
	Accepted(c, "")
}

// @Summary		List tags of a product
// @Description	Returns a list of tags for a specific product
// @Tags			pc
// @Produce		json
// @Param			id	path		string				true	"Product ID"
// @Success		200	{array}		string				"List of tags"
// @Failure		404	{object}	map[string]string	"Product not found"
// @Router			/categories/pc/{id} [get]
func (pch *pcHandler) list(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", pch.baseCacheKey, id)
	var output []string
	if err := GetJSONCache(ctx, pch.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = pch.pcStore.List(ctx, id)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, pch.cache, cacheKey, 12*time.Hour, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}
