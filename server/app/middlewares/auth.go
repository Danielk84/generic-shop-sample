package middlewares

import (
	"context"
	"fmt"
	"generic-shop-sample/app"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/cache"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type userContextKey struct{}

var userKey = userContextKey{}

func AuthMiddleware(deps *app.ServiceDeps, log logger.Logger) gin.HandlerFunc {
	jwtToken := auth.JWTToken{Log: log, JWTSecretKey: []byte(deps.Config.JWTSecretKey)}
	store := queries.NewUserStore(deps.DB.GetSession(), log)
	cache := deps.Cache.GetCache(cache.UsersCache)
	return func(c *gin.Context) {
		tokenString, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatus(http.StatusUnauthorized)
			log.Debug("AuthMiddleware", "error", "invalid cookie")
			return
		}

		claims, err := jwtToken.Decoder(tokenString)
		if err != nil {
			log.Debug("AuthMiddleware:invalid token", "token", tokenString)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx := c.Request.Context()
		isValid := store.IsValidUser(ctx,
			queries.ValidUserRequest{
				ID:                claims.ID,
				EmailAddrRequest:  queries.EmailAddrRequest{Email: claims.Email},
				PermissionRequest: queries.PermissionRequest{PermissionType: claims.PermissionType}})
		if !isValid {
			log.Debug("AuthMiddleware:invalid user", "email", claims.Email)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		cacheKey := fmt.Sprintf("login:%s", claims.ID)
		exists := cache.Exists(ctx, cacheKey)
		err = exists.Err()
		count := exists.Val()
		if err != nil || count != 1 {
			log.Debug("AuthMiddleware:exists",
				"error", err,
				"count", count)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		ctx = context.WithValue(ctx, userKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetUserClaims(c *gin.Context) *auth.AuthClaims {
	claims, ok := c.Request.Context().Value(userKey).(*auth.AuthClaims)
	if ok {
		return claims
	}
	return nil
}
