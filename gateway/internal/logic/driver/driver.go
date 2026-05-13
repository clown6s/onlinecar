package driver

import (
	"context"
	"time"

	"onlinecar-gateway/common/ctxdata"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	driverpb "onlinecar-gateway/pb/driver"

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

func (l *RegisterLogic) Register(req *types.DriverRegisterReq) (*types.DriverResp, error) {
	l.Infof("Gateway->DriverRpc.Register, phone=%s, name=%s", req.Phone, req.Name)

	resp, err := l.svcCtx.DriverRpc.Register(l.ctx, &driverpb.RegisterReq{
		DriverName:     req.Name,
		DriverPhone:    req.Phone,
		VehicleNo:      req.VehicleNo,
		VehicleType:    req.VehicleType,
		CommercialType: req.CommercialType,
	})
	if err != nil {
		l.Errorf("DriverRpc.Register 失败, phone=%s, err=%v", req.Phone, err)
		return nil, err
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

func (l *LoginLogic) Login(req *types.DriverLoginReq) (map[string]any, error) {
	l.Infof("Gateway->DriverRpc.Login, phone=%s", req.Phone)

	resp, err := l.svcCtx.DriverRpc.Login(l.ctx, &driverpb.LoginReq{
		DriverPhone: req.Phone,
	})
	if err != nil {
		l.Errorf("DriverRpc.Login 失败, phone=%s, err=%v", req.Phone, err)
		return nil, err
	}

	conf := l.svcCtx.Config.AuthDriver
	claims := jwt.MapClaims{
		"uid":     resp.Id,
		"subject": "driver",
		"exp":     time.Now().Add(conf.AccessExpire).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(conf.AccessSecret))
	if err != nil {
		l.Errorf("JWT签名失败, driverId=%d, err=%v", resp.Id, err)
		return nil, err
	}

	l.Infof("司机登录成功, driverId=%d, name=%s", resp.Id, resp.DriverName)
	return map[string]any{
		"token":      tokenStr,
		"expire":     conf.AccessExpire.Seconds(),
		"driverId":   resp.Id,
		"driverName": resp.DriverName,
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
		return err
	}
	return nil
}
