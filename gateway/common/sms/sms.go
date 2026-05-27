package sms

import (
	"context"
	"log/slog"
)

// SmsSender 短信发送接口，后续对接真实短信服务只需实现此接口
type SmsSender interface {
	Send(ctx context.Context, phone, code string) error
}

// MockSmsSender Mock 实现，仅打印日志
type MockSmsSender struct{}

func NewMockSmsSender() *MockSmsSender {
	return &MockSmsSender{}
}

func (m *MockSmsSender) Send(_ context.Context, phone, code string) error {
	slog.Info("mock sms sent", "phone", phone, "code_len", len(code))
	return nil
}
