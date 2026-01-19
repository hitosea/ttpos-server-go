package constant

// 采购订单状态常量
const (
	PurchaseOrderStatusDraft              = 0 // 草稿 - 待提交
	PurchaseOrderStatusPending            = 1 // 待审核
	PurchaseOrderStatusApproved           = 2 // 已审核 - 采购中 - 待收货
	PurchaseOrderStatusRejected           = 3 // 已驳回
	PurchaseOrderStatusCompleted          = 4 // 已完成 - 全部收货 - 已收货
	PurchaseOrderStatusHeadquarterPending = 5 // 待总部审核
)

// 采购类型常量
const (
	PurchaseTypeExternal = 1 // 外部采购
	PurchaseTypeBrand    = 2 // 内部采购（品牌采购）
)

// 总部状态常量
const (
	HeadquarterStatusDraft     = 0 // 待提交
	HeadquarterStatusPending   = 1 // 待审核
	HeadquarterStatusApproved  = 2 // 已审核 - 采购中 - 待收货
	HeadquarterStatusRejected  = 3 // 已驳回
	HeadquarterStatusCompleted = 4 // 已完成 - 全部收货 - 已收货
)

// 收货单状态常量
const (
	ReceiptOrderStatusPending  = 0 // 待收货
	ReceiptOrderStatusReceived = 1 // 已收货
	ReceiptOrderStatusRejected = 2 // 已取消
)
