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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

// 创建行程（事务：验证乘客 + 检查活跃行程 + 插入行程）
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

	// 验证坐标范围（含 NaN/Inf 检查）
	if math.IsNaN(in.OriginLng) || math.IsInf(in.OriginLng, 0) || in.OriginLng < -180 || in.OriginLng > 180 ||
		math.IsNaN(in.OriginLat) || math.IsInf(in.OriginLat, 0) || in.OriginLat < -90 || in.OriginLat > 90 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "起点坐标无效")
	}
	if math.IsNaN(in.DestLng) || math.IsInf(in.DestLng, 0) || in.DestLng < -180 || in.DestLng > 180 ||
		math.IsNaN(in.DestLat) || math.IsInf(in.DestLat, 0) || in.DestLat < -90 || in.DestLat > 90 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "终点坐标无效")
	}

	var tripId int64

	// 事务内原子执行：验证乘客 + 检查活跃行程 + 插入行程
	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 验证乘客是否存在且未禁用
		passenger, err := l.svcCtx.PassengerUserModel.FindOne(l.ctx, in.PassengerId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return errorx.NewCodeError(errorx.ErrCodeUserNotFound, "乘客不存在")
			}
			return errorx.ErrServer
		}
		if passenger.State != nil && *passenger.State == 1 {
			return errorx.ErrUserDisabled
		}

		// 2. 检查是否有未完成的行程（事务内查询，避免 TOCTOU 竞态）
		activeTrips, err := l.svcCtx.TripModel.FindActiveTripsWithConn(ctx, session, in.PassengerId,
			[]int8{model.TripStatusCreated, model.TripStatusAccepted})
		if err != nil {
			return errorx.ErrServer
		}
		if len(activeTrips) > 0 {
			return errorx.NewCodeError(errorx.ErrCodeTripStatusInvalid, "您有未完成的行程，请等待完成后再创建新行程")
		}

		// 3. 插入行程
		trip := &model.Trip{
			PassengerId: in.PassengerId,
			Origin:      in.Origin,
			Destination: in.Destination,
			OriginLng:   in.OriginLng,
			OriginLat:   in.OriginLat,
			DestLng:     in.DestLng,
			DestLat:     in.DestLat,
			Status:      model.TripStatusCreated,
			Fare:        0,
		}

		result, err := l.svcCtx.TripModel.InsertWithConn(ctx, session, trip)
		if err != nil {
			return errorx.ErrTripCreateFailed
		}

		id, err := result.LastInsertId()
		if err != nil {
			return errorx.ErrTripCreateFailed
		}
		tripId = id
		return nil
	})
	if err != nil {
		var codeErr *errorx.CodeError
		if errors.As(err, &codeErr) {
			return nil, codeErr
		}
		logx.Errorf("创建行程失败, passengerId=%d, err=%v", in.PassengerId, err)
		return nil, errorx.ErrServer
	}

	l.Infof("行程创建成功, tripId=%d", tripId)
	return &pb.TripResp{
		Id:          tripId,
		PassengerId: in.PassengerId,
		Origin:      in.Origin,
		Destination: in.Destination,
		Status:      int64(model.TripStatusCreated),
	}, nil
}
