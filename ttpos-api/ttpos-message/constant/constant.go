// Package constant 定义 ttpos-message 服务相关的常量
package constant

// Topic 常量定义
const (
	// TopicMessageSend 消息发送队列
	// 用于 ttpos-message 服务的消息发送
	TopicMessageSend = "message.send"

	// TopicMessageRetry 消息重试队列
	// 用于消息发送失败后的重试
	TopicMessageRetry = "message.retry"

	// TopicMessageStatusChange 消息状态变更通知队列
	// 用于通知消息状态变更（发送成功/失败）
	TopicMessageStatusChange = "message.status.change"
)

// 消息类型常量
const (
	// MessageTypeEmail 邮件消息
	MessageTypeEmail = "email"

	// MessageTypeSMS 短信消息
	MessageTypeSMS = "sms"

	// MessageTypePush 推送消息（预留）
	MessageTypePush = "push"

	// MessageTypeWechat 微信消息（预留）
	MessageTypeWechat = "wechat"
)

// 消息状态常量
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

// GetAllTopics 获取所有 ttpos-message 的 Topic
func GetAllTopics() []string {
	return []string{
		TopicMessageSend,
		TopicMessageRetry,
		TopicMessageStatusChange,
	}
}

// IsValidTopic 检查 Topic 是否有效
func IsValidTopic(topic string) bool {
	for _, t := range GetAllTopics() {
		if t == topic {
			return true
		}
	}
	return false
}

// GetAllMessageTypes 获取所有消息类型
func GetAllMessageTypes() []string {
	return []string{
		MessageTypeEmail,
		MessageTypeSMS,
		MessageTypePush,
		MessageTypeWechat,
	}
}

// IsValidMessageType 检查消息类型是否有效
func IsValidMessageType(messageType string) bool {
	for _, t := range GetAllMessageTypes() {
		if t == messageType {
			return true
		}
	}
	return false
}

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
