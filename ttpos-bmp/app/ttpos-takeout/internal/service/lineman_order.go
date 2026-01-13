// Package service 提供服务接口定义
package service

import (
	"context"
	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
)

// ILinemanOrder LINE MAN 订单处理服务接口
type ILinemanOrder interface {
	// HandlePlaceOrder 处理 LINE MAN 订单创建 Webhook
	HandlePlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) error
	// HandleOrderUpdate 处理 LINE MAN 订单更新 Webhook
	HandleOrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) error
}

var localLinemanOrder ILinemanOrder

// LinemanOrder 获取 LINE MAN 订单服务实例
func LinemanOrder() ILinemanOrder {
	if localLinemanOrder == nil {
		panic("implement not found for interface ILinemanOrder, forgot register?")
	}
	return localLinemanOrder
}

// RegisterLinemanOrder 注册 LINE MAN 订单服务实现
func RegisterLinemanOrder(i ILinemanOrder) {
	localLinemanOrder = i
}
