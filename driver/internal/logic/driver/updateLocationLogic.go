package driverlogic

import (
	"context"
	"fmt"

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

	if in.Id <= 0 {
		return nil, fmt.Errorf("无效的司机ID")
	}

	key := "driver:location"
	member := fmt.Sprintf("%d", in.Id)
	// go-zero Redis 封装未直接暴露 GeoAdd，通过 Eval 执行 Redis GEOADD 命令
	script := "redis.call('GEOADD', KEYS[1], ARGV[1], ARGV[2], ARGV[3])"
	_, err := l.svcCtx.Redis.Eval(script, []string{key}, in.Lng, in.Lat, member)
	if err != nil {
		l.Errorf("Redis GEOADD 失败, driverId=%d, err=%v", in.Id, err)
		return nil, fmt.Errorf("更新位置失败")
	}

	return &pb.Empty{}, nil
}
