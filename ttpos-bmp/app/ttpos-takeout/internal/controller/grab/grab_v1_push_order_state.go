package grab

import (
	"context"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// PushOrderState 处理 Grab 订单状态变更 Webhook
// GrabFood 在订单状态变更时调用此端点推送状态通知
func (c *ControllerV1) PushOrderState(ctx context.Context, req *v1.PushOrderStateReq) (res *v1.PushOrderStateRes, err error) {
	// 调用 Service 层处理订单状态变更
	// 签名验证已由中间件完成
	err = service.Grab().HandlePushOrderState(ctx, req.OrderStateRequest)
	if err != nil {
		return nil, err
	}

	// Webhook 成功处理，返回空的响应体
	return &v1.PushOrderStateRes{}, nil
}
