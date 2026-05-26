package passenger

import (
	"context"
	"testing"
	"time"

	"onlinecar-gateway/internal/config"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	passengerpb "onlinecar-gateway/pb/passenger"

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

func newTestSvcCtx(client passengerpb.PassengerClient) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			AuthPassenger: config.JwtConf{
				AccessSecret: "test-secret",
				AccessExpire: time.Hour,
			},
		},
		PassengerRpc: client,
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
	mockClient := &mockPassengerClient{
		loginFunc: func(ctx context.Context, in *passengerpb.LoginReq, opts ...grpc.CallOption) (*passengerpb.LoginResp, error) {
			return &passengerpb.LoginResp{Id: 1, PassengerName: "李四", State: 0}, nil
		},
	}

	logic := NewLoginLogic(ctx, newTestSvcCtx(mockClient))
	resp, err := logic.Login(&types.PassengerLoginReq{Phone: "13900000001"})
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
