package driverlogic

import (
	"context"
	"fmt"

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

	driverUser, err := l.svcCtx.DriverUserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("查询司机失败, driverId=%d, err=%v", in.Id, err)
		return nil, fmt.Errorf("查询失败，请稍后重试")
	}
	if driverUser == nil {
		return nil, fmt.Errorf("司机不存在")
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

	// 查询绑定的车辆信息
	binding, err := l.svcCtx.DriverCarBindingModel.FindByDriverId(l.ctx, in.Id)
	if err != nil {
		l.Errorf("查询司机车辆绑定失败, driverId=%d, err=%v", in.Id, err)
		// 绑定查询失败不影响基本信息返回，仅日志记录
	}
	if binding != nil && binding.CarId != nil {
		car, err := l.svcCtx.CarModel.FindOne(l.ctx, *binding.CarId)
		if err != nil {
			l.Errorf("查询车辆信息失败, carId=%d, err=%v", *binding.CarId, err)
		}
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
	}

	// 查询工作状态
	workStatus, err := l.svcCtx.DriverWorkStatusModel.FindByDriverId(l.ctx, in.Id)
	if err != nil {
		l.Errorf("查询司机工作状态失败, driverId=%d, err=%v", in.Id, err)
	}
	if workStatus != nil && workStatus.WorkStatus != nil {
		info.WorkStatus = int64(*workStatus.WorkStatus)
	}

	return info, nil
}
