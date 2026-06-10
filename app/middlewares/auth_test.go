package middlewares_test

import (
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/auth"
	tu "generic-shop-sample/internal/testutils"
	"generic-shop-sample/storage/database"
	"generic-shop-sample/storage/queries"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware(t *testing.T) {
	db := tu.DBManagerSetup(t.Context())
	defer db.Close()

	us := queries.NewUserStore(database.GetSession())
	if err := us.Create(t.Context(), &queries.CreateUserRequest{
		LoginRequest:          queries.LoginRequest{Username: "auth_user", Password: "securePassword"},
		UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Admin, IsActive: true},
	}); err != nil {
		t.Errorf("failed to create user, %s", err)
	}
	user, _ := us.Get(t.Context(), "auth_user")

	baseClaims := auth.AuthClaims{
		ID:             user.ID,
		Username:       user.Username,
		PermissionType: user.PermissionType,
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
	router := gin.Default()
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

	// test authorization header
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(`failed to auth user with authorization header, status="%d"`, w.Code)
	}

	// test __Host-auth-token cookie
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", fmt.Sprintf("__Host-auth-token=%s", tokenString))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(`failed to auth user with cookie, status=%d`, w.Code)
	}
}
