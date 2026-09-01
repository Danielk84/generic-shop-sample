package api

import (
	"generic-shop-sample/app"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SearchRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	session := deps.DB.GetSession()
	h := searchHandler{
		store:        queries.NewSearchStore(session, log),
		productStore: queries.NewProductStore(session, log),
		cache:        deps.Cache.GetCache(cache.ProductsCache),
		log:          log,
		pagination:   deps.Config.Pagination,
	}

	rl := md.NewRateLimiter(deps.Ctx, 50, 30*time.Minute, 60*time.Second)

	RegisterRoutesWith(router, []gin.HandlerFunc{rl.RateLimiterMiddleware()}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{h.search}},
	})
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
		{http.MethodGet, "/reindex/:product_id", []gin.HandlerFunc{h.reindex}},
	})
}

type searchRequest struct {
	QueryStr string `json:"query_str" binding:"required,max=500"`
}

type searchHandler struct {
	store        queries.SearchStore
	productStore queries.ProductStore
	cache        cache.CacheClient
	log          logger.Logger
	pagination   int
}

func (h *searchHandler) reindex(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	product_id := c.Param("product_id")
	ctx := c.Request.Context()
	if err := h.store.Reindex(ctx, product_id); err != nil {
		NotFound(c, "")
		return
	}
}

func (h *searchHandler) search(c *gin.Context) {
	var input searchRequest
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	output, err := h.store.Search(ctx, input.QueryStr, h.pagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "search",
		pagination: h.pagination,
		getMaxPage: h.productStore.MaxPage,
	})
	c.JSON(http.StatusOK, output)
}
