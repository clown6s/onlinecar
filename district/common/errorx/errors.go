package errorx

import "fmt"

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewCodeError(code int, msg string) *CodeError {
	return &CodeError{Code: code, Msg: msg}
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("code:%d, msg:%s", e.Code, e.Msg)
}

// ========== HTTP 通用错误 ==========
var (
	ErrServer    = NewCodeError(500, "服务器内部错误")
	ErrParam     = NewCodeError(400, "参数错误")
	ErrNotFound  = NewCodeError(404, "资源不存在")
	ErrAuth      = NewCodeError(401, "未授权")
	ErrForbidden = NewCodeError(403, "无权限访问")
)

// ========== 业务错误码 ==========
const (
	ErrCodeDistrictNotFound = 10010
)

var (
	ErrDistrictNotFound = NewCodeError(ErrCodeDistrictNotFound, "行政区划不存在")
)
