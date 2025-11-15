// Package util 提供验证工具函数
package util

import (
	"regexp"
	"ttpos-api/common/message"
)

// ValidateMessage 验证消息
// 提供统一的消息验证方法
func ValidateMessage(msg message.Message) error {
	return msg.Validate()
}

// IsValidEmail 验证邮箱地址格式
func IsValidEmail(email string) bool {
	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhone 验证手机号格式（中国大陆）
func IsValidPhone(phone string) bool {
	// 简单的手机号格式验证（11位数字，1开头）
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidUUID 验证 UUID 格式
func IsValidUUID(uuid string) bool {
	if uuid == "" {
		return false
	}
	// UUID v4 格式验证
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}

// IsValidMessageType 验证消息类型
func IsValidMessageType(messageType string) bool {
	validTypes := []string{"email", "sms", "push", "wechat"}
	for _, t := range validTypes {
		if t == messageType {
			return true
		}
	}
	return false
}
