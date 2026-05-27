package trip

import (
	"net/http"

	"onlinecar-gateway/common/errorx"
	"onlinecar-gateway/common/response"
	"onlinecar-gateway/internal/logic/trip"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateTripHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateTripReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		l := trip.NewCreateTripLogic(r.Context(), svcCtx)
		resp, err := l.CreateTrip(&req)
		response.Response(w, resp, err)
	}
}
