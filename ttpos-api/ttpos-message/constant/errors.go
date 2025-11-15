// Package constant 定义 ttpos-message 模块相关的错误
package constant

import "errors"

var (
	// 消息发送相关错误

	// ErrMessageUUIDRequired 消息UUID不能为空
	ErrMessageUUIDRequired = errors.New("消息UUID不能为空")

	// ErrMessageTypeRequired 消息类型不能为空
	ErrMessageTypeRequired = errors.New("消息类型不能为空")

	// ErrInvalidMessageType 无效的消息类型
	ErrInvalidMessageType = errors.New("无效的消息类型")

	// ErrRecipientRequired 接收人不能为空
	ErrRecipientRequired = errors.New("接收人不能为空")
)
