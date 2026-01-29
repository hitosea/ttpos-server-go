package lineman

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

func (c *ControllerV1) OrderStatusUpdate(ctx context.Context, req *v1.OrderStatusUpdateReq) (res *v1.OrderStatusUpdateRes, err error) {
	// 调用 Service 层处理订单状态更新
	err = service.Lineman().HandleOrderStatusUpdate(ctx, req)
	if err != nil {
		// 返回失败响应
		return &v1.OrderStatusUpdateRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Code:    "500",
				Message: err.Error(),
			},
		}, nil // 返回 nil error，让 GoFrame 返回 HTTP 200
	}

	// 返回成功响应
	return &v1.OrderStatusUpdateRes{
		LinemanCommonResData: v1.LinemanCommonResData{
			Status:  "ok",
			Code:    "200",
			Message: "Order status updated successfully",
		},
	}, nil
}
