package api

import (
	"context"
	"errors"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func CommentsRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := commentsHandler{
		store:           queries.NewCommentStore(deps.DB.GetSession(), log),
		cache:           deps.Cache.GetCache(cache.PublicCache),
		baseCacheKey:    "comments",
		cacheExpiration: 1 * time.Hour,
		log:             log,
		pagination:      deps.Config.Pagination,
	}

	router.GET("/", h.list)

	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
		{http.MethodPost, "/", []gin.HandlerFunc{h.create}},
		{http.MethodGet, "/full", []gin.HandlerFunc{h.fullList}},
		{http.MethodGet, "/find/:id", []gin.HandlerFunc{h.find}},
		{http.MethodGet, "/overview/:id", []gin.HandlerFunc{h.get}},
		{http.MethodPut, "/set-active/:id", []gin.HandlerFunc{h.setActive}},
		{http.MethodDelete, "/:id", []gin.HandlerFunc{h.delete}},
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
	pagination      int
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
	if err := h.store.Create(c.Request.Context(), claims.ID, claims.Name, &input); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (h *commentsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if HasPermissions(nil, claims.PermissionType, queries.BlockUser) {
		Forbidden(c, "")
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:get:%s", h.baseCacheKey, id)

	output, err := JsonCache(JsonCacheInput[queries.RelatedCommentResponse]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) (output queries.RelatedCommentResponse, err error) {
			output, err = h.store.Get(sCtx, id)
			if err != nil {
				return
			}
			if output.UserID != claims.ID {
				err = ErrForbiddenAccess
			}
			return
		},
	})
	if err != nil {
		if errors.Is(err, ErrForbiddenAccess) {
			Forbidden(c, "")
		} else {
			NotFound(c, "")
		}
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
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:list:%s:%s:%d", h.baseCacheKey, input.Referrer, input.Parent, page)

	output, err := JsonCache(JsonCacheInput[[]queries.CommentResponse]{
		Ctx:        ctx,
		CacheKey:   cacheKey,
		Client:     h.cache,
		Expiration: h.cacheExpiration,
		Log:        h.log,
		Fn: func(sCtx context.Context) ([]queries.CommentResponse, error) {
			return h.store.List(sCtx, input.Parent, url.QueryEscape(input.Referrer), h.pagination, page)
		},
	})
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "comments-list",
		pagination: h.pagination,
		getMaxPage: h.store.MaxListPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *commentsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	userID := claims.ID
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		userID = ""
	}

	ctx := c.Request.Context()
	page := GetPage(c)
	cacheKey := fmt.Sprintf("%s:full:%s:%d", h.baseCacheKey, userID, page)

	var output []queries.RelatedCommentResponse
	var err error
	if userID == "" {
		output, err = h.store.FullList(ctx, userID, h.pagination, page)
	} else {
		output, err = JsonCache(JsonCacheInput[[]queries.RelatedCommentResponse]{
			Ctx:        ctx,
			CacheKey:   cacheKey,
			Client:     h.cache,
			Expiration: h.cacheExpiration,
			Log:        h.log,
			Fn: func(sCtx context.Context) ([]queries.RelatedCommentResponse, error) {
				return h.store.FullList(sCtx, userID, h.pagination, page)
			},
		})
	}

	if err != nil {
		NotFound(c, "")
		return
	}
	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "comments-fullList",
		pagination: h.pagination,
		getMaxPage: h.store.MaxFullListPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *commentsHandler) find(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin) {
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()
	page := GetPage(c)

	output, err := h.store.Find(ctx, id, h.pagination, page)
	if err != nil {
		NotFound(c, "")
		return
	}

	SetPageHeader(c, CacheMaxPageInput{
		ctx:        ctx,
		client:     h.cache,
		name:       "comments-fullList",
		pagination: h.pagination,
		getMaxPage: h.store.MaxFullListPage,
	})
	c.JSON(http.StatusOK, output)
}

func (h *commentsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	userID := claims.ID
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		userID = ""
	}
	id := c.Param("id")

	ctx := c.Request.Context()
	output, err := h.store.Get(ctx, id)
	if err != nil {
		NotFound(c, "")
		return
	}

	if userID != "" && userID != output.UserID {
		Forbidden(c, "")
		return
	}
	referrer, err := h.store.Delete(ctx, output.ID)
	if err != nil {
		BadRequest(c, "")
		return
	}
	err = background.SendCacheCleaner(ctx, h.cache, background.CacheCleanerMessage{
		CacheDB: cache.PublicCache,
		Keys: []string{
			fmt.Sprintf("%s:get:%s", h.baseCacheKey, id),
			fmt.Sprintf("%s:list:%s", h.baseCacheKey, referrer),
			fmt.Sprintf("%s:full:%s", h.baseCacheKey, userID),
		},
	})
	if err != nil {
		h.log.Warn("commentsHandler.delete", "error", err)
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
