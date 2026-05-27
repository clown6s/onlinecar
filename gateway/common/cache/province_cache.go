package cache

import (
	"context"
	"sync"
	"time"

	"onlinecar-gateway/internal/types"
	districtpb "onlinecar-gateway/pb/district"

	"github.com/zeromicro/go-zero/core/logx"
)

// ProvinceCache 省级行政区本地内存缓存
// 采用异步刷新：后台 goroutine 定期拉取，请求路径只做 RLock 读，零阻塞
type ProvinceCache struct {
	client districtpb.DistrictServiceClient
	data   []*types.DistrictNodeResp
	mu     sync.RWMutex
	ttl    time.Duration
	cancel context.CancelFunc
}

func NewProvinceCache(client districtpb.DistrictServiceClient, ttl time.Duration) *ProvinceCache {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &ProvinceCache{
		client: client,
		ttl:    ttl,
	}
}

// Get 获取省份列表，纯内存读取，不走 RPC
func (c *ProvinceCache) Get(_ context.Context) ([]*types.DistrictNodeResp, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.data == nil {
		return []*types.DistrictNodeResp{}, nil
	}
	result := make([]*types.DistrictNodeResp, len(c.data))
	copy(result, c.data)
	return result, nil
}

// Start 启动后台异步刷新 goroutine
func (c *ProvinceCache) Start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)

	// 启动时同步加载，确保首次请求不空
	c.load(ctx)

	go func() {
		ticker := time.NewTicker(c.ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.load(ctx)
			}
		}
	}()
	logx.Infof("ProvinceCache 异步刷新已启动, ttl=%v", c.ttl)
}

// Stop 停止后台刷新
func (c *ProvinceCache) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *ProvinceCache) load(ctx context.Context) {
	resp, err := c.client.GetProvinces(ctx, &districtpb.GetProvincesReq{})
	if err != nil {
		logx.Errorf("ProvinceCache 刷新失败, err=%v", err)
		return
	}

	list := make([]*types.DistrictNodeResp, 0, len(resp.List))
	for _, node := range resp.List {
		list = append(list, &types.DistrictNodeResp{
			Id:     node.Id,
			AdCode: node.AdCode,
			Name:   node.Name,
			Level:  node.Level,
			Center: node.Center,
		})
	}

	c.mu.Lock()
	c.data = list
	c.mu.Unlock()
	logx.Infof("ProvinceCache 刷新成功, count=%d", len(list))
}
