package skootar

type GetDriverLocationInp struct {
	ReqBase
	SKootarId string `json:"skootarId"` // 司机 skootar id
}

type GetDriverLocationOut struct {
	ResponseCode string  `json:"responseCode"`
	ResponseDesc string  `json:"responseDesc"`
	Lat          float64 `json:"lat"` // 纬度
	Lng          float64 `json:"lng"` // 经度
}
