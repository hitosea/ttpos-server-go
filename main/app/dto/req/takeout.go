package req

type TakeoutAddress struct {
	AddressName string
	Address     string
	Lat         string
	Lng         string
}

type TakeoutDistanceReq struct {
	ProviderName string            // 外送渠道
	Address      []*TakeoutAddress // 起止地址和经纬度
}

type TakeoutLocation struct {
	TakeoutAddress
	ContactName  string
	ContactPhone string
}

type CreateTakeoutOrderReq struct {
	ProviderName     string           // 外送渠道
	CustomerLocation *TakeoutLocation // 顾客地址
	MerchantLocation *TakeoutLocation // 商户地址
	Remark           string           // 备注
	CallbackUrl      string           // 回调地址
	ShopOrderUuid    string           // 商户订单号
}

type ConfirmTakeoutOrderReq struct {
	ShopOrderUuid string // 商户订单号
}

type GetDriverInfoReq struct {
	ShopOrderUuid string // 商户订单号
}
