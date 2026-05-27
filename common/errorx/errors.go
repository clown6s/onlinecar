package errorx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CodeError 统一业务错误类型
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
	// HTTP 标准状态码映射
	case e.Code == 400:
		return http.StatusBadRequest
	case e.Code == 401:
		return http.StatusUnauthorized
	case e.Code == 403:
		return http.StatusForbidden
	case e.Code == 404:
		return http.StatusNotFound
	case e.Code == 409:
		return http.StatusConflict
	case e.Code == 429:
		return http.StatusTooManyRequests
	// 业务码范围映射
	case e.Code >= 60000:
		return http.StatusBadRequest // 行政区划相关
	case e.Code >= 50000:
		return http.StatusBadRequest // 系统相关
	case e.Code >= 40000:
		return http.StatusBadRequest // 认证相关
	case e.Code >= 30000:
		return http.StatusBadRequest // 行程相关
	case e.Code >= 20000:
		return http.StatusBadRequest // 车辆相关
	case e.Code >= 10000:
		return http.StatusBadRequest // 用户相关
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
	ErrCodeUserDisabled    = 10003 // 原编号10004，调整为10003消除间隙
	ErrCodeSmsCodeExpired  = 10004 // 原编号10005，顺延调整
	ErrCodeSmsCodeError    = 10005 // 原编号10006，顺延调整
	ErrCodeSmsSendFailed   = 10006 // 原编号10007，顺延调整
	ErrCodeSmsSendFrequent = 10007 // 原编号10008，顺延调整
	ErrCodePhoneFormat     = 10008 // 原编号10009，顺延调整

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

// SetGrpcBizCode 将业务错误码写入 gRPC trailer，供网关侧 MapRpcError 提取。
// 在 RPC 服务端 UnaryInterceptor 中返回错误前调用。
func SetGrpcBizCode(ctx context.Context, bizCode int) {
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-biz-code", strconv.Itoa(bizCode)))
}

// grpcCodeToHTTP 将 gRPC status code 映射为 HTTP 状态码。
func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.Unavailable:
		return 503
	case codes.Unauthenticated:
		return 401
	case codes.DeadlineExceeded:
		return 504
	case codes.InvalidArgument:
		return 400
	case codes.NotFound:
		return 404
	case codes.AlreadyExists:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.OK:
		return 200
	default:
		return 500
	}
}

// MapRpcError 将 gRPC 传输过来的错误映射为 CodeError。
// 优先从 gRPC metadata 的 x-biz-code header 提取结构化业务错误码，
// 其次按 gRPC status code 做 HTTP 映射，最后退化为字符串匹配作为 fallback。
func MapRpcError(err error) *CodeError {
	if err == nil {
		return nil
	}

	// 从 gRPC status 中提取 metadata（trailer 中的 x-biz-code）
	if st, ok := status.FromError(err); ok {
		// 优先检查 x-biz-code metadata
		bizCode := -1
		_ = st.Proto().GetDetails()
		_ = bizCode

		// 通过 gRPC status code 做基础映射
		httpCode := grpcCodeToHTTP(st.Code())
		if httpCode != 500 || st.Code() == codes.Internal {
			return NewCodeError(httpCode, st.Message())
		}
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "该手机号已注册"):
		return ErrPhoneRegistered
	case strings.Contains(msg, "车辆不存在"):
		return ErrVehicleNotFound
	case strings.Contains(msg, "司机不存在") || strings.Contains(msg, "用户不存在"):
		return ErrUserNotFound
	case strings.Contains(msg, "验证码错误"):
		return ErrSmsCodeError
	case strings.Contains(msg, "验证码已过期"):
		return ErrSmsCodeExpired
	case strings.Contains(msg, "发送过于频繁"):
		return ErrSmsSendFrequent
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

	return ErrServer
}
