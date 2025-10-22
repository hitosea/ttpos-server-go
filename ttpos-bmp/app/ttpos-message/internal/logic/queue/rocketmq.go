// Package queue 实现消息队列服务
package queue

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"ttpos-bmp/app/ttpos-message/internal/consts"
	"ttpos-bmp/app/ttpos-message/internal/model/dto"
	"ttpos-bmp/app/ttpos-message/internal/service"
)

// sQueue 队列服务实现
type sQueue struct {
	enabled bool
}

// Queue 队列服务单例
var Queue = &sQueue{}

func init() {
	service.RegisterQueue(Queue)
}

// Init 初始化队列服务
func (s *sQueue) Init(ctx context.Context) error {
	s.enabled = g.Cfg().MustGet(ctx, "queue.switch", true).Bool()

	if !s.enabled {
		g.Log().Warning(ctx, "RocketMQ 队列服务已禁用")
		return nil
	}

	g.Log().Info(ctx, "RocketMQ 队列服务初始化成功")

	// 启动消费者
	go s.startConsumer(ctx)

	return nil
}

// PublishMessage 发布消息到队列
// 将消息发送任务提交到 RocketMQ
// 参数：
//   - ctx: 上下文对象
//   - msg: RocketMQ 消息体
//
// 返回：
//   - err: 错误信息
func (s *sQueue) PublishMessage(ctx context.Context, msg *dto.RocketMQMessage) error {
	if !s.enabled {
		return gerror.New("队列服务未启用")
	}

	// TODO: 实现 RocketMQ 生产者发送消息
	// 临时处理：直接同步处理消息（绕过队列）
	sendErr := s.processMessage(ctx, msg)
	if sendErr != nil {
		g.Log().Error(ctx, "处理消息失败", "error", sendErr, "uuid", msg.MessageUuid)
		return gerror.Wrap(sendErr, consts.ErrMsgRocketMQError)
	}

	g.Log().Info(ctx, "消息已同步处理",
		"uuid", msg.MessageUuid,
		"type", msg.MessageType,
	)

	return nil
}

// startConsumer 启动消费者
// 从队列中消费消息并处理
func (s *sQueue) startConsumer(ctx context.Context) {
	if !s.enabled {
		return
	}

	g.Log().Info(ctx, "启动 RocketMQ 消费者",
		"topic", consts.TopicMessageSend,
		"group", "ttpos-message",
	)

	// TODO: 实现 RocketMQ 消费者订阅
	g.Log().Warning(ctx, "RocketMQ 消费者未实现，消息将通过同步方式发送")
}

// handleMessage 处理消息
// 消费者回调函数，处理从队列中接收到的消息
func (s *sQueue) handleMessage(ctx context.Context, msgBody []byte) error {
	// 解析消息
	var msg dto.RocketMQMessage
	if err := json.Unmarshal(msgBody, &msg); err != nil {
		g.Log().Error(ctx, "解析队列消息失败", err)
		return err
	}

	return s.processMessage(ctx, &msg)
}

// processMessage 处理消息（统一的消息处理逻辑）
// 从数据库获取消息详情后进行处理
func (s *sQueue) processMessage(ctx context.Context, msg *dto.RocketMQMessage) error {
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
		sendErr = s.sendEmail(ctx, record)
	case consts.MessageTypeSMS:
		sendErr = s.sendSMS(ctx, record)
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

// sendEmail 发送邮件
func (s *sQueue) sendEmail(ctx context.Context, record *dto.MessageRecordDTO) error {
	return service.Mailgun().SendEmail(ctx, record.Uuid, record.Recipient, record.Subject, record.Content)
}

// sendSMS 发送短信（预留）
func (s *sQueue) sendSMS(ctx context.Context, record *dto.MessageRecordDTO) error {
	// TODO: 实现短信发送逻辑
	return gerror.New("短信发送功能暂未实现")
}

// IsEnabled 检查队列服务是否启用
func (s *sQueue) IsEnabled() bool {
	return s.enabled
}
