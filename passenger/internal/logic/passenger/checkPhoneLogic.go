package passengerlogic

import (
	"context"
	"errors"

	"onlinecar-common/errorx"
	"onlinecar-common/util"
	"onlinecar-common/validate"
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
	if !validate.IsValidPhone(in.PassengerPhone) {
		return nil, errorx.ErrPhoneFormat
	}

	// 统一返回 exists=false，防止攻击者批量枚举已注册手机号
	// 前端流程：用户输入手机号 → 发送验证码 → 登录/注册（根据是否有返回的 user_id 判断）
	_, err := l.svcCtx.PassengerUserModel.FindByPhone(l.ctx, in.PassengerPhone)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Errorf("查询乘客失败, phone=%s, err=%v", util.MaskPhone(in.PassengerPhone), err)
		return nil, errorx.ErrServer
	}

	return &pb.CheckPhoneResp{Exists: false}, nil
}
