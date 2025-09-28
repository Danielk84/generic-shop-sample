package api

import (
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal/auth"
	md "generic-shop-sample/middlewares"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UsersRouter(router *gin.RouterGroup) {
	uh := usersHandler{queries.NewUserStore(db.NewSession())}

	router.Use(md.AuthMiddleware())
	router.GET("/", uh.list)
	router.GET("/:username", uh.get)
	router.POST("/", uh.createUserByAdmin)
	router.PUT("/:id", uh.updateUserPermission)
	router.PUT("/set-email", uh.setEmail)
	router.DELETE("/", uh.delete)
}

type usersHandler struct {
	us queries.UserStore
}

func (uh *usersHandler) createUserByAdmin(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	var json queries.CreateUserRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user data"})
		return
	}

	if uh.us.IsUsernameExists(c.Request.Context(), json.Username) {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}
	var err error
	json.Password, err = auth.PasswordHash(json.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password string"})
		return
	}
	if err = uh.us.Create(c.Request.Context(), &json); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func (uh *usersHandler) list(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	users, err := uh.us.List(c.Request.Context(), defaultPagination, getOffsetFromPageNum(c.Query("page")))
	if err == nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (uh *usersHandler) get(c *gin.Context) {
	username := c.Param("username")
	user, err := uh.us.Get(c.Request.Context(), username)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uh *usersHandler) updateUserPermission(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var json queries.UserPermissionRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission_id or is_active"})
		return
	}

	if err := uh.us.UpdatePermission(c.Request.Context(), int32(id), &json); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func (uh *usersHandler) delete(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if err := uh.us.Delete(c.Request.Context(), claims.ID); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func (uh *usersHandler) setEmail(c *gin.Context) {
	var json queries.EmailAddrRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
		return
	}
	claims := md.GetUserClaims(c)
	if err := uh.us.SetEmail(c.Request.Context(), claims.ID, &json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}
	c.Status(http.StatusNoContent)
}
