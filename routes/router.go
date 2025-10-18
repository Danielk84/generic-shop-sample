package routes

import (
	"context"
	"generic-shop-sample/routes/api"

	"github.com/gin-gonic/gin"
)

func APIRouter(ctx context.Context, router *gin.RouterGroup) {
	api.LoginRouter(ctx, router.Group("/auth"))
	api.UserProfileRouter(router.Group("/users/profile"))
	api.UsersRouter(router.Group("/users"))
	api.ProductImagesRouter(router.Group("/products/images"))
	api.ProductsRouter(router.Group("/products"))
	api.PCRouter(router.Group("categories/pc"))
	api.CategoriesRouter(router.Group("/categories"))
	api.CommentsRouter(router.Group("/comments"))
	api.OrderItemsRouter(router.Group("/orders/items"))
	api.OrderRouter(router.Group("/orders"))
	api.OrderRouter(router.Group("/payment"))
}
