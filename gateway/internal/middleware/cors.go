package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
)

type CORSMiddleware struct {
	allowOrigins []string
}

func NewCORSMiddleware(origins ...string) *CORSMiddleware {
	if len(origins) == 0 {
		// 安全警告：无配置时不允许跨域，仅记录日志
		logx.Severe("[CORS] 未配置 AllowOrigins，跨域请求将被拒绝")
		origins = []string{} // 空列表 = 拒绝所有跨域
	}
	return &CORSMiddleware{allowOrigins: origins}
}

func (m *CORSMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range m.allowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			resolvedOrigin := m.resolveOrigin(origin)
			if resolvedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", resolvedOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (m *CORSMiddleware) resolveOrigin(origin string) string {
	for _, o := range m.allowOrigins {
		if o == "*" {
			// Credentials=true 时不能返回 *，回显请求的 Origin
			return origin
		}
		if o == origin {
			return origin
		}
	}
	return ""
}
