package middlewares_test

import (
	ts "generic-shop-sample/internal/testutils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	router := ts.RouterSetup()

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
			t.Errorf("expected %s=%s but got %s", key, value, got)
		}
	}
}
