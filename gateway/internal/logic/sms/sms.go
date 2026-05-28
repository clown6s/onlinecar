package sms

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"onlinecar-common/errorx"
	"onlinecar-common/validate"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	codePrefix        = "sms:code:"
	limitPrefix       = "sms:limit:"
	verifyPrefix      = "sms:verify:"
	codeTTL           = 5 * time.Minute
	limitTTL          = 60 * time.Second
	verifyTTL         = 5 * time.Minute
	maxVerifyAttempts = 5
)

type SendSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	role   string // "driver" 或 "passenger"
}

func NewSendSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext, role string) *SendSmsLogic {
	return &SendSmsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx, role: role}
}

func (l *SendSmsLogic) SendSms(phone string) error {
	// 频率限制检查（原子操作，避免 TOCTOU 竞态）
	limitKey := fmt.Sprintf("%s%s:%s", limitPrefix, l.role, phone)
	ok, err := l.svcCtx.Redis.SetnxEx(limitKey, "1", int(limitTTL.Seconds()))
	if err != nil {
		l.Errorf("Redis.SetnxEx 失败, key=%s, err=%v", limitKey, err)
		return errorx.ErrSmsSendFailed
	}
	if !ok {
		return errorx.NewCodeError(errorx.ErrCodeSmsSendFrequent, "发送过于频繁，请60秒后再试")
	}

	// 发送新验证码时，不删除旧的 verify 计数 key（防止重置暴力破解计数器）
	// verifyKey 仅在验证成功或 TTL 自然过期时清除

	// 生成 6 位随机验证码
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	code := fmt.Sprintf("%06d", n.Int64())

	// 存验证码到 Redis（带角色前缀）
	codeKey := fmt.Sprintf("%s%s:%s", codePrefix, l.role, phone)
	if err := l.svcCtx.Redis.Setex(codeKey, code, int(codeTTL.Seconds())); err != nil {
		l.Errorf("Redis.Setex 失败, key=%s, err=%v", codeKey, err)
		return errorx.ErrSmsSendFailed
	}

	// 设频率限制
	// Setnx 已在上面原子设置，此处无需再次 Setex

	// 发送短信
	if err := l.svcCtx.SmsSender.Send(l.ctx, phone, code); err != nil {
		l.Errorf("SmsSender.Send 失败, phone=%s, err=%v", phone, err)
		_, _ = l.svcCtx.Redis.Del(codeKey)
		_, _ = l.svcCtx.Redis.Del(limitKey) // 回滚限流 key，允许用户重试
		return errorx.ErrSmsSendFailed
	}

	l.Infof("验证码发送成功, role=%s, phone=%s", l.role, phone)
	return nil
}

// checkAndIncrVerifyScript 原子检查+递增验证失败次数（防止 TOCTOU 竞态）
// 返回 -2 表示已超过最大尝试次数，否则返回当前计数
const checkAndIncrVerifyScript = `
local count = redis.call('GET', KEYS[1])
if count and tonumber(count) >= tonumber(ARGV[2]) then
	return -2
end
count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`
const verifyCodeScript = `
local stored = redis.call('GET', KEYS[1])
if not stored or stored == '' then
	return -1
end
if stored ~= ARGV[1] then
	return 0
end
redis.call('DEL', KEYS[1])
return 1
`

// VerifyCode 校验验证码，成功后自动删除（原子操作，防重放）
func VerifyCode(ctx context.Context, svcCtx *svc.ServiceContext, role, phone, code string) error {
	if !validate.IsValidPhone(phone) {
		return errorx.ErrPhoneFormat
	}
	if code == "" {
		return errorx.ErrParam
	}

	// 暴力破解防护：verifyKey 在验证失败时原子检查+递增（见下方 case 0）
	// 避免先 GET 后 INCR 的 TOCTOU 竞态
	verifyKey := fmt.Sprintf("%s%s:%s", verifyPrefix, role, phone)

	codeKey := fmt.Sprintf("%s%s:%s", codePrefix, role, phone)

	result, err := svcCtx.Redis.Eval(verifyCodeScript, []string{codeKey}, code)
	if err != nil {
		logx.Errorf("[SMS] 验证码校验Lua执行失败, key=%s, err=%v", codeKey, err)
		return errorx.ErrSmsCodeExpired
	}

	// Eval 返回值类型为 int64: -1=不存在, 0=不匹配, 1=成功
	switch val := result.(type) {
	case int64:
		switch val {
		case -1:
			return errorx.ErrSmsCodeExpired
		case 0:
			checkResult, checkErr := svcCtx.Redis.Eval(checkAndIncrVerifyScript, []string{verifyKey}, int(verifyTTL.Seconds()), maxVerifyAttempts)
			if checkErr != nil {
				logx.Errorf("[SMS] 暴力破解检查+递增失败, verifyKey=%s, err=%v", verifyKey, checkErr)
			}
			if cr, ok := checkResult.(int64); ok && cr == -2 {
				return errorx.ErrSmsTooManyAttempts
			}
			if cs, ok := checkResult.(string); ok && cs == "-2" {
				return errorx.ErrSmsTooManyAttempts
			}
			return errorx.ErrSmsCodeError
		case 1:
			if _, delErr := svcCtx.Redis.Del(verifyKey); delErr != nil {
				logx.Errorf("[SMS] 删除验证计数器失败(成功后), verifyKey=%s, err=%v", verifyKey, delErr)
			}
			return nil
		}
	case string:
		// go-zero 的 Eval 可能返回字符串形式的数字
		switch val {
		case "-1":
			return errorx.ErrSmsCodeExpired
		case "0":
			checkResult, checkErr := svcCtx.Redis.Eval(checkAndIncrVerifyScript, []string{verifyKey}, int(verifyTTL.Seconds()), maxVerifyAttempts)
			if checkErr != nil {
				logx.Errorf("[SMS] 暴力破解检查+递增失败(字符串返回), verifyKey=%s, err=%v", verifyKey, checkErr)
			}
			if cr, ok := checkResult.(int64); ok && cr == -2 {
				return errorx.ErrSmsTooManyAttempts
			}
			if cs, ok := checkResult.(string); ok && cs == "-2" {
				return errorx.ErrSmsTooManyAttempts
			}
			return errorx.ErrSmsCodeError
		case "1":
			if _, delErr := svcCtx.Redis.Del(verifyKey); delErr != nil {
				logx.Errorf("[SMS] 删除验证计数器失败(字符串成功后), verifyKey=%s, err=%v", verifyKey, delErr)
			}
			return nil
		}
	}

	logx.Errorf("[SMS] 验证码校验返回异常值, key=%s, result=%v", codeKey, result)
	return errorx.ErrSmsCodeExpired
}
