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
		{http.MethodPut, "incr/:id", []gin.HandlerFunc{ph.incrBy}},
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
	ps queries.ProductStore
}

func (ph *productsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.CreateProductRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	if err := ph.ps.Create(c.Request.Context(), claims.ID, &json); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (ph *productsHandler) list(c *gin.Context) {
	products, err := ph.ps.List(c.Request.Context(), defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, products)
}

func (ph *productsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := claims.ID
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		id = 0
	}
	products, err := ph.ps.FullList(c.Request.Context(), id, defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, products)
}

func (ph *productsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	product, err := ph.ps.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}
	if !product.IsActive {
		if claims == nil {
			Unauthorized(c, "")
			return
		}
		if claims.ID != product.UserID && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
			Forbidden(c, "")
			return
		}
	}
	c.JSON(http.StatusOK, product)
}

func (ph *productsHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		slog.Debug(err.Error())
		BadRequest(c, "")
		return
	}
	if err := ph.ps.Update(c.Request.Context(), claims.ID, &json); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (ph *productsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	if err := ph.ps.Delete(c.Request.Context(), id, claims.ID); err != nil {
		NotFound(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func (ph *productsHandler) incrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json AvailableQuantity
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := ph.ps.IncrBy(c.Request.Context(), id, claims.ID, json.Num); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (ph *productsHandler) decrBy(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json AvailableQuantity
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")
	if err := ph.ps.DecrBy(c.Request.Context(), id, claims.ID, json.Num); err != nil {
		slog.Debug("why", "error", err)
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
	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := ph.ps.SetAvailable(c.Request.Context(), id, json.Accepted); err != nil {
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
	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}

	id := c.Param("id")
	if err := ph.ps.SetActive(c.Request.Context(), id, json.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func ProductImagesRouter(router *gin.RouterGroup) {
	session := db.NewSession()
	pih := productImagesHandler{
		ps:  queries.NewProductStore(session),
		pis: queries.NewProductImagesStore(session),
	}

	router.GET("/:productID", pih.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/:productID", []gin.HandlerFunc{pih.create}},
		{http.MethodDelete, "/:productID/:id", []gin.HandlerFunc{pih.delete}},
	})
}

type productImagesHandler struct {
	ps  queries.ProductStore
	pis queries.ProductImagesStore
}

func (pih *productImagesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	productID := c.Param("productID")
	product, err := pih.ps.Get(c.Request.Context(), productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	if product.UserID != claims.ID {
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
	if err := pih.pis.Create(c.Request.Context(), productID, resultPath); err != nil {
		NotFound(c, "")
		return
	}
	Created(c, "")
}

func (pih *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("productID")
	items, err := pih.pis.List(c.Request.Context(), productID)
	if err != nil {
		slog.Debug("hear is", "error", err)
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (pih *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	productID := c.Param("productID")
	product, err := pih.ps.Get(c.Request.Context(), productID)
	if err != nil {
		slog.Debug("not found product", "error", err)
		NotFound(c, "")
		return
	}
	if product.UserID != claims.ID {
		Forbidden(c, "")
		return
	}
	id := c.Param("id")
	imgPath, err := pih.pis.Delete(c.Request.Context(), id)
	if err != nil {
		slog.Debug("not found imgPath", "error", err)
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
