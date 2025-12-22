// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/order"
)

type (
	IOrder interface {
		// GetOrderInfo 获取订单信息
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 获取订单信息请求，包含 shop_uuid 和 order_uuid
		// 返回：
		//   - res: 订单信息响应
		//   - err: 错误信息
		GetOrderInfo(ctx context.Context, req *api.GetOrderInfoReq) (res *api.GetOrderInfoResp, err error)
	}
)

var (
	localOrder IOrder
)

func Order() IOrder {
	if localOrder == nil {
		panic("implement not found for interface IOrder, forgot register?")
	}
	return localOrder
}

func RegisterOrder(i IOrder) {
	localOrder = i
}
