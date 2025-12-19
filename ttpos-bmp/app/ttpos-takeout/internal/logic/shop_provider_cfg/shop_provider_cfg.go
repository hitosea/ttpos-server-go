// Package shop_provider_cfg 提供门店第三方配置服务的业务逻辑
package shop_provider_cfg

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	"ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"

	"ttpos-bmp/app/ttpos-takeout/internal/service"
	"ttpos-bmp/internal/pkg/queue"
	"ttpos-bmp/utility/uuid"
)

const (
	// TopicStoreIntegrationState 门店集成状态变更 Topic
	TopicStoreIntegrationState = "takeout_store_integration_state"
)

var (
	ShopProviderCfg = new(sShopProviderCfg)
)

// sShopProviderCfg 门店第三方配置服务
type sShopProviderCfg struct{}

func init() {
	service.RegisterShopProviderCfg(ShopProviderCfg)
}

// UpsertShopProviderCfg 更新或插入门店第三方配置（幂等）
// shopUUID: 门店 UUID
// providerName: 第三方名称（如 grab）
// merchantID: 第三方商户 ID
// status: 集成状态
func (s *sShopProviderCfg) UpsertShopProviderCfg(ctx context.Context, shopUUID uint64, providerName string, merchantID string, status consts.ProviderShopStatus) error {
	now := gtime.Timestamp()

	// 先查询是否存在
	existing, err := s.GetShopProviderCfg(ctx, shopUUID, providerName)
	if err != nil {
		return err
	}

	if existing != nil {
		// 更新（不更新 uuid 和 created_at）
		_, err = dao.ShopProviderCfg.Ctx(ctx).
			Where(dao.ShopProviderCfg.Columns().Uuid, existing.Uuid).
			Data(g.Map{
				dao.ShopProviderCfg.Columns().ProviderMerchantId: merchantID,
				dao.ShopProviderCfg.Columns().ProviderShopStatus: string(status),
				dao.ShopProviderCfg.Columns().UpdatedAt:          now,
			}).Update()
	} else {
		// 插入新记录
		_, err = dao.ShopProviderCfg.Ctx(ctx).Data(do.ShopProviderCfg{
			Uuid:               uuid.MustGetID(),
			ShopUuid:           shopUUID,
			ProviderName:       providerName,
			ProviderMerchantId: merchantID,
			ProviderShopStatus: string(status),
			CreatedAt:          now,
			UpdatedAt:          now,
		}).Insert()
	}

	if err != nil {
		g.Log().Errorf(ctx, "[ShopProviderCfg] 更新失败: shop_uuid=%d, provider=%s, status=%s, error: %v",
			shopUUID, providerName, status, err)
		return gerror.Wrap(err, "更新门店第三方配置失败")
	}

	g.Log().Infof(ctx, "[ShopProviderCfg] 更新成功: shop_uuid=%d, provider=%s, merchant_id=%s, status=%s",
		shopUUID, providerName, merchantID, status)
	return nil
}

// GetShopProviderCfg 查询门店第三方配置
// shopUUID: 门店 UUID
// providerName: 第三方名称（如 grab），为空默认 grab
func (s *sShopProviderCfg) GetShopProviderCfg(ctx context.Context, shopUUID uint64, providerName string) (*entity.ShopProviderCfg, error) {
	if providerName == "" {
		providerName = string(consts.ProviderGrab)
	}

	var cfg entity.ShopProviderCfg
	err := dao.ShopProviderCfg.Ctx(ctx).
		Where(dao.ShopProviderCfg.Columns().ShopUuid, shopUUID).
		Where(dao.ShopProviderCfg.Columns().ProviderName, providerName).
		Where(dao.ShopProviderCfg.Columns().DeletedAt, 0).
		Scan(&cfg)

	if err != nil {
		g.Log().Errorf(ctx, "[ShopProviderCfg] 查询失败: shop_uuid=%d, provider=%s, error: %v",
			shopUUID, providerName, err)
		return nil, nil
	}

	// 检查是否找到记录
	if cfg.Id == 0 {
		g.Log().Debugf(ctx, "[ShopProviderCfg] 未找到: shop_uuid=%d, provider=%s", shopUUID, providerName)
		return nil, nil
	}

	return &cfg, nil
}

// GetShopProviderCfgByMerchantID 通过 MerchantID 查询门店第三方配置
// merchantID: 第三方商户 ID（如 Grab MerchantID）
// providerName: 第三方名称（如 grab），为空默认 grab
func (s *sShopProviderCfg) GetShopProviderCfgByMerchantID(ctx context.Context, merchantID string, providerName string) (*entity.ShopProviderCfg, error) {
	if providerName == "" {
		providerName = string(consts.ProviderGrab)
	}

	var cfg entity.ShopProviderCfg
	err := dao.ShopProviderCfg.Ctx(ctx).
		Where(dao.ShopProviderCfg.Columns().ProviderMerchantId, merchantID).
		Where(dao.ShopProviderCfg.Columns().ProviderName, providerName).
		Where(dao.ShopProviderCfg.Columns().DeletedAt, 0).
		Scan(&cfg)

	if err != nil {
		g.Log().Errorf(ctx, "[ShopProviderCfg] 通过 MerchantID 查询失败: merchant_id=%s, provider=%s, error: %v",
			merchantID, providerName, err)
		return nil, nil
	}

	if cfg.Id == 0 {
		g.Log().Debugf(ctx, "[ShopProviderCfg] 通过 MerchantID 未找到: merchant_id=%s, provider=%s", merchantID, providerName)
		return nil, nil
	}

	return &cfg, nil
}

