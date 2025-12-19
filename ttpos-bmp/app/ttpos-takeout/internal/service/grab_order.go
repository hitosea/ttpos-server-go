// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

type (
	IGrabOrder interface {
		// HandleSubmitOrder 处理 Grab 提交订单 Webhook
		// 签名验证已由中间件完成，此处只处理业务逻辑
		// 使用 SDK grabfood.SubmitOrderRequest 替换自定义 DTO
		HandleSubmitOrder(ctx context.Context, req *grabfood.SubmitOrderRequest) error
		// HandlePushOrderState 处理订单状态变更 Webhook
		// 签名验证已由中间件完成，此处只处理业务逻辑
		// 使用 SDK grabfood.OrderStateRequest 替换自定义 DTO
		HandlePushOrderState(ctx context.Context, body []byte) error
	}
)

var (
	localGrabOrder IGrabOrder
)

func GrabOrder() IGrabOrder {
	if localGrabOrder == nil {
		panic("implement not found for interface IGrabOrder, forgot register?")
	}
	return localGrabOrder
}

func RegisterGrabOrder(i IGrabOrder) {
	localGrabOrder = i
}
