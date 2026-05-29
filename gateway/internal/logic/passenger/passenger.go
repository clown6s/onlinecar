package passenger

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"onlinecar-common/errorx"
	"onlinecar-common/util"
	"onlinecar-gateway/internal/logic/sms"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"
	passengerpb "onlinecar-gateway/pb/passenger"

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

func (l *RegisterLogic) Register(req *types.PassengerRegisterReq) (*types.PassengerResp, error) {
	l.Infof("Gateway->PassengerRpc.Register, phone=%s, name=%s", util.MaskPhone(req.Phone), util.MaskName(req.Name))

	// 注册前校验验证码，防止为任意手机号注册账号
	if req.Code == "" {
		return nil, errorx.ErrParam
	}
	if err := sms.VerifyCode(l.ctx, l.svcCtx, "passenger", req.Phone, req.Code); err != nil {
		return nil, err
	}

	// 直接调用 Register RPC，由 RPC 服务端数据库 UNIQUE 约束保证手机号唯一性
	resp, err := l.svcCtx.PassengerRpc.Register(l.ctx, &passengerpb.RegisterReq{
		PassengerName:   req.Name,
		PassengerPhone:  req.Phone,
		PassengerGender: req.Gender,
	})
	if err != nil {
		l.Errorf("PassengerRpc.Register 失败, phone=%s, err=%v", util.MaskPhone(req.Phone), err)
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

type LoginResp struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

func (l *LoginLogic) Login(req *types.PassengerLoginReq) (*LoginResp, error) {
	l.Infof("Gateway->PassengerRpc.Login, phone=%s", util.MaskPhone(req.Phone))

	// 校验验证码
	if err := sms.VerifyCode(l.ctx, l.svcCtx, "passenger", req.Phone, req.Code); err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.PassengerRpc.Login(l.ctx, &passengerpb.LoginReq{
		PassengerPhone: req.Phone,
	})
	if err != nil {
		l.Errorf("PassengerRpc.Login 失败, phone=%s, err=%v", util.MaskPhone(req.Phone), err)
		return nil, errorx.MapRpcError(err)
	}
	if resp == nil {
		return nil, errorx.ErrServer
	}

	conf := l.svcCtx.Config.AuthPassenger
	var secret string
	secret, err = conf.GetAccessSecret("ONLINECAR_PASSENGER_JWT_SECRET")
	if err != nil {
		l.Errorf("获取 JWT secret 失败: %v", err)
		return nil, errorx.ErrServer
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("passenger:%d", resp.Id),
		"jti":  uuid.New().String(),
		"uid":  resp.Id,
		"role": "passenger",
		"exp":  now.Add(conf.AccessExpire).Unix(),
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		l.Errorf("JWT签名失败, passengerId=%d, err=%v", resp.Id, err)
		return nil, err
	}

	// 设备数限制：最多 3 个同时在线设备，超出时踢出最早的 token
	l.manageDeviceSessions(secret, resp.Id, tokenStr, conf.AccessExpire)

	l.Infof("乘客登录成功, passengerId=%d, name=%s", resp.Id, resp.PassengerName)
	return &LoginResp{
		Token:     tokenStr,
		ExpiresIn: max(int64(conf.AccessExpire.Seconds()), 1),
	}, nil
}

const maxDevicesPerUser = 3

// manageDeviceSessions 管理用户多设备会话，最多保留 maxDevicesPerUser 个 token
func (l *LoginLogic) manageDeviceSessions(secret string, uid int64, tokenStr string, expire time.Duration) {
	if l.svcCtx.Redis == nil {
		return
	}

	deviceKey := fmt.Sprintf("user:tokens:passenger:%d", uid)

	l.svcCtx.Redis.Lpush(deviceKey, tokenStr)
	l.svcCtx.Redis.Expire(deviceKey, int(expire.Seconds()))

	length, err := l.svcCtx.Redis.Llen(deviceKey)
	if err != nil || length <= maxDevicesPerUser {
		return
	}

	excess := length - maxDevicesPerUser
	oldTokens, err := l.svcCtx.Redis.Lrange(deviceKey, -excess, -1)
	if err != nil || len(oldTokens) == 0 {
		return
	}

	for _, oldToken := range oldTokens {
		if oldToken == "" || oldToken == tokenStr {
			continue
		}
		oldHash := fmt.Sprintf("%x", sha256.Sum256([]byte(oldToken)))
		blacklistKey := "token:blacklist:" + oldHash
		l.svcCtx.Redis.SetexCtx(context.Background(), blacklistKey, "1", int(expire.Seconds()))
		logx.Infof("踢出旧设备token, uid=%d, hash=%s", uid, oldHash)
	}

	l.svcCtx.Redis.Ltrim(deviceKey, 0, maxDevicesPerUser-1)
}
