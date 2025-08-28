package queue

import (
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Push 推送队列
func Push(topic string, data interface{}) error {
	q, err := InstanceProducer()
	if err != nil {
		return gerror.Wrap(err, "初始化消息生产者失败")
	}
	body, err := gjson.EncodeString(data)
	if err != nil {
		return gerror.Wrap(err, "消息体序列化失败")
	}
	mqMsg, err := q.SendMsg(topic, body)
	ProducerLog(ctx, topic, mqMsg, err)
	return nil
}

// DelayPush 推送延迟队列
// redis delay 传入 秒。如：10代表延迟10秒
// rocketmq delay 传入 延迟时间，单位秒。
func DelayPush(topic string, data interface{}, delay time.Duration) (err error) {
	q, err := InstanceProducer()
	if err != nil {
		return
	}
	body, err := gjson.EncodeString(data)
	if err != nil {
		return gerror.Wrap(err, "消息体序列化失败")
	}
	mqMsg, err := q.SendDelayMsg(topic, body, delay)
	ProducerLog(ctx, topic, mqMsg, err)
	return
}
