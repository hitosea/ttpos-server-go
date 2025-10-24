// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-message/internal/model/dto"
)

type (
	IMessage interface {
		// SendMessage 发送消息
		// 接收消息发送请求，创建消息记录并提交到消息队列
		// 参数：
		//   - ctx: 上下文对象
		//   - in: 发送消息输入参数
		//
		// 返回：
		//   - out: 发送结果
		//   - err: 错误信息
		SendMessage(ctx context.Context, in *dto.SendMessageInput) (out *dto.SendMessageOutput, err error)
		// GetMessageStatus 查询消息状态
		// 根据消息UUID查询消息的发送状态和详细信息
		// 参数：
		//   - ctx: 上下文对象
		//   - in: 查询消息状态输入参数
		//
		// 返回：
		//   - out: 消息状态信息
		//   - err: 错误信息
		GetMessageStatus(ctx context.Context, in *dto.GetMessageStatusInput) (out *dto.GetMessageStatusOutput, err error)
		// ResendMessage 重发消息
		// 针对发送失败的消息，支持重新发送
		// 参数：
		//   - ctx: 上下文对象
		//   - in: 重发消息输入参数
		//
		// 返回：
		//   - out: 重发结果
		//   - err: 错误信息
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
)

var (
	localMessage IMessage
)

func Message() IMessage {
	if localMessage == nil {
		panic("implement not found for interface IMessage, forgot register?")
	}
	return localMessage
}

func RegisterMessage(i IMessage) {
	localMessage = i
}
