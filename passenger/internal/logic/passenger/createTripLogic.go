package passengerlogic

import (
	"context"
	"fmt"

	"onlinecar-passenger/internal/model"
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

	if in.PassengerId <= 0 {
		return nil, fmt.Errorf("无效的乘客ID")
	}
	if in.Origin == "" {
		return nil, fmt.Errorf("起点不能为空")
	}
	if in.Destination == "" {
		return nil, fmt.Errorf("终点不能为空")
	}

	trip := &model.Trip{
		PassengerId: in.PassengerId,
		Origin:      in.Origin,
		Destination: in.Destination,
		OriginLng:   in.OriginLng,
		OriginLat:   in.OriginLat,
		DestLng:     in.DestLng,
		DestLat:     in.DestLat,
		Status:      0, // 待接单
		Fare:        0,
	}

	result, err := l.svcCtx.TripModel.Insert(l.ctx, trip)
	if err != nil {
		l.Errorf("创建行程失败, passengerId=%d, err=%v", in.PassengerId, err)
		return nil, fmt.Errorf("创建行程失败")
	}

	tripId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取行程ID失败, err=%v", err)
		return nil, fmt.Errorf("创建行程失败")
	}

	l.Infof("行程创建成功, tripId=%d", tripId)
	return &pb.TripResp{
		Id:          tripId,
		PassengerId: in.PassengerId,
		Origin:      in.Origin,
		Destination: in.Destination,
		Status:      0,
	}, nil
}
