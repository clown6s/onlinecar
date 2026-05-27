package errorx

import (
	"fmt"
	"net/http"
	"strings"
)

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

// HttpCode 返回对应的 HTTP 状态码
func (e *CodeError) HttpCode() int {
	switch {
	case e.Code >= 50000:
		return http.StatusServiceUnavailable
	case e.Code >= 40000:
		return http.StatusUnauthorized
	case e.Code >= 30000:
		return http.StatusBadRequest
	case e.Code >= 20000:
		return http.StatusBadRequest
	case e.Code >= 10000:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
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
	// 通用
	ErrCodeParam = 400

	// 用户相关 10001-10099
	ErrCodePhoneRegistered = 10001
	ErrCodeUserNotFound    = 10002
	ErrCodeUserDisabled    = 10004
	ErrCodeSmsCodeExpired  = 10005
	ErrCodeSmsCodeError    = 10006
	ErrCodeSmsSendFailed   = 10007
	ErrCodeSmsSendFrequent = 10008
	ErrCodePhoneFormat     = 10009

	// 车辆相关 20001-20099
	ErrCodeVehicleNotFound = 20001
	ErrCodeVehicleBound    = 20002

	// 行程相关 30001-30099
	ErrCodeTripNotFound      = 30001
	ErrCodeTripStatusInvalid = 30002
	ErrCodeTripCreateFailed  = 30003

	// 行政区划相关 60001-60099
	ErrCodeDistrictNotFound = 60001

	// 认证相关 40001-40099
	ErrCodeTokenExpired        = 40001
	ErrCodeTokenInvalid        = 40002
	ErrCodeAuthHeaderMissing   = 40003
	ErrCodeAuthHeaderMalformed = 40004

	// 系统相关 50001-50099
	ErrCodeServiceUnavailable = 50001
	ErrCodeRpcCallFailed      = 50002
	ErrCodeDatabaseError      = 50003
)

var (
	// 用户
	ErrPhoneRegistered = NewCodeError(ErrCodePhoneRegistered, "该手机号已注册")
	ErrUserNotFound    = NewCodeError(ErrCodeUserNotFound, "用户不存在")
	ErrUserDisabled    = NewCodeError(ErrCodeUserDisabled, "用户已禁用")
	ErrSmsCodeExpired  = NewCodeError(ErrCodeSmsCodeExpired, "验证码已过期")
	ErrSmsCodeError    = NewCodeError(ErrCodeSmsCodeError, "验证码错误")
	ErrSmsSendFailed   = NewCodeError(ErrCodeSmsSendFailed, "发送验证码失败")
	ErrSmsSendFrequent = NewCodeError(ErrCodeSmsSendFrequent, "发送过于频繁，请60秒后重试")
	ErrPhoneFormat     = NewCodeError(ErrCodePhoneFormat, "手机号格式不正确")

	// 车辆
	ErrVehicleNotFound = NewCodeError(ErrCodeVehicleNotFound, "车辆不存在")
	ErrVehicleBound    = NewCodeError(ErrCodeVehicleBound, "车辆已绑定其他司机")

	// 行程
	ErrTripNotFound      = NewCodeError(ErrCodeTripNotFound, "行程不存在")
	ErrTripStatusInvalid = NewCodeError(ErrCodeTripStatusInvalid, "行程状态不允许此操作")
	ErrTripCreateFailed  = NewCodeError(ErrCodeTripCreateFailed, "创建行程失败")

	// 行政区划
	ErrDistrictNotFound = NewCodeError(ErrCodeDistrictNotFound, "行政区划不存在")

	// 认证
	ErrTokenExpired        = NewCodeError(ErrCodeTokenExpired, "token已过期")
	ErrTokenInvalid        = NewCodeError(ErrCodeTokenInvalid, "token无效")
	ErrAuthHeaderMissing   = NewCodeError(ErrCodeAuthHeaderMissing, "缺少Authorization header")
	ErrAuthHeaderMalformed = NewCodeError(ErrCodeAuthHeaderMalformed, "Authorization格式错误")

	// 系统
	ErrServiceUnavailable = NewCodeError(ErrCodeServiceUnavailable, "服务不可用")
	ErrRpcCallFailed      = NewCodeError(ErrCodeRpcCallFailed, "RPC调用失败")
	ErrDatabaseError      = NewCodeError(ErrCodeDatabaseError, "数据库错误")
)

// MapRpcError 将 gRPC 传输过来的错误映射为 CodeError。
// 跨 gRPC 传输后 CodeError 类型会丢失为普通 error，通过错误消息中的关键词匹配恢复业务错误码。
func MapRpcError(err error) *CodeError {
	if err == nil {
		return nil
	}
	msg := err.Error()

	switch {
	case strings.Contains(msg, "该手机号已注册"):
		return ErrPhoneRegistered
	case strings.Contains(msg, "不存在"):
		return ErrUserNotFound
	case strings.Contains(msg, "已禁用"):
		return ErrUserDisabled
	case strings.Contains(msg, "无效的"):
		return ErrParam
	case strings.Contains(msg, "不能为空"):
		return ErrParam
	case strings.Contains(msg, "创建行程失败"):
		return ErrTripCreateFailed
	case strings.Contains(msg, "坐标无效"):
		return ErrParam
	case strings.Contains(msg, "乘客账号已禁用"):
		return ErrUserDisabled
	case strings.Contains(msg, "行政区划不存在"):
		return ErrDistrictNotFound
	}

	// 默认返回服务不可用（不暴露内部错误细节）
	return ErrServiceUnavailable
}
