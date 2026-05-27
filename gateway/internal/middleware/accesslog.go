package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type AccessLogMiddleware struct{}

func NewAccessLogMiddleware() *AccessLogMiddleware {
	return &AccessLogMiddleware{}
}

func (m *AccessLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: 200}
		next(rw, r)

		duration := time.Since(start)
		rid := r.Header.Get("X-Request-ID")

		// 慢请求（>200ms）或非 2xx 响应才打 Info，其余 Debug
		if duration > 200*time.Millisecond || rw.statusCode >= 300 {
			logx.Infof("[ACCESS] method=%s path=%s status=%d duration=%v remote=%s request_id=%s",
				r.Method, r.URL.Path, rw.statusCode, duration, r.RemoteAddr, rid)
		} else {
			logx.Debugf("[ACCESS] method=%s path=%s status=%d duration=%v remote=%s request_id=%s",
				r.Method, r.URL.Path, rw.statusCode, duration, r.RemoteAddr, rid)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
