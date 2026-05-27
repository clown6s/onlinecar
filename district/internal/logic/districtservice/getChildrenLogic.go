package districtservicelogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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

	cacheKey := fmt.Sprintf("%s%d", districtCachePrefix, in.ParentAdCode)

	// 1. 尝试读缓存
	cached, err := l.svcCtx.Redis.Get(cacheKey)
	if err != nil && err != redis.Nil {
		l.Errorf("Redis 读取失败, key=%s, err=%v", cacheKey, err)
	}
	if err == nil && cached != "" {
		var resp pb.GetChildrenResp
		if err = json.Unmarshal([]byte(cached), &resp); err == nil {
			l.Infof("命中缓存, key=%s, count=%d", cacheKey, len(resp.List))
			return &resp, nil
		}
		l.Errorf("缓存反序列化失败, key=%s, err=%v", cacheKey, err)
	}

	// 2. 缓存未命中，使用 singleflight 防止击穿
	v, sfErr, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
		// double check：再次读缓存，可能其他 goroutine 已填充
		if cached, err := l.svcCtx.Redis.Get(cacheKey); err == nil && cached != "" {
			var resp pb.GetChildrenResp
			if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
				return &resp, nil
			}
		}

		// 查数据库（parentAdCode 转 string 给 model）
		list, err := l.svcCtx.DistrictModel.FindByParentAdCode(l.ctx, strconv.FormatUint(uint64(in.ParentAdCode), 10))
		if err != nil {
			l.Errorf("查询子级行政区失败, parentAdCode=%d, err=%v", in.ParentAdCode, err)
			return nil, err
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
			if setErr := l.svcCtx.Redis.Setex(cacheKey, string(data), int(cacheTTL().Seconds())); setErr != nil {
				l.Errorf("写入缓存失败, key=%s, err=%v", cacheKey, setErr)
			}
		}

		return resp, nil
	})
	if sfErr != nil {
		return nil, sfErr
	}

	l.Infof("查询子级行政区成功, parentAdCode=%d, count=%d", in.ParentAdCode, v.(*pb.GetChildrenResp).List)
	return v.(*pb.GetChildrenResp), nil
}
