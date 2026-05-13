package passengerlogic

import (
	"context"

	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTripLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTripLogic {
	return &CreateTripLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建行程
func (l *CreateTripLogic) CreateTrip(in *pb.CreateTripReq) (*pb.TripResp, error) {
	l.Infof("创建行程, passengerId=%d, origin=%s, destination=%s",
		in.PassengerId, in.Origin, in.Destination)

	// TODO: 需要创建 trip 表后实现
	// 目前返回空响应

	return &pb.TripResp{}, nil
}
