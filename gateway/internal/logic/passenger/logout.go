package passenger

import (
	"net/http"
	"strings"

	"onlinecar-common/ctxdata"
	"onlinecar-gateway/internal/middleware"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    interface{}
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx interface{}, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{ctx: ctx, svcCtx: svcCtx}
}

// Logout 乘客登出：将当前 token 加入黑名单使其立即失效
func (l *LogoutLogic) Logout(r *http.Request) error {
	uid, _ := ctxdata.GetUID(r.Context())
	if uid <= 0 {
		return nil
	}

	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}

	conf := l.svcCtx.Config.AuthPassenger
	secret, err := conf.GetAccessSecret("ONLINECAR_PASSENGER_JWT_SECRET")
	if err != nil {
		logx.Errorf("获取 JWT secret 失败: %v", err)
		return nil
	}

	if err := middleware.InvalidateToken(l.svcCtx.Redis, secret, parts[1]); err != nil {
		logx.Errorf("乘客登出失效token失败, uid=%d, err=%v", uid, err)
	}

	logx.Infof("乘客登出成功, uid=%d", uid)
	return nil
}
