package driverlogic

import (
	"context"
	"fmt"

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
	l.Infof("司机登录, phone=%s", in.DriverPhone)

	driverUser, err := l.svcCtx.DriverUserModel.FindByPhone(l.ctx, in.DriverPhone)
	if err != nil {
		l.Errorf("查询司机失败, phone=%s, err=%v", in.DriverPhone, err)
		return nil, fmt.Errorf("查询失败: %v", err)
	}
	if driverUser == nil {
		l.Errorf("司机不存在, phone=%s", in.DriverPhone)
		return nil, fmt.Errorf("该手机号未注册")
	}

	// TODO: 密码验证（当前数据库无 password 字段，后续加列后接入）

	var state int64
	if driverUser.State != nil {
		state = *driverUser.State
	}

	l.Infof("司机登录成功, driverId=%d, name=%s", driverUser.Id, driverUser.DriverName)
	return &pb.LoginResp{
		Id:         driverUser.Id,
		DriverName: driverUser.DriverName,
		State:      state,
	}, nil
}
