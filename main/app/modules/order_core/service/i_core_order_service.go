package service

import (
	"context"
	"ttpos-server-go/app/modules/order_core/dto"
)

// ICoreOrderService 核心订单服务接口
type ICoreOrderService interface {
	// 创建订单核心数据
	CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.CreateOrderResp, error)
	// 支付成功回调 (更新状态 + 发布事件)
	MarkAsPaid(ctx context.Context, billUuid uint64) error
	// 完成订单
	FinishOrder(ctx context.Context, billUuid uint64) error
	// 取消订单
	CancelOrder(ctx context.Context, billUuid uint64) error
}
