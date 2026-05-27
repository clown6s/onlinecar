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
	districtCachePrefix = "district:list:"
	districtCacheTTL    = 24 * time.Hour
)

// 获取省级行政区列表（带 Redis 缓存）
func (l *GetProvincesLogic) GetProvinces(in *pb.GetProvincesReq) (*pb.GetProvincesResp, error) {
	l.Info("获取省级行政区列表")

	cacheKey := fmt.Sprintf("%s%s", districtCachePrefix, "province")

	// 1. 尝试读缓存
	cached, err := l.svcCtx.Redis.Get(cacheKey)
	if err == nil && cached != "" {
		var resp pb.GetProvincesResp
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
			l.Infof("命中缓存, key=%s, count=%d", cacheKey, len(resp.List))
			return &resp, nil
		}
	}

	// 2. 缓存未命中，查数据库（省级 parent_adcode 为空）
	list, err := l.svcCtx.DistrictModel.FindByParentAdCode(l.ctx, "")
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

	// 3. 写入缓存（失败不影响响应）
	if data, marshalErr := json.Marshal(resp); marshalErr == nil {
		if setErr := l.svcCtx.Redis.Setex(cacheKey, string(data), int(districtCacheTTL.Seconds())); setErr != nil {
			l.Errorf("写入缓存失败, key=%s, err=%v", cacheKey, setErr)
		}
	}

	l.Infof("查询省级行政区成功, count=%d", len(resp.List))
	return resp, nil
}
