package districtservicelogic

import (
	"context"

	"district/internal/svc"
	"district/pb"

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

func (l *GetChildrenLogic) GetChildren(in *pb.GetChildrenReq) (*pb.GetChildrenResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetChildrenResp{}, nil
}
