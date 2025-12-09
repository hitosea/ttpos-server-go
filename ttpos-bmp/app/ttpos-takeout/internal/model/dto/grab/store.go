// Package grab 定义 GrabFood API v1.1.3 交互的数据结构
//
// Deprecated: 结构体已废弃，请使用 github.com/grab/grabfood-api-sdk-go SDK Model
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

// StoreHoursResponse 门店营业时间响应
type StoreHoursResponse struct {
	MerchantID  string       `json:"merchantID"`  // 商户 ID
	OpenPeriods []OpenPeriod `json:"openPeriods"` // 营业时段
}

// OpenPeriod 营业时段
type OpenPeriod struct {
	Day     string   `json:"day"`     // 星期
	Periods []Period `json:"periods"` // 时间段
}

// UpdateStoreHoursRequest 更新门店营业时间请求
type UpdateStoreHoursRequest struct {
	MerchantID  string       `json:"merchantID"`  // 商户 ID
	OpenPeriods []OpenPeriod `json:"openPeriods"` // 营业时段
}

// SpecialHoursRequest 特殊营业时间请求
type SpecialHoursRequest struct {
	MerchantID   string           `json:"merchantID"`   // 商户 ID
	SpecialHours []SpecialHourDay `json:"specialHours"` // 特殊日期
}

// SpecialHourDay 特殊日期时间
type SpecialHourDay struct {
	Date    string   `json:"date"`    // 日期 YYYY-MM-DD
	Periods []Period `json:"periods"` // 时间段，空数组表示休息
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
