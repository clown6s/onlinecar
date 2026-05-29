package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"onlinecar-common/ctxdata"

	"github.com/google/uuid"
)

// validRequestID 匹配字母数字、连字符、下划线，防止日志注入
var validRequestID = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (m *RequestIDMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" || len(requestID) > 64 || !validRequestID.MatchString(requestID) {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), ctxdata.RequestIDKey, requestID)
		next(w, r.WithContext(ctx))
	}
}
