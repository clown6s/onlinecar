package driver

import (
	"context"
	"testing"
	"time"

	"onlinecar-common/ctxdata"
	"onlinecar-gateway/common/sms"
	"onlinecar-gateway/internal/config"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	driverpb "onlinecar-gateway/pb/driver"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
)

type mockDriverClient struct {
	registerFunc       func(ctx context.Context, in *driverpb.RegisterReq, opts ...grpc.CallOption) (*driverpb.RegisterResp, error)
	loginFunc          func(ctx context.Context, in *driverpb.LoginReq, opts ...grpc.CallOption) (*driverpb.LoginResp, error)
	updateLocationFunc func(ctx context.Context, in *driverpb.UpdateLocationReq, opts ...grpc.CallOption) (*driverpb.Empty, error)
	checkPhoneFunc     func(ctx context.Context, in *driverpb.CheckPhoneReq, opts ...grpc.CallOption) (*driverpb.CheckPhoneResp, error)
}

func (m *mockDriverClient) Register(ctx context.Context, in *driverpb.RegisterReq, opts ...grpc.CallOption) (*driverpb.RegisterResp, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockDriverClient) Login(ctx context.Context, in *driverpb.LoginReq, opts ...grpc.CallOption) (*driverpb.LoginResp, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockDriverClient) UpdateLocation(ctx context.Context, in *driverpb.UpdateLocationReq, opts ...grpc.CallOption) (*driverpb.Empty, error) {
	if m.updateLocationFunc != nil {
		return m.updateLocationFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockDriverClient) GetDriver(ctx context.Context, in *driverpb.GetDriverReq, opts ...grpc.CallOption) (*driverpb.DriverInfo, error) {
	return nil, nil
}

func (m *mockDriverClient) GetDriverWorkStatus(ctx context.Context, in *driverpb.GetDriverWorkStatusReq, opts ...grpc.CallOption) (*driverpb.DriverWorkStatusInfo, error) {
	return nil, nil
}

func (m *mockDriverClient) UpdateDriverWorkStatus(ctx context.Context, in *driverpb.UpdateDriverWorkStatusReq, opts ...grpc.CallOption) (*driverpb.Empty, error) {
	return nil, nil
}

func (m *mockDriverClient) CheckPhone(ctx context.Context, in *driverpb.CheckPhoneReq, opts ...grpc.CallOption) (*driverpb.CheckPhoneResp, error) {
	if m.checkPhoneFunc != nil {
		return m.checkPhoneFunc(ctx, in, opts...)
	}
	return &driverpb.CheckPhoneResp{Exists: true}, nil
}

func newTestSvcCtx(client driverpb.DriverClient) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			AuthDriver: config.JwtConf{
				AccessSecret: "test-secret",
				AccessExpire: time.Hour,
			},
		},
		DriverRpc: client,
		Redis:     redis.Nil, // 使用 nil Redis（mock 模式）
		SmsSender: sms.NewMockSmsSender(),
	}
}

func TestRegisterLogic_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockDriverClient{
		checkPhoneFunc: func(ctx context.Context, in *driverpb.CheckPhoneReq, opts ...grpc.CallOption) (*driverpb.CheckPhoneResp, error) {
			return &driverpb.CheckPhoneResp{Exists: false}, nil
		},
		registerFunc: func(ctx context.Context, in *driverpb.RegisterReq, opts ...grpc.CallOption) (*driverpb.RegisterResp, error) {
			return &driverpb.RegisterResp{Id: 1, DriverName: "张三", DriverPhone: "13800000001", State: 0}, nil
		},
	}

	logic := NewRegisterLogic(ctx, newTestSvcCtx(mockClient))
	resp, err := logic.Register(&types.DriverRegisterReq{
		Name:  "张三",
		Phone: "13800000001",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
	if resp.Name != "张三" {
		t.Errorf("expected name=张三, got: %s", resp.Name)
	}
}

func TestLoginLogic_Success(t *testing.T) {
	ctx := context.Background()

	// 先发验证码（写入 Redis）
	svcCtx := newTestSvcCtx(&mockDriverClient{
		loginFunc: func(ctx context.Context, in *driverpb.LoginReq, opts ...grpc.CallOption) (*driverpb.LoginResp, error) {
			return &driverpb.LoginResp{Id: 1, DriverName: "张三"}, nil
		},
	})
	codeKey := "sms:code:driver:13800000001"
	svcCtx.Redis.Setex(codeKey, "123456", 300)

	// 再登录
	logic := NewLoginLogic(ctx, svcCtx)
	resp, err := logic.Login(&types.DriverLoginReq{Phone: "13800000001", Code: "123456"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("expected expiresIn=3600, got: %d", resp.ExpiresIn)
	}
}

func TestLoginLogic_WrongCode(t *testing.T) {
	ctx := context.Background()
	svcCtx := newTestSvcCtx(&mockDriverClient{})
	codeKey := "sms:code:driver:13800000001"
	svcCtx.Redis.Setex(codeKey, "123456", 300)

	logic := NewLoginLogic(ctx, svcCtx)
	_, err := logic.Login(&types.DriverLoginReq{Phone: "13800000001", Code: "654321"})
	if err == nil {
		t.Fatal("expected error for wrong code, got nil")
	}
}

func TestUpdateLocationLogic_Success(t *testing.T) {
	ctx := context.Background()
	uidCtx := ctxdata.SetUID(ctx, 1)
	mockClient := &mockDriverClient{
		updateLocationFunc: func(ctx context.Context, in *driverpb.UpdateLocationReq, opts ...grpc.CallOption) (*driverpb.Empty, error) {
			return &driverpb.Empty{}, nil
		},
	}

	logic := NewUpdateLocationLogic(uidCtx, newTestSvcCtx(mockClient))
	err := logic.UpdateLocation(&types.DriverLocationReq{Lng: 116.397428, Lat: 39.90923})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
