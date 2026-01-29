package queue

import (
	"context"
	"encoding/json"
	"errors"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// HandleIntegrationStatus 处理门店集成状态变更消息
func HandleIntegrationStatus(ctx context.Context, msg *primitive.MessageExt) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("处理门店集成状态变更消息失败", zap.Any("err", err))
		}
	}()

	var event request.TakeoutIntegrationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.Logger.Error("解析门店集成状态变更消息失败",
			zap.String("msg_id", msg.MsgId),
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		return err
	}

	logger.Logger.Info("收到门店集成状态变更事件",
		zap.Uint64("shop_uuid", event.ShopUuid),
		zap.String("provider_name", event.ProviderName),
		zap.String("provider_shop_status", event.ProviderShopStatus),
		zap.String("provider_merchant_id", event.ProviderMerchantId),
		zap.Int64("updated_at", event.UpdatedAt))

	// 创建gin上下文
	dbm := database.GetDBManager(config.Database)
	// 获取数据库连接
	db := dbm.GetDB(event.ShopUuid)
	if db == nil {
		logger.Logger.Error("获取数据库连接失败", zap.Uint64("shop_uuid", event.ShopUuid))
		return errors.New("获取数据库连接失败")
	}
	// 获取公司信息
	companyInfo, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(event.ShopUuid)
	if err != nil {
		logger.Logger.Error("获取公司信息失败", zap.Error(err))
		return errors.New("获取公司信息失败")
	}
	// 设置上下文
	ctxHelper := ttposContext.NewDefaultContext()
	ctxHelper.SetBasicInfo(db, companyInfo)

	// 处理门店集成状态变更
	err = application.NewTakeoutAppService(dbm).HandleIntegrationStatus(ctxHelper, event)
	if err != nil {
		logger.Logger.Error("处理门店集成状态变更失败", zap.Error(err), zap.Uint64("shop_uuid", event.ShopUuid))
		return errors.New("处理门店集成状态变更失败")
	}

	return nil
}
