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
	router.Use(md.AuthMiddleware())
	router.POST("/", createUserByAdminEndpoint)
	router.GET("/:username", getUserEndpoint)
	router.PUT("/:id", updateUserPermissionEndpoint)
	router.DELETE("/", deleteUserEndpoint)
	router.PUT("/set-email", setEmailEndpoint)
}

type UserPermission struct {
	PermissionType queries.PermissionType `json:"permission_type" binding:"required,number,gt=0,lte=4"`
	IsActive       bool                   `json:"is_active" binding:"required,boolean"`
}

type UserCreator struct {
	Login
	UserPermission
}

type Email struct {
	Email string `json:"email" binding:"required,min=10,max=256,email"`
}

func createUserByAdminEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	if claims.PermissionType != queries.Admin {
		c.Status(http.StatusForbidden)
		return
	}
	var json UserCreator
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user data"})
		return
	}

	us := queries.NewUserStore(db.NewSession())
	if us.IsUsernameExists(c.Request.Context(), json.Username) {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}
	passwordHash, err := auth.PasswordHash(json.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password string"})
		return
	}
	if err := us.Create(
		c.Request.Context(),
		&queries.User{
			Username:       json.Username,
			PasswordHash:   &passwordHash,
			PermissionType: json.PermissionType,
			IsActive:       json.IsActive,
		},
	); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusNoContent)
}

func getUserEndpoint(c *gin.Context) {
	username := c.Param("username")
	us := queries.NewUserStore(db.NewSession())
	user, err := us.Get(c.Request.Context(), username)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, user)
}

func updateUserPermissionEndpoint(c *gin.Context) {
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
	var json UserPermission
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission_id or is_active"})
		return
	}

	us := queries.NewUserStore(db.NewSession())
	if err := us.Update(
		c.Request.Context(),
		&queries.User{ID: int32(id), PermissionType: json.PermissionType, IsActive: json.IsActive},
	); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func deleteUserEndpoint(c *gin.Context) {
	claims := md.GetUserClaims(c)
	us := queries.NewUserStore(db.NewSession())
	if err := us.Delete(c.Request.Context(), claims.ID); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func setEmailEndpoint(c *gin.Context) {
	var json Email
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}
	claims := md.GetUserClaims(c)

	us := queries.NewUserStore(db.NewSession())
	if err := us.SetEmail(c.Request.Context(), &queries.User{ID: claims.ID, Email: &json.Email}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}
	c.Status(http.StatusNoContent)
}
