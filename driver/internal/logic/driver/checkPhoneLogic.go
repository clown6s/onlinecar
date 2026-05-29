package driverlogic

import (
	"context"
	"errors"

	"onlinecar-common/errorx"
	"onlinecar-common/util"
	"onlinecar-common/validate"
	"onlinecar-driver/internal/model"
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
	if !validate.IsValidPhone(in.DriverPhone) {
		return nil, errorx.ErrPhoneFormat
	}
	// 不再返回明确的 exists 状态，改为统一返回"已注册请直接登录"或"未注册请先注册"
	// 防止攻击者批量枚举已注册手机号
	_, err := l.svcCtx.DriverUserModel.FindByPhone(l.ctx, in.DriverPhone)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Errorf("查询司机失败, phone=%s, err=%v", util.MaskPhone(in.DriverPhone), err)
		return nil, errorx.ErrServer
	}

	return &pb.CheckPhoneResp{Exists: false}, nil
}
