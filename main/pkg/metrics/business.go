package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ttpos",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"status", "method", "route"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "ttpos",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	OrdersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ttpos",
			Name:      "orders_total",
			Help:      "Total number of orders created",
		},
		[]string{"source", "order_type"},
	)
	PaymentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "ttpos",
			Name:      "payments_total",
			Help:      "Total number of payments processed",
		},
		[]string{"payment_method", "status"},
	)
)

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			OrdersTotal,
			PaymentsTotal,
		)
	})
}

// Middleware returns a Gin middleware that records HTTP metrics.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmapped"
		}
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		elapsed := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(status, method, route).Inc()
		httpRequestDuration.WithLabelValues(method, route).Observe(elapsed)
	}
}

// Handler returns the promhttp handler as a Gin handler.
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
