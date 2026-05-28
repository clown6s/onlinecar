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
	lastUpdateKey := "driver:last_update"
	member := fmt.Sprintf("%d", in.Id)

	if l.svcCtx.Redis == nil {
		l.Errorf("Redis 未配置, driverId=%d", in.Id)
		return nil, fmt.Errorf("更新位置失败")
	}

	// GEOADD 写入位置，使用 HSET 记录每个司机的最后更新时间戳
	// 不再使用全局 EXPIRE，改由后台 goroutine 根据 driver:last_update 清理过期数据
	script := `
redis.call('GEOADD', KEYS[1], ARGV[1], ARGV[2], ARGV[3])
redis.call('HSET', KEYS[2], ARGV[3], ARGV[4])
return 1
`
	// 不再使用全局 EXPIRE，改用 HSET 记录每个司机最后更新时间戳
	// 后台 goroutine 根据 driver:last_update 中的时间戳清理过期数据（24h 阈值）
	now := time.Now().Unix()
	result, err := l.svcCtx.Redis.Eval(script, []string{key, lastUpdateKey}, in.Lng, in.Lat, member, now)
	if err != nil {
		l.Errorf("Redis GEOADD 失败, driverId=%d, err=%v", in.Id, err)
		return nil, fmt.Errorf("更新位置失败")
	}
	if result == nil {
		l.Errorf("Redis Lua 返回 nil, driverId=%d", in.Id)
		return nil, fmt.Errorf("更新位置失败")
	}
	// Lua 返回值类型可能是 int64 或 string，安全处理
	var resultCode int64
	switch v := result.(type) {
	case int64:
		resultCode = v
	case string:
		l.Errorf("Redis Lua 返回非预期类型 string: %s, driverId=%d", v, in.Id)
		return nil, fmt.Errorf("更新位置失败")
	default:
		l.Errorf("Redis Lua 返回非预期类型 %T: %v, driverId=%d", result, result, in.Id)
		return nil, fmt.Errorf("更新位置失败")
	}
	if resultCode != 1 {
		l.Errorf("更新司机位置返回异常: driverId=%d, result=%d", in.Id, resultCode)
		return nil, fmt.Errorf("更新位置失败")
	}

	return &pb.Empty{}, nil
}
