package req

// SaveStoreScanOrderSettingReq 保存门店点餐配置请求
type SaveStoreScanOrderSettingReq struct {
	IsEnabled         int `json:"is_enabled"`          // 启用状态：0-关闭，1-开启
	EnableDelivery    int `json:"enable_delivery"`     // 外送服务：0-关闭，1-开启
	EnableSelfPickup  int `json:"enable_self_pickup"`  // 到店自取：0-关闭，1-开启
}
