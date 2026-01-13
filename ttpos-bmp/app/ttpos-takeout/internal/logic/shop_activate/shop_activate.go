// Package shop_activate 提供门店多渠道激活服务的业务逻辑
package shop_activate

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"ttpos-bmp/app/ttpos-takeout/api/shop"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// sShopActivate 门店激活服务
type sShopActivate struct{}

func init() {
	service.RegisterShopActivate(New())
}

// New 创建门店激活服务实例
func New() *sShopActivate {
	return &sShopActivate{}
}

// ActivateShop 激活门店外卖渠道（多渠道路由）
// 参数：
//   - ctx: 上下文对象
//   - req: 激活请求参数，包含 provider_name、shop_uuid、request_id
//
// 返回：
//   - res: 激活结果，包含 shop_uuid、provider_name、updated_at，Grab 渠道额外返回 self_serve_url
//   - err: 操作过程中产生的错误（若有）
func (s *sShopActivate) ActivateShop(ctx context.Context, req *shop.ActivateShopReq) (*shop.ActivateShopResp, error) {
	// 1. 参数校验
	if req.ProviderName == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "provider_name 不能为空")
	}
	if req.ShopUuid == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "shop_uuid 不能为空")
	}

	shopUUID := gconv.Uint64(req.ShopUuid)
	if shopUUID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "shop_uuid 格式无效")
	}

	// 2. 渠道路由
	switch req.ProviderName {
	case string(consts.ProviderLineman):
		return s.activateLineman(ctx, shopUUID, req.RequestId)
	case string(consts.ProviderGrab):
		return s.activateGrab(ctx, req)
	default:
		return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "不支持的渠道: %s", req.ProviderName)
	}
}

// activateLineman LINE MAN 渠道激活（直接创建配置）
// 参数：
//   - ctx: 上下文对象
//   - shopUUID: 门店 UUID
//   - requestID: 追踪 ID（可选）
//
// 返回：
//   - res: 激活结果
//   - err: 操作过程中产生的错误（若有）
func (s *sShopActivate) activateLineman(ctx context.Context, shopUUID uint64, requestID string) (*shop.ActivateShopResp, error) {
	// 创建配置记录（状态 ACTIVE） 直接激活
	err := service.ShopProviderCfg().UpsertShopProviderCfg(ctx, shopUUID, string(consts.ProviderLineman), "", consts.ProviderShopStatusActive)
	if err != nil {
		g.Log().Errorf(ctx, "[ShopActivate] LINE MAN 渠道激活失败: shop_uuid=%d, error=%v", shopUUID, err)
		return nil, gerror.Wrap(err, "LINE MAN 渠道激活失败")
	}

	g.Log().Infof(ctx, "[ShopActivate] LINE MAN 渠道激活成功: shop_uuid=%d", shopUUID)

	// 查询配置返回
	cfg, err := service.ShopProviderCfg().GetShopProviderCfg(ctx, shopUUID, string(consts.ProviderLineman))
	if err != nil {
		return nil, gerror.Wrap(err, "查询 LINE MAN 配置失败")
	}

	// 检查配置是否存在
	if cfg == nil {
		g.Log().Errorf(ctx, "[ShopActivate] LINE MAN 配置不存在: shop_uuid=%d", shopUUID)
		return nil, gerror.NewCode(gcode.CodeNotFound, "LINE MAN 配置不存在")
	}

	return &shop.ActivateShopResp{
		ShopUuid:     cfg.ShopUuid,
		ProviderName: cfg.ProviderName,
		RequestId:    requestID,
		UpdatedAt:    int64(cfg.UpdatedAt),
	}, nil
}

// activateGrab Grab 渠道激活（调用自助激活服务）
// 参数：
//   - ctx: 上下文对象
//   - req: 激活请求参数
//
// 返回：
//   - res: 激活结果，包含 self_serve_url
//   - err: 操作过程中产生的错误（若有）
func (s *sShopActivate) activateGrab(ctx context.Context, req *shop.ActivateShopReq) (*shop.ActivateShopResp, error) {
	// shop_uuid 即为 Grab MerchantID
	merchantID := req.ShopUuid

	// 调用 Grab 自助激活服务
	activationURL, requestID, err := service.Grab().CreateSelfServeJourney(ctx, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "[ShopActivate] Grab 渠道激活失败: merchant_id=%s, error=%v", merchantID, err)
		return nil, gerror.Wrap(err, "Grab 渠道激活失败")
	}

	g.Log().Infof(ctx, "[ShopActivate] Grab 渠道激活成功: merchant_id=%s, self_serve_url=%s", merchantID, activationURL)

	return &shop.ActivateShopResp{
		ShopUuid:     gconv.Uint64(req.ShopUuid),
		ProviderName: req.ProviderName,
		SelfServeUrl: activationURL,
		RequestId:    requestID,
		UpdatedAt:    0, // 激活时尚未创建配置记录
	}, nil
}
