// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/order"
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
		// CheckOrderCancelable 检查订单是否可取消
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体
		//
		// 返回：
		//   - res: 检查订单可取消性响应
		//   - err: 错误信息
		CheckOrderCancelable(ctx context.Context, orderEntity *entity.Order) (*api.CheckOrderCancelableResp, error)
		// CancelOrder 取消订单
		// 参数：
		//   - ctx: 上下文对象
		//   - orderEntity: 订单实体
		//   - cancelCode: 取消原因码（字符串格式，可根据不同平台传入不同的编码）
		//
		// 返回：
		//   - res: 取消订单响应
		//   - err: 错误信息
		CancelOrder(ctx context.Context, orderEntity *entity.Order, cancelCode string) (res *api.CancelOrderResp, err error)
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
