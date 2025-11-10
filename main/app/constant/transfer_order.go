package constant

// 调拨类型
const (
	TransferTypeIn  = 1 // 调入
	TransferTypeOut = 2 // 调出
)

// 调拨单状态
const (
	TransferOrderStatusDraft     = 0 // 待提交
	TransferOrderStatusPending   = 1 // 待审核
	TransferOrderStatusRejected  = 2 // 已驳回
	TransferOrderStatusReceiving = 3 // 待收货
	TransferOrderStatusCompleted = 4 // 已完成
)

// 审批状态
const (
	TransferApprovalPending  = 0 // 待审批
	TransferApprovalApproved = 1 // 已通过
	TransferApprovalRejected = 2 // 已驳回
	TransferApprovalSkipped  = 3 // 已跳过, 无需审核（兼容旧常量）
)

// 审批类型
const (
	TransferApprovalTypeSender         = "sender"          // 发货门店
	TransferApprovalTypeSenderParent   = "sender_parent"   // 发货门店上级
	TransferApprovalTypeReceiver       = "receiver"        // 收货门店
	TransferApprovalTypeReceiverParent = "receiver_parent" // 收货门店上级
)

// 收货状态
const (
	TransferReceivePending   = 0 // 待收货
	TransferReceiveCompleted = 1 // 已收货
)

// 操作动作
const (
	TransferActionCreate  = "create"  // 创建
	TransferActionSubmit  = "submit"  // 提交
	TransferActionApprove = "approve" // 审核通过
	TransferActionReject  = "reject"  // 驳回
	TransferActionReceive = "receive" // 收货
	TransferActionDelete  = "delete"  // 删除
)

// 操作角色/审批类型
const (
	TransferRoleInitiator      = "initiator"       // 发起人公司
	TransferRoleSender         = "sender"          // 发货门店
	TransferRoleSenderParent   = "sender_parent"   // 发货门店上级
	TransferRoleReceiverParent = "receiver_parent" // 收货门店上级
	TransferRoleReceiver       = "receiver"        // 收货门店
)

// 调拨单状态文本映射
var TransferOrderStatusText = map[int]string{
	TransferOrderStatusDraft:     "待提交",
	TransferOrderStatusPending:   "待审核",
	TransferOrderStatusRejected:  "已驳回",
	TransferOrderStatusReceiving: "待收货",
	TransferOrderStatusCompleted: "已完成",
}

// 调拨类型文本映射
var TransferTypeText = map[int]string{
	TransferTypeIn:  "调入",
	TransferTypeOut: "调出",
}

// 审批状态文本映射
var TransferApprovalStatusText = map[int]string{
	TransferApprovalPending:  "待审批",
	TransferApprovalApproved: "已通过",
	TransferApprovalRejected: "已驳回",
	TransferApprovalSkipped:  "已跳过",
}

// 是否必须审批
const (
	TransferApprovalNotRequired = 0 // 非必须
	TransferApprovalRequired    = 1 // 必须
)
