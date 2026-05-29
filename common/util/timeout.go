package util

import (
	"context"
	"time"
)

// WithTimeout 创建带超时的子 context。
// 默认超时 5 秒，适用于单个 DB/Redis 操作。
func WithTimeout(parent context.Context, timeout ...time.Duration) (context.Context, context.CancelFunc) {
	d := 5 * time.Second
	if len(timeout) > 0 && timeout[0] > 0 {
		d = timeout[0]
	}
	return context.WithTimeout(parent, d)
}
