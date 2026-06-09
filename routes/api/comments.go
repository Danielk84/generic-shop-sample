package api

import (
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/database"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func CommentsRouter(router *gin.RouterGroup) {
	ch := commentsHandler{
		store:           queries.NewCommentStore(database.GetSession()),
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

// @Summary		Create a new comment
// @Description	Users can create comments unless blocked
// @Tags			comments
// @Accept			json
// @Produce		json
// @Param			input	body		queries.CommentRequest	true	"Comment input"
// @Success		201		{object}	map[string]string		"Created"
// @Failure		400		{object}	map[string]string		"Bad Request"
// @Failure		403		{object}	map[string]string		"Forbidden"
// @Security		CookieAuth
// @Router			/comments/ [post]
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

// @Summary		Get a comment by ID
// @Description	Only the owner or Admin can view the comment
// @Tags			comments
// @Produce		json
// @Param			id	path		string	true	"Comment ID"
// @Success		200	{object}	queries.CommentResponse
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/comments/overview/{id} [get]
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

// @Summary		List comments by parent and referrer
// @Description	Returns comments with caching
// @Tags			comments
// @Produce		json
// @Param			parent		query		string	true	"Parent ID (UUID)"
// @Param			referrer	query		string	true	"Referrer ID (UUID)"
// @Success		200			{array}		queries.CommentResponse
// @Failure		400			{object}	map[string]string	"Bad Request"
// @Failure		404			{object}	map[string]string	"Not Found"
// @Router			/comments/ [get]
func (ch *commentsHandler) list(c *gin.Context) {
	var input RelatedCommentsRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("%s:list:%s:%s", ch.baseCacheKey, input.Parent, input.Referrer)
	var output []queries.CommentResponse
	if err := GetJSONCache(ctx, ch.cache, cacheKey, &output); err != nil {
		LogCacheErr("HGetAll", cacheKey, err)

		output, err = ch.store.List(ctx, input.Parent, url.QueryEscape(input.Referrer), defaultPagination, GetPage(c))
		if err != nil {
			NotFound(c, "")
			return
		}
		if err := SetJSONCacheEx(ctx, ch.cache, cacheKey, ch.cacheExpiration, output); err != nil {
			LogCacheErr("SetHCacheEx", cacheKey, err)
		}
	}
	c.JSON(http.StatusOK, output)
}

// @Summary		List all related comments for the user
// @Description	Returns full comment list. Admin can see all users
// @Tags			comments
// @Produce		json
// @Success		200	{array}		queries.RelatedCommentResponse
// @Failure		404	{object}	map[string]string	"Not Found"
// @Router			/comments/full [get]
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

// @Summary		Delete a comment
// @Description	Only the owner or admin can delete a comment
// @Tags			comments
// @Produce		json
// @Param			id	path	string	true	"Comment ID"
// @Success		204	"No Content"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/comments/{id} [delete]
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

// @Summary		Set comment active status
// @Description	Admin can approve or reject a comment
// @Tags			comments
// @Accept			json
// @Produce		json
// @Param			id		path		string				true	"Comment ID"
// @Param			input	body		SetFlag				true	"Accepted flag"
// @Success		202		{object}	map[string]string	"Accepted"
// @Failure		400		{object}	map[string]string	"Bad Request"
// @Failure		404		{object}	map[string]string	"Not Found"
// @Security		CookieAuth
// @Router			/comments/set-active/{id} [put]
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
