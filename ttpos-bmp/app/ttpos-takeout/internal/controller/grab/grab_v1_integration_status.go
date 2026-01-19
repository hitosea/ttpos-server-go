package grab

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// IntegrationStatus 处理 Grab 门店集成状态回调
// GrabFood 在门店集成状态变更时调用此端点推送状态通知
// 参考: https://developer.grab.com/docs/grabfood/api/#tag/push-integration-status-webhook
func (c *ControllerV1) IntegrationStatus(ctx context.Context, req *v1.IntegrationStatusReq) (res *v1.IntegrationStatusRes, err error) {
	if err := service.ShopProviderCfg().HandleIntegrationStatus(ctx, &req.PushIntegrationStatusWebhookRequest); err != nil {
		g.Log().Errorf(ctx, "[Grab] HandleIntegrationStatus failed: %v", err)
	}
	return &v1.IntegrationStatusRes{}, nil
}
