// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ================================================================================

package service

import (
	"context"

	"ttpos-bmp/app/ttpos-message/internal/model/dto"
)

// IMessage 消息服务接口
type IMessage interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, in *dto.SendMessageInput) (out *dto.SendMessageOutput, err error)
	// GetMessageStatus 查询消息状态
	GetMessageStatus(ctx context.Context, in *dto.GetMessageStatusInput) (out *dto.GetMessageStatusOutput, err error)
	// ResendMessage 重发消息
	ResendMessage(ctx context.Context, in *dto.ResendMessageInput) (out *dto.ResendMessageOutput, err error)
	// GetTemplateById 根据ID获取模板
	GetTemplateById(ctx context.Context, templateId uint64) (*dto.MessageTemplateDTO, error)
	// GetMessageByUuid 根据UUID获取消息记录
	GetMessageByUuid(ctx context.Context, uuid string) (*dto.MessageRecordDTO, error)
	// CreateMessageRecord 创建消息记录
	CreateMessageRecord(ctx context.Context, record *dto.MessageRecordDTO) error
	// UpdateMessageStatus 更新消息状态
	UpdateMessageStatus(ctx context.Context, uuid string, status int, errorMsg string) error
	// CreateSendLog 创建发送日志
	CreateSendLog(ctx context.Context, log *dto.MessageSendLogDTO) error
}

var localMessage IMessage

// Message 获取消息服务实例
func Message() IMessage {
	if localMessage == nil {
		panic("implement not found for interface IMessage, forgot register?")
	}
	return localMessage
}

// RegisterMessage 注册消息服务实现
func RegisterMessage(i IMessage) {
	localMessage = i
}
