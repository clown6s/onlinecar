package sms

import (
	"context"
	"log/slog"
	"os"
)

// AliyunSmsSender 阿里云短信服务实现（占位）
// 生产环境需替换为真实的阿里云 SDK 调用。
// 需要的环境变量：ALIYUM_SMS_ACCESS_KEY_ID, ALIYUM_SMS_ACCESS_KEY_SECRET, ALIYUM_SMS_SIGN_NAME, ALIYUM_SMS_TEMPLATE_CODE
type AliyunSmsSender struct{}

func NewAliyunSmsSender() *AliyunSmsSender {
	accessKeyId := os.Getenv("ALIYUM_SMS_ACCESS_KEY_ID")
	if accessKeyId == "" {
		slog.Warn("ALIYUM_SMS_ACCESS_KEY_ID 未设置，阿里云短信服务将无法正常工作")
	}
	return &AliyunSmsSender{}
}

func (a *AliyunSmsSender) Send(ctx context.Context, phone, code string) error {
	// TODO: 对接阿里云短信 SDK
	// 参考: https://help.aliyun.com/document_detail/101414.html
	slog.Info("aliyun sms send (stub)", "phone", phone, "code_len", len(code))
	return nil
}
