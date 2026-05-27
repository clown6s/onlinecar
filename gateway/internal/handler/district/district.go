package district

import (
	"net/http"
	"strconv"

	"onlinecar-gateway/common/errorx"
	"onlinecar-gateway/common/response"
	"onlinecar-gateway/internal/svc"
)

func GetProvincesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := NewGetProvincesLogic(r.Context(), svcCtx)
		resp, err := l.GetProvinces()
		response.Response(w, resp, err)
	}
}

func GetChildrenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adCodeStr := r.PathValue("parentAdCode")
		if adCodeStr == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "缺少 parentAdCode 参数")
			return
		}
		parentAdCode, err := strconv.ParseUint(adCodeStr, 10, 32)
		if err != nil {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "parentAdCode 格式错误")
			return
		}
		l := NewGetChildrenLogic(r.Context(), svcCtx)
		resp, err := l.GetChildren(uint32(parentAdCode))
		response.Response(w, resp, err)
	}
}
