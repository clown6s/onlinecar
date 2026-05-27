package middleware

import (
	"net/http"
	"strconv"
	"time"
)

type SecurityHeadersMiddleware struct{}

func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

func (m *SecurityHeadersMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 防止 MIME 类型嗅探
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// 防止点击劫持
		w.Header().Set("X-Frame-Options", "DENY")
		// XSS 保护
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// 控制 referrer 信息泄露
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// HSTS（仅 HTTPS）
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age="+strconv.Itoa(int(365*24*time.Hour.Seconds()))+"; includeSubDomains")
		}

		next(w, r)
	}
}
