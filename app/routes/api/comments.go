package api

import (
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func CommentsRouter(router *gin.RouterGroup) {
	log := logger.GetLogger()
	ch := commentsHandler{
		store:           queries.NewCommentStore(database.GetSession(), log),
		cache:           cache.GetCache(cache.PublicCache),
		baseCacheKey:    "comments",
		cacheExpiration: 1 * time.Hour,
		log:             log,
	}

	router.GET("/", ch.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{ch.create}},
		{http.MethodGet, "/full", []gin.HandlerFunc{ch.fullList}},
		{http.MethodGet, "/overview/:id", []gin.HandlerFunc{ch.get}},
		{http.MethodPut, "/set-active/:id", []gin.HandlerFunc{ch.setActive}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{ch.delete}},
	})
}

type RelatedCommentsRequest struct {
	Parent   string `form:"parent" binding:"required,uuid"`
	Referrer string `form:"referrer" binding:"required,uuid"`
}

type commentsHandler struct {
	store           queries.CommentStore
	cache           cache.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
	log             logger.Logger
}

func (h *commentsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.CommentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("commentsHandler.create", "error", err)
		BadRequest(c, "")
		return
	}
	if err := h.store.Create(c.Request.Context(), claims.Username, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (h *commentsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	output, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}

	if output.Username != claims.Username && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}
	c.JSON(http.StatusOK, output)
}

func (h *commentsHandler) list(c *gin.Context) {
	var input RelatedCommentsRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		h.log.Debug("commentsHandler.list", "error", err)
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:list:%s:%s", h.baseCacheKey, input.Parent, input.Referrer)
	var output []queries.CommentResponse
	if err := GetJSONCache(ctx, h.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = h.store.List(ctx, input.Parent, url.QueryEscape(input.Referrer), defaultPagination, GetPage(c))
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, h.cache, cacheKey, h.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *commentsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	username := claims.Username
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		username = ""
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:full:%s", h.baseCacheKey, username)
	var output []queries.RelatedCommentResponse
	if err := h.cache.HGetAll(ctx, cacheKey).Scan(&output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)
		output, err = h.store.FullList(ctx, username, defaultPagination, GetPage(c))
		if err != nil {
			NotFound(c, "")
			return
		}
	}
	c.JSON(http.StatusOK, output)
}

func (h *commentsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")

	ctx := c.Request.Context()
	output, err := h.store.Get(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}

	if output.Username != claims.Username && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}
	if err := h.store.Delete(ctx, output.ID); err != nil {
		BadRequest(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *commentsHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("commentsHandler.setActive", "error", err)
		BadRequest(c, "")
		return
	}
	id := c.Param("id")

	if err := h.store.SetActive(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
