package api

import (
	"context"
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	md "generic-shop-sample/middlewares"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LoginRouter(ctx context.Context, router *gin.RouterGroup) {
	ah := authHandler{
		queries.NewUserStore(db.NewSession()),
		db.NewCache(db.UsersCache),
	}

	rl := md.NewRateLimiter(ctx, 10, 30*time.Minute, 60*time.Second)
	router.Use(rl.RateLimiterMiddleware())
	router.POST("/login", ah.login)
}

type authHandler struct {
	us    queries.UserStore
	cache db.CacheClient
}

func (ah *authHandler) login(c *gin.Context) {
	var json queries.LoginRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password"})
		return
	}

	user, err := ah.us.Get(c.Request.Context(), json.Username)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !auth.ComparePassword(user.Password, json.Password) {
		c.Status(http.StatusUnauthorized)
		return
	}
	cacheKey := fmt.Sprintf("login:%d", user.ID)
	if val, err := ah.cache.Get(c.Request.Context(), cacheKey).Result(); err == nil {
		loginResponse(c, val)
		return
	}
	config := internal.NewConfig()
	authExpiration := time.Now().Add(config.AuthExpiration * time.Minute)
	claims := auth.AuthClaims{
		ID:             user.ID,
		Username:       user.Username,
		PermissionType: user.PermissionType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(authExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	tokenString, err := auth.TokenEncoder(claims)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	if err := ah.cache.Set(c.Request.Context(), cacheKey, tokenString, time.Until(authExpiration)).Err(); err == nil {
		log.Println("failed to set tokenString, ", err)
	}
	loginResponse(c, tokenString)
}

func loginResponse(c *gin.Context, tokenString string) {
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
