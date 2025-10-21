// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ================================================================================

package service

import (
	"context"

	"ttpos-bmp/app/ttpos-message/internal/model/dto"
)

// IQueue 队列服务接口
type IQueue interface {
	// Init 初始化服务
	Init(ctx context.Context) error
	// PublishMessage 发布消息到队列
	PublishMessage(ctx context.Context, msg *dto.RocketMQMessage) error
	// IsEnabled 检查队列服务是否启用
	IsEnabled() bool
}

var localQueue IQueue

// Queue 获取队列服务实例
func Queue() IQueue {
	if localQueue == nil {
		panic("implement not found for interface IQueue, forgot register?")
	}
	return localQueue
}

// RegisterQueue 注册队列服务实现
func RegisterQueue(i IQueue) {
	localQueue = i
}
