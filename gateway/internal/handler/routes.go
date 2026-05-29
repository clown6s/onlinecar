package handler

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	districthandler "onlinecar-gateway/internal/handler/district"
	driverhandler "onlinecar-gateway/internal/handler/driver"
	healthhandler "onlinecar-gateway/internal/handler/health"
	passengerhandler "onlinecar-gateway/internal/handler/passenger"
	smshandler "onlinecar-gateway/internal/handler/sms"
	triphandler "onlinecar-gateway/internal/handler/trip"
	"onlinecar-gateway/internal/middleware"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	// 全局中间件
	server.Use(middleware.NewPanicRecoveryMiddleware().Handle)
	server.Use(middleware.NewRequestIDMiddleware().Handle)
	server.Use(middleware.NewMetricsMiddleware().Handle)
	server.Use(middleware.NewAccessLogMiddleware().Handle)
	server.Use(middleware.NewCORSMiddleware(svcCtx.Config.CORS.AllowOrigins...).Handle)
	server.Use(middleware.NewSecurityHeadersMiddleware("10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle)

	// Prometheus metrics 端点（仅内网访问，无需认证）
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/metrics", Handler: promhttp.Handler().ServeHTTP},
	})

	// 健康检查（无需认证，宽松限流 10 req/s per IP）
	healthRateLimit := middleware.NewRateLimitMiddleware(10).Handle
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/healthz", Handler: healthRateLimit(healthhandler.HealthHandler(svcCtx))},
	})

	// 公开路由：不需要 JWT，带限流（20 req/s per IP）
	// 可信代理网段：仅信任内网代理（Nginx 等）的 X-Real-IP / X-Forwarded-For 头
	publicRateLimit := middleware.NewRateLimitMiddleware(20, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
	// 登录路由单独限流（1 req/s per IP），防止暴力破解验证码
	loginRateLimit := middleware.NewRateLimitMiddleware(1, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
	publicRoutes := []rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/driver/register", Handler: driverhandler.RegisterHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/driver/login", Handler: driverhandler.LoginHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/driver/sms/send", Handler: smshandler.DriverSendSmsHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/passenger/register", Handler: passengerhandler.RegisterHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/passenger/login", Handler: passengerhandler.LoginHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/passenger/sms/send", Handler: smshandler.PassengerSendSmsHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/v1/district/provinces", Handler: districthandler.GetProvincesHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/v1/district/children/:parentAdCode", Handler: districthandler.GetChildrenHandler(svcCtx)},
	}
	// 限制请求体大小 1MB，登录路由额外叠加严格限流
	for i := range publicRoutes {
		handler := publicRoutes[i].Handler
		// 登录接口使用更严格的限流
		if strings.HasSuffix(publicRoutes[i].Path, "/login") {
			handler = loginRateLimit(handler)
		} else {
			handler = publicRateLimit(handler)
		}
		publicRoutes[i].Handler = handler
	}
	server.AddRoutes(publicRoutes, rest.WithMaxBytes(1<<20))

	// 司机认证路由（带限流 100 req/s per IP）
	driverSecret, err := svcCtx.Config.AuthDriver.GetAccessSecret("ONLINECAR_DRIVER_JWT_SECRET")
	if err != nil {
		panic(err)
	}
	driverAuth := middleware.NewJWTAuthMiddleware(driverSecret, svcCtx.Redis, "driver")
	driverRateLimit := middleware.NewRateLimitMiddleware(100, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/driver/location", Handler: driverRateLimit(driverAuth.Handle(driverhandler.UpdateLocationHandler(svcCtx)))},
		{Method: http.MethodPost, Path: "/api/v1/driver/logout", Handler: driverAuth.Handle(driverhandler.LogoutHandler(svcCtx))},
	}, rest.WithMaxBytes(1<<20))

	// 乘客认证路由（带限流 50 req/s per IP）
	passengerSecret, err := svcCtx.Config.AuthPassenger.GetAccessSecret("ONLINECAR_PASSENGER_JWT_SECRET")
	if err != nil {
		panic(err)
	}
	passengerAuth := middleware.NewJWTAuthMiddleware(passengerSecret, svcCtx.Redis, "passenger")
	passengerRateLimit := middleware.NewRateLimitMiddleware(50, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/trip", Handler: passengerRateLimit(passengerAuth.Handle(triphandler.CreateTripHandler(svcCtx)))},
		{Method: http.MethodPost, Path: "/api/v1/passenger/logout", Handler: passengerAuth.Handle(passengerhandler.LogoutHandler(svcCtx))},
	}, rest.WithMaxBytes(1<<20))
}
