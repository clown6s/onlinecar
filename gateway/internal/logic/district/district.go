package district

import (
	"context"
	"strconv"

	"onlinecar-common/errorx"
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
	return l.svcCtx.ProvinceCache.Get(l.ctx)
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
func (l *GetChildrenLogic) GetChildren(parentAdCode string) ([]*types.DistrictNodeResp, error) {
	code, err := strconv.ParseUint(parentAdCode, 10, 32)
	if err != nil {
		l.Errorf("parentAdCode 格式错误, value=%s, err=%v", parentAdCode, err)
		return nil, errorx.NewCodeError(errorx.ErrCodeParam, "adCode格式错误")
	}

	resp, err := l.svcCtx.DistrictRpc.GetChildren(l.ctx, &districtpb.GetChildrenReq{
		ParentAdCode: uint32(code),
	})
	if err != nil {
		l.Errorf("DistrictRpc.GetChildren 失败, parentAdCode=%d, err=%v", uint32(code), err)
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

	return list, nil
}
