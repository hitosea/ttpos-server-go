// Package message 定义消息队列的消息结构体
package message

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BaseMessage 基础消息结构体
// 所有消息队列消息都应该包含这些基础字段
type BaseMessage struct {
	MessageID   string `json:"message_id"`   // 消息唯一标识（用于幂等性和追踪）
	Topic       string `json:"topic"`        // 消息主题/队列名称
	Timestamp   int64  `json:"timestamp"`    // 消息时间戳（Unix 时间戳，秒）
	Version     string `json:"version"`      // 消息版本号（用于兼容性处理）
	CompanyUUID string `json:"company_uuid"` // 公司UUID（多租户场景）
	TraceID     string `json:"trace_id"`     // 链路追踪ID（可选）
}

// NewBaseMessage 创建基础消息
// 自动生成 MessageID 和 Timestamp
func NewBaseMessage(topic string) BaseMessage {
	return BaseMessage{
		MessageID: uuid.New().String(),
		Topic:     topic,
		Timestamp: time.Now().Unix(),
		Version:   "1.0",
	}
}

// WithCompanyUUID 设置公司UUID
func (m *BaseMessage) WithCompanyUUID(companyUUID string) *BaseMessage {
	m.CompanyUUID = companyUUID
	return m
}

// WithTraceID 设置链路追踪ID
func (m *BaseMessage) WithTraceID(traceID string) *BaseMessage {
	m.TraceID = traceID
	return m
}

// WithVersion 设置消息版本
func (m *BaseMessage) WithVersion(version string) *BaseMessage {
	m.Version = version
	return m
}

// ToJSON 将消息序列化为 JSON
func (m *BaseMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从 JSON 反序列化
func (m *BaseMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// Validate 验证基础消息字段
func (m *BaseMessage) Validate() error {
	if m.MessageID == "" {
		return ErrMessageIDRequired
	}
	if m.Topic == "" {
		return ErrTopicRequired
	}
	if m.Timestamp == 0 {
		return ErrTimestampRequired
	}
	return nil
}

// Message 消息接口
// 所有消息结构体都应该实现这个接口
type Message interface {
	// GetMessageID 获取消息ID
	GetMessageID() string
	// GetTopic 获取消息主题
	GetTopic() string
	// GetTimestamp 获取消息时间戳
	GetTimestamp() int64
	// Validate 验证消息
	Validate() error
	// ToJSON 序列化为JSON
	ToJSON() ([]byte, error)
}

// GetMessageID 实现 Message 接口
func (m *BaseMessage) GetMessageID() string {
	return m.MessageID
}

// GetTopic 实现 Message 接口
func (m *BaseMessage) GetTopic() string {
	return m.Topic
}

// GetTimestamp 实现 Message 接口
func (m *BaseMessage) GetTimestamp() int64 {
	return m.Timestamp
}
