package api

import (
	"context"
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
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
	log := logger.GetLogger()
	h := authHandler{
		queries.NewUserStore(database.GetSession(), log),
		cache.GetCache(cache.UsersCache),
		log,
		time.Duration(config.Opt.AuthExpiration),
	}

	rl := md.NewRateLimiter(ctx, 10, 30*time.Minute, 60*time.Second)
	RegisterRoutesWith(router, []gin.HandlerFunc{rl.RateLimiterMiddleware()}, []RouteSpec{
		{http.MethodPost, "/login", []gin.HandlerFunc{h.login}},
	})
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware()}, []RouteSpec{
		{http.MethodGet, "/ping", []gin.HandlerFunc{h.ping}},
	})
}

type authHandler struct {
	us             queries.UserStore
	cache          cache.CacheClient
	log            logger.Logger
	authExpiration time.Duration
}

func (h *authHandler) login(c *gin.Context) {
	var input queries.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("authHandler.login", "step", "ShouldBindJSON", "error", err)
		BadRequest(c, "invalid username or password")
		return
	}

	ctx := c.Request.Context()
	user, err := h.us.Get(ctx, input.Username)
	if err != nil {
		NotFound(c, "")
		return
	}
	if !auth.ComparePassword(user.Password, input.Password) {
		Unauthorized(c, "")
		return
	}

	cacheKey := fmt.Sprintf("login:%d", user.ID)
	var maxAge time.Duration
	var output string
	if err := h.cache.Get(c.Request.Context(), cacheKey).Scan(&output); err != nil {
		LogCacheErr("Get", cacheKey, err)

		authExpiration := time.Now().Add(h.authExpiration * time.Minute)
		maxAge = time.Until(authExpiration)
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
		if err = h.cache.Set(ctx, cacheKey, output, maxAge).Err(); err != nil {
			LogCacheErr("Set", cacheKey, err)
		}
	} else {
		ttl := h.cache.TTL(ctx, cacheKey)
		if err := ttl.Err(); err != nil {
			LogCacheErr("TTL", cacheKey, err)
			maxAge = 0
		} else {
			maxAge = ttl.Val()
		}
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("__Host-auth-token", output, int(maxAge), "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// auth token validator
func (h *authHandler) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "PONG"})
}
