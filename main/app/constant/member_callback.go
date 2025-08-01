package constant

// 外送渠道方
const (
	ProviderNameSkootar = "skootar"
	ProviderNameGrab    = "grab"
)

// skootar 订单状态
const (
	SkootarCancel              = "0"  // 已取消 Canceled
	SkootarNew                 = "1"  // 新订单 New
	SkootarReviewing           = "2"  // 正在审核中 Under review
	SkootarApproved            = "3"  // 审核通过 Mission created
	SkootarMatchingDriver      = "4"  // 正在匹配司机 Matching
	SkootarDriverMatched       = "5"  // 已匹配司机 Assigned
	SkootarFetching            = "6"  // 正在前往取货 OTW to pickup
	SkootarDelivering          = "7"  // 正在前往送货 OTW to delivery
	SkootarPendingModification = "8"  // 等待修改 Pending
	SkootarCompleted           = "9"  // 已完成 Completed
	SkootarInvoiceGenerated    = "10" // 已生成发票 Invoice Created
	SkootarServiceFeePaid      = "11" // 已支付服务费 Paid
)
