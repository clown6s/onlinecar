package errorx

import (
	"context"
	"errors"
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
	// HTTP 标准状态码映射（仅限语义与 HTTP 一致的错误码）
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
// ErrCodeParam 参数错误（使用 10xxx 范围，避免与 HTTP 401/403/404 等标准状态码冲突）
// 注意：ErrAuth(401)、ErrForbidden(403)、ErrNotFound(404) 保留 HTTP 标准码，因其语义与 HTTP 一致
const (
	ErrCodeParam = 10011

	// 用户相关 10001-10099
	ErrCodePhoneRegistered    = 10001
	ErrCodeUserNotFound       = 10002
	ErrCodeUserDisabled       = 10003 // 原编号10004，调整为10003消除间隙
	ErrCodeSmsCodeExpired     = 10004 // 原编号10005，顺延调整
	ErrCodeSmsCodeError       = 10005 // 原编号10006，顺延调整
	ErrCodeSmsSendFailed      = 10006 // 原编号10007，顺延调整
	ErrCodeSmsSendFrequent    = 10007 // 原编号10008，顺延调整
	ErrCodePhoneFormat        = 10008 // 原编号10009，顺延调整
	ErrCodeSmsTooManyAttempts = 10009 // 验证码错误次数过多
	ErrCodeLoginFailed        = 10010 // 登录失败（手机号或验证码错误）

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
	ErrPhoneRegistered    = NewCodeError(ErrCodePhoneRegistered, "该手机号已注册")
	ErrUserNotFound       = NewCodeError(ErrCodeUserNotFound, "用户不存在")
	ErrUserDisabled       = NewCodeError(ErrCodeUserDisabled, "用户已禁用")
	ErrSmsCodeExpired     = NewCodeError(ErrCodeSmsCodeExpired, "验证码已过期")
	ErrSmsCodeError       = NewCodeError(ErrCodeSmsCodeError, "验证码错误")
	ErrSmsSendFailed      = NewCodeError(ErrCodeSmsSendFailed, "发送验证码失败")
	ErrSmsSendFrequent    = NewCodeError(ErrCodeSmsSendFrequent, "发送过于频繁，请60秒后重试")
	ErrPhoneFormat        = NewCodeError(ErrCodePhoneFormat, "手机号格式不正确")
	ErrSmsTooManyAttempts = NewCodeError(ErrCodeSmsTooManyAttempts, "验证码错误次数过多，请稍后再试")
	ErrLoginFailed        = NewCodeError(ErrCodeLoginFailed, "手机号或验证码错误")

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
// 在 RPC 服务端 UnaryInterceptor 中 handler 返回错误后调用。
// 使用 SetTrailer 而非 SetHeader，确保在 error 场景下 trailer 仍能被发送。
func SetGrpcBizCode(ctx context.Context, bizCode int) {
	_ = grpc.SetTrailer(ctx, metadata.Pairs("x-biz-code", strconv.Itoa(bizCode)))
}

// errorMsgToBizCode 错误消息 → 业务错误码映射（单一数据源，供 ErrorToBizCode 和 MapRpcError 共用）
// 新增错误类型时只需在此处添加一行，避免两处 switch 重复维护
var errorMsgToBizCode = []struct {
	substr  string
	bizCode int
}{
	{"该手机号已注册", ErrCodePhoneRegistered},
	{"车辆不存在", ErrCodeVehicleNotFound},
	{"司机不存在", ErrCodeUserNotFound},
	{"用户不存在", ErrCodeUserNotFound},
	{"已禁用", ErrCodeUserDisabled},
	{"乘客账号已禁用", ErrCodeUserDisabled},
	{"验证码错误", ErrCodeSmsCodeError},
	{"验证码已过期", ErrCodeSmsCodeExpired},
	{"发送过于频繁", ErrCodeSmsSendFrequent},
	{"创建行程失败", ErrCodeTripCreateFailed},
	{"行政区划不存在", ErrCodeDistrictNotFound},
	{"验证码错误次数过多", ErrCodeSmsTooManyAttempts},
	{"车辆已绑定", ErrCodeVehicleBound},
	{"手机号或验证码错误", ErrCodeLoginFailed},
	{"无效的", ErrCodeParam},
	{"不能为空", ErrCodeParam},
	{"坐标无效", ErrCodeParam},
}

// matchErrorBizCode 根据错误消息匹配业务错误码，返回 0 表示未匹配
func matchErrorBizCode(msg string) int {
	for _, m := range errorMsgToBizCode {
		if strings.Contains(msg, m.substr) {
			return m.bizCode
		}
	}
	return 0
}

// ErrorToBizCode 根据错误消息匹配业务错误码，供 RPC 服务端 UnaryInterceptor 使用。
// 返回 0 表示未匹配到已知业务错误码。
func ErrorToBizCode(err error) int {
	if err == nil {
		return 0
	}
	return matchErrorBizCode(err.Error())
}

// BizCodeToError 根据业务错误码返回对应的 CodeError，供网关侧 MapRpcError 使用。
// 返回 nil 表示未匹配到已知业务错误码。
func BizCodeToError(bizCode int) *CodeError {
	switch bizCode {
	case ErrCodeParam:
		return ErrParam
	case ErrCodePhoneRegistered:
		return ErrPhoneRegistered
	case ErrCodeUserNotFound:
		return ErrUserNotFound
	case ErrCodeUserDisabled:
		return ErrUserDisabled
	case ErrCodeSmsCodeExpired:
		return ErrSmsCodeExpired
	case ErrCodeSmsCodeError:
		return ErrSmsCodeError
	case ErrCodeSmsSendFailed:
		return ErrSmsSendFailed
	case ErrCodeSmsSendFrequent:
		return ErrSmsSendFrequent
	case ErrCodePhoneFormat:
		return ErrPhoneFormat
	case ErrCodeSmsTooManyAttempts:
		return ErrSmsTooManyAttempts
	case ErrCodeLoginFailed:
		return ErrLoginFailed
	case ErrCodeVehicleNotFound:
		return ErrVehicleNotFound
	case ErrCodeVehicleBound:
		return ErrVehicleBound
	case ErrCodeTripNotFound:
		return ErrTripNotFound
	case ErrCodeTripStatusInvalid:
		return ErrTripStatusInvalid
	case ErrCodeTripCreateFailed:
		return ErrTripCreateFailed
	case ErrCodeDistrictNotFound:
		return ErrDistrictNotFound
	case ErrCodeTokenExpired:
		return ErrTokenExpired
	case ErrCodeTokenInvalid:
		return ErrTokenInvalid
	case ErrCodeAuthHeaderMissing:
		return ErrAuthHeaderMissing
	case ErrCodeAuthHeaderMalformed:
		return ErrAuthHeaderMalformed
	case ErrCodeServiceUnavailable:
		return ErrServiceUnavailable
	case ErrCodeRpcCallFailed:
		return ErrRpcCallFailed
	case ErrCodeDatabaseError:
		return ErrDatabaseError
	default:
		return nil
	}
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
// 优先从 gRPC trailer 的 x-biz-code 提取结构化业务错误码，
// 其次按 gRPC status code 做 HTTP 映射，最后退化为字符串匹配作为 fallback。
//
// 注意：当前 go-zero zrpc 客户端不直接暴露 gRPC trailer 给调用方，
// 因此 x-biz-code 路径暂时无法在 Gateway 侧生效，实际走字符串匹配 fallback。
// RPC 服务端通过 SetGrpcBizCode 写入 trailer 的逻辑已就绪，
// 将来 zrpc 升级支持 trailer 透传后，只需在此处启用 trailer 读取即可。
func MapRpcError(err error) *CodeError {
	if err == nil {
		return nil
	}

	// 尝试从 gRPC trailer 提取 x-biz-code（预留，当前 zrpc 不透传 trailer）
	if st, ok := status.FromError(err); ok {
		// TODO: 当 zrpc 支持 trailer 透传后，取消注释以下代码
		// trailers := st.Trailers()
		// for _, md := range trailers {
		// 	if vals := md.Get("x-biz-code"); len(vals) > 0 {
		// 		if bizCode, parseErr := strconv.Atoi(vals[0]); parseErr == nil {
		// 			if ce := BizCodeToError(bizCode); ce != nil {
		// 				return ce
		// 			}
		// 		}
		// 	}
		// }

		// 通过 gRPC status code 做基础映射
		httpCode := grpcCodeToHTTP(st.Code())
		if httpCode != 500 || st.Code() == codes.Internal {
			return NewCodeError(httpCode, st.Message())
		}
	}

	// errors.As 解包：优先匹配被包装的 *CodeError（如 TransactCtx 包装后的错误）
	var codeErr *CodeError
	if errors.As(err, &codeErr) {
		return codeErr
	}

	// fallback: 字符串匹配，复用 matchErrorBizCode + BizCodeToError 避免重复维护
	if bizCode := matchErrorBizCode(err.Error()); bizCode != 0 {
		if ce := BizCodeToError(bizCode); ce != nil {
			return ce
		}
	}

	return ErrServer
}
