package lineman

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

func (c *ControllerV1) PlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) (res *v1.PlaceOrderRes, err error) {
	// 调用 Service 层处理订单
	err = service.Lineman().HandlePlaceOrder(ctx, req)
	if err != nil {
		// 返回失败响应
		return &v1.PlaceOrderRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Code:    "500",
				Message: err.Error(),
			},
		}, nil // 返回 nil error，让 GoFrame 返回 HTTP 200
	}

	// 返回成功响应
	return &v1.PlaceOrderRes{
		LinemanCommonResData: v1.LinemanCommonResData{
			Status:  "ok",
			Code:    "200",
			Message: "Order received successfully",
		},
	}, nil
}
