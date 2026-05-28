package districtservicelogic

import (
	"context"
	"encoding/json"
	"fmt"

	"onlinecar-district/internal/model"
	"onlinecar-district/internal/svc"
	"onlinecar-district/pb"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChildrenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChildrenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildrenLogic {
	return &GetChildrenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取指定行政区划的直接下级子节点（带 Redis 缓存）
func (l *GetChildrenLogic) GetChildren(in *pb.GetChildrenReq) (*pb.GetChildrenResp, error) {
	l.Infof("获取子级行政区, parentAdCode=%d", in.ParentAdCode)

	// parentAdCode=0 时复用省级缓存 key，避免同一份数据缓存两份
	cacheKey := fmt.Sprintf("%s%s", districtCachePrefix, "province")
	if in.ParentAdCode != 0 {
		cacheKey = fmt.Sprintf("%s%d", districtCachePrefix, in.ParentAdCode)
	}

	// 1. 尝试读缓存
	cached, err := l.svcCtx.Redis.Get(cacheKey)
	if err != nil && err != redis.Nil {
		l.Errorf("Redis 读取失败, key=%s, err=%v", cacheKey, err)
	}
	if err == nil && cached != "" {
		var resp pb.GetChildrenResp
		if err = json.Unmarshal([]byte(cached), &resp); err == nil {
			// 缓存命中，空数组也直接返回（防穿透）
			l.Infof("命中缓存, key=%s, count=%d", cacheKey, len(resp.List))
			return &resp, nil
		}
		l.Errorf("缓存反序列化失败, key=%s, err=%v", cacheKey, err)
	}

	// 2. 缓存未命中，使用 singleflight 防止击穿
	v, sfErr, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
		// 检查 context 是否已取消
		select {
		case <-l.ctx.Done():
			return nil, l.ctx.Err()
		default:
		}

		// double check：再次读缓存，可能其他 goroutine 已填充
		if cached, err := l.svcCtx.Redis.Get(cacheKey); err == nil && cached != "" {
			var resp pb.GetChildrenResp
			if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
				return &resp, nil
			}
		}

		var list []*model.District
		var dbErr error

		if in.ParentAdCode == 0 {
			// 复用省级查询逻辑
			list, dbErr = l.svcCtx.DistrictModel.FindProvinces(l.ctx)
		} else {
			list, dbErr = l.svcCtx.DistrictModel.FindByParentAdCode(l.ctx, in.ParentAdCode)
		}
		if dbErr != nil {
			l.Errorf("查询子级行政区失败, parentAdCode=%d, err=%v", in.ParentAdCode, dbErr)
			return nil, dbErr
		}

		resp := &pb.GetChildrenResp{}
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

		// 写入缓存（失败不影响响应）
		if data, marshalErr := json.Marshal(resp); marshalErr == nil {
			ttl := int(cacheTTL().Seconds())
			if len(list) == 0 {
				// 缓存空结果，防穿透，5分钟TTL
				ttl = 300
			}
			if setErr := l.svcCtx.Redis.Setex(cacheKey, string(data), ttl); setErr != nil {
				l.Errorf("写入缓存失败, key=%s, err=%v", cacheKey, setErr)
			}
		}

		return resp, nil
	})
	if sfErr != nil {
		return nil, sfErr
	}

	resp, ok := v.(*pb.GetChildrenResp)
	if !ok {
		return nil, fmt.Errorf("singleflight returned unexpected type: %T", v)
	}

	l.Infof("查询子级行政区成功, parentAdCode=%d, count=%d", in.ParentAdCode, resp.List)
	return resp, nil
}
