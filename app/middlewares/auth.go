package middlewares

import (
	"context"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type userContextKey struct{}

var userKey = userContextKey{}

func AuthMiddleware() gin.HandlerFunc {
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
				return
			}
		}
		claims, err := auth.TokenDecoder(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		log := logger.GetLogger()
		us := queries.NewUserStore(database.GetSession(), log)
		ctx := c.Request.Context()
		if !us.IsValidUser(ctx, &queries.ValidUserRquest{ID: claims.ID, Username: claims.Username, PermissionType: claims.PermissionType}) {
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
