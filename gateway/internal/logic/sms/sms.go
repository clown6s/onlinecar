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
	codePrefix  = "sms:code:"
	limitPrefix = "sms:limit:"
	codeTTL     = 5 * time.Minute
	limitTTL    = 60 * time.Second
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
	// 频率限制检查
	limitKey := fmt.Sprintf("%s%s:%s", limitPrefix, l.role, phone)
	limitVal, err := l.svcCtx.Redis.Get(limitKey)
	if err == nil && limitVal != "" {
		return errorx.ErrSmsSendFrequent
	}

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
	if err := l.svcCtx.Redis.Setex(limitKey, "1", int(limitTTL.Seconds())); err != nil {
		l.Errorf("Redis.Setex 频率限制失败, key=%s, err=%v", limitKey, err)
		// 频率限制写入失败时回滚验证码，防止限流失效
		_, _ = l.svcCtx.Redis.Del(codeKey)
		return errorx.ErrSmsSendFailed
	}

	// 发送短信
	if err := l.svcCtx.SmsSender.Send(l.ctx, phone, code); err != nil {
		l.Errorf("SmsSender.Send 失败, phone=%s, err=%v", phone, err)
		_, _ = l.svcCtx.Redis.Del(codeKey)
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
			return errorx.ErrSmsCodeError
		case 1:
			return nil
		}
	case string:
		// go-zero 的 Eval 可能返回字符串形式的数字
		switch val {
		case "-1":
			return errorx.ErrSmsCodeExpired
		case "0":
			return errorx.ErrSmsCodeError
		case "1":
			return nil
		}
	}

	logx.Errorf("[SMS] 验证码校验返回异常值, key=%s, result=%v", codeKey, result)
	return errorx.ErrSmsCodeExpired
}
