package driverlogic

import (
	"context"
	"fmt"
	"time"

	"onlinecar-common/errorx"
	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLocationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// updateLocationScript 原子写入司机位置 + 更新时间戳
// 注意：此脚本不为 Redis key 设置 TTL，过期数据由 driver.go 中的后台 goroutine
// 每小时扫描 driver:last_update 并清理超过 24h 的条目。这是有意设计：
// 如果服务重启，goroutine 会自动恢复清理，无需担心数据无限增长。
const updateLocationScript = `
redis.call('GEOADD', KEYS[1], ARGV[1], ARGV[2], ARGV[3])
redis.call('HSET', KEYS[2], ARGV[3], ARGV[4])
return 1
`

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
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "无效的司机ID")
	}

	key := "driver:location"
	lastUpdateKey := "driver:last_update"
	member := fmt.Sprintf("%d", in.Id)

	if l.svcCtx.Redis == nil {
		l.Errorf("Redis 未配置, driverId=%d", in.Id)
		return nil, errorx.ErrServer
	}

	// GEOADD 写入位置，使用 HSET 记录每个司机的最后更新时间戳
	// 不再使用全局 EXPIRE，改由后台 goroutine 根据 driver:last_update 清理过期数据（24h 阈值）
	now := time.Now().Unix()
	result, err := l.svcCtx.Redis.Eval(updateLocationScript, []string{key, lastUpdateKey}, in.Lng, in.Lat, member, now)
	if err != nil {
		l.Errorf("Redis GEOADD 失败, driverId=%d, err=%v", in.Id, err)
		return nil, errorx.ErrServer
	}
	if result == nil {
		l.Errorf("Redis Lua 返回 nil, driverId=%d", in.Id)
		return nil, errorx.ErrServer
	}
	// Lua 返回值类型可能是 int64 或 string，安全处理
	var resultCode int64
	switch v := result.(type) {
	case int64:
		resultCode = v
	case string:
		l.Errorf("Redis Lua 返回非预期类型 string: %s, driverId=%d", v, in.Id)
		return nil, errorx.ErrServer
	default:
		l.Errorf("Redis Lua 返回非预期类型 %T: %v, driverId=%d", result, result, in.Id)
		return nil, errorx.ErrServer
	}
	if resultCode != 1 {
		l.Errorf("更新司机位置返回异常: driverId=%d, result=%d", in.Id, resultCode)
		return nil, errorx.ErrServer
	}

	return &pb.Empty{}, nil
}
