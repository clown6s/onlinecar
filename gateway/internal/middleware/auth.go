package middleware

import (
	"errors"
	"fmt"
	"net/http"
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
			logx.Errorf("认证失败: token缺少uid, method=%s, path=%s, ip=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
			httpx.Error(w, errorx.ErrTokenInvalid)
			return
		}
		uid := int64(uidFloat)

		logx.Infof("认证通过, uid=%d, method=%s, path=%s", uid, r.Method, r.URL.Path)
		ctx := ctxdata.SetUID(r.Context(), uid)
		next(w, r.WithContext(ctx))
	}
}
