package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/app/background"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func LoginRouter(deps *app.ServiceDeps, router *gin.RouterGroup) {
	log := logger.GetLogger()
	h := authHandler{
		store:    queries.NewUserStore(deps.DB.GetSession(), log),
		cache:    deps.Cache.GetCache(cache.UsersCache),
		log:      log,
		jwtToken: auth.JWTToken{Log: log, JWTSecretKey: []byte(deps.Config.JWTSecretKey)},

		authExpiration:        time.Duration(deps.Config.Auth.AuthExpiration),
		passKeyExpiration:     2 * time.Minute,
		registerKeyExpiration: 10 * time.Minute,
	}

	rl := md.NewRateLimiter(deps.Ctx, 10, 30*time.Minute, 60*time.Second)
	RegisterRoutesWith(router, []gin.HandlerFunc{rl.RateLimiterMiddleware()}, []RouteSpec{
		{http.MethodPost, "/is-user-exists", []gin.HandlerFunc{h.isUserExists}},
		{http.MethodPost, "/login", []gin.HandlerFunc{h.login}},
		{http.MethodPost, "/register", []gin.HandlerFunc{h.register}},
		{http.MethodGet, "/refresh", []gin.HandlerFunc{h.refresh}},
		{http.MethodGet, "/healthz", []gin.HandlerFunc{h.healthz}},
	})
	RegisterRoutesWith(router, []gin.HandlerFunc{md.AuthMiddleware(deps, log)}, []RouteSpec{
		{http.MethodGet, "/logout", []gin.HandlerFunc{h.logout}},
	})
}

type randKeyRequest struct {
	PassKey string `json:"pass_key" binding:"required,length=7"`
}

type loginRequest struct {
	queries.EmailAddrRequest
	randKeyRequest
}

type registerRequest struct {
	queries.RegisterUserRequest
	randKeyRequest
}

type authHandler struct {
	store    queries.UserStore
	cache    cache.CacheClient
	log      logger.Logger
	jwtToken auth.JWTToken

	authExpiration        time.Duration
	registerKeyExpiration time.Duration
	passKeyExpiration     time.Duration
}

