// Package grab_store 提供 GrabFood 门店服务的业务逻辑
package grab_store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// sGrabStore 门店服务
type sGrabStore struct{}

func init() {
	service.RegisterGrabStore(New())
}

// New 创建门店服务实例
func New() *sGrabStore {
	return &sGrabStore{}
}

// HandleIntegrationStatus 处理门店集成状态回调
// 签名验证已由中间件完成，此处只处理业务逻辑
// 使用 SDK grabfood.PushIntegrationStatusWebhookRequest
func (s *sGrabStore) HandleIntegrationStatus(ctx context.Context, body []byte) error {
	// 1. 解析请求 - 使用 SDK Model
	var req grabfood.PushIntegrationStatusWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		g.Log().Errorf(ctx, "解析集成状态请求失败: %v", err)
		return fmt.Errorf("解析请求失败: %w", err)
	}

	// 2. 记录状态变更
	partnerMerchantID := req.GetPartnerMerchantID()
	integrationStatus := req.GetIntegrationStatus()
	grabMerchantID := req.GetGrabMerchantID()

	g.Log().Infof(ctx, "门店集成状态已变更: partner_merchant_id=%s, grab_merchant_id=%s, status=%s",
		partnerMerchantID, grabMerchantID, integrationStatus)

	// 3. 将 partnerMerchantID 转换为 shopUUID (uint64)
	shopUUID := g.NewVar(partnerMerchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Warningf(ctx, "[GrabStore] partnerMerchantID 格式无效，跳过更新: partnerMerchantID=%s", partnerMerchantID)
		return nil
	}

	// 4. 映射 Grab 状态到内部状态
	internalStatus := mapGrabIntegrationStatus(integrationStatus)

	// 5. 更新配置并发送通知
	if err := service.ShopProviderCfg().UpsertAndNotify(ctx, shopUUID, string(consts.ProviderGrab), grabMerchantID, internalStatus); err != nil {
		g.Log().Errorf(ctx, "[GrabStore] 更新门店第三方配置失败: shop_uuid=%d, status=%s, error: %v",
			shopUUID, internalStatus, err)
		// 记录错误但不中断流程，Webhook 应返回成功
	}

	return nil
}

// mapGrabIntegrationStatus 映射 Grab integrationStatus 到内部状态
func mapGrabIntegrationStatus(grabStatus string) consts.ProviderShopStatus {
	switch grabStatus {
	case "ACTIVE":
		return consts.ProviderShopStatusActive
	case "INACTIVE":
		return consts.ProviderShopStatusInactive
	case "SYNCING":
		return consts.ProviderShopStatusSyncing
	case "FAILED":
		return consts.ProviderShopStatusFailed
	default:
		// 未知状态默认为 INACTIVE
		return consts.ProviderShopStatusInactive
	}
}
