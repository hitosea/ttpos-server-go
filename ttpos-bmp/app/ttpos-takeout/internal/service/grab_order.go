// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"

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
		// 使用 SDK grabfood.OrderStateRequest
		HandlePushOrderState(ctx context.Context, req *grabfood.OrderStateRequest) error
		// PrepareOrder 准备订单（接受/拒绝）
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体
		//   - toState: 目标状态 (Accepted/Rejected)
		//
		// 返回：
		//   - err: 错误信息
		PrepareOrder(ctx context.Context, orderEntityInterface interface{}, toState string) error
		// MarkOrderReady 标记订单准备完成
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体，包含 ProviderOrderId 等信息
		//
		// 返回：
		//   - err: 错误信息
		MarkOrderReady(ctx context.Context, orderEntity *entity.Order) error
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
