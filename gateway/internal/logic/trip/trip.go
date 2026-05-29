package trip

import (
	"context"
	"fmt"

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

// formatFare 将 proto double 金额转为字符串格式，保留2位小数。
// proto 层使用 double 传输（float64 精度对存储/传输足够），
// HTTP API 层使用 string 输出，避免前端 JavaScript 浮点精度丢失（如 0.1+0.2≠0.3）。
// TODO: 当 proto fare 字段改为 string 后，移除此函数并直接赋值。
func formatFare(fare float64) string {
	return fmt.Sprintf("%.2f", fare)
}

func NewCreateTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTripLogic {
	return &CreateTripLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateTripLogic) CreateTrip(req *types.CreateTripReq) (*types.TripResp, error) {
	uid, ok := ctxdata.GetUID(l.ctx)
	if !ok || uid <= 0 {
		l.Errorf("CreateTrip: UID 未设置或无效, uid=%d", uid)
		return nil, errorx.ErrAuth
	}
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
	if resp == nil {
		l.Errorf("PassengerRpc.CreateTrip 返回 nil, passengerId=%d", uid)
		return nil, errorx.ErrServer
	}

	l.Infof("行程创建成功, tripId=%d, passengerId=%d", resp.Id, uid)
	return &types.TripResp{
		Id:          resp.Id,
		PassengerId: resp.PassengerId,
		DriverId:    resp.DriverId,
		Origin:      resp.Origin,
		Destination: resp.Destination,
		Status:      resp.Status,
		Fare:        formatFare(resp.Fare), // double→string，保留2位小数
	}, nil
}
