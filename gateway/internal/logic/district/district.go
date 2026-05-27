package district

import (
	"context"

	"onlinecar-gateway/common/errorx"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	districtpb "onlinecar-gateway/pb/district"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProvincesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProvincesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProvincesLogic {
	return &GetProvincesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetProvincesLogic) GetProvinces() ([]*types.DistrictNodeResp, error) {
	l.Info("Gateway->DistrictRpc.GetProvinces")

	resp, err := l.svcCtx.DistrictRpc.GetProvinces(l.ctx, &districtpb.GetProvincesReq{})
	if err != nil {
		l.Errorf("DistrictRpc.GetProvinces 失败, err=%v", err)
		return nil, errorx.MapRpcError(err)
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

	l.Infof("查询省级行政区成功, count=%d", len(list))
	return list, nil
}

type GetChildrenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChildrenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildrenLogic {
	return &GetChildrenLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetChildren 获取子级行政区
func (l *GetChildrenLogic) GetChildren(parentAdCode uint32) ([]*types.DistrictNodeResp, error) {
	l.Infof("Gateway->DistrictRpc.GetChildren, parentAdCode=%d", parentAdCode)

	resp, err := l.svcCtx.DistrictRpc.GetChildren(l.ctx, &districtpb.GetChildrenReq{
		ParentAdCode: parentAdCode,
	})
	if err != nil {
		l.Errorf("DistrictRpc.GetChildren 失败, parentAdCode=%d, err=%v", parentAdCode, err)
		return nil, errorx.MapRpcError(err)
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

	l.Infof("查询子级行政区成功, parentAdCode=%d, count=%d", parentAdCode, len(list))
	return list, nil
}
