package middlewares_test

import (
	md "generic-shop-sample/app/middlewares"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(md.SecurityHeadersMiddleware())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	header := w.Header()

	tests := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	for key, value := range tests {
		if got := header.Get(key); got != value {
			t.Errorf(`expected "%s=%s" but got "%s"`, key, value, got)
		}
	}
}
