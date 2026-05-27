package districtservicelogic

import (
	"context"

	"district/internal/svc"
	"district/pb"

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

func (l *GetProvincesLogic) GetProvinces(in *pb.GetProvincesReq) (*pb.GetProvincesResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetProvincesResp{}, nil
}
