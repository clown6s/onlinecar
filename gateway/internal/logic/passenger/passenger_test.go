package passenger

import (
	"context"
	"testing"
	"time"

	"onlinecar-gateway/common/sms"
	"onlinecar-gateway/internal/config"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	passengerpb "onlinecar-gateway/pb/passenger"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
)

type mockPassengerClient struct {
	registerFunc func(ctx context.Context, in *passengerpb.RegisterReq, opts ...grpc.CallOption) (*passengerpb.RegisterResp, error)
	loginFunc    func(ctx context.Context, in *passengerpb.LoginReq, opts ...grpc.CallOption) (*passengerpb.LoginResp, error)
}

func (m *mockPassengerClient) Register(ctx context.Context, in *passengerpb.RegisterReq, opts ...grpc.CallOption) (*passengerpb.RegisterResp, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockPassengerClient) Login(ctx context.Context, in *passengerpb.LoginReq, opts ...grpc.CallOption) (*passengerpb.LoginResp, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockPassengerClient) GetPassenger(ctx context.Context, in *passengerpb.GetPassengerReq, opts ...grpc.CallOption) (*passengerpb.PassengerInfo, error) {
	return nil, nil
}

func (m *mockPassengerClient) CreateTrip(ctx context.Context, in *passengerpb.CreateTripReq, opts ...grpc.CallOption) (*passengerpb.TripResp, error) {
	return nil, nil
}

func (m *mockPassengerClient) CheckPhone(ctx context.Context, in *passengerpb.CheckPhoneReq, opts ...grpc.CallOption) (*passengerpb.CheckPhoneResp, error) {
	return &passengerpb.CheckPhoneResp{Exists: true}, nil
}

func newTestSvcCtx(client passengerpb.PassengerClient) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			AuthPassenger: config.JwtConf{
				AccessSecret: "test-secret",
				AccessExpire: time.Hour,
			},
		},
		PassengerRpc: client,
		Redis:        redis.Nil,
		SmsSender:    sms.NewMockSmsSender(),
	}
}

func TestRegisterLogic_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockPassengerClient{
		registerFunc: func(ctx context.Context, in *passengerpb.RegisterReq, opts ...grpc.CallOption) (*passengerpb.RegisterResp, error) {
			return &passengerpb.RegisterResp{Id: 1, PassengerName: "李四", PassengerPhone: "13900000001", State: 0}, nil
		},
	}

	logic := NewRegisterLogic(ctx, newTestSvcCtx(mockClient))
	resp, err := logic.Register(&types.PassengerRegisterReq{
		Name:  "李四",
		Phone: "13900000001",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
	if resp.Name != "李四" {
		t.Errorf("expected name=李四, got: %s", resp.Name)
	}
}

func TestLoginLogic_Success(t *testing.T) {
	ctx := context.Background()

	svcCtx := newTestSvcCtx(&mockPassengerClient{})
	codeKey := "sms:code:passenger:13900000001"
	svcCtx.Redis.Setex(codeKey, "123456", 300)

	logic := NewLoginLogic(ctx, svcCtx)
	resp, err := logic.Login(&types.PassengerLoginReq{Phone: "13900000001", Code: "123456"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp["passengerId"].(int64) != 1 {
		t.Errorf("expected passengerId=1, got: %v", resp["passengerId"])
	}
	if resp["token"] == nil || resp["token"].(string) == "" {
		t.Error("expected non-empty token")
	}
}

func TestLoginLogic_WrongCode(t *testing.T) {
	ctx := context.Background()
	svcCtx := newTestSvcCtx(&mockPassengerClient{})
	codeKey := "sms:code:passenger:13900000001"
	svcCtx.Redis.Setex(codeKey, "123456", 300)

	logic := NewLoginLogic(ctx, svcCtx)
	_, err := logic.Login(&types.PassengerLoginReq{Phone: "13900000001", Code: "654321"})
	if err == nil {
		t.Fatal("expected error for wrong code, got nil")
	}
}
