package ctxdata

import "context"

type ctxKey string

const UID ctxKey = "uid"

func SetUID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, UID, uid)
}

func GetUID(ctx context.Context) int64 {
	if v, ok := ctx.Value(UID).(int64); ok {
		return v
	}
	return 0
}
