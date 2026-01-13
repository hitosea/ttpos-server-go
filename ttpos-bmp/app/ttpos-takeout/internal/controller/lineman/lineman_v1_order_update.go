package lineman

import (
	"context"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// OrderUpdate 接收 LINE MAN 订单更新通知
// PUT /v1/partners/{partnerId}/stores/{storeId}/orders
func (c *ControllerV1) OrderUpdate(ctx context.Context, req *v1.OrderUpdateReq) (res *v1.OrderUpdateRes, err error) {
	// 调用 Service 层处理订单更新
	err = service.LinemanOrder().HandleOrderUpdate(ctx, req)
	if err != nil {
		// 返回失败响应
		return &v1.OrderUpdateRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Message: err.Error(),
			},
		}, nil
	}

	// 返回成功响应
	return &v1.OrderUpdateRes{
		LinemanCommonResData: v1.LinemanCommonResData{
			Status:  "ok",
			Message: "订单更新成功",
		},
	}, nil
}
