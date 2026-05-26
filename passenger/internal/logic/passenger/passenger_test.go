package passengerlogic

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"onlinecar-passenger/internal/model"
	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"
)

// ========== Mock Model ==========

type mockPassengerUserModel struct {
	findByPhoneFunc func(ctx context.Context, phone string) (*model.PassengerUser, error)
	findOneFunc     func(ctx context.Context, id int64) (*model.PassengerUser, error)
	insertFunc      func(ctx context.Context, data *model.PassengerUser) (sql.Result, error)
}

func (m *mockPassengerUserModel) FindByPhone(ctx context.Context, phone string) (*model.PassengerUser, error) {
	if m.findByPhoneFunc != nil {
		return m.findByPhoneFunc(ctx, phone)
	}
	return nil, nil
}

func (m *mockPassengerUserModel) FindOne(ctx context.Context, id int64) (*model.PassengerUser, error) {
	if m.findOneFunc != nil {
		return m.findOneFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockPassengerUserModel) Insert(ctx context.Context, data *model.PassengerUser) (sql.Result, error) {
	if m.insertFunc != nil {
		return m.insertFunc(ctx, data)
	}
	return mockResult{1, 1}, nil
}

func (m *mockPassengerUserModel) Update(ctx context.Context, data *model.PassengerUser) error {
	return nil
}
func (m *mockPassengerUserModel) Delete(ctx context.Context, id int64) error { return nil }

type mockResult struct{ lastInsertID, rowsAffected int64 }

func (r mockResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r mockResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

// ========== Register Tests ==========

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockPassengerUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.PassengerUser, error) {
			return nil, nil
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewRegisterLogic(ctx, svcCtx)

	resp, err := logic.Register(&pb.RegisterReq{
		PassengerPhone: "13900000001",
		PassengerName:  "李四",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
	if resp.PassengerPhone != "13900000001" {
		t.Errorf("expected phone=13900000001, got: %s", resp.PassengerPhone)
	}
}

func TestRegister_PhoneExists(t *testing.T) {
	ctx := context.Background()
	phone := "13900000001"
	mock := &mockPassengerUserModel{
		findByPhoneFunc: func(ctx context.Context, p string) (*model.PassengerUser, error) {
			return &model.PassengerUser{Id: 1, PassengerPhone: &phone}, nil
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewRegisterLogic(ctx, svcCtx)

	_, err := logic.Register(&pb.RegisterReq{
		PassengerPhone: "13900000001",
		PassengerName:  "李四",
	})
	if err == nil {
		t.Fatal("expected error for existing phone")
	}
}

func TestRegister_DBError(t *testing.T) {
	ctx := context.Background()
	mock := &mockPassengerUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.PassengerUser, error) {
			return nil, errors.New("db error")
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewRegisterLogic(ctx, svcCtx)

	_, err := logic.Register(&pb.RegisterReq{
		PassengerPhone: "13900000001",
		PassengerName:  "李四",
	})
	if err == nil {
		t.Fatal("expected error for db failure")
	}
}

// ========== Login Tests ==========

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	phone := "13900000001"
	name := "李四"
	state := int8(0)
	mock := &mockPassengerUserModel{
		findByPhoneFunc: func(ctx context.Context, p string) (*model.PassengerUser, error) {
			return &model.PassengerUser{Id: 1, PassengerPhone: &phone, PassengerName: &name, State: &state}, nil
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewLoginLogic(ctx, svcCtx)

	resp, err := logic.Login(&pb.LoginReq{PassengerPhone: "13900000001"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Id != 1 {
		t.Errorf("expected id=1, got: %d", resp.Id)
	}
	if resp.PassengerName != "李四" {
		t.Errorf("expected name=李四, got: %s", resp.PassengerName)
	}
}

func TestLogin_NotFound(t *testing.T) {
	ctx := context.Background()
	mock := &mockPassengerUserModel{
		findByPhoneFunc: func(ctx context.Context, phone string) (*model.PassengerUser, error) {
			return nil, nil
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewLoginLogic(ctx, svcCtx)

	_, err := logic.Login(&pb.LoginReq{PassengerPhone: "13900000001"})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ========== GetPassenger Tests ==========

func TestGetPassenger_Success(t *testing.T) {
	ctx := context.Background()
	phone := "13900000001"
	name := "李四"
	mock := &mockPassengerUserModel{
		findOneFunc: func(ctx context.Context, id int64) (*model.PassengerUser, error) {
			return &model.PassengerUser{Id: id, PassengerPhone: &phone, PassengerName: &name}, nil
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewGetPassengerLogic(ctx, svcCtx)

	info, err := logic.GetPassenger(&pb.GetPassengerReq{Id: 1})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info.PassengerName != "李四" {
		t.Errorf("expected name=李四, got: %s", info.PassengerName)
	}
}

func TestGetPassenger_NotFound(t *testing.T) {
	ctx := context.Background()
	mock := &mockPassengerUserModel{
		findOneFunc: func(ctx context.Context, id int64) (*model.PassengerUser, error) {
			return nil, nil
		},
	}

	svcCtx := &svc.ServiceContext{PassengerUserModel: mock}
	logic := NewGetPassengerLogic(ctx, svcCtx)

	_, err := logic.GetPassenger(&pb.GetPassengerReq{Id: 999})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

// ========== CreateTrip Tests ==========

func TestCreateTrip_Success(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{}
	logic := NewCreateTripLogic(ctx, svcCtx)

	resp, err := logic.CreateTrip(&pb.CreateTripReq{
		PassengerId: 1,
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
