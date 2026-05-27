package driverlogic

import (
	"context"
	"errors"
	"fmt"

	"onlinecar-driver/internal/model"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverWorkStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverWorkStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverWorkStatusLogic {
	return &GetDriverWorkStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取司机工作状态
func (l *GetDriverWorkStatusLogic) GetDriverWorkStatus(in *pb.GetDriverWorkStatusReq) (*pb.DriverWorkStatusInfo, error) {
	l.Infof("获取司机工作状态, driverId=%d", in.DriverId)

	workStatus, err := l.svcCtx.DriverWorkStatusModel.FindByDriverId(l.ctx, in.DriverId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &pb.DriverWorkStatusInfo{
				DriverId:   in.DriverId,
				WorkStatus: 0, // 默认收车
			}, nil
		}
		l.Errorf("查询司机工作状态失败, driverId=%d, err=%v", in.DriverId, err)
		return nil, fmt.Errorf("查询失败，请稍后重试")
	}

	return &pb.DriverWorkStatusInfo{
		DriverId:   in.DriverId,
		WorkStatus: int64(*workStatus.WorkStatus),
	}, nil
}
