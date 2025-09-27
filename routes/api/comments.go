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
	router.GET("/", commentsListEndpoint)

	sRouter := router.Group("/p")
	sRouter.Use(md.AuthMiddleware())
	sRouter.POST("/", createCommentsEndpoint)
	sRouter.GET("/:id", getCommentEndpoint)
	sRouter.GET("/", commentsFullListEndpoint)
	sRouter.DELETE("/:id", deleteCommentEndpoint)
	sRouter.PUT("/set-active/:id", setCommentActiveEndpoint)
}

type CreateComment struct {
	Body     string `json:"body"`
	Parent   string `json:"parent"`
	Referrer string `json:"referrer"`
}

func createCommentsEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType > queries.Customer {
		c.Status(http.StatusForbidden)
		return
	}
	var json CreateComment
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cs := queries.NewCommentStore(db.NewSession())
	if err := cs.Create(c.Request.Context(), &queries.RelatedComment{
		Parent:   &json.Parent,
		Referrer: json.Referrer,
		Comment: queries.Comment{
			Username: claims.Username,
			Body:     json.Body,
		},
	}); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusCreated)
}

func getCommentEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")
	cs := queries.NewCommentStore(db.NewSession())
	comment, err := cs.Get(c.Request.Context(), id)
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

type CommentsList struct {
	Parent   string `json:"Parent"`
	Referrer string `json:"referrer"`
}

func commentsListEndpoint(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var json CommentsList
	if err := c.ShouldBindJSON(&json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var parent *string = nil
	if json.Parent != "" {
		parent = &json.Parent
	}

	cs := queries.NewCommentStore(db.NewSession())
	items, err := cs.List(c.Request.Context(), parent, json.Referrer, 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, items)
}

func commentsFullListEndpoint(c *gin.Context) {
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

	cs := queries.NewCommentStore(db.NewSession())
	items, err := cs.FullList(c.Request.Context(), username, 20, page)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, items)
}

func deleteCommentEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	id := c.Param("id")

	cs := queries.NewCommentStore(db.NewSession())
	comment, err := cs.Get(c.Request.Context(), id)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if !(comment.Username == claims.Username || claims.PermissionType == queries.Admin) {
		c.Status(http.StatusForbidden)
		return
	}
	if err := cs.Delete(c.Request.Context(), id); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
}

func setCommentActiveEndpoint(c *gin.Context) {
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

	cs := queries.NewCommentStore(db.NewSession())
	cs.SetActive(c.Request.Context(), id, json.Accepted)
}
