package driverlogic

import (
	"context"
	"errors"
	"testing"

	"onlinecar-driver/internal/model"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"
)

// ========== WorkStatus Mock ==========

type workStatusMockModel struct {
	upsertFunc         func(ctx context.Context, driverId, workStatus int64) error
	findByDriverIdFunc func(ctx context.Context, driverId int64) (*model.DriverWorkStatus, error)
}

func (m *workStatusMockModel) Upsert(ctx context.Context, driverId, workStatus int64) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, driverId, workStatus)
	}
	return nil
}

func (m *workStatusMockModel) FindByDriverId(ctx context.Context, driverId int64) (*model.DriverWorkStatus, error) {
	if m.findByDriverIdFunc != nil {
		return m.findByDriverIdFunc(ctx, driverId)
	}
	return nil, nil
}

func TestUpdateDriverWorkStatus_Success(t *testing.T) {
	ctx := context.Background()
	mockModel := &workStatusMockModel{
		upsertFunc: func(ctx context.Context, driverId, workStatus int64) error {
			return nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverWorkStatusModel: mockModel}
	logic := NewUpdateDriverWorkStatusLogic(ctx, svcCtx)

	_, err := logic.UpdateDriverWorkStatus(&pb.UpdateDriverWorkStatusReq{
		DriverId:   1,
		WorkStatus: 1,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestUpdateDriverWorkStatus_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{DriverWorkStatusModel: &workStatusMockModel{}}
	logic := NewUpdateDriverWorkStatusLogic(ctx, svcCtx)

	_, err := logic.UpdateDriverWorkStatus(&pb.UpdateDriverWorkStatusReq{
		DriverId:   1,
		WorkStatus: 5,
	})
	if err == nil {
		t.Fatal("expected error for invalid work status")
	}
}

func TestUpdateDriverWorkStatus_DBError(t *testing.T) {
	ctx := context.Background()
	mockModel := &workStatusMockModel{
		upsertFunc: func(ctx context.Context, driverId, workStatus int64) error {
			return errors.New("db error")
		},
	}

	svcCtx := &svc.ServiceContext{DriverWorkStatusModel: mockModel}
	logic := NewUpdateDriverWorkStatusLogic(ctx, svcCtx)

	_, err := logic.UpdateDriverWorkStatus(&pb.UpdateDriverWorkStatusReq{
		DriverId:   1,
		WorkStatus: 1,
	})
	if err == nil {
		t.Fatal("expected error for db failure")
	}
}

func TestGetDriverWorkStatus_Found(t *testing.T) {
	ctx := context.Background()
	status := 1
	mockModel := &workStatusMockModel{
		findByDriverIdFunc: func(ctx context.Context, driverId int64) (*model.DriverWorkStatus, error) {
			return &model.DriverWorkStatus{
				Id:         1,
				DriverId:   &driverId,
				WorkStatus: &status,
			}, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverWorkStatusModel: mockModel}
	logic := NewGetDriverWorkStatusLogic(ctx, svcCtx)

	resp, err := logic.GetDriverWorkStatus(&pb.GetDriverWorkStatusReq{DriverId: 1})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.WorkStatus != 1 {
		t.Errorf("expected work_status=1, got: %d", resp.WorkStatus)
	}
}

func TestGetDriverWorkStatus_Default(t *testing.T) {
	ctx := context.Background()
	mockModel := &workStatusMockModel{
		findByDriverIdFunc: func(ctx context.Context, driverId int64) (*model.DriverWorkStatus, error) {
			return nil, nil
		},
	}

	svcCtx := &svc.ServiceContext{DriverWorkStatusModel: mockModel}
	logic := NewGetDriverWorkStatusLogic(ctx, svcCtx)

	resp, err := logic.GetDriverWorkStatus(&pb.GetDriverWorkStatusReq{DriverId: 1})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.WorkStatus != 0 {
		t.Errorf("expected default work_status=0, got: %d", resp.WorkStatus)
	}
}
