package sms

import (
	"net/http"

	"onlinecar-common/errorx"
	"onlinecar-gateway/common/response"
	"onlinecar-gateway/common/validate"
	smslogic "onlinecar-gateway/internal/logic/sms"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DriverSendSmsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSmsReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if !validate.IsValidPhone(req.Phone) {
			response.Errorx(w, errorx.ErrPhoneFormat)
			return
		}
		l := smslogic.NewSendSmsLogic(r.Context(), svcCtx, "driver")
		err := l.SendSms(req.Phone)
		response.Response(w, nil, err)
	}
}

func PassengerSendSmsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSmsReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if !validate.IsValidPhone(req.Phone) {
			response.Errorx(w, errorx.ErrPhoneFormat)
			return
		}
		l := smslogic.NewSendSmsLogic(r.Context(), svcCtx, "passenger")
		err := l.SendSms(req.Phone)
		response.Response(w, nil, err)
	}
}
