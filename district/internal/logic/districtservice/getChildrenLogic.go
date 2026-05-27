package districtservicelogic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"onlinecar-district/internal/svc"
	"onlinecar-district/pb"

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
	if err == nil && cached != "" {
		var resp pb.GetChildrenResp
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
			l.Infof("命中缓存, key=%s, count=%d", cacheKey, len(resp.List))
			return &resp, nil
		}
	}

	// 2. 缓存未命中，查数据库
	list, err := l.svcCtx.DistrictModel.FindByParentAdCode(l.ctx, in.ParentAdCode)
	if err != nil {
		l.Errorf("查询子级行政区失败, parentAdCode=%d, err=%v", in.ParentAdCode, err)
		return nil, err
	}

	resp := &pb.GetChildrenResp{}
	for _, d := range list {
		resp.List = append(resp.List, &pb.DistrictNode{
			Id:     d.Id,
			AdCode: d.AdCode,
			Name:   d.Name,
			Level:  d.Level,
			Center: d.Center,
		})
	}

	// 3. 写入缓存（失败不影响响应）
	if data, marshalErr := json.Marshal(resp); marshalErr == nil {
		if setErr := l.svcCtx.Redis.Setex(cacheKey, string(data), int(districtCacheTTL.Seconds())); setErr != nil {
			l.Errorf("写入缓存失败, key=%s, err=%v", cacheKey, setErr)
		}
	}

	l.Infof("查询子级行政区成功, parentAdCode=%d, count=%d", in.ParentAdCode, len(resp.List))
	return resp, nil
}
