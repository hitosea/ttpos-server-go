package grab

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// PushGrabMenuWebhook 处理 Grab 菜单推送 webhook
// GrabFood 在 Self-Serve Activation 流程中调用此端点推送现有门店菜单
// 参考: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/push-grab-menu-webhook
func (c *ControllerV1) PushGrabMenuWebhook(ctx context.Context, req *v1.PushGrabMenuWebhookReq) (res *v1.PushGrabMenuWebhookRes, err error) {
	g.Log().Infof(ctx, "[Grab] Received PushGrabMenuWebhook: merchantID=%s, partnerMerchantID=%s",
		req.MerchantID, req.PartnerMerchantID)

	// 记录菜单统计信息
	g.Log().Infof(ctx, "[Grab] Menu stats: sellingTimes=%d, categories=%d",
		len(req.SellingTimes), len(req.Categories))

	// 构造 DTO
	dto := &grabDto.PushGrabMenuDTO{
		MerchantID:        req.MerchantID,
		PartnerMerchantID: req.PartnerMerchantID,
		Currency:          req.Currency,
		SellingTimes:      req.SellingTimes,
		Categories:        req.Categories,
	}

	// 调用服务处理
	if err := service.Grab().HandlePushGrabMenu(ctx, dto); err != nil {
		g.Log().Errorf(ctx, "[Grab] HandlePushGrabMenu failed: %v", err)
		// 注意: Grab 可能会重试，如果内部错误 (如 Redis/MQ) 应该返回 500
		return nil, err
	}

	// 返回 204 No Content (GoFrame 会自动处理空响应)
	return &v1.PushGrabMenuWebhookRes{}, nil
}
