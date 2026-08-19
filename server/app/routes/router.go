package routes

import (
	"generic-shop-sample/app"
	"generic-shop-sample/app/routes/api"

	"github.com/gin-gonic/gin"
)

func APIRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	api.LoginRouter(deps, router.Group("/auth"))
	api.ShopRouter(deps, router.Group("/users/shop"))
	api.UsersRouter(deps, router.Group("/users"))
	api.ProductImagesRouter(deps, router.Group("/products/images"))
	api.ProductsRouter(deps, router.Group("/products"))
	api.PCRouter(deps, router.Group("categories/pc"))
	api.CategoriesRouter(deps, router.Group("/categories"))
	api.CommentsRouter(deps, router.Group("/comments"))
	api.OrderItemsRouter(deps, router.Group("/orders/items"))
	api.OrderRouter(deps, router.Group("/orders"))
	api.PaymentRouter(deps, router.Group("/payment"))
	api.VendorOrderRouter(deps, router.Group("/vendors"))
}

func StaticRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	api.FileRouter(deps, router.Group("/"))
}
