package trip

import (
	"math"
	"net/http"

	"onlinecar-common/errorx"
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
		if req.Origin == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "出发地不能为空")
			return
		}
		if req.Destination == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "目的地不能为空")
			return
		}
		if math.IsNaN(req.OriginLng) || math.IsInf(req.OriginLng, 0) || req.OriginLng < -180 || req.OriginLng > 180 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "出发地经度无效")
			return
		}
		if math.IsNaN(req.OriginLat) || math.IsInf(req.OriginLat, 0) || req.OriginLat < -90 || req.OriginLat > 90 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "出发地纬度无效")
			return
		}
		if math.IsNaN(req.DestLng) || math.IsInf(req.DestLng, 0) || req.DestLng < -180 || req.DestLng > 180 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "目的地经度无效")
			return
		}
		if math.IsNaN(req.DestLat) || math.IsInf(req.DestLat, 0) || req.DestLat < -90 || req.DestLat > 90 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "目的地纬度无效")
			return
		}
		l := trip.NewCreateTripLogic(r.Context(), svcCtx)
		resp, err := l.CreateTrip(&req)
		response.Response(w, resp, err)
	}
}
