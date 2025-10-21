// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ================================================================================

package service

import (
	"context"
)

// IMailgun Mailgun 邮件发送服务接口
type IMailgun interface {
	// Init 初始化服务
	Init(ctx context.Context) error
	// SendEmail 发送邮件
	SendEmail(ctx context.Context, messageUuid, recipient, subject, content string) error
	// ValidateConfig 验证配置
	ValidateConfig(ctx context.Context) error
	// GetConfig 获取配置信息（用于调试）
	GetConfig() map[string]string
}

var localMailgun IMailgun

// Mailgun 获取 Mailgun 服务实例
func Mailgun() IMailgun {
	if localMailgun == nil {
		panic("implement not found for interface IMailgun, forgot register?")
	}
	return localMailgun
}

// RegisterMailgun 注册 Mailgun 服务实现
func RegisterMailgun(i IMailgun) {
	localMailgun = i
}
