package constant

// 采购订单状态常量
const (
	PurchaseOrderStatusDraft           = 0 // 草稿 - 待提交
	PurchaseOrderStatusPending         = 1 // 待审核
	PurchaseOrderStatusApproved        = 2 // 已审核
	PurchaseOrderStatusRejected        = 3 // 已驳回
	PurchaseOrderStatusPurchasing      = 4 // 采购中
	PurchaseOrderStatusPartialReceived = 5 // 部分到货
	PurchaseOrderStatusCompleted       = 6 // 已完成
	PurchaseOrderStatusCancelled       = 7 // 已取消
)

// 采购订单优先级常量
const (
	PurchaseOrderPriorityLow    = 1 // 低
	PurchaseOrderPriorityMedium = 2 // 中
	PurchaseOrderPriorityHigh   = 3 // 高
	PurchaseOrderPriorityUrgent = 4 // 紧急
)

// 收货质检状态常量
const (
	QualityStatusQualified        = 1 // 合格
	QualityStatusUnqualified      = 2 // 不合格
	QualityStatusPartialQualified = 3 // 部分合格
)

// 收货单状态常量
const (
	ReceiptOrderStatusPending  = 0 // 待收货
	ReceiptOrderStatusReceived = 1 // 已收货
)
