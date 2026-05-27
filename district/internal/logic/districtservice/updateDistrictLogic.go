package districtservicelogic

import (
	"context"

	"district/internal/svc"
	"district/pb"

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

func (l *UpdateDistrictLogic) UpdateDistrict(in *pb.UpdateDistrictReq) (*pb.UpdateDistrictResp, error) {
	// todo: add your logic here and delete this line

	return &pb.UpdateDistrictResp{}, nil
}
