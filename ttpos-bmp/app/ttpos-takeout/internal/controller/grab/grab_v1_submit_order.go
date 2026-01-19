package grab

import (
	"context"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

func (c *ControllerV1) SubmitOrder(ctx context.Context, req *v1.SubmitOrderReq) (res *v1.SubmitOrderRes, err error) {
	// 直接使用类型化的请求对象
	// GoFrame 已自动解析 JSON 到 req.SubmitOrderRequest

	// 调用 Service 层处理 Grab 提交订单 Webhook
	// 签名验证已由中间件完成
	err = service.Grab().HandleSubmitOrder(ctx, req.SubmitOrderRequest)
	if err != nil {
		return nil, err
	}

	// Webhook 成功处理，返回空的响应体
	return &v1.SubmitOrderRes{}, nil
}