// NotifyStoreIntegrationState 发送门店集成状态变更通知 (RocketMQ)
func (s *sShopProviderCfg) NotifyStoreIntegrationState(ctx context.Context, event *grabDto.ShopIntegrationStatusEvent) error {
	if err := queue.PushWithContext(ctx, TopicStoreIntegrationState, event); err != nil {
		g.Log().Errorf(ctx, "[ShopProviderCfg] 发送集成状态事件失败: %v", err)
		return gerror.Wrap(err, "发送门店集成状态变更通知失败")
	}

	g.Log().Infof(ctx, "[ShopProviderCfg] 集成状态事件已发送: topic=%s, shop_uuid=%d, status=%s",
		TopicStoreIntegrationState, event.ShopUuid, event.ProviderShopStatus)
	return nil
}

// UpsertAndNotify 更新配置并发送通知（组合方法）
func (s *sShopProviderCfg) UpsertAndNotify(ctx context.Context, shopUUID uint64, providerName string, merchantID string, status consts.ProviderShopStatus) error {
	// 1. Upsert 配置
	if err := s.UpsertShopProviderCfg(ctx, shopUUID, providerName, merchantID, status); err != nil {
		return err
	}

	// 2. 发送通知
	event := &grabDto.ShopIntegrationStatusEvent{
		ShopUuid:           shopUUID,
		ProviderName:       providerName,
		ProviderShopStatus: string(status),
		ProviderMerchantId: merchantID,
		UpdatedAt:          gtime.Timestamp(),
	}
	return s.NotifyStoreIntegrationState(ctx, event)
}

// GetShopProviderCfgResp 查询门店第三方配置（gRPC 响应格式）
func (s *sShopProviderCfg) GetShopProviderCfgResp(ctx context.Context, req *grab.GetShopProviderCfgReq) (*grab.GetShopProviderCfgResp, error) {
	// 参数校验
	if req.ShopUuid == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "shop_uuid 不能为空")
	}

	// 默认 provider_name 为 grab
	providerName := req.ProviderName
	if providerName == "" {
		providerName = string(consts.ProviderGrab)
	}

	// 查询配置
	cfg, err := s.GetShopProviderCfg(ctx, req.ShopUuid, providerName)
	if err != nil {
		g.Log().Errorf(ctx, "[ShopProviderCfg] 获取门店配置响应失败: shop_uuid=%d, provider=%s, error=%v",
			req.ShopUuid, providerName, err)
		return nil, gerror.Wrap(err, "查询门店配置失败")
	}

	// 未找到记录
	if cfg == nil {
		g.Log().Debugf(ctx, "[ShopProviderCfg] 获取门店配置响应未找到: shop_uuid=%d, provider=%s",
			req.ShopUuid, providerName)
		//没有就返回 未绑定
		return &grab.GetShopProviderCfgResp{
			ShopUuid:           req.ShopUuid,
			ProviderName:       req.ProviderName,
			ProviderShopStatus: "INACTIVE",
		}, nil
	}

	// 构建响应
	resp := &grab.GetShopProviderCfgResp{
		ShopUuid:           cfg.ShopUuid,
		ProviderName:       cfg.ProviderName,
		ProviderMerchantId: cfg.ProviderMerchantId,
		ProviderShopStatus: cfg.ProviderShopStatus,
		UpdatedAt:          int64(cfg.UpdatedAt),
	}

	g.Log().Debugf(ctx, "[ShopProviderCfg] 获取门店配置响应成功: shop_uuid=%d, provider=%s, status=%s",
		cfg.ShopUuid, cfg.ProviderName, cfg.ProviderShopStatus)
	return resp, nil
}

// HandleIntegrationStatus 处理 Grab 门店集成状态 Webhook
// 包含: partnerMerchantID 解析、状态映射、配置更新、通知发送
func (s *sShopProviderCfg) HandleIntegrationStatus(ctx context.Context, req *grabfood.PushIntegrationStatusWebhookRequest) error {
	partnerMerchantID := req.GetPartnerMerchantID()
	grabMerchantID := req.GetGrabMerchantID()
	integrationStatus := req.GetIntegrationStatus()

	g.Log().Infof(ctx, "[ShopProviderCfg] HandleIntegrationStatus: partnerMerchantID=%s, grabMerchantID=%s, status=%s",
		partnerMerchantID, grabMerchantID, integrationStatus)

	// 1. partnerMerchantID 转换为 shopUUID
	shopUUID := g.NewVar(partnerMerchantID).Uint64()
	if shopUUID == 0 {
		g.Log().Warningf(ctx, "[ShopProviderCfg] partnerMerchantID 无效: %s", partnerMerchantID)
		return nil
	}

	// 2. 映射 Grab 状态到内部状态
	internalStatus := consts.MapGrabIntegrationStatus(integrationStatus)

	// 3. 更新配置并发送通知
	return s.UpsertAndNotify(ctx, shopUUID, string(consts.ProviderGrab), grabMerchantID, internalStatus)
}
