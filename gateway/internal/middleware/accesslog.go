package middleware

import (
	"bufio"
	"fmt"
	"net"
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
		logMsg := fmt.Sprintf("[ACCESS] method=%s path=%s status=%d duration=%v remote=%s request_id=%s content_length=%d ua=%s",
			r.Method, r.URL.Path, rw.statusCode, duration, r.RemoteAddr, rid, r.ContentLength, r.UserAgent())
		if duration > 200*time.Millisecond || rw.statusCode >= 300 {
			logx.Info(logMsg)
		} else {
			logx.Debug(logMsg)
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

// Flush 转发 http.Flusher 接口，支持 SSE/流式响应
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack 转发 http.Hijacker 接口，支持 WebSocket 升级
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
