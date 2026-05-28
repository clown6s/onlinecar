package passengerlogic

import (
	"context"
	"errors"
	"fmt"

	"onlinecar-common/errorx"
	"onlinecar-passenger/internal/model"
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
	// 输入校验
	if !phoneRegex.MatchString(in.PassengerPhone) {
		return nil, errorx.ErrPhoneFormat
	}

	user, err := l.svcCtx.PassengerUserModel.FindByPhone(l.ctx, in.PassengerPhone)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Errorf("查询乘客失败, phone=%s, err=%v", in.PassengerPhone, err)
		return nil, fmt.Errorf("查询失败，请稍后重试")
	}

	return &pb.CheckPhoneResp{Exists: user != nil}, nil
}
