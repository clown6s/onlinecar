package handler

import (
	"net/http"

	driverhandler "onlinecar-gateway/internal/handler/driver"
	passengerhandler "onlinecar-gateway/internal/handler/passenger"
	triphandler "onlinecar-gateway/internal/handler/trip"
	"onlinecar-gateway/internal/middleware"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	// 公开路由：不需要 JWT
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/driver/register", Handler: driverhandler.RegisterHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/driver/login", Handler: driverhandler.LoginHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/passenger/register", Handler: passengerhandler.RegisterHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/passenger/login", Handler: passengerhandler.LoginHandler(svcCtx)},
	})

	// 司机认证路由
	driverAuth := middleware.NewJWTAuthMiddleware(svcCtx.Config.AuthDriver.AccessSecret)
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/driver/location", Handler: driverAuth.Handle(driverhandler.UpdateLocationHandler(svcCtx))},
	})

	// 乘客认证路由
	passengerAuth := middleware.NewJWTAuthMiddleware(svcCtx.Config.AuthPassenger.AccessSecret)
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/trip", Handler: passengerAuth.Handle(triphandler.CreateTripHandler(svcCtx))},
	})
}
