package api

import (
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(router *gin.RouterGroup) {
	ph := productsHandler{
		store:           queries.NewProductStore(database.GetSession()),
		cache:           db.NewCache(db.ProductsCache),
		baseCacheKey:    "products",
		cacheExpiration: 1 * time.Hour,
	}

	router.GET("/", ph.list)
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{ph.create}},
		{http.MethodPut, "/", []gin.HandlerFunc{ph.update}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{ph.delete}},
		{http.MethodGet, "/full", []gin.HandlerFunc{ph.fullList}},
		{http.MethodGet, "/overview/:id", []gin.HandlerFunc{ph.get}},
		{http.MethodPut, "/incr/:id", []gin.HandlerFunc{ph.incrBy}},
		{http.MethodPut, "/decr/:id", []gin.HandlerFunc{ph.decrBy}},
		{http.MethodPut, "set-available/:id", []gin.HandlerFunc{ph.setAvailable}},
		{http.MethodPut, "set-active/:id", []gin.HandlerFunc{ph.setActive}},
	})
	router.GET("/:id", ph.get)
}

type AvailableQuantity struct {
	Num int32 `json:"num" binding:"required,gt=0"`
}

type productsHandler struct {
	store           queries.ProductStore
	cache           db.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
}

// @Summary		Create product
// @Description	Creates a new product (admin/vendor only)
// @Tags			products
// @Accept			json
// @Produce		json
// @Param			product	body		queries.CreateProductRequest	true	"Product payload"
// @Success		201		{object}	map[string]string				"Created"
// @Failure		400		{object}	map[string]string				"Bad Request"
// @Failure		403		{object}	map[string]string				"Forbidden"
// @Security		CookieAuth
// @Router			/products/ [post]
func (ph *productsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input queries.CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	if err := ph.store.Create(c.Request.Context(), claims.ID, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

// @Summary		List products
// @Description	Retrieves a paginated list of products
// @Tags			products
// @Produce		json
// @Param			page	query		int	false	"Page number"	default(1)
// @Success		200		{array}		queries.ProductSummaryResponse
// @Failure		404		{object}	map[string]string	"Not Found"
// @Router			/products/ [get]
func (ph *productsHandler) list(c *gin.Context) {
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:list:%d", ph.baseCacheKey, page)
	var output []queries.ProductSummaryResponse
	if err := GetJSONCache(ctx, ph.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = ph.store.List(ctx, defaultPagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ph.cache, cacheKey, ph.cacheExpiration, output); err != nil {
			LogCacheErr("SetCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Full product list
// @Description	Retrieves full list of products (admin/vendor only)
// @Tags			products
// @Produce		json
// @Param			page	query		int	false	"Page number"	default(1)
// @Success		200		{array}		queries.ProductStatusResponse
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Router			/products/full [get]
// @Security		CookieAuth
func (ph *productsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := claims.ID
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		id = 0
	}
	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:full:%d:%d", ph.baseCacheKey, id, page)
	var output []queries.ProductStatusResponse
	if err := GetJSONCache(ctx, ph.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = ph.store.FullList(ctx, id, defaultPagination, page)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ph.cache, cacheKey, ph.cacheExpiration, output); err != nil {
			LogCacheErr("SetCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Get product
// @Description	Retrieves a product by ID, access restricted for inactive products
// @Tags			products
// @Produce		json
// @Param			id	path		string	true	"Product ID"
// @Success		200	{object}	queries.OwnedProductResponse
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Router			/products/{id} [get]
// @Router			/products/overview/{id} [get]
// @Security		CookieAuth
func (ph *productsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	id := c.Param("id")

	cacheKey := fmt.Sprintf("%s:%s", ph.baseCacheKey, id)
	var output *queries.OwnedProductResponse
	if err := GetJSONCache(ctx, ph.cache, cacheKey, output); err != nil {
		LogCacheErr("HGetALl", ph.baseCacheKey, err)

		output, err = ph.store.Get(ctx, id)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ph.cache, cacheKey, ph.cacheExpiration, *output); err != nil {
			LogCacheErr("SetHCacheEx", ph.baseCacheKey, err)
		}
	}

	if !output.IsActive {
		if claims == nil {
			Unauthorized(c, "")
			return
		}
		if claims.ID != output.UserID && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
			Forbidden(c, "")
			return
		}
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Update product
// @Description	Updates a product (admin/vendor only)
// @Tags			products
// @Accept			json
// @Produce		json
// @Param			product	body		queries.UpdateProductRequest	true	"Update payload"
// @Success		202		{object}	map[string]string				"Accepted"
// @Failure		400		{object}	map[string]string				"Bad Request"
// @Failure		404		{object}	map[string]string				"Not Found"
// @Security		CookieAuth
// @Router			/products/ [put]
func (ph *productsHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := ph.store.Update(ctx, claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", ph.baseCacheKey, input.ID)
	if _, err := ph.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	Accepted(c, "")
}

// @Summary		Delete product
// @Description	Deletes a product (admin/vendor only)
// @Tags			products
// @Produce		json
// @Param			id	path		string				true	"Product ID"
// @Success		204	{object}	map[string]string	"No Content"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/products/{id} [delete]
func (ph *productsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	if err := ph.store.Delete(ctx, id, claims.ID); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", ph.baseCacheKey, id)
	if _, err := ph.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	c.Status(http.StatusNoContent)
}

// @Summary	Increment product available quantity
// @Tags		products
// @Accept		json
// @Param		id			path		string				true	"Product ID"
// @Param		quantity	body		AvailableQuantity	true	"Quantity to increment"
// @Success	202			{object}	map[string]string	"Accepted"
// @Failure	400			{object}	map[string]string	"Bad Request"
// @Failure	404			{object}	map[string]string	"Not Found"
// @Security	CookieAuth
// @Router		/products/incr/{id} [put]
func (ph *productsHandler) incrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input AvailableQuantity
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := ph.store.IncrBy(c.Request.Context(), id, claims.ID, input.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

// DecrementProduct godoc
//
//	@Summary	Decrement product available quantity
//	@Tags		products
//	@Accept		json
//	@Param		id			path		string				true	"Product ID"
//	@Param		quantity	body		AvailableQuantity	true	"Quantity to decrement"
//	@Success	202			{object}	map[string]string	"Accepted"
//	@Failure	400			{object}	map[string]string	"Bad Request"
//	@Failure	404			{object}	map[string]string	"Not Found"
//	@Security	CookieAuth
//	@Router		/products/decr/{id} [put]
func (ph *productsHandler) decrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input AvailableQuantity
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := ph.store.DecrBy(c.Request.Context(), id, claims.ID, input.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

// @Summary	Set product available status
// @Tags		products
// @Accept		json
// @Param		id		path		string				true	"Product ID"
// @Param		flag	body		SetFlag				true	"Available flag"
// @Success	202		{object}	map[string]string	"Accepted"
// @Failure	400		{object}	map[string]string	"Bad Request"
// @Failure	404		{object}	map[string]string	"Not Found"
// @Security	CookieAuth
// @Router		/products/set-available/{id} [put]
func (ph *productsHandler) setAvailable(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := ph.store.SetAvailable(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

// SetActiveProduct godoc
//
//	@Summary	Set product active status
//	@Tags		products
//	@Accept		json
//	@Param		id		path		string				true	"Product ID"
//	@Param		flag	body		SetFlag				true	"Active flag"
//	@Success	202		{object}	map[string]string	"Accepted"
//	@Failure	400		{object}	map[string]string	"Bad Request"
//	@Failure	404		{object}	map[string]string	"Not Found"
//	@Security	CookieAuth
//	@Router		/products/set-active/{id} [put]
func (ph *productsHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := ph.store.SetActive(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func ProductImagesRouter(router *gin.RouterGroup) {
	config := internal.NewConfig()
	session := database.GetSession()
	pih := productImagesHandler{
		productStore:    queries.NewProductStore(session),
		imagesStore:     queries.NewProductImagesStore(session),
		cache:           db.NewCache(db.ProductsCache),
		baseCacheKey:    "images",
		cacheExpiration: 1 * time.Hour,
		uploadPath:      config.UploadPath,
	}

	router.GET("/:productID", pih.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/:productID", []gin.HandlerFunc{pih.create}},
		{http.MethodDelete, "/:productID/:id", []gin.HandlerFunc{pih.delete}},
	})
}

type productImagesHandler struct {
	productStore    queries.ProductStore
	imagesStore     queries.ProductImagesStore
	cache           db.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	uploadPath      string
}

// @Summary		Upload product image
// @Description	Uploads an image for a product (admin/vendor only). Accepts multipart/form-data.
// @Tags			product-images
// @Accept			multipart/form-data
// @Produce		json
// @Param			productID	path		string				true	"Product ID"
// @Param			file		formData	file				true	"Image file"
// @Success		201			{object}	map[string]string	"Created"
// @Failure		400			{object}	map[string]string	"Bad Request"
// @Failure		403			{object}	map[string]string	"Forbidden"
// @Failure		404			{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/products/images/{productID} [post]
func (pih *productImagesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	output, err := pih.productStore.Get(ctx, productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if output.UserID != claims.ID {
		Forbidden(c, "")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "")
		return
	}
	resultPath, err := UploadFile(file, claims, "product-images", "")
	if err != nil {
		BadRequest(c, "")
		return
	}
	if err := pih.imagesStore.Create(ctx, productID, resultPath); err != nil {
		NotFound(c, "")
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", pih.baseCacheKey, productID)
	if _, err := pih.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}
	Created(c, "")
}

// @Summary		List product images
// @Description	Retrieves all images for a given product
// @Tags			product-images
// @Produce		json
// @Param			productID	path		string	true	"Product ID"
// @Success		200			{array}		queries.ProductImageResponse
// @Failure		404			{object}	map[string]string	"Not Found"
// @Router			/products/images/{productID} [get]
func (pih *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("productID")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:%s", pih.baseCacheKey, productID)
	var output []queries.ProductImageResponse
	if err := GetJSONCache(ctx, pih.cache, cacheKey, &cacheKey); err != nil {
		LogCacheErr("HGetAll", pih.baseCacheKey, err)

		output, err = pih.imagesStore.List(ctx, productID)
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, pih.cache, cacheKey, pih.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		Delete product image
// @Description	Deletes a product image (admin/vendor only)
// @Tags			product-images
// @Produce		json
// @Param			productID	path		string				true	"Product ID"
// @Param			id			path		string				true	"Image ID"
// @Success		204			{object}	map[string]string	"No Content"
// @Failure		403			{object}	map[string]string	"Forbidden"
// @Failure		404			{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/products/images/{productID}/{id} [delete]
func (pih *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	ctx := c.Request.Context()
	output, err := pih.productStore.Get(ctx, productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if output.UserID != claims.ID {
		Forbidden(c, "")
		return
	}
	id := c.Param("id")
	imgPath, err := pih.imagesStore.Delete(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := os.Remove(fmt.Sprintf("%s/%s", pih.uploadPath, imgPath)); err != nil {
		if err != os.ErrNotExist {
			Forbidden(c, "")
			return
		}
		slog.Info("error on removing img", "img_path", imgPath, "error", err)
	}
	cacheKey := fmt.Sprintf("%s:%s", pih.baseCacheKey, productID)
	if _, err := pih.cache.Del(ctx, cacheKey).Result(); err != nil {
		LogCacheErr("Del", cacheKey, err)
	}

	c.Status(http.StatusNoContent)
}
