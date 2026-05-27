package driver

import (
	"math"
	"net/http"
	"regexp"

	"onlinecar-common/errorx"
	"onlinecar-gateway/common/response"
	"onlinecar-gateway/common/validate"
	driverlogic "onlinecar-gateway/internal/logic/driver"
	"onlinecar-gateway/internal/svc"
	"onlinecar-gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

var vehicleNoRegex = regexp.MustCompile(`^[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤川青藏琼宁][A-Z][A-Z0-9]{5,6}$`)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DriverRegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if !validate.IsValidPhone(req.Phone) {
			response.Errorx(w, errorx.ErrPhoneFormat)
			return
		}
		if len(req.Name) < 2 || len(req.Name) > 20 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "姓名长度须为2-20个字符")
			return
		}
		if !vehicleNoRegex.MatchString(req.VehicleNo) {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "车牌号格式不正确")
			return
		}
		if req.VehicleType == "" {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "车辆类型不能为空")
			return
		}
		if req.CommercialType < 1 || req.CommercialType > 3 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "营运类型须为1(网约车)、2(出租车)或3(其他)")
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
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if !validate.IsValidPhone(req.Phone) {
			response.Errorx(w, errorx.ErrPhoneFormat)
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
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, err.Error())
			return
		}
		if math.IsNaN(req.Lng) || math.IsInf(req.Lng, 0) || req.Lng < -180 || req.Lng > 180 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "经度无效，须在[-180,180]范围内")
			return
		}
		if math.IsNaN(req.Lat) || math.IsInf(req.Lat, 0) || req.Lat < -90 || req.Lat > 90 {
			response.Error(w, http.StatusBadRequest, errorx.ErrCodeParam, "纬度无效，须在[-90,90]范围内")
			return
		}
		l := driverlogic.NewUpdateLocationLogic(r.Context(), svcCtx)
		err := l.UpdateLocation(&req)
		response.Response(w, nil, err)
	}
}
