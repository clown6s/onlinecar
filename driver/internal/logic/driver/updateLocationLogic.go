package driverlogic

import (
	"context"

	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLocationLogic {
	return &UpdateLocationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新位置（写入 Redis Geo）
func (l *UpdateLocationLogic) UpdateLocation(in *pb.UpdateLocationReq) (*pb.Empty, error) {
	l.Infof("更新司机位置, driverId=%d, lng=%.6f, lat=%.6f", in.Id, in.Lng, in.Lat)

	// TODO: 使用 Redis GEO 写入位置信息
	// l.svcCtx.Redis.GeoAdd(l.ctx, "driver:location", &redis.GeoLocation{
	//     Name:      fmt.Sprintf("%d", in.Id),
	//     Longitude: in.Lng,
	//     Latitude:  in.Lat,
	// })

	return &pb.Empty{}, nil
}
