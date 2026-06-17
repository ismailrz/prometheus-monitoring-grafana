package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	appmetrics "github.com/prometheus-stack/backend/pkg/metrics"
)

// Metrics instruments every request with Prometheus counters/histograms.
// Uses gin's FullPath() so route params (e.g. /products/:id) are normalised —
// preventing high-cardinality label explosion from real IDs in the path.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		appmetrics.HttpRequestsInFlight.Inc()
		defer appmetrics.HttpRequestsInFlight.Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		size := float64(c.Writer.Size())
		if size < 0 {
			size = 0
		}

		appmetrics.HttpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		appmetrics.HttpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
		appmetrics.HttpResponseSize.WithLabelValues(c.Request.Method, path).Observe(size)
	}
}
