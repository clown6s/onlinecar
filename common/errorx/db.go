package errorx

import "errors"

// ErrDBNotFound 数据库记录不存在（供 driver/passenger/district 模块共用）
var ErrDBNotFound = errors.New("record not found")
