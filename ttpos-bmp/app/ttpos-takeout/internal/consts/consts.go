package consts

// ProviderName 外送供应商名称
type ProviderName string

const (
	ProviderSkootar ProviderName = "skootar"
	ProviderGrab    ProviderName = "grab"
)

// ProviderShopStatus 第三方门店集成状态
type ProviderShopStatus string

const (
	ProviderShopStatusInactive ProviderShopStatus = "INACTIVE"
	ProviderShopStatusActive   ProviderShopStatus = "ACTIVE"
	ProviderShopStatusSyncing  ProviderShopStatus = "SYNCING"
	ProviderShopStatusFailed   ProviderShopStatus = "FAILED"
)

// TTPOS_HEADER_CALLBACK_AUTH TTPOS回调Auth
const TTPOS_HEADER_CALLBACK_AUTH = "X-TTPOS-Callback-Auth"

// TTPOS_HEADER_SECRET TTPOS服务间认证头
const TTPOS_HEADER_SECRET = "X-TTPOS-SECRET"

// MapGrabIntegrationStatus 映射 Grab integrationStatus 到内部状态
func MapGrabIntegrationStatus(grabStatus string) ProviderShopStatus {
	switch grabStatus {
	case "ACTIVE":
		return ProviderShopStatusActive
	case "INACTIVE":
		return ProviderShopStatusInactive
	case "SYNCING":
		return ProviderShopStatusSyncing
	case "FAILED":
		return ProviderShopStatusFailed
	default:
		// 未知状态默认为 INACTIVE
		return ProviderShopStatusInactive
	}
}
