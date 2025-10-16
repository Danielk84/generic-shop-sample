package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(router *gin.RouterGroup) {
	ph := productsHandler{queries.NewProductStore(db.NewSession())}

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
	store queries.ProductStore
}

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

func (ph *productsHandler) list(c *gin.Context) {
	output, err := ph.store.List(c.Request.Context(), defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (ph *productsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := claims.ID
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		id = 0
	}
	output, err := ph.store.FullList(c.Request.Context(), id, defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (ph *productsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	output, err := ph.store.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
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

func (ph *productsHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if err := ph.store.Update(c.Request.Context(), claims.ID, &input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (ph *productsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	id := c.Param("id")
	if err := ph.store.Delete(c.Request.Context(), id, claims.ID); err != nil {
		NotFound(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

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
	session := db.NewSession()
	pih := productImagesHandler{
		productStore: queries.NewProductStore(session),
		imagesStore:  queries.NewProductImagesStore(session),
	}

	router.GET("/:productID", pih.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/:productID", []gin.HandlerFunc{pih.create}},
		{http.MethodDelete, "/:productID/:id", []gin.HandlerFunc{pih.delete}},
	})
}

type productImagesHandler struct {
	productStore queries.ProductStore
	imagesStore  queries.ProductImagesStore
}

func (pih *productImagesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	output, err := pih.productStore.Get(c.Request.Context(), productID)
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
	if err := pih.imagesStore.Create(c.Request.Context(), productID, resultPath); err != nil {
		NotFound(c, "")
		return
	}
	Created(c, "")
}

func (pih *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("productID")
	output, err := pih.imagesStore.List(c.Request.Context(), productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (pih *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor) {
		return
	}

	productID := c.Param("productID")
	output, err := pih.productStore.Get(c.Request.Context(), productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if output.UserID != claims.ID {
		Forbidden(c, "")
		return
	}
	id := c.Param("id")
	imgPath, err := pih.imagesStore.Delete(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if err := os.Remove(imgPath); err != nil {
		if err != os.ErrNotExist {
			Forbidden(c, "")
			return
		}
		slog.Info("error on removing img", "img_path", imgPath, "error", err)
	}
	c.Status(http.StatusNoContent)
}
