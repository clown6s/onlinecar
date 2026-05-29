package driver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"onlinecar-common/ctxdata"
	"onlinecar-common/errorx"
	"onlinecar-common/util"
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
	l.Infof("Gateway->DriverRpc.Register, phone=%s, name=%s", util.MaskPhone(req.Phone), util.MaskName(req.Name))

	// 注册前校验验证码，防止为任意手机号注册账号
	if req.Code == "" {
		return nil, errorx.ErrParam
	}
	if err := sms.VerifyCode(l.ctx, l.svcCtx, "driver", req.Phone, req.Code); err != nil {
		return nil, err
	}

	// 直接调用 Register RPC，由 RPC 服务端数据库 UNIQUE 约束保证手机号唯一性
	resp, err := l.svcCtx.DriverRpc.Register(l.ctx, &driverpb.RegisterReq{
		DriverName:     req.Name,
		DriverPhone:    req.Phone,
		VehicleNo:      req.VehicleNo,
		VehicleType:    req.VehicleType,
		CommercialType: req.CommercialType,
	})
	if err != nil {
		l.Errorf("DriverRpc.Register 失败, phone=%s, err=%v", util.MaskPhone(req.Phone), err)
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
	l.Infof("Gateway->DriverRpc.Login, phone=%s", util.MaskPhone(req.Phone))

	// 校验验证码
	if err := sms.VerifyCode(l.ctx, l.svcCtx, "driver", req.Phone, req.Code); err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.DriverRpc.Login(l.ctx, &driverpb.LoginReq{
		DriverPhone: req.Phone,
	})
	if err != nil {
		l.Errorf("DriverRpc.Login 失败, phone=%s, err=%v", util.MaskPhone(req.Phone), err)
		return nil, errorx.MapRpcError(err)
	}
	if resp == nil {
		return nil, errorx.ErrServer
	}

	conf := l.svcCtx.Config.AuthDriver
	var secret string
	secret, err = conf.GetAccessSecret("ONLINECAR_DRIVER_JWT_SECRET")
	if err != nil {
		l.Errorf("获取 JWT secret 失败: %v", err)
		return nil, errorx.ErrServer
	}
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

	// 设备数限制：最多 3 个同时在线设备，超出时踢出最早的 token
	l.manageDeviceSessions(secret, resp.Id, tokenStr, conf.AccessExpire)

	l.Infof("司机登录成功, driverId=%d, name=%s", resp.Id, resp.DriverName)
	return &LoginResp{
		Token:     tokenStr,
		ExpiresIn: max(int64(conf.AccessExpire.Seconds()), 1),
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
	uid, _ := ctxdata.GetUID(l.ctx)
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

const maxDevicesPerUser = 3

// manageDeviceSessions 管理用户多设备会话，最多保留 maxDevicesPerUser 个 token
// 使用 Redis LIST 维护 token 列表，超出时将最早的 token 加入黑名单
func (l *LoginLogic) manageDeviceSessions(secret string, uid int64, tokenStr string, expire time.Duration) {
	if l.svcCtx.Redis == nil {
		return
	}

	deviceKey := fmt.Sprintf("user:tokens:driver:%d", uid)

	// 将新 token 加入列表头部
	l.svcCtx.Redis.Lpush(deviceKey, tokenStr)
	l.svcCtx.Redis.Expire(deviceKey, int(expire.Seconds()))

	// 裁剪到最大设备数
	length, err := l.svcCtx.Redis.Llen(deviceKey)
	if err != nil || length <= maxDevicesPerUser {
		return
	}

	// 获取超出部分（列表尾部）并加入黑名单
	excess := length - maxDevicesPerUser
	oldTokens, err := l.svcCtx.Redis.Lrange(deviceKey, -excess, -1)
	if err != nil || len(oldTokens) == 0 {
		return
	}

	for _, oldToken := range oldTokens {
		if oldToken == "" || oldToken == tokenStr {
			continue
		}
		// 加入黑名单
		oldHash := fmt.Sprintf("%x", sha256.Sum256([]byte(oldToken)))
		blacklistKey := "token:blacklist:" + oldHash
		l.svcCtx.Redis.SetexCtx(context.Background(), blacklistKey, "1", int(expire.Seconds()))
		logx.Infof("踢出旧设备token, uid=%d, hash=%s", uid, oldHash)
	}

	// 裁剪列表
	l.svcCtx.Redis.Ltrim(deviceKey, 0, maxDevicesPerUser-1)
}
