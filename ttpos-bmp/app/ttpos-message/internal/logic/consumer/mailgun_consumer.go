package consumer

import (
	"context"
	"encoding/json"
	"ttpos-bmp/app/ttpos-message/internal/consts"
	"ttpos-bmp/app/ttpos-message/internal/model/dto"
	"ttpos-bmp/app/ttpos-message/internal/service"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type MailgunConsumer struct {
}

func (*MailgunConsumer) GetTopic() string {
	return consts.TopicMessageSend
}
func (*MailgunConsumer) GetConcurrency() int {
	return 10
}

func (*MailgunConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	// 解析消息
	var msg dto.RocketMQMessage
	if err := json.Unmarshal(mqMsg.Body, &msg); err != nil {
		g.Log().Error(ctx, "解析队列消息失败", err)
		return err
	}

	return processMessage(ctx, &msg)
}

// processMessage 处理消息（统一的消息处理逻辑）
// 从数据库获取消息详情后进行处理
func processMessage(ctx context.Context, msg *dto.RocketMQMessage) error {
	g.Log().Info(ctx, "开始处理消息",
		"uuid", msg.MessageUuid,
		"type", msg.MessageType,
	)

	// 从数据库获取消息详情
	record, err := service.Message().GetMessageByUuid(ctx, msg.MessageUuid)
	if err != nil {
		g.Log().Error(ctx, "获取消息详情失败", "error", err, "uuid", msg.MessageUuid)
		return gerror.Wrap(err, "获取消息详情失败")
	}

	// 更新消息状态为发送中
	err = service.Message().UpdateMessageStatus(ctx, msg.MessageUuid, consts.MessageStatusSending, "")
	if err != nil {
		g.Log().Error(ctx, "更新消息状态失败", err)
	}

	// 根据消息类型调用对应的发送服务
	var sendErr error
	switch record.MessageType {
	case consts.MessageTypeEmail:
		sendErr = service.Mailgun().SendEmail(ctx, record.Uuid, record.Recipient, record.Subject, record.Content)
	default:
		sendErr = gerror.Newf("不支持的消息类型: %s", record.MessageType)
	}

	// 更新消息状态
	if sendErr != nil {
		g.Log().Error(ctx, "消息发送失败",
			"uuid", msg.MessageUuid,
			"error", sendErr,
		)
		_ = service.Message().UpdateMessageStatus(ctx, msg.MessageUuid, consts.MessageStatusFailed, sendErr.Error())
		return sendErr
	}

	// 更新消息状态为发送成功
	err = service.Message().UpdateMessageStatus(ctx, msg.MessageUuid, consts.MessageStatusSuccess, "")
	if err != nil {
		g.Log().Error(ctx, "更新消息状态失败", err)
	}

	g.Log().Info(ctx, "消息处理完成",
		"uuid", msg.MessageUuid,
		"result", "success",
	)

	return nil
}
