package resp

type TakeoutDistanceResp struct {
	Distance     float64 // 预估距离。单位km
	TripDuration int     // 预估时间，单位秒
}

type CreateTakeoutOrderResp struct {
	TakeoutJobUuid string // 外送服务订单
	Status         string // 订单状态
	ShopOrderUuid  string // 商户订单
	TakeoutRefNo   string // 外送渠道订单，比如skootar订单
	FinishTime     string // 预计送达时间
}

type GetDriverInfoResp struct {
	Name   string  // 骑手姓名
	Phone  string  // 骑手电话
	Avatar string  // 骑手头像
	Rating float64 // 骑手评分
	Lat    float64 // 骑手纬度
	Lng    float64 // 骑手经度
}
