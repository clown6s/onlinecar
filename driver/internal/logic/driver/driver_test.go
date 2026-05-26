package driverlogic

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"onlinecar-driver/internal/model"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"
)

// ========== Mock Models ==========

type mockDriverCarBindingModel struct{}

func (m *mockDriverCarBindingModel) Insert(ctx context.Context, data *model.DriverCarBinding) (sql.Result, error) {
	return nil, nil
}
func (m *mockDriverCarBindingModel) FindByDriverId(ctx context.Context, driverId int64) (*model.DriverCarBinding, error) {
	return nil, nil
}
func (m *mockDriverCarBindingModel) Update(ctx context.Context, data *model.DriverCarBinding) error {
	return nil
}

type mockCarModel struct{}

func (m *mockCarModel) Insert(ctx context.Context, data *model.Car) (sql.Result, error) {
	return nil, nil
}
func (m *mockCarModel) FindOne(ctx context.Context, id int64) (*model.Car, error) { return nil, nil }
func (m *mockCarModel) FindByVehicleNo(ctx context.Context, vehicleNo string) (*model.Car, error) {
	return nil, nil
}
func (m *mockCarModel) Update(ctx context.Context, data *model.Car) error { return nil }
func (m *mockCarModel) Delete(ctx context.Context, id int64) error        { return nil }

type mockDriverWorkStatusModel struct{}

func (m *mockDriverWorkStatusModel) Upsert(ctx context.Context, driverId, workStatus int64) error {
	return nil
}
func (m *mockDriverWorkStatusModel) FindByDriverId(ctx context.Context, driverId int64) (*model.DriverWorkStatus, error) {
	return nil, nil
}

type mockDriverUserModel struct {
	findByPhoneFunc func(ctx context.Context, phone string) (*model.DriverUser, error)
	findOneFunc     func(ctx context.Context, id int64) (*model.DriverUser, error)
	insertFunc      func(ctx context.Context, data *model.DriverUser) (sql.Result, error)
}

func (m *mockDriverUserModel) FindByPhone(ctx context.Context, phone string) (*model.DriverUser, error) {
	if m.findByPhoneFunc != nil {
		return m.findByPhoneFunc(ctx, phone)
	}
	return nil, nil
}

func (m *mockDriverUserModel) FindOne(ctx context.Context, id int64) (*model.DriverUser, error) {
	if m.findOneFunc != nil {
		return m.findOneFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockDriverUserModel) Insert(ctx context.Context, data *model.DriverUser) (sql.Result, error) {
	if m.insertFunc != nil {
		return m.insertFunc(ctx, data)
	}
	return sqlmockResult{1, 1}, nil
}

func (m *mockDriverUserModel) Update(ctx context.Context, data *model.DriverUser) error { return nil }
func (m *mockDriverUserModel) Delete(ctx context.Context, id int64) error               { return nil }

type sqlmockResult struct{ lastInsertID, rowsAffected int64 }

func (r sqlmockResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r sqlmockResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

// ========== RegisterLogic Tests ==========

func TestRegisterLogic_Register_Success(t *testing.T) {
	ctx := context.Background()
	mockUser := &mockDriverUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.DriverUser, error) {
			return nil, nil // 不存在
		},
		insertFunc: func(ctx context.Context, data *model.DriverUser) (sql.Result, error) {
			return sqlmockResult{1, 1}, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewRegisterLogic(ctx, svcCtx)

	resp, err := logic.Register(&pb.RegisterReq{
		DriverName:  "张三",
		DriverPhone: "13800000001",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
	if resp.DriverName != "张三" {
		t.Errorf("expected name=张三, got: %s", resp.DriverName)
	}
}

func TestRegisterLogic_Register_PhoneExists(t *testing.T) {
	ctx := context.Background()
	mockUser := &mockDriverUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.DriverUser, error) {
			return &model.DriverUser{Id: 1, DriverPhone: "13800000001"}, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewRegisterLogic(ctx, svcCtx)

	_, err := logic.Register(&pb.RegisterReq{
		DriverName:  "张三",
		DriverPhone: "13800000001",
	})
	if err == nil {
		t.Fatal("expected error for existing phone")
	}
}

func TestRegisterLogic_Register_DBError(t *testing.T) {
	ctx := context.Background()
	mockUser := &mockDriverUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.DriverUser, error) {
			return nil, errors.New("db connection error")
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewRegisterLogic(ctx, svcCtx)

	_, err := logic.Register(&pb.RegisterReq{
		DriverName:  "张三",
		DriverPhone: "13800000001",
	})
	if err == nil {
		t.Fatal("expected error for db failure")
	}
}

// ========== LoginLogic Tests ==========

func TestLoginLogic_Login_Success(t *testing.T) {
	ctx := context.Background()
	state := int64(0)
	mockUser := &mockDriverUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.DriverUser, error) {
			return &model.DriverUser{
				Id:          1,
				DriverName:  "张三",
				DriverPhone: phone,
				State:       &state,
			}, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewLoginLogic(ctx, svcCtx)

	resp, err := logic.Login(&pb.LoginReq{DriverPhone: "13800000001"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
}

func TestLoginLogic_Login_NotFound(t *testing.T) {
	ctx := context.Background()
	mockUser := &mockDriverUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.DriverUser, error) {
			return nil, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewLoginLogic(ctx, svcCtx)

	_, err := logic.Login(&pb.LoginReq{DriverPhone: "13800000001"})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ========== GetDriverLogic Tests ==========

func TestGetDriverLogic_Success(t *testing.T) {
	ctx := context.Background()
	state := int64(0)
	mockUser := &mockDriverUserModel{
		findOneFunc: func(ctx context.Context, id int64) (*model.DriverUser, error) {
			return &model.DriverUser{
				Id:            1,
				DriverName:    "张三",
				DriverPhone:   "13800000001",
				LicenseId:     "320123199001011234",
				CertificateNo: "CERT001",
				State:         &state,
			}, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewGetDriverLogic(ctx, svcCtx)

	info, err := logic.GetDriver(&pb.GetDriverReq{Id: 1})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info.DriverName != "张三" {
		t.Errorf("expected name=张三, got: %s", info.DriverName)
	}
}

func TestGetDriverLogic_NotFound(t *testing.T) {
	ctx := context.Background()
	mockUser := &mockDriverUserModel{
		findOneFunc: func(ctx context.Context, id int64) (*model.DriverUser, error) {
			return nil, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverUserModel: mockUser, DriverCarBindingModel: &mockDriverCarBindingModel{}, CarModel: &mockCarModel{}, DriverWorkStatusModel: &mockDriverWorkStatusModel{}}
	logic := NewGetDriverLogic(ctx, svcCtx)

	_, err := logic.GetDriver(&pb.GetDriverReq{Id: 999})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ========== UpdateLocationLogic Tests ==========

func TestUpdateLocationLogic_Success(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}
	logic := NewUpdateLocationLogic(ctx, svcCtx)

	resp, err := logic.UpdateLocation(&pb.UpdateLocationReq{
		Id:  1,
		Lng: 116.397428,
		Lat: 39.90923,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}
