package middlewares

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsConfig is using for the CorsMiddleware.
// The first element in Origins (Origins[0]) is treated as the default origin.
// Methods and Headers are optional.
type CorsConfig struct {
	Origins, Methods, Headers []string
	Credentials               bool
}

// CorsMiddleware adds CORS headers to responses for browser clients.
func CorsMiddleware(cc *CorsConfig) gin.HandlerFunc {
	if len(cc.Methods) == 0 || len(cc.Origins) == 0 {
		panic("Methods and Origins must be set in CorsMiddleware")
	}

	origins := make(map[string]bool)
	for _, org := range cc.Origins {
		if _, err := url.ParseRequestURI(org); err != nil {
			panic(fmt.Errorf(`Invalid URI origin "%s": %s\n`, org, err))
		}
		origins[org] = true
	}

	credentials := strconv.FormatBool(cc.Credentials)
	methods := strings.Join(cc.Methods, ", ")
	headers := strings.Join(cc.Headers, ", ")

	return func(c *gin.Context) {
		originHeader := c.GetHeader("Origin")
		if _, err := url.ParseRequestURI(originHeader); err == nil && origins[originHeader] {
			c.Header("Access-Control-Allow-Origin", originHeader)
		} else {
			c.Header("Access-Control-Allow-Origin", cc.Origins[0])
		}
		c.Header("Access-Control-Allow-Credentials", credentials)
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			reqHeaders := c.GetHeader("Access-Control-Request-Headers")
			if reqHeaders != "" {
				c.Header("Access-Control-Allow-Headers", reqHeaders)
			} else if headers != "" {
				c.Header("Access-Control-Allow-Headers", headers)
			}

			reqMethod := c.GetHeader("Access-Control-Request-Method")
			if reqMethod != "" && slices.Contains(cc.Methods, reqMethod) {
				c.Header("Access-Control-Allow-Methods", reqMethod)
			} else if methods != "" {
				c.Header("Access-Control-Allow-Methods", methods)
			}

			c.Header("Access-Control-Max-Age", "600")
			c.Status(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
