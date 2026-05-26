package sms

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"onlinecar-gateway/common/errorx"
	"onlinecar-gateway/internal/svc"
	driverpb "onlinecar-gateway/pb/driver"
	passengerpb "onlinecar-gateway/pb/passenger"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	codePrefix  = "sms:code:"
	limitPrefix = "sms:limit:"
	codeTTL     = 5 * time.Minute
	limitTTL    = 60 * time.Second
	codeLength  = 6
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
	// 检查手机号是否已注册
	if err := l.checkPhoneExists(phone); err != nil {
		return err
	}

	// 频率限制检查
	limitKey := fmt.Sprintf("%s%s:%s", limitPrefix, l.role, phone)
	limitVal, err := l.svcCtx.Redis.Get(limitKey)
	if err == nil && limitVal != "" {
		return errorx.ErrSmsSendFrequent
	}

	// 生成 6 位随机验证码
	code := fmt.Sprintf("%0*d", codeLength, rand.IntN(1000000))

	// 存验证码到 Redis（带角色前缀）
	codeKey := fmt.Sprintf("%s%s:%s", codePrefix, l.role, phone)
	if err := l.svcCtx.Redis.Setex(codeKey, code, int(codeTTL.Seconds())); err != nil {
		l.Errorf("Redis.Setex 失败, key=%s, err=%v", codeKey, err)
		return errorx.ErrSmsSendFailed
	}

	// 设频率限制
	if err := l.svcCtx.Redis.Setex(limitKey, "1", int(limitTTL.Seconds())); err != nil {
		l.Errorf("Redis.Setex 频率限制失败, key=%s, err=%v", limitKey, err)
	}

	// 发送短信
	if err := l.svcCtx.SmsSender.Send(l.ctx, phone, code); err != nil {
		l.Errorf("SmsSender.Send 失败, phone=%s, err=%v", phone, err)
		return errorx.ErrSmsSendFailed
	}

	l.Infof("验证码发送成功, role=%s, phone=%s", l.role, phone)
	return nil
}

func (l *SendSmsLogic) checkPhoneExists(phone string) error {
	switch l.role {
	case "driver":
		resp, err := l.svcCtx.DriverRpc.CheckPhone(l.ctx, &driverpb.CheckPhoneReq{DriverPhone: phone})
		if err != nil {
			l.Errorf("DriverRpc.CheckPhone RPC异常, phone=%s, err=%v", phone, err)
			return errorx.ErrServiceUnavailable
		}
		if !resp.Exists {
			return errorx.ErrUserNotFound
		}
	case "passenger":
		resp, err := l.svcCtx.PassengerRpc.CheckPhone(l.ctx, &passengerpb.CheckPhoneReq{PassengerPhone: phone})
		if err != nil {
			l.Errorf("PassengerRpc.CheckPhone RPC异常, phone=%s, err=%v", phone, err)
			return errorx.ErrServiceUnavailable
		}
		if !resp.Exists {
			return errorx.ErrUserNotFound
		}
	}
	return nil
}

// VerifyCode 校验验证码，成功后自动删除（防重放）
func VerifyCode(ctx context.Context, rds *svc.ServiceContext, role, phone, code string) error {
	codeKey := fmt.Sprintf("%s%s:%s", codePrefix, role, phone)

	stored, err := rds.Redis.Get(codeKey)
	if err != nil || stored == "" {
		return errorx.ErrSmsCodeExpired
	}

	if stored != code {
		return errorx.ErrSmsCodeError
	}

	// 删除已使用的验证码
	rds.Redis.Del(codeKey)
	return nil
}
