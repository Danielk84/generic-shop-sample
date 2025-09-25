package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UpdateProduct struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Price       int64             `json:"price"`
	IsAvailable bool              `json:"is_available"`
	Description string            `json:"description"`
	Details     map[string]string `json:"details"`
}

type SetFlag struct {
	Accepted bool `json:"accepted"`
}

func ProductsRouter(router *gin.RouterGroup) {
	router.GET("/", productsListEndpoint)
	router.GET("/:id", getProductEndpoint)

	sRouter := router.Group("/p")
	sRouter.Use(md.AuthMiddleware())
	sRouter.GET("/", productsFullListEndpoint)
	sRouter.GET("/:id", getProductEndpoint)
	sRouter.PUT("/:id", updateProductEndpoint)
	sRouter.DELETE("/:id", deleteProductEndPoint)

	sRouter.PUT("/set-available/:id", setProductAvailableEndpoint)
	sRouter.PUT("/set-active/:id", setProductActiveEndpoint)
}

func productsListEndpoint(c *gin.Context) {
	page, err := strconv.Atoi((c.DefaultQuery("page", "1")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	ps := queries.NewProductStore(db.NewSession())
	products, err := ps.List(c.Request.Context(), 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, products)
}

func productsFullListEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	ps := queries.NewProductStore(db.NewSession())
	var id int32 = claims.ID
	if claims.PermissionType == queries.Admin {
		id = 0
	}

	products, err := ps.FullList(c.Request.Context(), id, 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, products)
}

func getProductEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	ps := queries.NewProductStore(db.NewSession())
	product, err := ps.Get(c.Request.Context(), id)
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

func updateProductEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var json UpdateProduct
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}
	product := &queries.OwnedProduct{
		UserID: claims.ID,
		Product: queries.Product{
			IsAvailable: json.IsAvailable,
			Description: &json.Description,
			Details:     json.Details,
			ProductSummary: queries.ProductSummary{
				ID:    json.ID,
				Name:  json.Name,
				Price: json.Price,
			},
		},
	}
	ps := queries.NewProductStore(db.NewSession())
	if err := ps.Update(c.Request.Context(), product); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func deleteProductEndPoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	ps := queries.NewProductStore(db.NewSession())
	if err := ps.Delete(c.Request.Context(), id, claims.ID); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func setProductAvailableEndpoint(c *gin.Context) {
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
	ps := queries.NewProductStore(db.NewSession())
	if err := ps.SetAvailable(c.Request.Context(), id, json.Accepted); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func setProductActiveEndpoint(c *gin.Context) {
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
	ps := queries.NewProductStore(db.NewSession())
	if err := ps.SetActive(c.Request.Context(), id, json.Accepted); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}
