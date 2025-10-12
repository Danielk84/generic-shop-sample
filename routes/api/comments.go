package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func CommentsRouter(router *gin.RouterGroup) {
	ch := commentsHandler{queries.NewCommentStore(db.NewSession())}

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
	cs queries.CommentStore
}

func (ch *commentsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if !HasPermissions(c, claims.PermissionType, queries.Admin, queries.Vendor, queries.Customer) {
		return
	}
	var json queries.CommentRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}

	if err := ch.cs.Create(c.Request.Context(), claims.Username, &json); err != nil {
		BadRequest(c, "")
		return
	}
	Created(c, "")
}

func (ch *commentsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	comment, err := ch.cs.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}

	if comment.Username != claims.Username && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}
	c.JSON(http.StatusOK, comment)
}

func (ch *commentsHandler) list(c *gin.Context) {
	var json RelatedCommentsRequest
	if err := c.ShouldBindQuery(&json); err != nil {
		BadRequest(c, "")
		return
	}

	comments, err := ch.cs.List(c.Request.Context(), json.Parent, url.QueryEscape(json.Referrer), defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (ch *commentsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	username := claims.Username
	if HasPermissions(nil, claims.PermissionType, queries.Admin) {
		username = ""
	}

	items, err := ch.cs.FullList(c.Request.Context(), username, defaultPagination, GetPage(c))
	if err != nil {
		NotFound(c, "")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (ch *commentsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")

	comment, err := ch.cs.Get(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "")
		return
	}

	if comment.Username != claims.Username && !HasPermissions(nil, claims.PermissionType, queries.Admin) {
		Forbidden(c, "")
		return
	}
	if err := ch.cs.Delete(c.Request.Context(), id); err != nil {
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
	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		BadRequest(c, "")
		return
	}
	id := c.Param("id")

	if err := ch.cs.SetActive(c.Request.Context(), id, json.Accepted); err != nil {
		NotFound(c, "")
		return
	}
	Accepted(c, "")
}
