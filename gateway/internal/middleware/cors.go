package middleware

import (
	"net/http"
)

type CORSMiddleware struct {
	allowOrigins []string
}

func NewCORSMiddleware(origins ...string) *CORSMiddleware {
	if len(origins) == 0 {
		origins = []string{"*"}
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
			if o := m.resolveOrigin(origin); o != "" {
				w.Header().Set("Access-Control-Allow-Origin", o)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
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
			return "*"
		}
		if o == origin {
			return origin
		}
	}
	return ""
}
