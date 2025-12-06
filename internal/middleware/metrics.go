package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP 请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 请求持续时间（秒）",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求大小
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP 请求大小（字节）",
			Buckets: []float64{100, 500, 1000, 2500, 5000, 10000, 25000, 50000, 100000},
		},
		[]string{"method", "path"},
	)

	// HTTP 响应大小
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP 响应大小（字节）",
			Buckets: []float64{100, 500, 1000, 2500, 5000, 10000, 25000, 50000, 100000},
		},
		[]string{"method", "path", "status"},
	)

	// 当前正在处理的请求数
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "当前正在处理的 HTTP 请求数",
		},
	)
)

// Metrics Prometheus 监控中间件
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查和监控端点
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 记录请求大小
		requestSize := float64(c.Request.ContentLength)
		if requestSize > 0 {
			httpRequestSize.WithLabelValues(method, path).Observe(requestSize)
		}

		// 增加正在处理的请求数
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// 处理请求
		c.Next()

		// 计算持续时间
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// 记录指标
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)

		// 记录响应大小
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, path, status).Observe(responseSize)
		}
	}
}
