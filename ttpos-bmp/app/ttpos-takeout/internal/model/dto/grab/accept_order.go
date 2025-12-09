// Package grab 定义 GrabFood API v1.1.3 交互的数据结构
//
// Deprecated: 结构体已废弃，请使用 github.com/grab/grabfood-api-sdk-go SDK Model
// 仅保留常量定义以兼容现有代码
package grab

// ============================================================================
// Accept/Reject Order API
// Deprecated: 使用 SDK grabfood.AcceptRejectOrderAPI
// ============================================================================

// AcceptOrderRequest 接受订单请求
// Deprecated: 使用 SDK
type AcceptOrderRequest struct {
	OrderID string `json:"orderID"` // 订单 ID
	State   string `json:"state"`   // ACCEPTED
}

// RejectOrderRequest 拒绝订单请求
type RejectOrderRequest struct {
	OrderID    string `json:"orderID"`    // 订单 ID
	State      string `json:"state"`      // REJECTED
	RejectCode int    `json:"rejectCode"` // 拒绝原因码
}

// RejectCode 拒绝原因码
const (
	RejectCodeOutOfStock           = 1  // 商品缺货
	RejectCodeStoreClose           = 2  // 门店已关闭
	RejectCodeDeliveryAddressIssue = 3  // 配送地址问题
	RejectCodeBusy                 = 4  // 太忙无法接单
	RejectCodeOther                = 99 // 其他原因
)

// ============================================================================
// Mark Order Ready API
// ============================================================================

// MarkOrderReadyRequest 标记订单准备完成请求
type MarkOrderReadyRequest struct {
	OrderID    string `json:"orderID"`    // 订单 ID
	MarkStatus string `json:"markStatus"` // READY, PICKED_UP (DineIn)
}

// MarkStatus 标记状态
const (
	MarkStatusReady    = "READY"
	MarkStatusPickedUp = "PICKED_UP"
)

// ============================================================================
// Update Order Ready Time API
// ============================================================================

// UpdateOrderReadyTimeRequest 更新订单准备时间请求
type UpdateOrderReadyTimeRequest struct {
	OrderID      string `json:"orderID"`      // 订单 ID
	NewReadyTime string `json:"newReadyTime"` // 新准备时间 (RFC3339)
}

// ============================================================================
// Update Delivery State API
// ============================================================================

// UpdateDeliveryStateRequest 更新配送状态请求
type UpdateDeliveryStateRequest struct {
	OrderID string `json:"orderID"` // 订单 ID
	State   string `json:"state"`   // COLLECTED, DELIVERED
}

// DeliveryState 配送状态
const (
	DeliveryStateCollected = "COLLECTED"
	DeliveryStateDelivered = "DELIVERED"
)

// ============================================================================
// Cancel Order API
// ============================================================================

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	OrderID    string `json:"orderID"`    // 订单 ID
	CancelCode int    `json:"cancelCode"` // 取消原因码
}

// CancelCode 取消原因码
const (
	CancelCodeOutOfStock      = 1001 // 商品缺货
	CancelCodeStoreClosed     = 1002 // 门店已关闭
	CancelCodeDuplicateOrder  = 1003 // 重复订单
	CancelCodePricingError    = 1004 // 价格错误
	CancelCodeCustomerRequest = 1005 // 顾客要求
	CancelCodeOther           = 1099 // 其他原因
)

// CheckOrderCancelableResponse 检查订单是否可取消响应
type CheckOrderCancelableResponse struct {
	Cancelable bool   `json:"cancelable"` // 是否可取消
	Reason     string `json:"reason"`     // 原因
}
