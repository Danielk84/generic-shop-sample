package middlewares_test

import (
	"fmt"
	md "generic-shop-sample/app/middlewares"
	"generic-shop-sample/internal/auth"
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"
	tu "generic-shop-sample/tests/internal/testutils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware(t *testing.T) {
	ctx := t.Context()
	config := tu.ConfigTestSetup()
	deps := tu.ServiceDepsTestSetup(ctx, config)
	defer tu.CloseServieDepsTestSetup(deps)

	j := auth.JWTToken{
		JWTSecretKey: []byte(config.JWTSecretKey),
	}

	session := deps.DB.GetSession()
	log := logger.SetLogger(logger.LevelDebug, os.Stdout)
	if _, err := session.Exec(ctx, "TRUNCATE user_s.users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	store := queries.NewUserStore(session, log)
	if err := store.Create(ctx, queries.CreateUserRequest{
		LoginRequest:          queries.LoginRequest{Username: "auth_user", Password: "securePassword"},
		UserPermissionRequest: queries.UserPermissionRequest{PermissionType: queries.Admin, IsActive: true},
	}); err != nil {
		t.Errorf("failed to create user, %s", err)
	}
	user, _ := store.Get(ctx, "auth_user")

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
	tokenString, err := j.Encoder(baseClaims)
	if err != nil {
		t.Errorf("failed to encode token string, %s", err)
	}
	router := gin.Default()
	router.Use(md.AuthMiddleware(deps, log))
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
