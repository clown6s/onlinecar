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

	// 验证乘客是否存在
	passenger, err := l.svcCtx.PassengerUserModel.FindOne(l.ctx, in.PassengerId)
	if err != nil {
		l.Errorf("查询乘客失败, passengerId=%d, err=%v", in.PassengerId, err)
		return nil, fmt.Errorf("查询乘客失败")
	}
	if passenger == nil {
		return nil, fmt.Errorf("乘客不存在")
	}
	if passenger.State != nil && *passenger.State == 1 {
		return nil, fmt.Errorf("乘客账号已禁用")
	}

	// 验证坐标范围
	if in.OriginLng < -180 || in.OriginLng > 180 || in.OriginLat < -90 || in.OriginLat > 90 {
		return nil, fmt.Errorf("起点坐标无效")
	}
	if in.DestLng < -180 || in.DestLng > 180 || in.DestLat < -90 || in.DestLat > 90 {
		return nil, fmt.Errorf("终点坐标无效")
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
