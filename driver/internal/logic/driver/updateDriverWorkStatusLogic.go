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

type UpdateDriverWorkStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDriverWorkStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDriverWorkStatusLogic {
	return &UpdateDriverWorkStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新司机工作状态（0收车 1出车 2暂停）
func (l *UpdateDriverWorkStatusLogic) UpdateDriverWorkStatus(in *pb.UpdateDriverWorkStatusReq) (*pb.Empty, error) {
	l.Infof("更新司机工作状态, driverId=%d, workStatus=%d", in.DriverId, in.WorkStatus)

	if in.WorkStatus < 0 || in.WorkStatus > 2 {
		l.Errorf("无效的工作状态, workStatus=%d", in.WorkStatus)
		return nil, fmt.Errorf("无效的工作状态，仅支持0(收车)/1(出车)/2(暂停)")
	}

	_, err := l.svcCtx.DriverUserModel.FindOne(l.ctx, in.DriverId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, fmt.Errorf("司机不存在")
		}
		l.Errorf("查询司机失败, driverId=%d, err=%v", in.DriverId, err)
		return nil, fmt.Errorf("查询失败，请稍后重试")
	}

	err = l.svcCtx.DriverWorkStatusModel.Upsert(l.ctx, in.DriverId, in.WorkStatus)
	if err != nil {
		l.Errorf("更新司机工作状态失败, driverId=%d, err=%v", in.DriverId, err)
		return nil, fmt.Errorf("更新失败，请稍后重试")
	}

	l.Infof("司机工作状态更新成功, driverId=%d, workStatus=%d", in.DriverId, in.WorkStatus)
	return &pb.Empty{}, nil
}
