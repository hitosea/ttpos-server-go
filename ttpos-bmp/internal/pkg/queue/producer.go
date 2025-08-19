package queue

import (
	"github.com/gogf/gf/v2/util/gconv"
	"time"
)

// Push 推送队列
func Push(topic string, data interface{}) (err error) {
	q, err := InstanceProducer()
	if err != nil {
		return
	}
	mqMsg, err := q.SendMsg(topic, gconv.String(data))
	ProducerLog(ctx, topic, mqMsg, err)
	return
}

// DelayPush 推送延迟队列
// redis delay 传入 秒。如：10代表延迟10秒
// rocketmq delay 传入 延迟时间，单位秒。
func DelayPush(topic string, data interface{}, delay time.Duration) (err error) {
	q, err := InstanceProducer()
	if err != nil {
		return
	}
	mqMsg, err := q.SendDelayMsg(topic, gconv.String(data), delay)
	ProducerLog(ctx, topic, mqMsg, err)
	return
}
