package districtservicelogic

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"onlinecar-district/internal/svc"
	"onlinecar-district/pb"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

type GetProvincesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProvincesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProvincesLogic {
	return &GetProvincesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

const (
	districtCachePrefix  = "district:list:"
	districtCacheBaseTTL = 24 * time.Hour
)

// sfGroup 用于缓存击穿防护，同一时刻只有一个 goroutine 查询 DB
var sfGroup singleflight.Group

// cacheTTL 返回带随机偏移的 TTL，防止缓存雪崩
func cacheTTL() time.Duration {
	// 基础 24h ± 随机 0~2h 偏移，防止雪崩
	jitter := time.Duration(rand.IntN(7200)) * time.Second // 0~2h
	return districtCacheBaseTTL + jitter - time.Hour       // 23~25h
}

// 获取省级行政区列表（带 Redis 缓存，Redis 不可用时降级为直查 DB）
func (l *GetProvincesLogic) GetProvinces(in *pb.GetProvincesReq) (*pb.GetProvincesResp, error) {
	l.Info("获取省级行政区列表")

	cacheKey := fmt.Sprintf("%s%s", districtCachePrefix, "province")

	// 1. 尝试读缓存（Redis 可用时）
	if l.svcCtx.Redis != nil {
		cached, err := l.svcCtx.Redis.Get(cacheKey)
		if err != nil && err != redis.Nil {
			l.Errorf("Redis 读取失败, key=%s, err=%v", cacheKey, err)
		}
		if err == nil && cached != "" {
			var resp pb.GetProvincesResp
			if err = json.Unmarshal([]byte(cached), &resp); err == nil {
				l.Infof("命中缓存, key=%s, count=%d", cacheKey, len(resp.List))
				return &resp, nil
			}
			l.Errorf("缓存反序列化失败, key=%s, err=%v", cacheKey, err)
		}
	}

	// 2. 缓存未命中，使用 singleflight 防止击穿
	v, sfErr, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
		// 检查 context 是否已取消
		select {
		case <-l.ctx.Done():
			return nil, l.ctx.Err()
		default:
		}

		// double check：再次读缓存（Redis 可用时）
		if l.svcCtx.Redis != nil {
			if cached, err := l.svcCtx.Redis.Get(cacheKey); err == nil && cached != "" {
				var resp pb.GetProvincesResp
				if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
					return &resp, nil
				}
			}
		}

		// 查数据库
		list, err := l.svcCtx.DistrictModel.FindProvinces(l.ctx)
		if err != nil {
			l.Errorf("查询省级行政区失败, err=%v", err)
			return nil, err
		}

		resp := &pb.GetProvincesResp{}
		for _, d := range list {
			node := &pb.DistrictNode{
				Id:     d.Id,
				Name:   d.Name,
				Level:  d.Level,
				Center: d.Center,
			}
			if d.AdCode != nil {
				node.AdCode = *d.AdCode
			}
			resp.List = append(resp.List, node)
		}

		// 写入缓存（失败不影响响应，Redis 不可用时跳过）
		if l.svcCtx.Redis != nil {
			if data, marshalErr := json.Marshal(resp); marshalErr == nil {
				if setErr := l.svcCtx.Redis.Setex(cacheKey, string(data), int(cacheTTL().Seconds())); setErr != nil {
					l.Errorf("写入缓存失败, key=%s, err=%v", cacheKey, setErr)
				}
			}
		}

		return resp, nil
	})
	if sfErr != nil {
		return nil, sfErr
	}

	resp, ok := v.(*pb.GetProvincesResp)
	if !ok {
		return nil, fmt.Errorf("singleflight returned unexpected type: %T", v)
	}

	l.Infof("查询省级行政区成功, count=%d", resp.List)
	return resp, nil
}
