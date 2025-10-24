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
	IQueue interface {
		// Init 初始化队列服务
		Init(ctx context.Context) error
		// PublishMessage 发布消息到队列
		// 将消息发送任务提交到 RocketMQ
		// 参数：
		//   - ctx: 上下文对象
		//   - msg: RocketMQ 消息体
		//
		// 返回：
		//   - err: 错误信息
		PublishMessage(ctx context.Context, msg *dto.RocketMQMessage) error
		// IsEnabled 检查队列服务是否启用
		IsEnabled() bool
	}
)

var (
	localQueue IQueue
)

func Queue() IQueue {
	if localQueue == nil {
		panic("implement not found for interface IQueue, forgot register?")
	}
	return localQueue
}

func RegisterQueue(i IQueue) {
	localQueue = i
}
