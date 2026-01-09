// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	grabclient "ttpos-bmp/app/ttpos-takeout/internal/client/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// ============================================================================
// Webhook 处理 (包装层 - 处理一些额外逻辑)
// ============================================================================

// VerifyWebhookSignature 验证 Grab Webhook 签名 (公开方法，供其他服务调用)
// signature: X-Grab-Signature 请求头值
// timestamp: X-Grab-Timestamp 请求头值
// body: 请求体原始字节
func (s *sGrab) VerifyWebhookSignature(ctx context.Context, signature, timestamp string, body []byte) error {
	return grabclient.Default().GetVerifier().VerifySignature(signature, timestamp, body)
}

// HandleGetMenuWrapper 处理 Grab 获取菜单请求的包装方法
// 此方法在实际的 HandleGetMenu 基础上添加了 MerchantID 和 PartnerMerchantID 的设置逻辑
// NOTE: 为避免与 grab_menu.go 中的 HandleGetMenu 冲突，使用不同的方法名
func (s *sGrab) HandleGetMenuWrapper(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error) {
	// 1. 将 merchantID (partnerMerchantID) 转换为 shopUUID
	shopUUID := g.NewVar(merchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] partnerMerchantID 格式无效: %s", merchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "partnerMerchantID 格式无效")
	}

	// 2. 调用内部方法获取菜单数据（不包含 MerchantID 和 PartnerMerchantID）
	menuResp, err := s.HandleGetMenu(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	// 3. 查询 ShopProviderCfg
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取门店第三方配置失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取门店第三方配置失败")
	}
	if cfg == nil {
		g.Log().Errorf(ctx, "[Grab] 门店第三方配置不存在: shopUUID=%d, provider=%s", shopUUID, consts.ProviderGrab)
		return nil, gerror.NewCode(gcode.CodeNotFound, "门店第三方配置不存在")
	}

	// 4. 设置 MerchantID 和 PartnerMerchantID
	merchantIDStr := cfg.ProviderMerchantId
	partnerMerchantIDStr := fmt.Sprintf("%d", shopUUID)
	menuResp.MerchantID = &merchantIDStr
	menuResp.PartnerMerchantID = &partnerMerchantIDStr

	g.Log().Infof(ctx, "[Grab] 获取菜单成功: merchantID=%v, partnerMerchantID=%v, categories=%d",
		menuResp.MerchantID, menuResp.PartnerMerchantID, len(menuResp.Categories))

	return menuResp, nil
}

// HandlePushGrabMenu 处理 Grab 菜单推送 Webhook
// 此方法组合了 SaveMenuSnapshot 和 NotifyMenuUpdate 两个操作
func (s *sGrab) HandlePushGrabMenu(ctx context.Context, dto *grabDto.PushGrabMenuDTO) error {
	// 1. Save Snapshot
	menuUuid, err := s.SaveMenuSnapshot(ctx, dto)
	if err != nil {
		return err
	}

	// 2. Notify
	event := &grabDto.ProviderMenuUpdateEvent{
		ProviderName:      string(consts.ProviderGrab),
		MerchantID:        dto.MerchantID,
		PartnerMerchantID: dto.PartnerMerchantID,
		ShopUuid:          dto.PartnerMerchantID,
		Uuid:              menuUuid,
		ReceivedAt:        gtime.Timestamp(),
	}
	return s.NotifyMenuUpdateEvent(ctx, event)
}
