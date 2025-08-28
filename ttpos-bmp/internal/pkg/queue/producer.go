package queue

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gctx"
)

// Push 推送队列消息
func Push(topic string, data interface{}) error {
	return PushWithContext(gctx.GetInitCtx(), topic, data)
}

// PushWithContext 使用指定Context推送队列消息
func PushWithContext(ctx context.Context, topic string, data interface{}) error {
	// 参数验证
	if topic == "" {
		return gerror.New("主题名称不能为空")
	}
	if data == nil {
		return gerror.New("消息内容不能为空")
	}

	// 初始化生产者
	producer, err := InstanceProducer()
	if err != nil {
		return gerror.Wrap(err, "初始化消息生产者失败")
	}

	// 序列化消息内容
	body, err := gjson.EncodeString(data)
	if err != nil {
		return gerror.Wrap(err, "消息体序列化失败")
	}

	// 发送消息
	mqMsg, err := producer.SendMsg(ctx, topic, body)
	// 记录日志
	ProducerLog(ctx, topic, mqMsg, err)
	return err
}

// DelayPush 推送延时队列消息
func DelayPush(topic string, data interface{}, delay time.Duration) error {
	return DelayPushWithContext(gctx.GetInitCtx(), topic, data, delay)
}

// DelayPushWithContext 使用指定Context推送延时队列消息
// delay: 延时时间，支持毫秒级精度
func DelayPushWithContext(ctx context.Context, topic string, data interface{}, delay time.Duration) error {
	// 参数验证
	if topic == "" {
		return gerror.New("主题名称不能为空")
	}
	if data == nil {
		return gerror.New("消息内容不能为空")
	}
	if delay < 0 {
		return gerror.New("延时时间不能为负数")
	}

	// 初始化生产者
	producer, err := InstanceProducer()
	if err != nil {
		return gerror.Wrap(err, "初始化消息生产者失败")
	}

	// 序列化消息内容
	body, err := gjson.EncodeString(data)
	if err != nil {
		return gerror.Wrap(err, "消息体序列化失败")
	}

	// 发送延时消息
	mqMsg, err := producer.SendDelayMsg(ctx, topic, body, delay)
	// 记录日志
	ProducerLog(ctx, topic, mqMsg, err)
	return err
}
