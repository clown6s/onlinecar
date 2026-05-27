package driverlogic

import (
	"context"
	"fmt"

	"onlinecar-driver/internal/svc"
	"onlinecar-driver/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPhoneLogic {
	return &CheckPhoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckPhoneLogic) CheckPhone(in *pb.CheckPhoneReq) (*pb.CheckPhoneResp, error) {
	user, err := l.svcCtx.DriverUserModel.FindByPhone(l.ctx, in.DriverPhone)
	if err != nil {
		l.Errorf("查询司机失败, phone=%s, err=%v", in.DriverPhone, err)
		return nil, fmt.Errorf("查询失败，请稍后重试")
	}

	return &pb.CheckPhoneResp{Exists: user != nil}, nil
}
