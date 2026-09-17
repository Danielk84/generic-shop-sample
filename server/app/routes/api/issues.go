package api

import (
	"generic-shop-sample/app"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IssuesRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	cache := deps.Cache.GetCache(cache.UsersCache)
	h := issuesHandler{
		store:      queries.NewIssuesStore(deps.DB.GetSession(), log),
		log:        log,
		pagination: deps.Config.Pagination,
	}

	rateLimiter := md.NewRateLimiter(deps.Ctx, md.RateLimiter{
		Cache:        cache,
		Log:          log,
		Scope:        "issues",
		RequestLimit: deps.Config.APIRateLimiter.IssuesRT,
		TTL:          deps.Config.APIRateLimiter.IssuesTTL,
	})
	router.Use(
		md.AuthMiddleware(deps, log),
		rateLimiter,
	)
	router.GET("/", h.userList)
	router.GET("/admin", h.adminList)
	router.GET("/:id/:user_id", h.get)
	router.POST("/", h.create)
	router.PUT("/", h.update)
	router.PUT("/set-response", h.setResponse)
}

type issuesHandler struct {
	store      queries.IssuesStore
	cache      cache.CacheClient
	log        logger.Logger
	pagination int
}

func (h *issuesHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Customer, queries.Vendor) {
		return
	}

	var input queries.CreateIssuesRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("issuesHandler.create", "error", err)
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	output, err := h.store.Create(ctx, input)
	if err != nil {
		Unprocessable(c, "")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": output})
}

func (h *issuesHandler) userList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Customer, queries.Vendor) {
		return
	}

	ctx := c.Request.Context()
	page := GetPage(c)
	output, err := h.store.UserList(ctx, claims.ID, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "issues-user-list",
		pagination: h.pagination,
		getMaxPage: h.store.UserMaxPage(claims.ID),
	})
	c.JSON(http.StatusOK, output)
}

func (h *issuesHandler) adminList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	ctx := c.Request.Context()
	page := GetPage(c)
	output, err := h.store.AdminList(ctx, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}
	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "issues-admin-list",
		pagination: h.pagination,
		getMaxPage: h.store.AdminMaxPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *issuesHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	var input queries.GetIssuesRequest
	if err := c.ShouldBindUri(&input); err != nil {
		h.log.Debug("issuesHandler.get", "error", err)
		BadRequest(c, "")
		return
	}
	if !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		if claims.ID != input.UserID {
			Forbidden(c, "Invalid user.")
			return
		}
	}
	ctx := c.Request.Context()
	output, err := h.store.Get(ctx, input)
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *issuesHandler) update(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, queries.Customer, queries.Vendor) {
		return
	}
	var input queries.UpdateIssuesRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("issuesHandler.update", "error", err)
		BadRequest(c, "")
		return
	}
	if claims.ID != input.UserID {
		Forbidden(c, "Invalid user.")
		return
	}
	ctx := c.Request.Context()
	if err := h.store.Update(ctx, input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}

func (h *issuesHandler) setResponse(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var input queries.SetIssuesResRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("issuesHandler.setResponse", "error", err)
		BadRequest(c, "")
		return
	}
	ctx := c.Request.Context()
	if err := h.store.SetResponse(ctx, input); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
