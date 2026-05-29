package middleware

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecurityHeadersMiddleware struct {
	// TrustedProxySubnets 可信代理网段（CIDR），仅信任这些来源的 X-Forwarded-Proto
	TrustedProxySubnets []string
	trustedCIDRs        []*net.IPNet
}

func NewSecurityHeadersMiddleware(trustedProxySubnets ...string) *SecurityHeadersMiddleware {
	m := &SecurityHeadersMiddleware{
		TrustedProxySubnets: trustedProxySubnets,
	}
	for _, cidr := range trustedProxySubnets {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			logx.Errorf("[SECURITY] 解析可信代理网段失败, cidr=%s, err=%v", cidr, err)
			continue
		}
		m.trustedCIDRs = append(m.trustedCIDRs, network)
	}
	return m
}

func (m *SecurityHeadersMiddleware) isTrustedProto(r *http.Request) bool {
	// 1. 直连 TLS
	if r.TLS != nil {
		return true
	}
	// 2. 未配置可信代理，不信任代理头
	if len(m.trustedCIDRs) == 0 {
		return false
	}
	// 3. 仅当来源 IP 属于可信代理时，才信任 X-Forwarded-Proto
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range m.trustedCIDRs {
		if network.Contains(ip) && r.Header.Get("X-Forwarded-Proto") == "https" {
			return true
		}
	}
	return false
}

func (m *SecurityHeadersMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 防止 MIME 类型嗅探
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// 防止点击劫持
		w.Header().Set("X-Frame-Options", "DENY")
		// 控制 referrer 信息泄露
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// 内容安全策略（默认宽松，生产环境根据前端资源按需收紧）
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		// HSTS（仅 HTTPS，需确认来源可信以防止伪造）
		if m.isTrustedProto(r) {
			w.Header().Set("Strict-Transport-Security", "max-age="+strconv.Itoa(int(365*24*time.Hour.Seconds()))+"; includeSubDomains")
		}

		next(w, r)
	}
}
