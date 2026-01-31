package queue

import (
	"context"
	"encoding/json"
	"strconv"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// TakeoutProviderOrderUpdateHandler 处理供应商订单更新消息
func TakeoutProviderOrderUpdateHandler(ctx context.Context, msg *primitive.MessageExt) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("处理供应商订单更新消息失败", zap.Any("err", err))
		}
	}()

	// 调试代码
	// // 构造测试事件
	// events := request.TakeoutOrderEvent{
	// 	Action:       "status_update",
	// 	ProviderName: value_object.TakeoutPlatformLineman,
	// 	MerchantId:   "",
	// 	ShopUuid:     "8609817471094784", // 使用实际存在的店铺UUID
	// 	OrderUuid:    "19yx3qw1p00dfz5wjmbtus92w0bz6s8a",
	// 	OrderId:      "LMF-260127-343306443",
	// 	Status:       "COMPLETED",
	// 	Message:      "",
	// 	Timestamp:    time.Now().Unix(),
	// }
	// // 序列化为JSON
	// body, err := json.Marshal(events)
	// if err != nil {
	// 	logger.Logger.Error("序列化供应商订单更新消息失败", zap.Error(err))
	// 	return err
	// }
	// msg.Body = body

	var event request.TakeoutOrderEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.Logger.Error("解析供应商订单更新消息失败",
			zap.String("msg_id", msg.MsgId),
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		return err
	}

	// 判断是否开发模式，如果是开发模式，则使用 6293997752320000
	if config.Server.Mode == "debug" && event.ProviderName == value_object.TakeoutPlatformLineman {
		shopUuid, err := strconv.ParseUint(event.ShopUuid, 10, 64)
		if err != nil {
			logger.Logger.Error("转换ShopUuid失败", zap.Error(err))
			return err
		}
		if config.Takeout.TakeoutLinemanStoreId == shopUuid {
			event.ShopUuid = "8609817471094784"
		}
	}

	logger.Logger.Info("收到供应商订单更新事件",
		zap.String("action", event.Action),
		zap.String("provider", event.ProviderName),
		zap.String("merchant_id", event.MerchantId),
		zap.String("shop_uuid", event.ShopUuid),
		zap.String("order_uuid", event.OrderUuid),
		zap.String("order_id", event.OrderId),
		zap.String("status", event.Status),
		zap.String("message", event.Message),
		zap.Int64("timestamp", event.Timestamp))

	// 创建gin上下文
	dbm := database.GetDBManager(config.Database)
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

	// 导入订单到TTPOS
	err = application.NewTakeoutOrderAppService(dbm).HandlePushOrderState(ctxHelper, event)
	if err != nil {
		logger.Logger.Error("导入订单到TTPOS失败", zap.Error(err), zap.String("provider", event.ProviderName), zap.String("order_uuid", event.OrderUuid))
		return err
	}

	return nil
}
