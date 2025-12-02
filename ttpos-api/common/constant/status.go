// Package constant 定义状态常量
package constant

// 消息状态常量
// 用于 ttpos-message 服务的消息状态定义
const (
	// MessageStatusPending 待发送
	MessageStatusPending = 0

	// MessageStatusSending 发送中
	MessageStatusSending = 1

	// MessageStatusSuccess 发送成功
	MessageStatusSuccess = 2

	// MessageStatusFailed 发送失败
	MessageStatusFailed = 3

	// MessageStatusRetrying 重试中
	MessageStatusRetrying = 4
)

// 订单状态常量
const (
	// OrderStatusPending 待支付
	OrderStatusPending = 0

	// OrderStatusPaid 已支付
	OrderStatusPaid = 1

	// OrderStatusCancelled 已取消
	OrderStatusCancelled = 2

	// OrderStatusCompleted 已完成
	OrderStatusCompleted = 3

	// OrderStatusRefunded 已退款
	OrderStatusRefunded = 4
)

// 会员状态常量
const (
	// MemberStatusActive 激活
	MemberStatusActive = 1

	// MemberStatusInactive 未激活
	MemberStatusInactive = 0

	// MemberStatusBlocked 已封禁
	MemberStatusBlocked = 2
)

// 桌台状态常量
const (
	// DeskStatusIdle 空闲
	DeskStatusIdle = 0

	// DeskStatusOccupied 使用中
	DeskStatusOccupied = 1

	// DeskStatusReserved 已预订
	DeskStatusReserved = 2

	// DeskStatusCleaning 清理中
	DeskStatusCleaning = 3
)

// GetMessageStatusText 获取消息状态文本
func GetMessageStatusText(status int) string {
	switch status {
	case MessageStatusPending:
		return "待发送"
	case MessageStatusSending:
		return "发送中"
	case MessageStatusSuccess:
		return "发送成功"
	case MessageStatusFailed:
		return "发送失败"
	case MessageStatusRetrying:
		return "重试中"
	default:
		return "未知状态"
	}
}

// GetOrderStatusText 获取订单状态文本
func GetOrderStatusText(status int) string {
	switch status {
	case OrderStatusPending:
		return "待支付"
	case OrderStatusPaid:
		return "已支付"
	case OrderStatusCancelled:
		return "已取消"
	case OrderStatusCompleted:
		return "已完成"
	case OrderStatusRefunded:
		return "已退款"
	default:
		return "未知状态"
	}
}

// GetMemberStatusText 获取会员状态文本
func GetMemberStatusText(status int) string {
	switch status {
	case MemberStatusActive:
		return "激活"
	case MemberStatusInactive:
		return "未激活"
	case MemberStatusBlocked:
		return "已封禁"
	default:
		return "未知状态"
	}
}

// GetDeskStatusText 获取桌台状态文本
func GetDeskStatusText(status int) string {
	switch status {
	case DeskStatusIdle:
		return "空闲"
	case DeskStatusOccupied:
		return "使用中"
	case DeskStatusReserved:
		return "已预订"
	case DeskStatusCleaning:
		return "清理中"
	default:
		return "未知状态"
	}
}
