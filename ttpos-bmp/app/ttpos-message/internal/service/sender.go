// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
)

type (
	IMailgun interface {
		// Init 初始化 Mailgun 服务
		// 从配置文件中加载 Mailgun 相关配置
		Init(ctx context.Context) error
		// SendEmail 发送邮件
		// 调用 Mailgun API 发送邮件
		// 参数：
		//   - ctx: 上下文对象
		//   - messageUuid: 消息UUID
		//   - recipient: 收件人邮箱地址
		//   - subject: 邮件主题
		//   - content: 邮件内容（HTML格式）
		//
		// 返回：
		//   - err: 错误信息
		SendEmail(ctx context.Context, messageUuid string, recipient string, subject string, content string) error
		// ValidateConfig 验证配置
		ValidateConfig(ctx context.Context) error
		// GetConfig 获取配置信息（用于调试）
		GetConfig() map[string]string
	}
)

var (
	localMailgun IMailgun
)

func Mailgun() IMailgun {
	if localMailgun == nil {
		panic("implement not found for interface IMailgun, forgot register?")
	}
	return localMailgun
}

func RegisterMailgun(i IMailgun) {
	localMailgun = i
}
