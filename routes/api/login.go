package api

import (
	"context"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal/auth"
	md "generic-shop-sample/middlewares"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LoginRouter(ctx context.Context, router *gin.RouterGroup) {
	rl := md.NewRateLimiter(ctx, 10, 30*time.Minute, 60*time.Second)
	router.Use(rl.RateLimiterMiddleware())
	router.POST("/login", loginEndpoint)
}

type Login struct {
	Username string `json:"username" binding:"required,min=4,max=128,alphanum"`
	Password string `json:"password" binding:"required,min=8,max=32,ascii"`
}

func loginEndpoint(c *gin.Context) {
	var json Login
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password"})
		return
	}

	us := queries.NewUserStore(db.NewSession())
	user, err := us.Get(c.Request.Context(), json.Username)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !auth.ComparePassword(*user.PasswordHash, json.Password) {
		c.Status(http.StatusUnauthorized)
		return
	}

	claims := auth.AuthClaims{
		ID:             user.ID,
		Username:       user.Username,
		PermissionType: user.PermissionType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(16 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	tokenString, err := auth.TokenEncoder(claims)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
