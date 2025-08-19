package constant

// 采购订单状态常量
const (
	PurchaseOrderStatusDraft     = 0 // 草稿 - 待提交
	PurchaseOrderStatusPending   = 1 // 待审核
	PurchaseOrderStatusApproved  = 2 // 已审核 - 采购中 - 待收货
	PurchaseOrderStatusRejected  = 3 // 已驳回
	PurchaseOrderStatusCompleted = 4 // 已完成 - 全部收货 - 已收货
)

// 收货单状态常量
const (
	ReceiptOrderStatusPending  = 0 // 待收货
	ReceiptOrderStatusReceived = 1 // 已收货
	ReceiptOrderStatusRejected = 2 // 已取消
)
