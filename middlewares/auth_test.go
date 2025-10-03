package middlewares_test

import (
	"fmt"
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal/auth"
	md "generic-shop-sample/middlewares"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware(t *testing.T) {
	baseClaims := auth.AuthClaims{
		ID:             1,
		Username:       "user",
		PermissionType: queries.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	tokenString, err := auth.TokenEncoder(baseClaims)
	if err != nil {
		t.Errorf("failed to encode token string, %s", err)
	}
	router := gin.New()
	router.Use(md.AuthMiddleware())
	router.GET("/", func(c *gin.Context) {
		claims := md.GetUserClaims(c)
		if claims.ID == baseClaims.ID &&
			claims.Username == baseClaims.Username &&
			claims.PermissionType == baseClaims.PermissionType {
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusForbidden)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(`failed to auth user, status="%d"`, w.Code)
	}
}
