package grab

// ShopIntegrationStatusEvent 门店配置事件
type ShopIntegrationStatusEvent struct {
	ShopUuid           uint64
	ProviderName       string
	ProviderShopStatus string
	ProviderMerchantId string
	UpdatedAt          int64
}
