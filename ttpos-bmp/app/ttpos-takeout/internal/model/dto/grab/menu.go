package grab

import (
	"context"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// ============================================================================
// Push Menu Webhook (POST /partner/v1/merchant/menu)
// ============================================================================

// PushGrabMenuDTO 推送菜单数据传输对象
type PushGrabMenuDTO struct {
	MerchantID        string                  `json:"merchantID"`
	PartnerMerchantID string                  `json:"partnerMerchantID"`
	Currency          grabfood.Currency       `json:"currency"`
	SellingTimes      []grabfood.SellingTime  `json:"sellingTimes"`
	Categories        []grabfood.MenuCategory `json:"categories"`
}

// ============================================================================
// Update Menu (POST /partner/v1/merchant/menu/notification)
// ============================================================================

// UpdateMenuNotifyRequest 菜单更新通知请求
type UpdateMenuNotifyRequest struct {
	MerchantID string `json:"merchantID"` // 商户 ID
}

// UpdateMenuNotifyResponse 菜单更新通知响应
type UpdateMenuNotifyResponse struct {
	RequestID string `json:"requestID"` // 请求 ID
}

// ============================================================================
// Menu Sync State Webhook (POST /partner/v1/merchant/menu/sync/state)
// Deprecated: 使用 SDK grabfood.MenuSyncWebhookRequest
// ============================================================================

// MenuSyncStateRequest 菜单同步状态回调
// Deprecated: 使用 SDK grabfood.MenuSyncWebhookRequest
type MenuSyncStateRequest struct {
	MerchantID        string          `json:"merchantID"`        // 商户 ID
	PartnerMerchantID string          `json:"partnerMerchantID"` // 合作商户 ID
	RequestID         string          `json:"requestID"`         // 请求 ID
	Status            string          `json:"status"`            // 状态: SUCCESS, QUEUED, FAIL
	Errors            []MenuSyncError `json:"errors,omitempty"`  // 错误信息
}

// MenuSyncError 菜单同步错误
type MenuSyncError struct {
	Field   string `json:"field"`   // 字段
	Code    string `json:"code"`    // 错误码
	Message string `json:"message"` // 消息
}

// MenuSyncStatus 菜单同步状态枚举
const (
	MenuSyncStatusQueued     = "QUEUED"
	MenuSyncStatusProcessing = "PROCESSING"
	MenuSyncStatusSuccess    = "SUCCESS"
	MenuSyncStatusFail       = "FAIL"
)

// ProviderMenuUpdateEvent 供应商菜单更新事件 (for RocketMQ)
type ProviderMenuUpdateEvent struct {
	ProviderName      string `json:"provider_name"`       // 供应商名称 (e.g., grab)
	MerchantID        string `json:"merchant_id"`         // 平台商户 ID
	PartnerMerchantID string `json:"partner_merchant_id"` // 合作商户 ID
	ShopUuid          string `json:"shop_uuid"`           // TTPOS 门店 UUID
	Uuid              uint64 `json:"uuid"`                // 唯一ID
	ReceivedAt        int64  `json:"received_at"`         // 接收时间戳
}

// ============================================================================
// Menu Notifier Interface
// ============================================================================

// MenuNotifier 菜单通知接口 (用于 SyncMenu)
type MenuNotifier interface {
	// NotifyMenuUpdate 通知 Grab 菜单已更新
	NotifyMenuUpdate(ctx context.Context, merchantID string) (requestID string, err error)
}
