package api

import (
	"context"
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LoginRouter(ctx context.Context, router *gin.RouterGroup) {
	config := internal.GetConfig()
	ah := authHandler{
		queries.NewUserStore(database.GetSession()),
		cache.GetCache(cache.UsersCache),
		config.Opt.AuthExpiration,
	}

	rl := md.NewRateLimiter(ctx, 10, 30*time.Minute, 60*time.Second)
	router.Use(rl.RateLimiterMiddleware())
	router.POST("/login", ah.login)
}

type authHandler struct {
	us             queries.UserStore
	cache          cache.CacheClient
	authExpiration time.Duration
}

func (ah *authHandler) login(c *gin.Context) {
	var input queries.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "invalid username or password")
		return
	}

	ctx := c.Request.Context()
	user, err := ah.us.Get(ctx, input.Username)
	if err != nil {
		NotFound(c, "")
		return
	}
	if !auth.ComparePassword(user.Password, input.Password) {
		Unauthorized(c, "")
		return
	}
	authExpiration := time.Now().Add(ah.authExpiration * time.Minute)
	maxAge := time.Until(authExpiration)
	cacheKey := fmt.Sprintf("login:%d", user.ID)
	var output string
	if err := ah.cache.Get(c.Request.Context(), cacheKey).Scan(&output); err != nil {
		LogCacheErr("Get", cacheKey, err)

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
		output, err = auth.TokenEncoder(claims)
		if err != nil {
			BadRequest(c, "")
			return
		}
		if err = ah.cache.Set(ctx, cacheKey, output, maxAge).Err(); err != nil {
			LogCacheErr("Set", cacheKey, err)
		}
	}
	loginResponse(c, output, int(maxAge))
}

func loginResponse(c *gin.Context, tokenString string, maxAge int) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("__Host-auth-token", tokenString, maxAge, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"token": "Bearer " + tokenString})
}
