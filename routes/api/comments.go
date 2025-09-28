package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	md "generic-shop-sample/middlewares"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CommentsRouter(router *gin.RouterGroup) {
	ch := commentsHandler{queries.NewCommentStore(db.NewSession())}

	router.GET("/", ch.list)

	sRouter := router.Group("/p")
	sRouter.Use(md.AuthMiddleware())
	sRouter.GET("/", ch.fullList)
	sRouter.GET("/:id", ch.get)
	sRouter.POST("/", ch.create)
	sRouter.PUT("/set-active/:id", ch.setActive)
	sRouter.DELETE("/:id", ch.delete)
}

type RelatedCommentsRequest struct {
	Parent   string `json:"Parent"`
	Referrer string `json:"referrer"`
}

type commentsHandler struct {
	cs queries.CommentStore
}

func (ch *commentsHandler) create(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType > queries.Customer {
		c.Status(http.StatusForbidden)
		return
	}
	var json queries.CommentRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := ch.cs.Create(c.Request.Context(), &json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusCreated)
}

func (ch *commentsHandler) get(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	comment, err := ch.cs.Get(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if !(comment.Username == claims.Username || claims.PermissionType == queries.Admin) {
		c.Status(http.StatusForbidden)
		return
	}
	c.JSON(http.StatusOK, comment)
}

func (ch *commentsHandler) list(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var json RelatedCommentsRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var parent *string = nil
	if json.Parent != "" {
		parent = &json.Parent
	}

	items, err := ch.cs.List(c.Request.Context(), parent, json.Referrer, 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (ch *commentsHandler) fullList(c *gin.Context) {
	claims := md.GetUserClaims(c)
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	username := claims.Username
	if claims.PermissionType == queries.Admin {
		username = ""
	}

	items, err := ch.cs.FullList(c.Request.Context(), username, 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (ch *commentsHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")

	comment, err := ch.cs.Get(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if !(comment.Username == claims.Username || claims.PermissionType == queries.Admin) {
		c.Status(http.StatusForbidden)
		return
	}
	if err := ch.cs.Delete(c.Request.Context(), id); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ch *commentsHandler) setActive(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	var json SetFlag
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	id := c.Param("id")

	if err := ch.cs.SetActive(c.Request.Context(), id, json.Accepted); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}
