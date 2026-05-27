package driverlogic

import (
	"context"
	"fmt"
	"time"

	"onlinecar-driver/common/errorx"
	"onlinecar-driver/internal/model"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

// 注册司机（事务：检查手机号 + 插入司机 + 车辆 + 绑定关系）
func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	l.Infof("司机注册, phone=%s, name=%s", in.DriverPhone, in.DriverName)

	commercialType := in.CommercialType
	if commercialType <= 0 {
		commercialType = 1 // 默认网约车
	}

	var driverId int64

	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 事务内检查手机号是否已注册（避免 TOCTOU 竞态）
		existing, err := l.svcCtx.DriverUserModel.FindByPhoneWithConn(ctx, session, in.DriverPhone)
		if err != nil {
			return fmt.Errorf("查询司机失败: %w", err)
		}
		if existing != nil {
			return errorx.ErrPhoneRegistered
		}

		// 2. 插入司机
		driverUser := &model.DriverUser{
			DriverName:     in.DriverName,
			DriverPhone:    in.DriverPhone,
			CommercialType: &commercialType,
		}
		result, err := l.svcCtx.DriverUserModel.InsertWithConn(ctx, session, driverUser)
		if err != nil {
			return fmt.Errorf("插入司机失败: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("获取司机ID失败: %w", err)
		}
		driverId = id

		// 3. 如果有车辆信息，插入车辆并绑定
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
			carResult, err := l.svcCtx.CarModel.InsertWithConn(ctx, session, car)
			if err != nil {
				return fmt.Errorf("插入车辆失败: %w", err)
			}
			carId, err := carResult.LastInsertId()
			if err != nil {
				return fmt.Errorf("获取车辆ID失败: %w", err)
			}

			// 4. 插入绑定关系
			bindState := 1
			now := time.Now()
			binding := &model.DriverCarBinding{
				DriverId:    &driverId,
				CarId:       &carId,
				BindState:   &bindState,
				BindingTime: &now,
			}
			if _, err := l.svcCtx.DriverCarBindingModel.InsertWithConn(ctx, session, binding); err != nil {
				return fmt.Errorf("绑定车辆失败: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		l.Errorf("司机注册事务失败, phone=%s, err=%v", in.DriverPhone, err)
		return nil, err
	}

	l.Infof("司机注册成功, driverId=%d", driverId)
	return &pb.RegisterResp{
		Id:          driverId,
		DriverName:  in.DriverName,
		DriverPhone: in.DriverPhone,
		State:       0,
	}, nil
}
