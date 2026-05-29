package driverlogic

import (
	"context"
	"fmt"
	"math"
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
// 使用 per-driver key（driver:last_update:{id}）而非单个 hash，避免活跃司机重置所有历史记录的 TTL。
// GEO key 和 per-driver key 均设置 48h EXPIRE 作为兜底清理机制。
const updateLocationScript = `
redis.call('GEOADD', KEYS[1], ARGV[1], ARGV[2], ARGV[3])
redis.call('SET', KEYS[2], ARGV[4], 'EX', ARGV[5])
redis.call('EXPIRE', KEYS[1], ARGV[6])
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

	// 经纬度合法性校验
	if math.IsNaN(in.Lng) || math.IsInf(in.Lng, 0) || math.IsNaN(in.Lat) || math.IsInf(in.Lat, 0) {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "坐标无效，经纬度不能为NaN或无穷大")
	}
	if in.Lng < -180 || in.Lng > 180 || in.Lat < -90 || in.Lat > 90 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "坐标无效，经度范围[-180,180]，纬度范围[-90,90]")
	}

	key := "driver:location"
	lastUpdateKey := fmt.Sprintf("driver:last_update:%d", in.Id)
	member := fmt.Sprintf("%d", in.Id)

	if l.svcCtx.Redis == nil {
		l.Errorf("Redis 未配置, driverId=%d", in.Id)
		return nil, errorx.ErrServer
	}

	// GEOADD 写入位置，SET per-driver 更新时间戳（各自独立 TTL）
	now := time.Now().Unix()
	ttlSeconds := int64(48 * time.Hour.Seconds())
	result, err := l.svcCtx.Redis.Eval(updateLocationScript, []string{key, lastUpdateKey}, in.Lng, in.Lat, member, now, ttlSeconds, ttlSeconds)
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
