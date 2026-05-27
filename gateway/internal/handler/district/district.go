package district

import (
	"net/http"

	"onlinecar-common/errorx"
	"onlinecar-gateway/common/response"
	districtlogic "onlinecar-gateway/internal/logic/district"
	"onlinecar-gateway/internal/svc"

	"github.com/zeromicro/go-zero/rest/pathvar"
)

func GetProvincesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := districtlogic.NewGetProvincesLogic(r.Context(), svcCtx)
		resp, err := l.GetProvinces()
		response.Response(w, resp, err)
	}
}

func GetChildrenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		parentAdCode := vars["parentAdCode"]
		if parentAdCode == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "缺少 parentAdCode 参数")
			return
		}
		l := districtlogic.NewGetChildrenLogic(r.Context(), svcCtx)
		resp, err := l.GetChildren(parentAdCode)
		response.Response(w, resp, err)
	}
}
