package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ProductsRouter(router *gin.RouterGroup) {
	ph := productHandler{queries.NewProductStore(db.NewSession())}

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

type SetFlag struct {
	Accepted bool `json:"accepted"`
}

type productHandler struct {
	ps queries.ProductStore
}

func (ph *productHandler) list(c *gin.Context) {
	page, err := strconv.Atoi((c.DefaultQuery("page", "1")))
	if err != nil {
		page = 1
	}

	products, err := ph.ps.List(c.Request.Context(), 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, products)
}

func (ph *productHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	id := claims.ID
	if claims.PermissionType == queries.Admin {
		id = 0
	}
	products, err := ph.ps.FullList(c.Request.Context(), id, 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, products)
}

func (ph *productHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	product, err := ph.ps.Get(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !product.IsActive {
		if claims == nil {
			c.Status(http.StatusUnauthorized)
			return
		}
		if claims.PermissionType != queries.Admin || claims.ID != product.UserID {
			c.Status(http.StatusForbidden)
			return
		}
	}
	c.JSON(http.StatusOK, product)
}

func (ph *productHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json queries.UpdateProductRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}
	if err := ph.ps.Update(c.Request.Context(), claims.ID, &json); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func (ph *productHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	if err := ph.ps.Delete(c.Request.Context(), id, claims.ID); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ph *productHandler) setAvailable(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	id := c.Param("id")
	if err := ph.ps.SetAvailable(c.Request.Context(), id, json.Accepted); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func (ph *productHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	id := c.Param("id")
	if err := ph.ps.SetActive(c.Request.Context(), id, json.Accepted); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func CategoriesRouter(router *gin.RouterGroup) {
	router.GET("/", categoriesListEndpoint)

	sRouter := router.Group("")
	sRouter.Use(md.AuthMiddleware())
	sRouter.POST("/", createCategoriesEndpoint)
	sRouter.DELETE("/:id", deleteCategoryEndpoint)
}

func createCategoriesEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	var json queries.CategoryTag
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cs := queries.NewCategoryStore(db.NewSession())
	if err := cs.Create(c.Request.Context(), json.Tag); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusCreated)
}

func categoriesListEndpoint(c *gin.Context) {
	cs := queries.NewCategoryStore(db.NewSession())
	categories, err := cs.List(c.Request.Context())
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, categories)
}

func deleteCategoryEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	id, err := strconv.Atoi(c.DefaultQuery("id", "0"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cs := queries.NewCategoryStore(db.NewSession())
	if err := cs.Delete(c.Request.Context(), int32(id)); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func PCRouter(router *gin.RouterGroup) {
	router.GET("/:id", pcListEndpoint)
	router.POST("/set-tags", md.AuthMiddleware(), setPCTagsEndpoint)
}

type PC struct {
	Tags []string `json:"tags"`
}

func setPCTagsEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType > queries.Vendor {
		c.Status(http.StatusForbidden)
		return
	}
	var json PC
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	id := c.Param("id")

	ps := queries.NewProductStore(db.NewSession())
	product, err := ps.Get(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !(product.UserID == claims.ID || claims.PermissionType == queries.Admin) {
		c.Status(http.StatusForbidden)
		return
	}

	pcs := queries.NewPCStore(db.NewSession())
	if err := pcs.SetTags(c.Request.Context(), product.ID, json.Tags); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func pcListEndpoint(c *gin.Context) {
	pcs := queries.NewPCStore(db.NewSession())
	id := c.Param("id")
	items, err := pcs.List(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, items)
}
