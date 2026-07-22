package middlewares

import (
	"context"
	"generic-shop-sample/app"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type userContextKey struct{}

var userKey = userContextKey{}

func AuthMiddleware(deps *app.ServiceDeps, log logger.Logger) gin.HandlerFunc {
	jwtToken := auth.JWTToken{JWTSecretKey: []byte(deps.Config.JWTSecretKey)}
	return func(c *gin.Context) {
		var tokenString string
		if cookie, err := c.Cookie("__Host-auth-token"); err == nil {
			tokenString = cookie
		} else {
			var found bool
			tokenString, found = strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
			if !found {
				c.Header("WWW-Authenticate", "Bearer")
				c.AbortWithStatus(http.StatusUnauthorized)
				log.Debug("AuthMiddleware", "error", "invalid cookie")
				return
			}
		}
		claims, err := jwtToken.Decoder(tokenString)
		if err != nil {
			log.Debug("AuthMiddleware:invalid token", "token", tokenString)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		store := queries.NewUserStore(deps.DB.GetSession(), log)
		ctx := c.Request.Context()
		if !store.IsValidUser(ctx, queries.ValidUserRequest{ID: claims.ID, Username: claims.Username, PermissionType: claims.PermissionType}) {
			c.AbortWithStatus(http.StatusNotFound)
			log.Debug("AuthMiddleware:invalid user", "username", claims.Username)
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
