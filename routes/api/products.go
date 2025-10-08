package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(router *gin.RouterGroup) {
	ph := productsHandler{queries.NewProductStore(db.NewSession())}

	router.GET("/", ph.list)
	router.GET("/:id", ph.get)

	sRouter := router.Group("/user")
	sRouter.Use(md.AuthMiddleware())
	sRouter.GET("/", ph.fullList)
	sRouter.GET("/:id", ph.get)
	sRouter.PUT("/", ph.update)
	sRouter.PUT("/set-available/:id", ph.setAvailable)
	sRouter.PUT("/set-active/:id", ph.setActive)
	sRouter.DELETE("/:id", ph.delete)
}

type productsHandler struct {
	ps queries.ProductStore
}

func (ph *productsHandler) list(c *gin.Context) {
	products, err := ph.ps.List(c.Request.Context(), defaultPagination, getOffsetFromPageNum(c.Query("page")))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, products)
}

func (ph *productsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := claims.ID
	if claims.PermissionType == queries.Admin {
		id = 0
	}
	products, err := ph.ps.FullList(c.Request.Context(), id, defaultPagination, getOffsetFromPageNum(c.Query("page")))
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
		if claims.PermissionType != queries.Admin || claims.ID != product.UserID {
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

func (ph *productsHandler) setAvailable(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType > queries.Vendor {
		Forbidden(c, "")
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
	if claims.PermissionType != queries.Admin {
		Forbidden(c, "")
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

func CategoriesRouter(router *gin.RouterGroup) {
	cs := categoriesHandler{queries.NewCategoryStore(db.NewSession())}

	router.GET("/", cs.list)

	sRouter := router.Group("user")
	sRouter.Use(md.AuthMiddleware())
	sRouter.POST("/", cs.create)
	sRouter.DELETE("/:id", cs.delete)
}

type categoriesHandler struct {
	cs queries.CategoryStore
}

func (ch *categoriesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		Forbidden(c, "")
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
	if claims.PermissionType != queries.Admin {
		Forbidden(c, "")
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
	if claims.PermissionType > queries.Vendor {
		Forbidden(c, "")
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
	if !(product.UserID == claims.ID || claims.PermissionType == queries.Admin) {
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

func ProductImagesRouter(router *gin.RouterGroup) {
	session := db.NewSession()
	pih := productImagesHandler{
		ps:  queries.NewProductStore(session),
		pis: queries.NewProductImagesStore(session),
	}

	router.GET("/:id", pih.list)

	sRouter := router.Group("/user")
	sRouter.POST("/", pih.create)
	sRouter.DELETE("/:productID/:id", pih.delete)
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
	pih.pis.Create(c.Request.Context(), productID, resultPath)
	Accepted(c, "")
}

func (pih *productImagesHandler) list(c *gin.Context) {
	productID := c.Param("ProductID")
	items, err := pih.pis.List(c.Request.Context(), productID)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (pih *productImagesHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	productID := c.Param("ProductID")
	product, err := pih.ps.Get(c.Request.Context(), productID)
	if err != nil {
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
		NotFound(c, "")
		return
	}
	if err := os.Remove(imgPath); err != nil {
		if err != os.ErrNotExist {
			Forbidden(c, "")
			return
		}
	}
	c.Status(http.StatusNoContent)
}
