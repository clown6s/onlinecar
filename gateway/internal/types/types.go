package types

// ========== 司机 ==========

type DriverRegisterReq struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Code           string `json:"code"`
	VehicleNo      string `json:"vehicleNo"`
	VehicleType    string `json:"vehicleType"`
	CommercialType int64  `json:"commercialType,omitempty"`
}

type DriverLoginReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type DriverLocationReq struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type DriverResp struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	VehicleNo      string `json:"vehicleNo"`
	VehicleType    string `json:"vehicleType"`
	Brand          string `json:"brand"`
	Model          string `json:"model"`
	CommercialType int64  `json:"commercialType"`
	State          int64  `json:"state"`
	WorkStatus     int64  `json:"workStatus"`
}

// ========== 乘客 ==========

type PassengerRegisterReq struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Code   string `json:"code"`
	Gender int64  `json:"gender,omitempty"`
}

type PassengerLoginReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type PassengerResp struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Gender       int64  `json:"gender"`
	State        int64  `json:"state"`
	ProfilePhoto string `json:"profilePhoto"`
}

// ========== 短信验证码 ==========

type SendSmsReq struct {
	Phone string `json:"phone"`
}

// ========== 行程 ==========

type CreateTripReq struct {
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	OriginLng   float64 `json:"originLng"`
	OriginLat   float64 `json:"originLat"`
	DestLng     float64 `json:"destLng"`
	DestLat     float64 `json:"destLat"`
}

type TripResp struct {
	Id          int64   `json:"id"`
	PassengerId int64   `json:"passengerId"`
	DriverId    int64   `json:"driverId"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	Status      int64   `json:"status"`
	Fare        float64 `json:"fare"`
}

// ========== 行政区划 ==========

type DistrictNodeResp struct {
	Id     int64  `json:"id"`
	AdCode uint32 `json:"adCode"`
	Name   string `json:"name"`
	Level  string `json:"level"`
	Center string `json:"center"`
}
