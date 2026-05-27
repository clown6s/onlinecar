package driver

import (
	"context"
	"fmt"
	"time"

	"onlinecar-common/ctxdata"
	"onlinecar-common/errorx"
	"onlinecar-gateway/internal/logic/sms"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	driverpb "onlinecar-gateway/pb/driver"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
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

func (l *RegisterLogic) Register(req *types.DriverRegisterReq) (*types.DriverResp, error) {
	l.Infof("Gateway->DriverRpc.Register, phone=%s, name=%s", req.Phone, req.Name)

	// 直接调用 Register RPC，由 RPC 服务端数据库 UNIQUE 约束保证手机号唯一性
	resp, err := l.svcCtx.DriverRpc.Register(l.ctx, &driverpb.RegisterReq{
		DriverName:     req.Name,
		DriverPhone:    req.Phone,
		VehicleNo:      req.VehicleNo,
		VehicleType:    req.VehicleType,
		CommercialType: req.CommercialType,
	})
	if err != nil {
		l.Errorf("DriverRpc.Register 失败, phone=%s, err=%v", req.Phone, err)
		return nil, errorx.MapRpcError(err)
	}

	l.Infof("司机注册成功, driverId=%d", resp.Id)
	return &types.DriverResp{
		Id:    resp.Id,
		Name:  resp.DriverName,
		Phone: resp.DriverPhone,
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

type LoginResp struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

func (l *LoginLogic) Login(req *types.DriverLoginReq) (*LoginResp, error) {
	l.Infof("Gateway->DriverRpc.Login, phone=%s", req.Phone)

	// 校验验证码
	if err := sms.VerifyCode(l.ctx, l.svcCtx, "driver", req.Phone, req.Code); err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.DriverRpc.Login(l.ctx, &driverpb.LoginReq{
		DriverPhone: req.Phone,
	})
	if err != nil {
		l.Errorf("DriverRpc.Login 失败, phone=%s, err=%v", req.Phone, err)
		return nil, errorx.MapRpcError(err)
	}
	if resp == nil {
		return nil, errorx.ErrServer
	}

	conf := l.svcCtx.Config.AuthDriver
	secret := conf.GetAccessSecret("ONLINECAR_DRIVER_JWT_SECRET")
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("driver:%d", resp.Id),
		"jti":  uuid.New().String(),
		"uid":  resp.Id,
		"role": "driver",
		"exp":  now.Add(conf.AccessExpire).Unix(),
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		l.Errorf("JWT签名失败, driverId=%d, err=%v", resp.Id, err)
		return nil, err
	}

	l.Infof("司机登录成功, driverId=%d, name=%s", resp.Id, resp.DriverName)
	return &LoginResp{
		Token:     tokenStr,
		ExpiresIn: int64(conf.AccessExpire.Seconds()),
	}, nil
}

type UpdateLocationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLocationLogic {
	return &UpdateLocationLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateLocationLogic) UpdateLocation(req *types.DriverLocationReq) error {
	uid := ctxdata.GetUID(l.ctx)
	l.Infof("Gateway->DriverRpc.UpdateLocation, driverId=%d, lng=%.6f, lat=%.6f", uid, req.Lng, req.Lat)

	_, err := l.svcCtx.DriverRpc.UpdateLocation(l.ctx, &driverpb.UpdateLocationReq{
		Id:  uid,
		Lng: req.Lng,
		Lat: req.Lat,
	})
	if err != nil {
		l.Errorf("DriverRpc.UpdateLocation 失败, driverId=%d, err=%v", uid, err)
		return errorx.MapRpcError(err)
	}
	return nil
}
