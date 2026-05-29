package ctxdata

import "context"

type ctxKey string

const UID ctxKey = "uid"

const RequestIDKey ctxKey = "requestId"

func SetUID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, UID, uid)
}

// GetUID 从 context 中提取用户 ID。
// 返回 (uid, true) 表示已认证，(0, false) 表示未登录或 context 中无值。
// 业务约定：uid >= 1 为有效用户 ID，0 表示未设置。
func GetUID(ctx context.Context) (int64, bool) {
	if v, ok := ctx.Value(UID).(int64); ok && v > 0 {
		return v, true
	}
	return 0, false
}

func SetRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, RequestIDKey, rid)
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}
