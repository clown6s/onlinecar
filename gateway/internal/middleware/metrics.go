package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestsTotal HTTP 请求总数（按 method、path、status）
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "onlinecar",
		Subsystem: "gateway",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	// httpRequestDurationSeconds HTTP 请求耗时（histogram）
	httpRequestDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "onlinecar",
		Subsystem: "gateway",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration in seconds",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	// httpRequestsInFlight 当前正在处理的请求数
	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "onlinecar",
		Subsystem: "gateway",
		Name:      "http_requests_in_flight",
		Help:      "Number of HTTP requests currently being processed",
	})

	// httpResponseSizeBytes 响应体大小
	httpResponseSizeBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "onlinecar",
		Subsystem: "gateway",
		Name:      "http_response_size_bytes",
		Help:      "HTTP response size in bytes",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B ~ 100MB
	}, []string{"method", "path"})
)

// MetricsMiddleware Prometheus 指标采集中间件
type MetricsMiddleware struct{}

func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (w *metricsResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += n
	return n, err
}

func (m *MetricsMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mrw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		next(mrw, r)

		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(mrw.statusCode)

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDurationSeconds.WithLabelValues(r.Method, path).Observe(duration)
		httpResponseSizeBytes.WithLabelValues(r.Method, path).Observe(float64(mrw.written))
	}
}

// normalizePath 将路径参数归一化，避免 label explosion
// 使用段数兜底策略：超过 4 段的路径统一归一化，防止随机路径参数打爆指标
func normalizePath(path string) string {
	// 固定模式优先匹配（精确度最高）
	patterns := []struct {
		prefix string
		result string
	}{
		{"/api/v1/district/children/", "/api/v1/district/children/:parentAdCode"},
		{"/api/v1/driver/", "/api/v1/driver/:id"},
		{"/api/v1/passenger/", "/api/v1/passenger/:id"},
		{"/api/v1/trip/", "/api/v1/trip/:id"},
	}
	for _, p := range patterns {
		if len(path) > len(p.prefix) && path[:len(p.prefix)] == p.prefix {
			return p.result
		}
	}
	// 兜底：路径段数超过 4 的归一化（如 /api/v1/xxx/yyy/zzz/... → /api/v1/:wildcard）
	segments := strings.Split(path, "/")
	if len(segments) > 5 { // "/api/v1/a/b/c" → segments=["", "api", "v1", "a", "b", "c"] → len=6 > 5
		// 保留前3段作为路径骨架
		return strings.Join(segments[:4], "/") + "/:wildcard"
	}
	return path
}
