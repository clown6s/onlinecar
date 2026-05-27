package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"onlinecar-common/ctxdata"
	"onlinecar-common/errorx"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type JWTAuthMiddleware struct {
	Secret string
}

func NewJWTAuthMiddleware(secret string) *JWTAuthMiddleware {
	if secret == "" {
		logx.Must(fmt.Errorf("JWT secret must not be empty"))
	}
	return &JWTAuthMiddleware{Secret: secret}
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
