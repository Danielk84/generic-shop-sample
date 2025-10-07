package api

import (
	"context"
	"fmt"
	"generic-shop-sample/db"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	md "generic-shop-sample/middlewares"
	"log/slog"
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
		slog.Debug("failed to validate json", "error", err)
		BadRequest(c, "invalid username or password")
		return
	}

	user, err := ah.us.Get(c.Request.Context(), json.Username)
	if err != nil {
		slog.Debug("failed to found user", "error", err)
		NotFound(c, "")
		return
	}
	if !auth.ComparePassword(user.Password, json.Password) {
		Unauthorized(c, "")
		return
	}
	config := internal.NewConfig()
	authExpiration := time.Now().Add(config.AuthExpiration * time.Minute)
	maxAge := time.Until(authExpiration).Seconds()
	cacheKey := fmt.Sprintf("login:%d", user.ID)
	if val, err := ah.cache.Get(c.Request.Context(), cacheKey).Result(); err == nil {
		loginResponse(c, val, int(maxAge))
		return
	} else {
		slog.Debug("failed to get jwt token",
			"cacheKey", cacheKey,
			"error", err)
	}
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
		BadRequest(c, "")
		return
	}
	bearerToken := "Bearer " + tokenString
	if err := ah.cache.Set(c.Request.Context(), cacheKey, bearerToken, time.Duration(maxAge)).Err(); err != nil {
		slog.Warn("failed to set tokenString", "error", err)
	}
	loginResponse(c, bearerToken, int(maxAge))
}

func loginResponse(c *gin.Context, tokenString string, maxAge int) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("__Host-auth-token", tokenString, maxAge, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
