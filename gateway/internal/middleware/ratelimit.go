package middleware

import (
	"net/http"
	"time"

	"onlinecar-gateway/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type RateLimitMiddleware struct {
	rate    int           // 每秒允许的请求数
	allowed chan struct{} // 令牌桶
}

func NewRateLimitMiddleware(rate int) *RateLimitMiddleware {
	rl := &RateLimitMiddleware{
		rate:    rate,
		allowed: make(chan struct{}, rate),
	}
	// 预填充令牌
	for i := 0; i < rate; i++ {
		rl.allowed <- struct{}{}
	}
	// 定时补充令牌
	go rl.refill()
	return rl
}

func (m *RateLimitMiddleware) refill() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		for i := 0; i < m.rate; i++ {
			select {
			case m.allowed <- struct{}{}:
			default:
			}
		}
	}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-m.allowed:
			next(w, r)
		default:
			logx.Infof("[RATELIMIT] 请求被限流, method=%s, path=%s, remote=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
			response.Error(w, 429, "请求过于频繁，请稍后再试")
		}
	}
}
