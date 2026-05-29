package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"onlinecar-common/ctxdata"
	"onlinecar-common/errorx"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type JWTAuthMiddleware struct {
	Secret string
	Role   string // "driver" 或 "passenger"，为空时不校验角色
	Redis  *redis.Redis
}

func NewJWTAuthMiddleware(secret string, r *redis.Redis, role string) *JWTAuthMiddleware {
	if secret == "" {
		logx.Must(fmt.Errorf("JWT secret must not be empty"))
	}
	return &JWTAuthMiddleware{Secret: secret, Role: role, Redis: r}
}

func (m *JWTAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logx.Errorf("认证失败: 缺少Authorization header, method=%s, path=%s, ip=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
			httpx.Error(w, errorx.ErrAuthHeaderMissing)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logx.Errorf("认证失败: Authorization格式错误, method=%s, path=%s, ip=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
			httpx.Error(w, errorx.ErrAuthHeaderMalformed)
			return
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
			// 校验签名算法，防止 alg:none 绕过
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(m.Secret), nil
		})
		if err != nil || !token.Valid {
			logx.Errorf("认证失败: token无效或已过期, method=%s, path=%s, ip=%s, err=%v",
				r.Method, r.URL.Path, r.RemoteAddr, err)
			httpx.Error(w, errorx.ErrTokenInvalid)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logx.Errorf("认证失败: token claims解析失败, method=%s, path=%s, ip=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
			httpx.Error(w, errorx.ErrTokenInvalid)
			return
		}

		// 纵深防御：校验 role claim（防止 JWT secret 配置错误导致跨角色访问）
		if m.Role != "" {
			roleVal, ok := claims["role"].(string)
			if !ok || roleVal != m.Role {
				logx.Errorf("认证失败: token role不匹配, expect=%s, got=%s, method=%s, path=%s",
					m.Role, roleVal, r.Method, r.URL.Path)
				httpx.Error(w, errorx.ErrTokenInvalid)
				return
			}
		}

		// JWT 黑名单检查（支持主动失效 token，如用户禁用、强制下线等）
		if m.Redis != nil {
			tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(parts[1])))
			blacklistKey := "token:blacklist:" + tokenHash
			val, err := m.Redis.Get(blacklistKey)
			if err != nil {
				// Redis 不可用时选择安全策略：拒绝请求，防止黑名单失效
				if !strings.Contains(err.Error(), "nil") {
					logx.Errorf("Redis blacklist check failed: %v, method=%s, path=%s", err, r.Method, r.URL.Path)
					httpx.Error(w, errorx.ErrServer)
					return
				}
				// redis.Nil 表示 key 不存在，放行
			} else if val != "" {
				logx.Infof("认证失败: token已失效(黑名单), method=%s, path=%s, ip=%s",
					r.Method, r.URL.Path, r.RemoteAddr)
				httpx.Error(w, errorx.ErrTokenInvalid)
				return
			}
		}

		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			// 兼容多种 uid 类型，防止大 ID 精度丢失
			var uid int64
			switch v := claims["uid"].(type) {
			case float64:
				uid = int64(v)
			case json.Number:
				uid, _ = v.Int64()
			case int64:
				uid = v
			case string:
				uid, _ = strconv.ParseInt(v, 10, 64)
			default:
				logx.Errorf("认证失败: token uid类型异常, method=%s, path=%s, ip=%s, uidType=%T",
					r.Method, r.URL.Path, r.RemoteAddr, claims["uid"])
				httpx.Error(w, errorx.ErrTokenInvalid)
				return
			}

			logx.Infof("认证通过, uid=%d, method=%s, path=%s", uid, r.Method, r.URL.Path)
			ctx := ctxdata.SetUID(r.Context(), uid)
			next(w, r.WithContext(ctx))
			return
		}
		uid := int64(uidFloat)

		logx.Infof("认证通过, uid=%d, method=%s, path=%s", uid, r.Method, r.URL.Path)
		ctx := ctxdata.SetUID(r.Context(), uid)
		next(w, r.WithContext(ctx))
	}
}

// InvalidateToken 将 token 加入黑名单，使其立即失效。
// TTL 自动设为 token 的剩余有效时间（从 JWT exp claim 计算）。
// 调用方示例（用户禁用/强制下线时）：
//
//	middleware.InvalidateToken(redis, secret, tokenStr)
func InvalidateToken(rdb *redis.Redis, secret, tokenStr string) error {
	if rdb == nil || tokenStr == "" {
		return nil
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("token missing exp claim")
	}
	remaining := time.Until(time.Unix(int64(expFloat), 0))
	if remaining <= 0 {
		return nil // token 已过期，无需加入黑名单
	}
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenStr)))
	blacklistKey := "token:blacklist:" + tokenHash
	return rdb.SetexCtx(context.Background(), blacklistKey, "1", int(remaining.Seconds()))
}
