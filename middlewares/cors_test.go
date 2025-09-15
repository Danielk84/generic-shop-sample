package middlewares_test

import (
	md "generic-shop-sample/middlewares"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseMethod := []string{http.MethodGet, http.MethodPost}
	baseHeaders := []string{"Content-Length"}

	cc := &md.CorsConfig{
		Origins:     []string{"http://example.com", "http://localhost:3000"},
		Methods:     baseMethod,
		Headers:     baseHeaders,
		Credentials: true,
	}

	router := gin.New()
	router.Use(md.CorsMiddleware(cc))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.reqMethod, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.wantMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.wantMethod)
			}
			if tt.wantHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tt.wantHeaders)
			}

			router.ServeHTTP(w, req)
			if w.Code != tt.code {
				t.Errorf("expected status code %d, but got %d", tt.code, w.Code)
			}

			header := w.Header()
			if header.Get("Access-Control-Allow-Credentials") != tt.credentials {
				t.Errorf(`Invalid allow credentials, expected "%s"`, tt.credentials)
			}
			if got := header.Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf(`expected allow origin "%s", but got "%s"`, tt.wantOrigin, got)
			}

			if tt.reqMethod == http.MethodOptions {
				if got := header.Get("Access-Control-Allow-Headers"); tt.wantHeaders != "" && got != tt.wantHeaders {
					t.Errorf(`expected allow headers "%s", but got "%s"`, tt.wantHeaders, got)
				} else if tt.wantHeaders == "" && got != strings.Join(baseHeaders, ", ") {
					t.Errorf(`invalid allow headers "%s"`, got)
				}

				if got := header.Get("Access-Control-Allow-Methods"); tt.wantMethod != "" && got != tt.gotMethods {
					t.Errorf(`expected allow methods "%s", but got "%s"`, tt.gotMethods, got)
				} else if tt.wantMethod == "" && got != tt.gotMethods {
					t.Errorf(`invalid allow methods "%s"`, got)
				}

				if got := header.Get("Access-Control-Max-Age"); got != "600" {
					t.Errorf(`invalid cors max age "%s"`, got)
				}
			}

		})
	}
}

var tests = []struct {
	name        string
	reqMethod   string
	origin      string
	wantOrigin  string
	wantMethod  string
	gotMethods  string
	wantHeaders string
	credentials string
	code        int
}{
	{
		"test with method OPTIONS, origin, method, headers",
		http.MethodOptions,
		"http://localhost:3000",
		"http://localhost:3000",
		"POST",
		"POST",
		"Accept-Encoding",
		"true",
		http.StatusNoContent,
	},
	{
		"test with method OPTIONS and without origin, method, headers",
		http.MethodOptions,
		"",
		"http://example.com",
		"",
		"GET, POST",
		"",
		"true",
		http.StatusNoContent,
	},
	{
		"test with method GET and origin",
		http.MethodGet,
		"http://example.com",
		"http://example.com",
		"",
		"",
		"",
		"true",
		http.StatusOK,
	},
	{
		"test with invalid mwthod and origin",
		http.MethodConnect,
		"http://invalid-origin.fun",
		"http://example.com",
		"",
		"",
		"",
		"true",
		http.StatusNotFound,
	},
}
