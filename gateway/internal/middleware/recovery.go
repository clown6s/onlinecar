package middleware

import (
	"net/http"
	"runtime/debug"

	"onlinecar-common/errorx"
	"onlinecar-gateway/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type PanicRecoveryMiddleware struct{}

func NewPanicRecoveryMiddleware() *PanicRecoveryMiddleware {
	return &PanicRecoveryMiddleware{}
}

func (m *PanicRecoveryMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logx.Errorf("panic recovered: %v\n%s", rec, debug.Stack())
				response.Errorx(w, errorx.ErrServer)
			}
		}()
		next(w, r)
	}
}
