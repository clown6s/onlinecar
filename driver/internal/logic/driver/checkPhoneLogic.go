package driverlogic

import (
	"context"

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
		// 优雅降级：DB 异常时不暴露错误，返回 exists=false
		// 用户拿到"手机号未注册"而不是"服务不可用"
		l.Errorf("查询司机失败, phone=%s, err=%v", in.DriverPhone, err)
		return &pb.CheckPhoneResp{Exists: false}, nil
	}

	return &pb.CheckPhoneResp{Exists: user != nil}, nil
}
