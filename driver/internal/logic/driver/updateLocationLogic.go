package driverlogic

import (
	"context"
	"fmt"
	"time"

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
	// 原子操作：GEOADD + EXPIRE（司机超过 24h 未更新位置则自动清理）
	// EXPIRE 作用于整个 key，每次有司机更新位置都会刷新 TTL
	script := `
redis.call('GEOADD', KEYS[1], ARGV[1], ARGV[2], ARGV[3])
redis.call('EXPIRE', KEYS[1], ARGV[4])
return 1
`
	driverLocationTTL := int((24 * time.Hour).Seconds())
	_, err := l.svcCtx.Redis.Eval(script, []string{key}, in.Lng, in.Lat, member, driverLocationTTL)
	if err != nil {
		l.Errorf("Redis GEOADD 失败, driverId=%d, err=%v", in.Id, err)
		return nil, fmt.Errorf("更新位置失败")
	}

	return &pb.Empty{}, nil
}