func (h *authHandler) isUserExists(c *gin.Context) {
	var input queries.EmailAddrRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("authHandler.init", "step", "ShouldBindJSON", "error", err)
		BadRequest(c, "invalid email")
		return
	}
	ctx := c.Request.Context()
	randKey := strings.ToLower(rand.Text()[:8])

	err := background.SendMail(ctx, h.cache, background.MailMessage{
		To:      []string{input.Email},
		Subject: "pass key",
		Msg:     randKey,
	})
	if err != nil {
		h.log.Debug("authHandler.init", "step", "SendMail", "error", err)
		Unprocessable(c, "failed to send passkey")
		return
	}

	if !h.store.IsUserExists(ctx, input.Email) {
		cacheKey := fmt.Sprintf("register-key:%s", input.Email)
		if err := h.cache.Set(ctx, cacheKey, randKey, h.registerKeyExpiration).Err(); err != nil {
			h.log.Debug("authHandler.init", "step", "Set register", "error", err)
			Unprocessable(c, "")
			return
		}
		NotFound(c, "register")
	} else {
		cacheKey := fmt.Sprintf("pass-key:%s", input.Email)
		if err := h.cache.Set(ctx, cacheKey, randKey, h.passKeyExpiration).Err(); err != nil {
			h.log.Debug("authHandler.init", "step", "Set login", "error", err)
			Unprocessable(c, "")
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func (h *authHandler) login(c *gin.Context) {
	var input loginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("authHandler.login", "step", "ShouldBindJSON", "error", err)
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	var randKey string
	cacheKey := fmt.Sprintf("pass-key:%s", input.Email)
	if err := h.cache.GetDel(ctx, cacheKey).Scan(&randKey); err != nil {
		LogCacheErr("Get", "authHandler.login", err)
		BadRequest(c, "Invalid email")
		return
	}
	if input.PassKey != randKey {
		BadRequest(c, "Invalid pass-key")
		return
	}
	user, err := h.store.Get(ctx, input.Email)
	if err != nil {
		NotFound(c, "")
		return
	}

	authToken, err := h.getAuthToken(ctx, user)
	if err != nil {
		BadRequest(c, "")
		return
	}
	var refreshKey string
	var refreshToken string
	var maxAge time.Duration
	if err = h.cache.Get(ctx, refreshKey).Scan(&refreshToken); err != nil {
		LogCacheErr("Get", "authHandler.getRefreshToken", err)
		u, err := uuid.NewV7()
		if err != nil {
			h.log.Warn("authHandler.login", "error", err)
			refreshKey = ""
			maxAge = 0
		} else {
			refreshToken = u.String()
			refreshKey = fmt.Sprintf("refresh:%s", refreshToken)
			authExpiration := time.Now().Add(h.authExpiration * time.Minute)
			maxAge = time.Until(authExpiration)
			if err = SetJSONCache(ctx, h.cache, refreshKey, maxAge, user); err != nil {
				LogCacheErr("Get", "authHandler.getRefreshToken", err)
			}
		}
	} else {
		ttl := h.cache.TTL(ctx, refreshKey)
		if err := ttl.Err(); err != nil {
			LogCacheErr("TTL", "authHandler.login", err)
			maxAge = 0
		} else {
			maxAge = ttl.Val()
		}
	}

	c.SetCookie("__Host-Http-Refresh", refreshKey, int(maxAge), "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"token": authToken})
}

func (h *authHandler) register(c *gin.Context) {
	var input registerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.log.Debug("authHandler.register", "step", "ShouldBindJSON", "error", err)
		BadRequest(c, "")
		return
	}

	ctx := c.Request.Context()
	var randKey string
	cacheKey := fmt.Sprintf("register-key:%s", input.Email)
	if err := h.cache.GetDel(ctx, cacheKey).Scan(&randKey); err != nil {
		LogCacheErr("Get", "authHandler.register", err)
		BadRequest(c, "Invalid email")
		return
	}
	if input.PassKey != randKey {
		BadRequest(c, "Invalid register-key")
		return
	}
	if err := h.store.Register(ctx, input.RegisterUserRequest); err != nil {
		Unprocessable(c, "failed to create request")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (h *authHandler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("__Host-Http-Refresh")
	if err != nil {
		Forbidden(c, "")
	}
	ctx := c.Request.Context()
	refreshKey := fmt.Sprintf("refresh:%s", refreshToken)
	var user queries.UserInfoResponse
	if err := GetJSONCache(ctx, h.cache, refreshKey, &user); err != nil {
		Unauthorized(c, "")
	}

	token, err := h.getAuthToken(ctx, user)
	if err != nil {
		BadRequest(c, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *authHandler) getAuthToken(ctx context.Context, user queries.UserInfoResponse) (token string, err error) {
	key := fmt.Sprintf("login:%s", user.ID)
	if err = h.cache.Get(ctx, key).Scan(&token); err != nil {
		LogCacheErr("Get", "authHandler.getToken", err)

		authExpiration := time.Now().Add(h.authExpiration * time.Minute)
		maxAge := time.Until(authExpiration)
		claims := auth.AuthClaims{
			ID:             user.ID,
			Email:          user.Email,
			PermissionType: user.PermissionType,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(authExpiration),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		}
		token, err = h.jwtToken.Encoder(claims)
		if err != nil {
			return
		}
		if err = h.cache.Set(ctx, key, token, maxAge).Err(); err != nil {
			LogCacheErr("Set", "authHandler.getToken", err)
		}
	}
	return
}

func (h *authHandler) logout(c *gin.Context) {
	claims := md.GetUserClaims(c)
	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("login:%s", claims.ID)
	if err := h.cache.Del(ctx, cacheKey).Err(); err != nil {
		LogCacheErr("Del", "authHandler.logout", err)
	}
	c.Status(http.StatusNoContent)
}

func (h *authHandler) healthz(c *gin.Context) {
	c.Status(http.StatusOK)
}
