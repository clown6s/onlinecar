package passengerlogic

import (
	"context"

	"onlinecar-passenger/internal/svc"
	"onlinecar-passenger/pb"

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
	user, err := l.svcCtx.PassengerUserModel.FindByPhone(l.ctx, in.PassengerPhone)
	if err != nil {
		// 优雅降级：DB 异常时不暴露错误，返回 exists=false
		l.Errorf("查询乘客失败, phone=%s, err=%v", in.PassengerPhone, err)
		return &pb.CheckPhoneResp{Exists: false}, nil
	}

	return &pb.CheckPhoneResp{Exists: user != nil}, nil
}
