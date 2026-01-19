// Package grab 定义 GrabFood API v1.1.3 交互的数据结构
//
// 仅保留常量定义以兼容现有代码
package grab

// ============================================================================
// Store Management APIs
// Deprecated: 使用 SDK grabfood.PauseStoreAPI, grabfood.GetStoreStatusAPI 等
// ============================================================================

// PauseStoreRequest 暂停门店请求
// Deprecated: 使用 SDK
type PauseStoreRequest struct {
	MerchantID string `json:"merchantID"` // 商户 ID
	IsPause    bool   `json:"isPause"`    // 是否暂停
	Duration   int    `json:"duration"`   // 暂停时长(分钟), 最大 1440 (24小时)
}

// GetStoreStatusResponse 获取门店状态响应
type GetStoreStatusResponse struct {
	MerchantID   string `json:"merchantID"`   // 商户 ID
	IsActive     bool   `json:"isActive"`     // 是否营业
	ClosingTime  string `json:"closingTime"`  // 关闭时间
	OpeningTime  string `json:"openingTime"`  // 开放时间
	ScheduleType string `json:"scheduleType"` // 调度类型
}

// ============================================================================
// Integration Status Webhook
// ============================================================================

// IntegrationStatusRequest 门店集成状态回调
type IntegrationStatusRequest struct {
	MerchantID        string `json:"merchantID"`        // 商户 ID
	PartnerMerchantID string `json:"partnerMerchantID"` // 合作商户 ID
	Status            string `json:"status"`            // 状态: ACTIVE, INACTIVE
}

// IntegrationStatus 集成状态枚举
const (
	IntegrationStatusActive   = "ACTIVE"
	IntegrationStatusInactive = "INACTIVE"
)
