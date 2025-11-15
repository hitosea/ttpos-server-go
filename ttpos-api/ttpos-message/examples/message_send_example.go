// Package examples 提供使用示例
package examples

import (
	"fmt"
	"ttpos-api/common/util"
	msgConstant "ttpos-api/ttpos-message/constant"
	msgMessage "ttpos-api/ttpos-message/message"
)

// MessageSendExample 消息发送示例
func MessageSendExample() {
	// 创建消息发送消息
	msg := msgMessage.NewMessageSendMessage("msg-uuid-123", msgConstant.MessageTypeEmail)

	// 设置公司UUID和链路追踪ID
	msg.WithCompanyUUID("company-uuid-456")
	msg.WithTraceID("trace-id-789")

	// 验证消息
	if err := msg.Validate(); err != nil {
		fmt.Printf("消息验证失败: %v\n", err)
		return
	}

	// 序列化为 JSON
	jsonData, err := msg.ToJSON()
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}

	fmt.Printf("消息JSON: %s\n", string(jsonData))

	// 发送到消息队列（伪代码）
	// mq.Publish(msgConstant.TopicMessageSend, jsonData)
}

// MessageReceiveExample 消息接收示例
func MessageReceiveExample(jsonData []byte) {
	// 反序列化消息
	var msg msgMessage.MessageSendMessage
	if err := msg.FromJSON(jsonData); err != nil {
		fmt.Printf("反序列化失败: %v\n", err)
		return
	}

	// 验证消息
	if err := msg.Validate(); err != nil {
		fmt.Printf("消息验证失败: %v\n", err)
		return
	}

	// 处理消息
	fmt.Printf("收到消息: MessageUUID=%s, MessageType=%s\n", msg.MessageUUID, msg.MessageType)

	// 获取消息详情（从数据库）
	// record := getMessageFromDB(msg.MessageUUID)
	// sendEmail(record)
}

// MessageRetryExample 消息重试示例
func MessageRetryExample() {
	// 创建重试消息
	msg := msgMessage.NewMessageRetryMessage(
		"msg-uuid-123",
		msgConstant.MessageTypeEmail,
		1,
		"发送失败: 网络超时",
	)

	msg.WithCompanyUUID("company-uuid-456")

	// 验证消息
	if err := msg.Validate(); err != nil {
		fmt.Printf("消息验证失败: %v\n", err)
		return
	}

	// 序列化为 JSON
	jsonData, err := msg.ToJSON()
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}

	fmt.Printf("重试消息JSON: %s\n", string(jsonData))

	// 发送到重试队列
	// mq.Publish(msgConstant.TopicMessageRetry, jsonData)
}

// MessageStatusChangeExample 消息状态变更示例
func MessageStatusChangeExample() {
	// 创建状态变更消息
	msg := msgMessage.NewMessageStatusChangeMessage(
		"msg-uuid-123",
		msgConstant.MessageTypeEmail,
		msgConstant.MessageStatusSending,
		msgConstant.MessageStatusSuccess,
	)

	msg.WithCompanyUUID("company-uuid-456")
	msg.SendTime = util.GetCurrentTimestamp()

	// 验证消息
	if err := msg.Validate(); err != nil {
		fmt.Printf("消息验证失败: %v\n", err)
		return
	}

	// 序列化为 JSON（格式化输出）
	jsonStr, err := util.ToJSONStringPretty(msg)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}

	fmt.Printf("状态变更消息JSON:\n%s\n", jsonStr)

	// 发送到状态变更队列
	// mq.Publish(msgConstant.TopicMessageStatusChange, []byte(jsonStr))
}

// ValidatorExample 验证工具示例
func ValidatorExample() {
	// 验证邮箱
	email := "test@example.com"
	if util.IsValidEmail(email) {
		fmt.Printf("邮箱格式正确: %s\n", email)
	}

	// 验证手机号
	phone := "13800138000"
	if util.IsValidPhone(phone) {
		fmt.Printf("手机号格式正确: %s\n", phone)
	}

	// 验证消息类型
	messageType := msgConstant.MessageTypeEmail
	if msgConstant.IsValidMessageType(messageType) {
		fmt.Printf("消息类型有效: %s\n", messageType)
	}

	// 验证 Topic
	topic := msgConstant.TopicMessageSend
	if msgConstant.IsValidTopic(topic) {
		fmt.Printf("Topic 有效: %s\n", topic)
	}
}

// HelperExample 辅助工具示例
func HelperExample() {
	// 获取当前时间戳
	timestamp := util.GetCurrentTimestamp()
	fmt.Printf("当前时间戳: %d\n", timestamp)

	// 格式化时间戳
	timeStr := util.FormatTimestamp(timestamp, "2006-01-02 15:04:05")
	fmt.Printf("格式化时间: %s\n", timeStr)

	// 生成消息ID
	messageID := util.GenerateMessageID("msg")
	fmt.Printf("生成的消息ID: %s\n", messageID)

	// 计算 MD5
	text := "hello world"
	hash := util.MD5Hash(text)
	fmt.Printf("MD5 哈希: %s\n", hash)
}
