package passengerlogic

import (
	"context"
	"errors"

	"onlinecar-common/errorx"
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
	l.Infof("乘客登录, phone=%s", in.PassengerPhone)

	// 输入校验
	if !validate.IsValidPhone(in.PassengerPhone) {
		return nil, errorx.ErrPhoneFormat
	}

	passengerUser, err := l.svcCtx.PassengerUserModel.FindByPhone(l.ctx, in.PassengerPhone)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 统一返回模糊信息，避免暴露注册状态（与 driver 对齐）
			return nil, errorx.NewCodeError(errorx.ErrAuth.Code, "手机号或验证码错误")
		}
		if codeErr, ok := err.(*errorx.CodeError); ok {
			return nil, codeErr
		}
		l.Errorf("查询乘客失败, phone=%s, err=%v", in.PassengerPhone, err)
		return nil, errorx.ErrServer
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

	l.Infof("乘客登录成功, passengerId=%d, name=%s", passengerUser.Id, resp.PassengerName)
	return resp, nil
}
