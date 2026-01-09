// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"
	"encoding/json"
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

// HandleGetMenu 处理 Grab 获取菜单请求 (Partner Endpoint)
// 签名验证已由中间件完成
func (s *sGrab) HandleGetMenu(ctx context.Context, merchantID string) (*grabfood.GetMenuNewResponse, error) {
	g.Log().Infof(ctx, "[Grab] 收到获取菜单请求: partnerMerchantID=%s", merchantID)

	// 1. 将 partnerMerchantID 转换为 shopUUID (uint64)
	// 假设 partnerMerchantID 是数字字符串格式的 shopUUID
	shopUUID := g.NewVar(merchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Errorf(ctx, "[Grab] partnerMerchantID 格式无效: %s", merchantID)
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "partnerMerchantID 格式无效")
	}

	// 2. 从本地快照读取菜单
	menuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取渠道菜单失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取渠道菜单失败")
	}

	// 3. 如果本地快照为空，返回错误
	if menuJSON == "" {
		g.Log().Errorf(ctx, "[Grab] 本地菜单快照不存在: shopUUID=%d", shopUUID)
		return nil, gerror.NewCode(gcode.CodeNotFound, "菜单不存在")
	}

	// 4. 解析本地快照 JSON 为 PushGrabMenuDTO (使用 SDK 类型)
	var pushDTO grabDto.PushGrabMenuDTO
	if err := json.Unmarshal([]byte(menuJSON), &pushDTO); err != nil {
		g.Log().Errorf(ctx, "[Grab] 解析菜单 JSON 失败: error: %v", err)
		return nil, gerror.Wrap(err, "解析菜单数据失败")
	}

	// 5. 构建响应结构
	menuResp := &grabfood.GetMenuNewResponse{
		Currency:     pushDTO.Currency,
		SellingTimes: pushDTO.SellingTimes,
		Categories:   pushDTO.Categories,
	}

	g.Log().Infof(ctx, "[Grab] 获取菜单成功（来自本地）: categories=%d", len(menuResp.Categories))

	// 6. 查询 ShopProviderCfg 获取 Grab MerchantID
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderGrab))
	if err != nil {
		g.Log().Errorf(ctx, "[Grab] 获取门店第三方配置失败: shopUUID=%d, error: %v", shopUUID, err)
		return nil, gerror.Wrap(err, "获取门店第三方配置失败")
	}
	if cfg == nil {
		g.Log().Errorf(ctx, "[Grab] 门店第三方配置不存在: shopUUID=%d, provider=%s", shopUUID, consts.ProviderGrab)
		return nil, gerror.NewCode(gcode.CodeNotFound, "门店第三方配置不存在")
	}

	// 7. 设置 MerchantID 和 PartnerMerchantID
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
