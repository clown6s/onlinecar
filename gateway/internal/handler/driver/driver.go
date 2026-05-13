package driver

import (
	"net/http"

	"onlinecar-gateway/common/response"
	driverlogic "onlinecar-gateway/internal/logic/driver"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DriverRegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		l := driverlogic.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		response.Response(w, resp, err)
	}
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DriverLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		l := driverlogic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		response.Response(w, resp, err)
	}
}

func UpdateLocationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DriverLocationReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		l := driverlogic.NewUpdateLocationLogic(r.Context(), svcCtx)
		err := l.UpdateLocation(&req)
		response.Response(w, nil, err)
	}
}
