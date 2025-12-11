// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
)

type (
	IGrabStore interface {
		// HandleIntegrationStatus 处理门店集成状态回调
		// 签名验证已由中间件完成，此处只处理业务逻辑
		// 使用 SDK grabfood.PushIntegrationStatusWebhookRequest
		HandleIntegrationStatus(ctx context.Context, body []byte) error
	}
)

var (
	localGrabStore IGrabStore
)

func GrabStore() IGrabStore {
	if localGrabStore == nil {
		panic("implement not found for interface IGrabStore, forgot register?")
	}
	return localGrabStore
}

func RegisterGrabStore(i IGrabStore) {
	localGrabStore = i
}
