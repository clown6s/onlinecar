package trip

import (
	"context"

	"onlinecar-common/ctxdata"
	"onlinecar-common/errorx"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	passengerpb "onlinecar-gateway/pb/passenger"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTripLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTripLogic {
	return &CreateTripLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateTripLogic) CreateTrip(req *types.CreateTripReq) (*types.TripResp, error) {
	uid := ctxdata.GetUID(l.ctx)
	l.Infof("Gateway->PassengerRpc.CreateTrip, passengerId=%d, origin=%s, destination=%s",
		uid, req.Origin, req.Destination)

	resp, err := l.svcCtx.PassengerRpc.CreateTrip(l.ctx, &passengerpb.CreateTripReq{
		PassengerId: uid,
		Origin:      req.Origin,
		Destination: req.Destination,
		OriginLng:   req.OriginLng,
		OriginLat:   req.OriginLat,
		DestLng:     req.DestLng,
		DestLat:     req.DestLat,
	})
	if err != nil {
		l.Errorf("PassengerRpc.CreateTrip 失败, passengerId=%d, err=%v", uid, err)
		return nil, errorx.MapRpcError(err)
	}

	l.Infof("行程创建成功, tripId=%d, passengerId=%d", resp.Id, uid)
	return &types.TripResp{
		Id:          resp.Id,
		PassengerId: resp.PassengerId,
		DriverId:    resp.DriverId,
		Origin:      resp.Origin,
		Destination: resp.Destination,
		Status:      resp.Status,
		Fare:        resp.Fare,
	}, nil
}
