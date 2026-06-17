package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status_code"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	HttpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served.",
		},
	)

	HttpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"method", "path"},
	)

	// Database metrics
	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds.",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "table"},
	)

	DbConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "Number of open database connections.",
		},
	)

	DbConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of database connections currently in use.",
		},
	)

	DbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections.",
		},
	)

	DbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors.",
		},
		[]string{"operation", "table"},
	)

	// Business metrics
	ProductsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_created_total",
			Help: "Total number of products created.",
		},
	)

	ProductsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_deleted_total",
			Help: "Total number of products deleted.",
		},
	)

	OrdersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created.",
		},
	)

	OrderValueTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "order_value_total",
			Help: "Total value of all orders placed.",
		},
	)

	BusinessErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_errors_total",
			Help: "Total number of business logic errors.",
		},
		[]string{"type"},
	)

	// App info
	AppInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application build information.",
		},
		[]string{"version", "env"},
	)
)

// Init sets the app info metric.
func Init(version, env string) {
	AppInfo.WithLabelValues(version, env).Set(1)
}

// DBTimer measures database query duration.
type DBTimer struct {
	operation string
	table     string
	start     time.Time
}

func NewDBTimer(operation, table string) *DBTimer {
	return &DBTimer{operation: operation, table: table, start: time.Now()}
}

func (t *DBTimer) ObserveDuration() {
	DbQueryDuration.WithLabelValues(t.operation, t.table).Observe(time.Since(t.start).Seconds())
}
