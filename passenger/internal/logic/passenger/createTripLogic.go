package passengerlogic

import (
	"context"
	"errors"
	"math"
	"unicode/utf8"

	"onlinecar-common/errorx"
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
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "无效的乘客ID")
	}
	if in.Origin == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "起点不能为空")
	}
	if in.Destination == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "终点不能为空")
	}

	// 输入校验：地址长度限制
	if utf8.RuneCountInString(in.Origin) > 200 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "起点地址过长")
	}
	if utf8.RuneCountInString(in.Destination) > 200 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "终点地址过长")
	}

	// 验证乘客是否存在
	passenger, err := l.svcCtx.PassengerUserModel.FindOne(l.ctx, in.PassengerId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "乘客不存在")
		}
		logx.Errorf("查询乘客失败, passengerId=%d, err=%v", in.PassengerId, err)
		return nil, errorx.ErrServer
	}
	if passenger.State != nil && *passenger.State == 1 {
		return nil, errorx.ErrUserDisabled
	}

	// 验证坐标范围（含 NaN/Inf 检查，NaN < -180 为 false 可绕过范围校验）
	if math.IsNaN(in.OriginLng) || math.IsInf(in.OriginLng, 0) || in.OriginLng < -180 || in.OriginLng > 180 ||
		math.IsNaN(in.OriginLat) || math.IsInf(in.OriginLat, 0) || in.OriginLat < -90 || in.OriginLat > 90 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "起点坐标无效")
	}
	if math.IsNaN(in.DestLng) || math.IsInf(in.DestLng, 0) || in.DestLng < -180 || in.DestLng > 180 ||
		math.IsNaN(in.DestLat) || math.IsInf(in.DestLat, 0) || in.DestLat < -90 || in.DestLat > 90 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "终点坐标无效")
	}

	// 检查是否有未完成的行程（status=0 待接单 或 status=1 进行中），一次查询
	activeTrips, err := l.svcCtx.TripModel.FindActiveTrips(l.ctx, in.PassengerId, []int8{0, 1})
	if err != nil {
		logx.Errorf("查询未完成行程失败, passengerId=%d, err=%v", in.PassengerId, err)
		return nil, errorx.ErrServer
	}
	if len(activeTrips) > 0 {
		return nil, errorx.NewCodeError(errorx.ErrCodeTripStatusInvalid, "您有未完成的行程，请等待完成后再创建新行程")
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
		logx.Errorf("创建行程失败, passengerId=%d, err=%v", in.PassengerId, err)
		return nil, errorx.ErrTripCreateFailed
	}

	tripId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("获取行程ID失败, err=%v", err)
		return nil, errorx.ErrTripCreateFailed
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
