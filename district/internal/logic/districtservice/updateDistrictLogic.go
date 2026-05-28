package districtservicelogic

import (
	"context"

	"onlinecar-district/internal/svc"
	"onlinecar-district/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDistrictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDistrictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDistrictLogic {
	return &UpdateDistrictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateDistrict 更新行政区划数据
// TODO: 当前为空实现预留接口，行政区划数据通过 SQL 直接导入，暂无在线更新需求
func (l *UpdateDistrictLogic) UpdateDistrict(in *pb.UpdateDistrictReq) (*pb.UpdateDistrictResp, error) {
	l.Info("更新行政区划数据")

	// UpdateDistrictReq 当前为空消息，预留接口
	return &pb.UpdateDistrictResp{}, nil
}
