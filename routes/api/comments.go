package api

import (
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func CommentsRouter(router *gin.RouterGroup) {
	ch := commentsHandler{
		store:           queries.NewCommentStore(db.NewSession()),
		cache:           db.NewCache(db.PublicCache),
		baseCacheKey:    "comments",
		cacheExpiration: 1 * time.Hour,
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
	cache           db.CacheClient
	baseCacheKey    string
	cacheExpiration time.Duration
}

func (ch *commentsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	var input queries.CommentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	if err := ch.store.Create(c.Request.Context(), claims.Username, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (ch *commentsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	output, err := ch.store.Get(c.Request.Context(), id)
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

func (ch *commentsHandler) list(c *gin.Context) {
	var input RelatedCommentsRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:list:%s:%s", ch.baseCacheKey, input.Parent, input.Referrer)
	var output []queries.CommentResponse
	if err := ch.cache.HGetAll(ctx, cacheKey).Scan(&output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = ch.store.List(ctx, input.Parent, url.QueryEscape(input.Referrer), defaultPagination, GetPage(c))
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetHCacheEx(ctx, ch.cache, cacheKey, ch.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

func (ch *commentsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	username := claims.Username
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		username = ""
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:full:%s", ch.baseCacheKey, username)
	var output []queries.RelatedCommentResponse
	if err := ch.cache.HGetAll(ctx, cacheKey).Scan(&output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)
		output, err = ch.store.FullList(ctx, username, defaultPagination, GetPage(c))
		if err != nil {
			NotFound(c, "")
			return
		}
	}
	c.JSON(http.StatusOK, output)
}

func (ch *commentsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")

	ctx := c.Request.Context()
	output, err := ch.store.Get(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}

	if output.Username != claims.Username && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}
	if err := ch.store.Delete(ctx, output.ID); err != nil {
		BadRequest(c, "")
		return
	}
	c.Status(http.StatusNoContent)
}

func (ch *commentsHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}
	var input SetFlag
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")

	if err := ch.store.SetActive(c.Request.Context(), id, input.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
