package consts

// ProviderName 外送供应商名称
type ProviderName string

const (
	ProviderSkootar ProviderName = "skootar"
	ProviderGrab    ProviderName = "grab"
	ProviderLineman ProviderName = "lineman"
)

// ProviderShopStatus 第三方门店集成状态
type ProviderShopStatus string

const (
	ProviderShopStatusInactive ProviderShopStatus = "INACTIVE"
	ProviderShopStatusActive   ProviderShopStatus = "ACTIVE"
	ProviderShopStatusSyncing  ProviderShopStatus = "SYNCING"
	ProviderShopStatusFailed   ProviderShopStatus = "FAILED"
)

// OrderPrepareState 订单准备状态（接受/拒绝）
type OrderPrepareState string

const (
	// OrderPrepareStateAccepted 接受订单
	OrderPrepareStateAccepted OrderPrepareState = "Accepted"
	// OrderPrepareStateRejected 拒绝订单
	OrderPrepareStateRejected OrderPrepareState = "Rejected"
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

// MenuSyncType 菜单同步类型（用于 menu_log 表的 sync_type 字段）
type MenuSyncType string

const (
	// MenuSyncTypeFull 全量菜单同步
	MenuSyncTypeFull MenuSyncType = "FULL"
	// MenuSyncTypeBatchUpdateItem 批量更新商品
	MenuSyncTypeBatchUpdateItem MenuSyncType = "BATCH_UPDATE_ITEM"
	// MenuSyncTypeBatchUpdateModifier 批量更新修饰符
	MenuSyncTypeBatchUpdateModifier MenuSyncType = "BATCH_UPDATE_MODIFIER"
)
