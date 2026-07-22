package middlewares_test

import (
	md "generic-shop-sample/app/middlewares"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestLimit := 5
	timePeriod := 5 * time.Second
	cleanupPeriod := 10 * time.Second

	rl := md.NewRateLimiter(t.Context(), requestLimit, timePeriod, cleanupPeriod)

	router := gin.New()
	router.Use(rl.RateLimiterMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	for range requestLimit {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf(`expected status code "%d", bug got "%d"`, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("bad rate Limiter")
	}

	time.Sleep(6 * time.Second)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(`expected ending of limitation, but got "%d"`, w.Code)
	}

	time.Sleep(20 * time.Second)
	if c := rl.GetLen(); c != 0 {
		t.Errorf(`bad removeExpired worker, count of stored user "%d"`, c)
	}
}
