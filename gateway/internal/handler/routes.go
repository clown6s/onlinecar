package handler

import (
	"net/http"

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
	server.Use(middleware.NewAccessLogMiddleware().Handle)
	server.Use(middleware.NewCORSMiddleware(svcCtx.Config.CORS.AllowOrigins...).Handle)
	server.Use(middleware.NewSecurityHeadersMiddleware("10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle)

	// 健康检查（无需认证，宽松限流 10 req/s per IP）
	healthRateLimit := middleware.NewRateLimitMiddleware(10).Handle
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/healthz", Handler: healthRateLimit(healthhandler.HealthHandler(svcCtx))},
	})

	// 公开路由：不需要 JWT，带限流（20 req/s per IP）
	// 可信代理网段：仅信任内网代理（Nginx 等）的 X-Real-IP / X-Forwarded-For 头
	publicRateLimit := middleware.NewRateLimitMiddleware(20, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
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
	// 限制请求体大小 1MB
	for i := range publicRoutes {
		publicRoutes[i].Handler = publicRateLimit(publicRoutes[i].Handler)
	}
	server.AddRoutes(publicRoutes, rest.WithMaxBytes(1<<20))

	// 司机认证路由（带限流 100 req/s per IP）
	driverAuth := middleware.NewJWTAuthMiddleware(svcCtx.Config.AuthDriver.GetAccessSecret("ONLINECAR_DRIVER_JWT_SECRET"))
	driverRateLimit := middleware.NewRateLimitMiddleware(100, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/driver/location", Handler: driverRateLimit(driverAuth.Handle(driverhandler.UpdateLocationHandler(svcCtx)))},
	}, rest.WithMaxBytes(1<<20))

	// 乘客认证路由（带限流 50 req/s per IP）
	passengerAuth := middleware.NewJWTAuthMiddleware(svcCtx.Config.AuthPassenger.GetAccessSecret("ONLINECAR_PASSENGER_JWT_SECRET"))
	passengerRateLimit := middleware.NewRateLimitMiddleware(50, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16").Handle
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/trip", Handler: passengerRateLimit(passengerAuth.Handle(triphandler.CreateTripHandler(svcCtx)))},
	}, rest.WithMaxBytes(1<<20))
}
