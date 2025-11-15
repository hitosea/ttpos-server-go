// Package util 提供工具函数
package util

import (
	"encoding/json"
	"ttpos-api/common/message"
)

// MarshalMessage 序列化消息为 JSON
// 提供统一的消息序列化方法
func MarshalMessage(msg message.Message) ([]byte, error) {
	return json.Marshal(msg)
}

// UnmarshalMessage 反序列化 JSON 为消息
// 提供统一的消息反序列化方法
func UnmarshalMessage(data []byte, msg interface{}) error {
	return json.Unmarshal(data, msg)
}

// MarshalMessagePretty 序列化消息为格式化的 JSON
// 用于调试和日志输出
func MarshalMessagePretty(msg message.Message) ([]byte, error) {
	return json.MarshalIndent(msg, "", "  ")
}

// ToJSONString 将消息转换为 JSON 字符串
func ToJSONString(msg message.Message) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONStringPretty 将消息转换为格式化的 JSON 字符串
func ToJSONStringPretty(msg message.Message) (string, error) {
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsValidJSON 检查字符串是否为有效的 JSON
func IsValidJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
