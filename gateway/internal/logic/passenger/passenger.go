package passenger

import (
	"context"
	"time"

	"onlinecar-gateway/common/errorx"
	"onlinecar-gateway/internal/logic/sms"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	passengerpb "onlinecar-gateway/pb/passenger"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterLogic) Register(req *types.PassengerRegisterReq) (*types.PassengerResp, error) {
	l.Infof("Gateway->PassengerRpc.Register, phone=%s, name=%s", req.Phone, req.Name)

	// 先检查手机号是否已注册
	checkResp, err := l.svcCtx.PassengerRpc.CheckPhone(l.ctx, &passengerpb.CheckPhoneReq{PassengerPhone: req.Phone})
	if err != nil {
		l.Errorf("PassengerRpc.CheckPhone RPC异常, phone=%s, err=%v", req.Phone, err)
		return nil, errorx.ErrServiceUnavailable
	}
	if checkResp.Exists {
		return nil, errorx.ErrPhoneRegistered
	}

	resp, err := l.svcCtx.PassengerRpc.Register(l.ctx, &passengerpb.RegisterReq{
		PassengerName:   req.Name,
		PassengerPhone:  req.Phone,
		PassengerGender: req.Gender,
	})
	if err != nil {
		l.Errorf("PassengerRpc.Register 失败, phone=%s, err=%v", req.Phone, err)
		return nil, errorx.MapRpcError(err)
	}

	l.Infof("乘客注册成功, passengerId=%d", resp.Id)
	return &types.PassengerResp{
		Id:    resp.Id,
		Name:  resp.PassengerName,
		Phone: resp.PassengerPhone,
		State: resp.State,
	}, nil
}

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *LoginLogic) Login(req *types.PassengerLoginReq) (map[string]any, error) {
	l.Infof("Gateway->PassengerRpc.Login, phone=%s", req.Phone)

	// 校验验证码
	if err := sms.VerifyCode(l.ctx, l.svcCtx, "passenger", req.Phone, req.Code); err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.PassengerRpc.Login(l.ctx, &passengerpb.LoginReq{
		PassengerPhone: req.Phone,
	})
	if err != nil {
		l.Errorf("PassengerRpc.Login 失败, phone=%s, err=%v", req.Phone, err)
		return nil, errorx.MapRpcError(err)
	}

	conf := l.svcCtx.Config.AuthPassenger
	secret := conf.GetAccessSecret("ONLINECAR_PASSENGER_JWT_SECRET")
	claims := jwt.MapClaims{
		"uid":     resp.Id,
		"subject": "passenger",
		"exp":     time.Now().Add(conf.AccessExpire).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		l.Errorf("JWT签名失败, passengerId=%d, err=%v", resp.Id, err)
		return nil, err
	}

	l.Infof("乘客登录成功, passengerId=%d, name=%s", resp.Id, resp.PassengerName)
	return map[string]any{
		"token":         tokenStr,
		"expire":        conf.AccessExpire.Seconds(),
		"passengerId":   resp.Id,
		"passengerName": resp.PassengerName,
	}, nil
}
