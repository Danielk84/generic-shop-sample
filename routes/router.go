package routes

import (
	"context"
	"generic-shop-sample/routes/api"

	"github.com/gin-gonic/gin"
)

func APIRouter(ctx context.Context, router *gin.RouterGroup) {
	api.LoginRouter(ctx, router.Group("/auth"))
	api.UsersRouter(router.Group("/users"))
	api.ProductsRouter(router.Group("/products"))
	api.CategoriesRouter(router.Group("/category"))
	api.PCRouter(router.Group("/pc"))
	api.CommentsRouter(router.Group("/comments"))
}
