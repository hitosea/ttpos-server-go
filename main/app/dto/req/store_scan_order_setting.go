package req

// SaveStoreScanOrderSettingReq 保存门店点餐配置请求
type SaveStoreScanOrderSettingReq struct {
	IsEnabled            int `json:"is_enabled"`                                               // 启用状态：0-关闭，1-开启
	EnableDelivery       int `json:"enable_delivery"`                                          // 外送服务：0-关闭，1-开启
	EnableSelfPickup     int `json:"enable_self_pickup"`                                       // 到店自取：0-关闭，1-开启
	IsOrderFirstPayLater int `json:"is_order_first_pay_later" form:"is_order_first_pay_later"` // 先下单后付：0-先付后下单(默认)，1-先下单后付
}
