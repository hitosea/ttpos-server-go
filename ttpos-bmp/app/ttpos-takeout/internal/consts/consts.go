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

// OrderType 订单类型
type OrderType string

const (
	// OrderTypeDeliveryByLineman LINE MAN 外送
	OrderTypeDeliveryByLineman OrderType = "DeliveryByLineman"
	// OrderTypeDeliveryByGrab Grab 外送
	OrderTypeDeliveryByGrab OrderType = "DeliveryByGrab"
	// OrderTypeDeliveryByProvider 第三方平台外送（通用）
	OrderTypeDeliveryByProvider OrderType = "DeliveryByProvider"
	// OrderTypeDeliveryByRestaurant 餐厅自配送
	OrderTypeDeliveryByRestaurant OrderType = "DeliveryByRestaurant"
	// OrderTypeTakeAway 自取
	OrderTypeTakeAway OrderType = "TakeAway"
	// OrderTypeDineIn 堂食
	OrderTypeDineIn OrderType = "DineIn"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	// OrderStatusPending 待接单
	OrderStatusPending OrderStatus = "PENDING"
	// OrderStatusAccepted 已接单
	OrderStatusAccepted OrderStatus = "ACCEPTED"
	// OrderStatusPreparing 准备中
	OrderStatusPreparing OrderStatus = "PREPARING"
	// OrderStatusReady 已完成（待取餐）
	OrderStatusReady OrderStatus = "READY"
	// OrderStatusCompleted 已完成
	OrderStatusCompleted OrderStatus = "COMPLETED"
	// OrderStatusCancelled 已取消
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// OrderAction 订单事件动作
type OrderAction string

const (
	// OrderActionCreate 创建订单
	OrderActionCreate OrderAction = "create"
	// OrderActionUpdate 订单内容更新
	OrderActionUpdate OrderAction = "update"
	// OrderActionStatusUpdate 订单状态更新
	OrderActionStatusUpdate OrderAction = "status_update"
	// OrderActionCancel 取消订单
	OrderActionCancel OrderAction = "cancel"
)

// LinemanCustomerType LINE MAN 订单类型
type LinemanCustomerType string

const (
	// LinemanCustomerTypeDelivery 外送
	LinemanCustomerTypeDelivery LinemanCustomerType = "DELIVERY"
	// LinemanCustomerTypePickup 自取
	LinemanCustomerTypePickup LinemanCustomerType = "PICKUP"
)
