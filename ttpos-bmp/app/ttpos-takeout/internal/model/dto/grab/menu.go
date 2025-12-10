package grab

import (
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
// Get Menu Webhook (GET /partner/v1/merchant/menu)
// ============================================================================

// GetMenuRequest 获取菜单请求
type GetMenuRequest struct {
	MerchantID        string `json:"merchantID"`        // 商户 ID
	PartnerMerchantID string `json:"partnerMerchantID"` // 合作商户 ID
}

// GetMenuResponse 菜单响应 (Grab Menu 结构)
type GetMenuResponse struct {
	MerchantID        string        `json:"merchantID"`        // 商户 ID
	PartnerMerchantID string        `json:"partnerMerchantID"` // 合作商户 ID
	Currency          Currency      `json:"currency"`          // 货币
	SellingTimes      []SellingTime `json:"sellingTimes"`      // 销售时间
	Categories        []Category    `json:"categories"`        // 分类
}

// SellingTime 销售时间
type SellingTime struct {
	ID           string       `json:"id"`           // 销售时间 ID
	Name         string       `json:"name"`         // 名称
	ServiceHours ServiceHours `json:"serviceHours"` // 服务时间
}

// ServiceHours 服务时间
type ServiceHours struct {
	Mon *DayHours `json:"mon,omitempty"`
	Tue *DayHours `json:"tue,omitempty"`
	Wed *DayHours `json:"wed,omitempty"`
	Thu *DayHours `json:"thu,omitempty"`
	Fri *DayHours `json:"fri,omitempty"`
	Sat *DayHours `json:"sat,omitempty"`
	Sun *DayHours `json:"sun,omitempty"`
}

// DayHours 单日时间段
type DayHours struct {
	Periods []Period `json:"periods"`
}

// Period 时间段
type Period struct {
	StartTime string `json:"startTime"` // HH:MM
	EndTime   string `json:"endTime"`   // HH:MM
}

// Category 分类
type Category struct {
	ID              string            `json:"id"`                        // 分类 ID
	Name            string            `json:"name"`                      // 名称
	NameTranslation map[string]string `json:"nameTranslation,omitempty"` // 名称翻译
	AvailableStatus string            `json:"availableStatus"`           // 可用状态: AVAILABLE, UNAVAILABLE
	SellingTimeID   string            `json:"sellingTimeID"`             // 销售时间 ID
	Items           []MenuItem        `json:"items"`                     // 菜品
}

// MenuItem 菜品
type MenuItem struct {
	ID                     string            `json:"id"`                               // 商品 ID
	Name                   string            `json:"name"`                             // 名称
	NameTranslation        map[string]string `json:"nameTranslation,omitempty"`        // 名称翻译
	Description            string            `json:"description"`                      // 描述
	DescriptionTranslation map[string]string `json:"descriptionTranslation,omitempty"` // 描述翻译
	AvailableStatus        string            `json:"availableStatus"`                  // 可用状态
	Price                  int64             `json:"price"`                            // 价格 (最小单位)
	Photos                 []string          `json:"photos,omitempty"`                 // 图片 URL
	SpecialType            string            `json:"specialType,omitempty"`            // 特殊类型: alcohol, tobacco
	Taxable                bool              `json:"taxable"`                          // 是否含税
	MaxStock               int               `json:"maxStock,omitempty"`               // 最大库存
	ModifierGroups         []ModifierGroup   `json:"modifierGroups,omitempty"`         // 修改项组
}

// ModifierGroup 修改项组
type ModifierGroup struct {
	ID                string            `json:"id"`                        // 组 ID
	Name              string            `json:"name"`                      // 名称
	NameTranslation   map[string]string `json:"nameTranslation,omitempty"` // 名称翻译
	AvailableStatus   string            `json:"availableStatus"`           // 可用状态
	SelectionRangeMin int               `json:"selectionRangeMin"`         // 最少选择
	SelectionRangeMax int               `json:"selectionRangeMax"`         // 最多选择
	Modifiers         []MenuModifier    `json:"modifiers"`                 // 修改项
}

// MenuModifier 菜单修改项
type MenuModifier struct {
	ID              string            `json:"id"`                        // 修改项 ID
	Name            string            `json:"name"`                      // 名称
	NameTranslation map[string]string `json:"nameTranslation,omitempty"` // 名称翻译
	AvailableStatus string            `json:"availableStatus"`           // 可用状态
	Price           int64             `json:"price"`                     // 价格
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
	StorageKey        string `json:"storage_key"`         // 存储 Key (Redis)
	ReceivedAt        int64  `json:"received_at"`         // 接收时间戳
}
