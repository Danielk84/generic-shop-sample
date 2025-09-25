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
		tokenString, ok := getToken(c.GetHeader("Authorization"), "Bearer ")
		if !ok {
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

func getToken(s string, prefix string) (string, bool) {
	if strings.HasPrefix(s, prefix) {
		return strings.TrimPrefix(s, prefix), true
	}
	return "", false
}

func GetUserClaims(c *gin.Context) *auth.AuthClaims {
	return c.Request.Context().Value(userKey).(*auth.AuthClaims)
}
