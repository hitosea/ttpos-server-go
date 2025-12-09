package grab

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
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

	// TODO: 实现菜单存储逻辑
	// 1. 验证商户映射关系
	// 2. 将 Grab 菜单结构转换为 POS 菜单结构
	// 3. 存储菜单数据供后续菜单同步使用

	// 返回 204 No Content (GoFrame 会自动处理空响应)
	return &v1.PushGrabMenuWebhookRes{}, nil
}
