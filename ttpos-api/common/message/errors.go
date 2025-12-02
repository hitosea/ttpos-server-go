// Package message 定义消息相关的公共错误
package message

import (
	"errors"
	"fmt"
)

var (
	// ErrMessageIDRequired 消息ID不能为空
	ErrMessageIDRequired = errors.New("消息ID不能为空")

	// ErrTopicRequired 消息主题不能为空
	ErrTopicRequired = errors.New("消息主题不能为空")

	// ErrTimestampRequired 消息时间戳不能为空
	ErrTimestampRequired = errors.New("消息时间戳不能为空")

	// ErrInvalidJSON JSON格式无效
	ErrInvalidJSON = errors.New("JSON格式无效")
)

// NewValidationError 创建验证错误
// 用于生成带有字段名和详细信息的验证错误
func NewValidationError(field, message string, args ...interface{}) error {
	if len(args) > 0 {
		field = fmt.Sprintf(field, args...)
	}
	return fmt.Errorf("字段 %s 验证失败: %s", field, message)
}
