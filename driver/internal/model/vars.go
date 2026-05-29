package model

import "onlinecar-common/errorx"

// ErrNotFound 记录不存在（使用 common 包统一定义，跨模块 errors.Is 兼容）
var ErrNotFound = errorx.ErrDBNotFound
