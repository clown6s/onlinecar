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

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录（手机号直接登录）
func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	l.Infof("乘客登录, phone=%s", util.MaskPhone(in.PassengerPhone))

	// 输入校验
	if !validate.IsValidPhone(in.PassengerPhone) {
		return nil, errorx.ErrPhoneFormat
	}

	passengerUser, err := l.svcCtx.PassengerUserModel.FindByPhone(l.ctx, in.PassengerPhone)
	if err != nil {
		var codeErr *errorx.CodeError
		if errors.As(err, &codeErr) {
			return nil, codeErr
		}
		if errors.Is(err, model.ErrNotFound) {
			// 统一返回模糊信息，避免暴露注册状态（与 driver 对齐）
			return nil, errorx.ErrLoginFailed
		}
		l.Errorf("查询乘客失败, phone=%s, err=%v", util.MaskPhone(in.PassengerPhone), err)
		return nil, errorx.ErrServer
	}

	// 先检查用户是否被禁用，再填充响应数据（防止禁用用户的 PII 泄露）
	if passengerUser.State != nil && *passengerUser.State == 1 {
		l.Infof("乘客已禁用, passengerId=%d", passengerUser.Id)
		return nil, errorx.ErrUserDisabled
	}

	resp := &pb.LoginResp{
		Id: passengerUser.Id,
	}

	if passengerUser.PassengerName != nil {
		resp.PassengerName = *passengerUser.PassengerName
	}
	if passengerUser.State != nil {
		resp.State = int64(*passengerUser.State)
	}
	if passengerUser.ProfilePhoto != nil {
		resp.ProfilePhoto = *passengerUser.ProfilePhoto
	}

	l.Infof("乘客登录成功, passengerId=%d", passengerUser.Id)
	return resp, nil
}
