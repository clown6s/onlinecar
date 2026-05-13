package passenger

import (
	"net/http"

	"onlinecar-gateway/common/response"
	passengerlogic "onlinecar-gateway/internal/logic/passenger"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PassengerRegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		l := passengerlogic.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		response.Response(w, resp, err)
	}
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PassengerLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, 400, err.Error())
			return
		}
		l := passengerlogic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		response.Response(w, resp, err)
	}
}
