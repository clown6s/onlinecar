package sms

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"onlinecar-common/errorx"
	"onlinecar-gateway/common/validate"
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

	// 发送新验证码时删除旧的 verify 计数 key，防止旧验证码继续试
	verifyKey := fmt.Sprintf("%s%s:%s", verifyPrefix, l.role, phone)
	_, _ = l.svcCtx.Redis.Del(verifyKey)

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

// verifyCodeScript 原子化验证码校验+删除（防并发重放）
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

	// 暴力破解防护：检查验证失败次数
	verifyKey := fmt.Sprintf("%s%s:%s", verifyPrefix, role, phone)
	verifyCount, err := svcCtx.Redis.Get(verifyKey)
	if err == nil && verifyCount != "" {
		var count int
		fmt.Sscanf(verifyCount, "%d", &count)
		if count >= maxVerifyAttempts {
			return errorx.NewCodeError(errorx.ErrCodeSmsCodeError+900, "验证码错误次数过多，请稍后再试")
		}
	}

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
			svcCtx.Redis.Incr(verifyKey)
			svcCtx.Redis.Expire(verifyKey, int(verifyTTL.Seconds()))
			return errorx.ErrSmsCodeError
		case 1:
			svcCtx.Redis.Del(verifyKey)
			return nil
		}
	case string:
		// go-zero 的 Eval 可能返回字符串形式的数字
		switch val {
		case "-1":
			return errorx.ErrSmsCodeExpired
		case "0":
			svcCtx.Redis.Incr(verifyKey)
			svcCtx.Redis.Expire(verifyKey, int(verifyTTL.Seconds()))
			return errorx.ErrSmsCodeError
		case "1":
			svcCtx.Redis.Del(verifyKey)
			return nil
		}
	}

	logx.Errorf("[SMS] 验证码校验返回异常值, key=%s, result=%v", codeKey, result)
	return errorx.ErrSmsCodeExpired
}
