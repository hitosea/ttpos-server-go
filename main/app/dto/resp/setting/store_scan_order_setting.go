package setting

// StoreScanOrderSettingResp 门店点餐配置响应
type StoreScanOrderSettingResp struct {
	IsEnabled            int `json:"is_enabled"`              // 启用状态：0-关闭，1-开启
	EnableDelivery       int `json:"enable_delivery"`         // 外送服务：0-关闭，1-开启
	EnableSelfPickup     int `json:"enable_self_pickup"`      // 到店自取：0-关闭，1-开启
	DeliveryAvailable    int `json:"delivery_available"`      // 外送服务是否可用（云平台是否开启）：0-不可用，1-可用
	SelfPickupAvailable  int `json:"self_pickup_available"`   // 到店自取是否可用（云平台是否开启）：0-不可用，1-可用
}
