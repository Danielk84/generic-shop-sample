package middlewares

import (
	"context"
	"generic-shop-sample/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type userContextKey struct{}

var userKey = userContextKey{}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := auth.TokenDecoder(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		ctx := context.WithValue(c.Request.Context(), userKey, claims)
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
