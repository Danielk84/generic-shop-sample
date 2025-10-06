package middlewares

import (
	"fmt"
	"generic-shop-sample/internal"

	"github.com/gin-gonic/gin"
)

func RequestLoggerMiddleware(fp string) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			return fmt.Sprintf(`[%s] | %s | %d | %s | size="%d" | latency="%s" | err="%s" \n`,
				params.TimeStamp,
				params.Method,
				params.StatusCode,
				params.Path,
				params.BodySize,
				params.Latency,
				params.ErrorMessage,
			)
		},
		Output: internal.CreateLogFile(fp),
	})
}
