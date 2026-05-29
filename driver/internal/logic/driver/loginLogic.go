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

// 登录
func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	l.Infof("司机登录, phone=%s", util.MaskPhone(in.DriverPhone))

	// 手机号格式校验
	if !validate.IsValidPhone(in.DriverPhone) {
		return nil, errorx.ErrPhoneFormat
	}

	driverUser, err := l.svcCtx.DriverUserModel.FindByPhone(l.ctx, in.DriverPhone)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("司机不存在, phone=%s", util.MaskPhone(in.DriverPhone))
		} else {
			l.Errorf("查询司机失败, phone=%s, err=%v", util.MaskPhone(in.DriverPhone), err)
		}
		// 统一返回模糊信息，避免暴露注册状态
		return nil, errorx.ErrLoginFailed
	}

	// 验证码在 Gateway 层已校验，此处仅查询用户

	var state int64
	if driverUser.State != nil {
		state = *driverUser.State
	}

	// 检查用户是否被禁用
	if state == 1 {
		l.Infof("司机已禁用, driverId=%d", driverUser.Id)
		return nil, errorx.ErrUserDisabled
	}

	l.Infof("司机登录成功, driverId=%d", driverUser.Id)
	return &pb.LoginResp{
		Id:         driverUser.Id,
		DriverName: driverUser.DriverName,
		State:      state,
	}, nil
}
