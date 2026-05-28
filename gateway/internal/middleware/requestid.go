package middleware

import (
	"context"
	"net/http"
	"strings"

	"onlinecar-common/ctxdata"

	"github.com/google/uuid"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (m *RequestIDMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" || len(requestID) > 64 {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), ctxdata.RequestIDKey, requestID)
		next(w, r.WithContext(ctx))
	}
}
