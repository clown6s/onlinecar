package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"onlinecar-gateway/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

// rateLimitStoppers 收集所有限流中间件的 Stop 方法，用于优雅关闭
var (
	rateLimitStoppers   []func()
	rateLimitStoppersMu sync.Mutex
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
	maxEntries      int           // map 最大条目数（防止 DDoS 下 OOM）
	trustedSubnets  []*net.IPNet  // 可信代理网段
	ctx             context.Context
	cancel          context.CancelFunc
}

const defaultMaxEntries = 100000 // 10 万 IP 上限

func NewRateLimitMiddleware(r int, trustedSubnets ...string) *RateLimitMiddleware {
	if r <= 0 {
		r = 1
	}
	burst := r * 2 // 允许短时突发，避免正常快速操作被误限流
	ctx, cancel := context.WithCancel(context.Background())
	m := &RateLimitMiddleware{
		rate:            r,
		burst:           burst,
		limiters:        make(map[string]*ipLimiter),
		cleanupInterval: 5 * time.Minute,
		maxIdle:         10 * time.Minute,
		maxEntries:      defaultMaxEntries,
		ctx:             ctx,
		cancel:          cancel,
	}
	// 解析可信代理网段（CIDR 格式）
	for _, cidr := range trustedSubnets {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			m.trustedSubnets = append(m.trustedSubnets, network)
		} else {
			logx.Errorf("[RATELIMIT] 解析可信代理网段失败, cidr=%s, err=%v", cidr, err)
		}
	}
	go m.cleanupLoop()
	rateLimitStoppersMu.Lock()
	rateLimitStoppers = append(rateLimitStoppers, m.Stop)
	rateLimitStoppersMu.Unlock()
	return m
}

// isTrustedProxy 检查请求来源 IP 是否属于可信代理网段
func (m *RateLimitMiddleware) isTrustedProxy(remoteAddr string) bool {
	if len(m.trustedSubnets) == 0 {
		// 未配置可信代理时，不信任任何代理头
		return false
	}
	// 去除端口号
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range m.trustedSubnets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (m *RateLimitMiddleware) getLimiter(ip string) *ipLimiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 防止 DDoS 下 map 无限增长：淘汰最旧条目（LRU），而非直接拒绝新 IP
	if len(m.limiters) >= m.maxEntries {
		var oldestIP string
		var oldestTime time.Time
		for ip, l := range m.limiters {
			if oldestIP == "" || l.lastRef.Before(oldestTime) {
				oldestIP = ip
				oldestTime = l.lastRef
			}
		}
		delete(m.limiters, oldestIP)
	}

	limiter, ok := m.limiters[ip]
	if !ok {
		limiter = &ipLimiter{
			limiter: rate.NewLimiter(rate.Limit(m.rate), m.burst),
			lastRef: time.Now(),
		}
		m.limiters[ip] = limiter
	} else {
		limiter.lastRef = time.Now()
	}
	return limiter
}

// cleanupLoop 定期清理长时间无请求的 IP limiter
func (m *RateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
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
}

// Stop 停止 cleanupLoop goroutine
func (m *RateLimitMiddleware) Stop() {
	m.cancel()
}

// StopRateLimiters 停止所有已注册的限流中间件的 cleanupLoop goroutine
func StopRateLimiters() {
	rateLimitStoppersMu.Lock()
	stoppers := make([]func(), len(rateLimitStoppers))
	copy(stoppers, rateLimitStoppers)
	rateLimitStoppersMu.Unlock()
	for _, stop := range stoppers {
		stop()
	}
}

// GetClientIP 获取客户端真实 IP
// 仅当请求来源属于可信代理网段时，才使用 X-Real-IP / X-Forwarded-For 头
// 否则直接使用 RemoteAddr，防止客户端伪造代理头绕过限流
func GetClientIP(r *http.Request, trustedSubnets ...*net.IPNet) string {
	isTrusted := len(trustedSubnets) > 0
	if isTrusted {
		// 去除端口号
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		ip := net.ParseIP(host)
		if ip != nil {
			isTrusted = false
			for _, network := range trustedSubnets {
				if network.Contains(ip) {
					isTrusted = true
					break
				}
			}
		}
	}

	if isTrusted {
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
	}

	// 不可信来源或无代理头时，使用 RemoteAddr（去除端口号）
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return ip
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r, m.trustedSubnets...)
		limiter := m.getLimiter(ip)

		if limiter == nil || !limiter.limiter.Allow() {
			logx.Infof("[RATELIMIT] 请求被限流, method=%s, path=%s, ip=%s",
				r.Method, r.URL.Path, ip)
			response.Error(w, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
			return
		}
		next(w, r)
	}
}
