// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

type (
	IShopProviderCfg interface {
		// UpsertShopProviderCfg 更新或插入门店第三方配置（幂等）
		// shopUUID: 门店 UUID
		// providerName: 第三方名称（如 grab）
		// merchantID: 第三方商户 ID
		// status: 集成状态
		UpsertShopProviderCfg(ctx context.Context, shopUUID uint64, providerName string, merchantID string, status consts.ProviderShopStatus) error
		// GetShopProviderCfg 查询门店第三方配置
		// shopUUID: 门店 UUID
		// providerName: 第三方名称（如 grab），为空默认 grab
		GetShopProviderCfg(ctx context.Context, shopUUID uint64, providerName string) (*entity.ShopProviderCfg, error)
		// NotifyStoreIntegrationState 发送门店集成状态变更通知 (RocketMQ)
		NotifyStoreIntegrationState(ctx context.Context, event *grabDto.ShopIntegrationStatusEvent) error
		// UpsertAndNotify 更新配置并发送通知（组合方法）
		UpsertAndNotify(ctx context.Context, shopUUID uint64, providerName string, merchantID string, status consts.ProviderShopStatus) error
		// GetShopProviderCfgResp 查询门店第三方配置（gRPC 响应格式）
		GetShopProviderCfgResp(ctx context.Context, req *grab.GetShopProviderCfgReq) (*grab.GetShopProviderCfgResp, error)
		// HandleIntegrationStatus 处理 Grab 门店集成状态 Webhook
		// 包含: partnerMerchantID 解析、状态映射、配置更新、通知发送
		HandleIntegrationStatus(ctx context.Context, req *grabfood.PushIntegrationStatusWebhookRequest) error
	}
)

var (
	localShopProviderCfg IShopProviderCfg
)

func ShopProviderCfg() IShopProviderCfg {
	if localShopProviderCfg == nil {
		panic("implement not found for interface IShopProviderCfg, forgot register?")
	}
	return localShopProviderCfg
}

func RegisterShopProviderCfg(i IShopProviderCfg) {
	localShopProviderCfg = i
}
