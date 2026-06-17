package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger writes a structured one-line log per request to stdout.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		fmt.Printf("time=%s level=info method=%s path=%s status=%d latency=%s ip=%s\n",
			time.Now().Format(time.RFC3339),
			method, path, status, latency, ip,
		)
	}
}
