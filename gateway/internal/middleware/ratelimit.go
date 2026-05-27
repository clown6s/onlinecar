package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"onlinecar-gateway/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter *rate.Limiter
	lastRef time.Time
}

type RateLimitMiddleware struct {
	rate            int // 每秒每个 IP 允许的请求数
	burst           int // 突发容量
	mu              sync.Mutex
	limiters        map[string]*ipLimiter
	cleanupInterval time.Duration
	maxIdle         time.Duration // IP 无请求多久后清理
}

func NewRateLimitMiddleware(r int) *RateLimitMiddleware {
	if r <= 0 {
		r = 1
	}
	burst := r
	m := &RateLimitMiddleware{
		rate:            r,
		burst:           burst,
		limiters:        make(map[string]*ipLimiter),
		cleanupInterval: 5 * time.Minute,
		maxIdle:         10 * time.Minute,
	}
	go m.cleanupLoop()
	return m
}

func (m *RateLimitMiddleware) getLimiter(ip string) *ipLimiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, ok := m.limiters[ip]
	if !ok {
		limiter = &ipLimiter{
			limiter: rate.NewLimiter(rate.Limit(m.rate), m.burst),
			lastRef: time.Now(),
		}
		m.limiters[ip] = limiter
	}
	return limiter
}

// cleanupLoop 定期清理长时间无请求的 IP limiter
func (m *RateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		count := 0
		for ip, limiter := range m.limiters {
			if now.Sub(limiter.lastRef) > m.maxIdle {
				delete(m.limiters, ip)
				count++
			}
		}
		m.mu.Unlock()
		if count > 0 {
			logx.Infof("[RATELIMIT] 清理了 %d 个过期 IP 限流器, 剩余 %d 个",
				count, len(m.limiters))
		}
	}
}

// GetClientIP 获取客户端真实 IP，支持 X-Forwarded-For 和 X-Real-IP
func GetClientIP(r *http.Request) string {
	// 优先 X-Real-IP（Nginx 常用）
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	// 其次 X-Forwarded-For（取第一个合法 IP）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.IndexByte(xff, ','); idx >= 0 {
			ip := strings.TrimSpace(xff[:idx])
			if net.ParseIP(ip) != nil {
				return ip
			}
		} else {
			if net.ParseIP(strings.TrimSpace(xff)) != nil {
				return strings.TrimSpace(xff)
			}
		}
	}
	return r.RemoteAddr
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		limiter := m.getLimiter(ip)

		m.mu.Lock()
		limiter.lastRef = time.Now()
		m.mu.Unlock()

		if limiter.limiter.Allow() {
			next(w, r)
		} else {
			logx.Infof("[RATELIMIT] 请求被限流, method=%s, path=%s, ip=%s",
				r.Method, r.URL.Path, ip)
			response.Error(w, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
		}
	}
}
