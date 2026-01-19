package grab

// ShopIntegrationStatusEvent 门店配置事件
type ShopIntegrationStatusEvent struct {
	ShopUuid           uint64
	ProviderName       string
	ProviderShopStatus string
	ProviderMerchantId string
	UpdatedAt          int64
}

// OrderEvent 订单事件
type OrderEvent struct {
	Action       string `json:"action,omitempty"`     // create, status_update, cancel
	ProviderName string `json:"providerName"`         // grab
	ShopUUID     string `json:"shopUuid"`             // TTPOS 店铺 UUID
	OrderUUID    string `json:"orderUuid"`            // 订单 UUID
	OrderID      string `json:"orderId,omitempty"`    // 平台订单 ID
	MerchantID   string `json:"merchantId,omitempty"` // 商户 ID
	Status       string `json:"status,omitempty"`     // 当前状态
	Timestamp    int64  `json:"timestamp,omitempty"`  // 事件时间戳
	Message      string `json:"message,omitempty"`    // 附加消息消息
}
