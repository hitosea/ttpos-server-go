// Package message 定义 ttpos-message 服务相关的消息结构体
package message

import (
	"encoding/json"
	"ttpos-api/common/message"
	"ttpos-api/ttpos-message/constant"
)

// MessageSendMessage 消息发送队列消息
// 用于 ttpos-message 服务的消息发送队列
// 对应 RocketMQ Topic: message.send
type MessageSendMessage struct {
	message.BaseMessage
	MessageUUID string `json:"message_uuid"` // 消息UUID（业务消息的唯一标识）
	MessageType string `json:"message_type"` // 消息类型（email/sms）
}

// NewMessageSendMessage 创建消息发送消息
func NewMessageSendMessage(messageUUID, messageType string) *MessageSendMessage {
	return &MessageSendMessage{
		BaseMessage: message.NewBaseMessage("message.send"),
		MessageUUID: messageUUID,
		MessageType: messageType,
	}
}

// Validate 验证消息字段
func (m *MessageSendMessage) Validate() error {
	// 验证基础字段
	if err := m.BaseMessage.Validate(); err != nil {
		return err
	}

	// 验证业务字段
	if m.MessageUUID == "" {
		return constant.ErrMessageUUIDRequired
	}
	if m.MessageType == "" {
		return constant.ErrMessageTypeRequired
	}
	if m.MessageType != "email" && m.MessageType != "sms" {
		return constant.ErrInvalidMessageType
	}

	return nil
}

// ToJSON 序列化为 JSON
func (m *MessageSendMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从 JSON 反序列化
func (m *MessageSendMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// MessageRetryMessage 消息重试队列消息
// 用于消息发送失败后的重试
// 对应 RocketMQ Topic: message.retry
type MessageRetryMessage struct {
	message.BaseMessage
	MessageUUID  string `json:"message_uuid"`  // 消息UUID
	MessageType  string `json:"message_type"`  // 消息类型
	RetryCount   int    `json:"retry_count"`   // 当前重试次数
	ErrorMessage string `json:"error_message"` // 错误信息
}

// NewMessageRetryMessage 创建消息重试消息
func NewMessageRetryMessage(messageUUID, messageType string, retryCount int, errorMessage string) *MessageRetryMessage {
	return &MessageRetryMessage{
		BaseMessage:  message.NewBaseMessage("message.retry"),
		MessageUUID:  messageUUID,
		MessageType:  messageType,
		RetryCount:   retryCount,
		ErrorMessage: errorMessage,
	}
}

// Validate 验证消息字段
func (m *MessageRetryMessage) Validate() error {
	// 验证基础字段
	if err := m.BaseMessage.Validate(); err != nil {
		return err
	}

	// 验证业务字段
	if m.MessageUUID == "" {
		return constant.ErrMessageUUIDRequired
	}
	if m.MessageType == "" {
		return constant.ErrMessageTypeRequired
	}

	return nil
}

// ToJSON 序列化为 JSON
func (m *MessageRetryMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从 JSON 反序列化
func (m *MessageRetryMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// MessageStatusChangeMessage 消息状态变更通知消息
// 用于通知消息状态变更（发送成功/失败）
// 对应 RocketMQ Topic: message.status.change
type MessageStatusChangeMessage struct {
	message.BaseMessage
	MessageUUID  string `json:"message_uuid"`  // 消息UUID
	MessageType  string `json:"message_type"`  // 消息类型
	OldStatus    int    `json:"old_status"`    // 旧状态
	NewStatus    int    `json:"new_status"`    // 新状态
	ErrorMessage string `json:"error_message"` // 错误信息（如果有）
	SendTime     int64  `json:"send_time"`     // 发送时间
}

// NewMessageStatusChangeMessage 创建消息状态变更消息
func NewMessageStatusChangeMessage(messageUUID, messageType string, oldStatus, newStatus int) *MessageStatusChangeMessage {
	return &MessageStatusChangeMessage{
		BaseMessage: message.NewBaseMessage("message.status.change"),
		MessageUUID: messageUUID,
		MessageType: messageType,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
	}
}

// Validate 验证消息字段
func (m *MessageStatusChangeMessage) Validate() error {
	// 验证基础字段
	if err := m.BaseMessage.Validate(); err != nil {
		return err
	}

	// 验证业务字段
	if m.MessageUUID == "" {
		return constant.ErrMessageUUIDRequired
	}

	return nil
}

// ToJSON 序列化为 JSON
func (m *MessageStatusChangeMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从 JSON 反序列化
func (m *MessageStatusChangeMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}
