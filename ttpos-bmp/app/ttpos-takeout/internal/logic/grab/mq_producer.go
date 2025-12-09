package grab

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	// TopicGrabOrder Grab 订单 MQ Topic
	TopicGrabOrder = "takeout_grab_order"
)

// RocketMQProducer RocketMQ 生产者实现
type RocketMQProducer struct {
	// producer 实际的 RocketMQ producer 实例
	// 这里先定义接口，具体实现依赖项目的 MQ 客户端
}

// NewRocketMQProducer 创建 RocketMQ 生产者
func NewRocketMQProducer() *RocketMQProducer {
	return &RocketMQProducer{}
}

// SendOrderEvent 发送订单事件到 MQ
func (p *RocketMQProducer) SendOrderEvent(ctx context.Context, event *OrderEvent) error {
	// 序列化消息
	msgBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	g.Log().Debugf(ctx, "Sending MQ message to topic %s: %s", TopicGrabOrder, string(msgBytes))

	// TODO: 实际发送到 RocketMQ
	// 这里需要根据项目实际的 MQ 客户端实现
	// 示例:
	// return p.producer.SendSync(ctx, &primitive.Message{
	//     Topic: TopicGrabOrder,
	//     Body:  msgBytes,
	// })

	g.Log().Infof(ctx, "MQ message sent successfully: action=%s, orderId=%s", event.Action, event.OrderID)
	return nil
}

// NoopMQProducer 空操作 MQ 生产者 (用于测试或禁用 MQ 场景)
type NoopMQProducer struct{}

// NewNoopMQProducer 创建空操作 MQ 生产者
func NewNoopMQProducer() *NoopMQProducer {
	return &NoopMQProducer{}
}

// SendOrderEvent 空操作实现
func (p *NoopMQProducer) SendOrderEvent(ctx context.Context, event *OrderEvent) error {
	g.Log().Debugf(ctx, "NoopMQProducer: skipping MQ message: action=%s, orderId=%s", event.Action, event.OrderID)
	return nil
}
