package driverlogic

import (
	"context"
	"errors"

	"onlinecar-common/errorx"
	"onlinecar-driver/internal/model"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverLogic {
	return &GetDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取司机信息（含车辆和工作状态）
func (l *GetDriverLogic) GetDriver(in *pb.GetDriverReq) (*pb.DriverInfo, error) {
	l.Infof("获取司机信息, driverId=%d", in.Id)

	driverUser, _, car, err := l.svcCtx.DriverUserModel.FindDriverDetail(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.ErrUserNotFound
		}
		l.Errorf("查询司机失败, driverId=%d, err=%v", in.Id, err)
		return nil, errorx.ErrServer
	}

	info := &pb.DriverInfo{
		Id:            driverUser.Id,
		DriverName:    driverUser.DriverName,
		DriverPhone:   driverUser.DriverPhone,
		LicenseId:     driverUser.LicenseId,
		CertificateNo: driverUser.CertificateNo,
	}

	if driverUser.DriverGender != nil {
		info.DriverGender = *driverUser.DriverGender
	}
	if driverUser.CommercialType != nil {
		info.CommercialType = *driverUser.CommercialType
	}
	if driverUser.State != nil {
		info.State = *driverUser.State
	}

	// 填充车辆信息
	if car != nil {
		info.VehicleNo = car.VehicleNo
		if car.VehicleType != nil {
			info.VehicleType = *car.VehicleType
		}
		if car.Brand != nil {
			info.Brand = *car.Brand
		}
		if car.Model != nil {
			info.Model = *car.Model
		}
	}

	// 查询工作状态（独立查询，保留）
	workStatus, err := l.svcCtx.DriverWorkStatusModel.FindByDriverId(l.ctx, in.Id)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Errorf("查询司机工作状态失败, driverId=%d, err=%v", in.Id, err)
	}
	if workStatus != nil && workStatus.WorkStatus != nil {
		info.WorkStatus = int64(*workStatus.WorkStatus)
	}

	return info, nil
}
