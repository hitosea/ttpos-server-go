// Package constant 定义消息队列相关的常量
package constant

// MQ Topic 常量定义
// 所有消息队列的 Topic 统一在这里定义，确保各服务使用一致的 Topic 名称
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

// GetAllTopics 获取所有定义的 Topic
// 用于初始化时创建所有队列
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
