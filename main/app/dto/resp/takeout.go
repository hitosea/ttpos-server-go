package resp

type TakeoutDistanceResp struct {
	Distance     float64 // 预估距离
	TripDuration int     // 预估时间，单位秒
}

type CreateTakeoutOrderResp struct {
	TakeoutJobUuid string // 外送服务订单
	Status         string // 订单状态
	ShopOrderUuid  string // 商户订单
	TakeoutRefNo   string // 外送渠道订单，比如skootar订单
}

type GetDriverInfoResp struct {
	Name   string
	Phone  string
	Avatar string
	Rating float32
	Lat    float32
	Lng    float32
}
