package trip

import (
	"context"
	"testing"

	"onlinecar-gateway/common/ctxdata"
	"onlinecar-gateway/internal/config"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	passengerpb "onlinecar-gateway/pb/passenger"

	"google.golang.org/grpc"
)

type mockPassengerClient struct {
	createTripFunc func(ctx context.Context, in *passengerpb.CreateTripReq, opts ...grpc.CallOption) (*passengerpb.TripResp, error)
}

func (m *mockPassengerClient) Register(ctx context.Context, in *passengerpb.RegisterReq, opts ...grpc.CallOption) (*passengerpb.RegisterResp, error) {
	return nil, nil
}

func (m *mockPassengerClient) Login(ctx context.Context, in *passengerpb.LoginReq, opts ...grpc.CallOption) (*passengerpb.LoginResp, error) {
	return nil, nil
}

func (m *mockPassengerClient) GetPassenger(ctx context.Context, in *passengerpb.GetPassengerReq, opts ...grpc.CallOption) (*passengerpb.PassengerInfo, error) {
	return nil, nil
}

func (m *mockPassengerClient) CreateTrip(ctx context.Context, in *passengerpb.CreateTripReq, opts ...grpc.CallOption) (*passengerpb.TripResp, error) {
	if m.createTripFunc != nil {
		return m.createTripFunc(ctx, in, opts...)
	}
	return nil, nil
}

func newTestSvcCtx(client passengerpb.PassengerClient) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			AuthPassenger: config.JwtConf{
				AccessSecret: "test-secret",
				AccessExpire: 86400e9,
			},
		},
		PassengerRpc: client,
	}
}

func TestCreateTripLogic_Success(t *testing.T) {
	ctx := context.Background()
	uidCtx := ctxdata.SetUID(ctx, 1)
	mockClient := &mockPassengerClient{
		createTripFunc: func(ctx context.Context, in *passengerpb.CreateTripReq, opts ...grpc.CallOption) (*passengerpb.TripResp, error) {
			return &passengerpb.TripResp{
				Id: 1, PassengerId: 1, Origin: "天安门", Destination: "北京南站", Status: 1,
			}, nil
		},
	}

	logic := NewCreateTripLogic(uidCtx, newTestSvcCtx(mockClient))
	resp, err := logic.CreateTrip(&types.CreateTripReq{
		Origin:      "天安门",
		Destination: "北京南站",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
	if resp.Origin != "天安门" {
		t.Errorf("expected origin=天安门, got: %s", resp.Origin)
	}
}

func TestCreateTripLogic_EmptyResp(t *testing.T) {
	ctx := context.Background()
	uidCtx := ctxdata.SetUID(ctx, 1)
	mockClient := &mockPassengerClient{
		createTripFunc: func(ctx context.Context, in *passengerpb.CreateTripReq, opts ...grpc.CallOption) (*passengerpb.TripResp, error) {
			return &passengerpb.TripResp{}, nil // TODO: 当前 RPC 返回空
		},
	}

	logic := NewCreateTripLogic(uidCtx, newTestSvcCtx(mockClient))
	resp, err := logic.CreateTrip(&types.CreateTripReq{
		Origin:      "天安门",
		Destination: "北京南站",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}
