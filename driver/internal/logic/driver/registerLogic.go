package driverlogic

import (
	"context"
	"fmt"
	"time"

	"onlinecar-driver/internal/model"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册司机
func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	l.Infof("司机注册, phone=%s, name=%s", in.DriverPhone, in.DriverName)

	// 检查手机号是否已注册
	existing, err := l.svcCtx.DriverUserModel.FindByPhone(l.ctx, in.DriverPhone)
	if err != nil {
		l.Errorf("查询司机失败, phone=%s, err=%v", in.DriverPhone, err)
		return nil, fmt.Errorf("查询司机失败: %v", err)
	}
	if existing != nil {
		l.Errorf("手机号已注册, phone=%s", in.DriverPhone)
		return nil, fmt.Errorf("该手机号已注册")
	}

	commercialType := in.CommercialType
	if commercialType <= 0 {
		commercialType = 1 // 默认网约车
	}

	driverUser := &model.DriverUser{
		DriverName:     in.DriverName,
		DriverPhone:    in.DriverPhone,
		CommercialType: &commercialType,
	}

	// 插入司机
	result, err := l.svcCtx.DriverUserModel.Insert(l.ctx, driverUser)
	if err != nil {
		l.Errorf("插入司机失败, phone=%s, err=%v", in.DriverPhone, err)
		return nil, fmt.Errorf("注册失败: %v", err)
	}
	driverId, _ := result.LastInsertId()

	// 如果有车辆信息，同时插入车辆并绑定
	if in.VehicleNo != "" {
		var vehicleType string
		if in.VehicleType != "" {
			vehicleType = in.VehicleType
		}
		ct := int(commercialType)
		car := &model.Car{
			VehicleNo:      in.VehicleNo,
			VehicleType:    &vehicleType,
			CommercialType: &ct,
		}
		carResult, err := l.svcCtx.CarModel.Insert(l.ctx, car)
		if err != nil {
			l.Errorf("插入车辆失败, vehicleNo=%s, err=%v", in.VehicleNo, err)
		} else {
			carId, _ := carResult.LastInsertId()
			bindState := 1
			now := time.Now()
			binding := &model.DriverCarBinding{
				Id:          driverId,
				DriverId:    &driverId,
				CarId:       &carId,
				BindState:   &bindState,
				BindingTime: &now,
			}
			if _, err := l.svcCtx.DriverCarBindingModel.Insert(l.ctx, binding); err != nil {
				l.Errorf("绑定车辆失败, driverId=%d, carId=%d, err=%v", driverId, carId, err)
			}
		}
	}

	l.Infof("司机注册成功, driverId=%d", driverId)
	return &pb.RegisterResp{
		Id:          driverId,
		DriverName:  in.DriverName,
		DriverPhone: in.DriverPhone,
		State:       0,
	}, nil
}
