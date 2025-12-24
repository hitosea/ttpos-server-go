package queue

import (
	"context"
	"encoding/json"
	"strconv"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// ProviderMenuUpdateEvent 供应商菜单更新事件
type ProviderMenuUpdateEvent struct {
	ProviderName      string `json:"provider_name"`       // 供应商名称 (e.g., grab)
	MerchantID        string `json:"merchant_id"`         // 平台商户 ID
	PartnerMerchantID string `json:"partner_merchant_id"` // 合作商户 ID
	ShopUuid          string `json:"shop_uuid"`           // TTPOS 门店 UUID
	Uuid              uint64 `json:"uuid"`                // 唯一ID
	ReceivedAt        int64  `json:"received_at"`         // 接收时间戳
}

// TakeoutProviderMenuUpdateHandler 处理供应商菜单更新消息
func TakeoutProviderMenuUpdateHandler(ctx context.Context, msg *primitive.MessageExt) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("处理供应商菜单更新消息失败", zap.Any("err", err))
		}
	}()

	var event ProviderMenuUpdateEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.Logger.Error("解析供应商菜单更新消息失败",
			zap.String("msg_id", msg.MsgId),
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		return err
	}

	logger.Logger.Info("收到供应商菜单更新事件",
		zap.String("provider", event.ProviderName),
		zap.String("merchant_id", event.MerchantID),
		zap.String("partner_merchant_id", event.PartnerMerchantID),
		zap.String("shop_uuid", event.ShopUuid),
		zap.Uint64("uuid", event.Uuid),
		zap.Int64("received_at", event.ReceivedAt))

	// 创建gin上下文
	dbm := database.GetDBManager(config.Database)
	cache := cache.NewCache(cache.Redis, cache.Config{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	settingSrv := setting.NewSrv(dbm, cache)
	localeSrv := service.NewLocaleSrv()
	productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, service.NewTranslateSrv(dbm, cache))
	translateSrv := service.NewTranslateSrv(dbm, cache)
	productTakeoutSrv := service.NewProductTakeoutSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
	takeoutSrv := takeoutService.NewTakeoutSrv(dbm, cache, productSrv, productTakeoutSrv, translateSrv, settingSrv)

	shopUuid, err := strconv.ParseUint(event.ShopUuid, 10, 64)
	if err != nil {
		logger.Logger.Error("转换ShopUuid失败", zap.Error(err))
		return err
	}
	// 获取数据库连接
	db := dbm.GetDB(shopUuid)
	if db == nil {
		logger.Logger.Error("获取数据库连接失败", zap.Uint64("shop_uuid", shopUuid))
		return err
	}
	// 获取公司信息
	companyInfo, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(shopUuid)
	if err != nil {
		logger.Logger.Error("获取公司信息失败", zap.Error(err))
		return err
	}
	// 设置上下文
	ctxHelper := ttposContext.NewDefaultContext()
	ctxHelper.SetBasicInfo(db, companyInfo)
	logger.Logger.Info("导入菜单到TTPOS开始", zap.String("provider", event.ProviderName), zap.String("merchant_id", event.MerchantID), zap.String("partner_merchant_id", event.PartnerMerchantID), zap.String("shop_uuid", event.ShopUuid), zap.Uint64("uuid", event.Uuid), zap.Int64("received_at", event.ReceivedAt))

	// 导入菜单到TTPOS
	_, err = takeoutSrv.ImportMenuToTTPOS(ctxHelper)
	if err != nil {
		logger.Logger.Error("导入菜单到TTPOS失败", zap.Error(err))
		return err
	}

	logger.Logger.Info("导入菜单到TTPOS成功", zap.String("provider", event.ProviderName), zap.String("merchant_id", event.MerchantID), zap.String("partner_merchant_id", event.PartnerMerchantID), zap.String("shop_uuid", event.ShopUuid), zap.Uint64("uuid", event.Uuid), zap.Int64("received_at", event.ReceivedAt))

	return nil
}
