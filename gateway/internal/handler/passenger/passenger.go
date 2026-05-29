package passenger

import (
	"net/http"
	"unicode/utf8"

	"onlinecar-common/errorx"
	"onlinecar-common/validate"
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
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if !validate.IsValidPhone(req.Phone) {
			response.Errorx(w, errorx.ErrPhoneFormat)
			return
		}
		nameLen := utf8.RuneCountInString(req.Name)
		if nameLen < 2 || nameLen > 20 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "姓名长度需在2-20个字符之间")
			return
		}
		if req.Gender < 0 || req.Gender > 2 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "性别取值无效（0=未知, 1=男, 2=女）")
			return
		}
		if req.Code == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "验证码不能为空")
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
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if !validate.IsValidPhone(req.Phone) {
			response.Errorx(w, errorx.ErrPhoneFormat)
			return
		}
		if req.Code == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "验证码不能为空")
			return
		}
		l := passengerlogic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		response.Response(w, resp, err)
	}
}

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := passengerlogic.NewLogoutLogic(r.Context(), svcCtx)
		err := l.Logout(r)
		response.Response(w, nil, err)
	}
}
