package queue

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// 日志格式常量
const (
	// 消费者日志格式
	ConsumerLogErrFormat     = "消费失败 [%s] - 消息ID: %s, 错误: %+v"
	ConsumerLogSuccessFormat = "消费成功 [%s] - 消息ID: %s, 耗时: %v"
	// 生产者日志格式
	ProducerLogErrFormat     = "生产失败 [%s] - 消息长度: %d, 错误: %+v"
	ProducerLogSuccessFormat = "生产成功 [%s] - 消息ID: %s, 耗时: %v"
)

// Logger 获取队列专用日志记录器
func Logger() *glog.Logger {
	return g.Log("queue")
}

// ConsumerLog 消费者日志记录
func ConsumerLog(ctx context.Context, topic string, mqMsg MqMsg, err error) {
	ConsumerLogWithDuration(ctx, topic, mqMsg, err, 0)
}

// ConsumerLogWithDuration 带执行耗时的消费者日志记录
func ConsumerLogWithDuration(ctx context.Context, topic string, mqMsg MqMsg, err error, duration time.Duration) {
	if err != nil {
		// 记录错误日志
		Logger().Errorf(ctx, ConsumerLogErrFormat, topic, mqMsg.MsgId, err)
		// 记录详细信息供调试
		Logger().Debugf(ctx, "消费错误详情 [%s] - 消息内容: %s, 耗时: %v",
			topic, mqMsg.BodyString(), duration)
	} else {
		// 记录成功日志（Debug级别）
		Logger().Debugf(ctx, ConsumerLogSuccessFormat, topic, mqMsg.MsgId, duration)
	}

	// 记录性能监控信息
	if duration > 5*time.Second {
		Logger().Warningf(ctx, "消费者处理耗时过长 [%s] - 消息ID: %s, 耗时: %v",
			topic, mqMsg.MsgId, duration)
	}
}

// ProducerLog 生产者日志记录
func ProducerLog(ctx context.Context, topic string, mqMsg MqMsg, err error) {
	ProducerLogWithDuration(ctx, topic, mqMsg, err, 0)
}

// ProducerLogWithDuration 带执行耗时的生产者日志记录
func ProducerLogWithDuration(ctx context.Context, topic string, mqMsg MqMsg, err error, duration time.Duration) {
	if err != nil {
		// 记录错误日志
		Logger().Errorf(ctx, ProducerLogErrFormat, topic, len(mqMsg.Body), err)
		// 记录详细信息供调试
		Logger().Debugf(ctx, "生产错误详情 [%s] - 消息内容: %s, 耗时: %v",
			topic, mqMsg.BodyString(), duration)
	} else {
		// 记录成功日志（Debug级别）
		Logger().Debugf(ctx, ProducerLogSuccessFormat, topic, mqMsg.MsgId, duration)
	}

	// 记录性能监控信息
	if duration > 1*time.Second {
		Logger().Warningf(ctx, "生产者发送耗时过长 [%s] - 消息ID: %s, 耗时: %v",
			topic, mqMsg.MsgId, duration)
	}
}

// LogQueueMetrics 记录队列指标信息
func LogQueueMetrics(ctx context.Context, topic string, messageCount int64, avgProcessTime time.Duration) {
	Logger().Infof(ctx, "队列指标 [%s] - 消息数量: %d, 平均处理时间: %v",
		topic, messageCount, avgProcessTime)
}

// LogQueueError 记录队列系统错误
func LogQueueError(ctx context.Context, operation string, err error) {
	Logger().Errorf(ctx, "队列系统错误 - 操作: %s, 错误: %+v", operation, err)
}
