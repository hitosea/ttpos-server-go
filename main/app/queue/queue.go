package queue

import (
	"context"
	websocketConstant "ttpos-api/ttpos-websocket/constant"
	baseQueue "ttpos-server-go/app/queue/base"
	erpQueue "ttpos-server-go/app/queue/erp"
	printerQueue "ttpos-server-go/app/queue/printer"
	takeoutQueue "ttpos-server-go/app/queue/takeout"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/rocketmq"

	"github.com/apache/rocketmq-client-go/v2/rlog"
	"go.uber.org/zap"
)

const TAKEOUT = "takeout"

const (
	TopicItemChange                    = "erp-item-change"
	TopicDocChange                     = "erp-doc-change"
	TopicErpCancelInvoiceCallback      = "erp-invoice-cancel"              // ERP 取消发票成功回调
	TopicErpSalesInvoiceCallback       = "erp-sales-invoice-callback"      // ERP SI 异步回调（BMP → Main）
)

const (
	TopicProviderMenuUpdate    = "takeout_provider_menu_update"
	TopicProviderOrderUpdate   = "takeout_grab_order"
	TopicStoreIntegrationState = "takeout_store_integration_state"
)

var manager *rocketmq.Manager

func Init() {
	baseQueue.InitMemberOrderCancel()

	baseQueue.InitOperationDurationQueue()
	// baseQueue.InitDBPoolStatsQueue() // 连接池监控采集已关闭：利用率极低，无需持续采集

	manager = rocketmq.NewManager(logger.Logger)
	manager.RegisterConsumer(config.Rocketmq.GroupName, &config.Rocketmq)
	rlog.SetLogLevel(config.Rocketmq.LogLevel)

	service.Queue.RegisterManager(manager)

	err := manager.StartAll()
	if err != nil {
		logger.Logger.Error("启动 RocketMQ 消费者失败", zap.Error(err))
	}

	//订阅消息RocketmqRocketmq
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicItemChange, baseQueue.ErpItemChangeHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicItemChange))
	}

	// 订阅 LAN 打印机上报消息
	err = manager.Subscribe(config.Rocketmq.GroupName, websocketConstant.TopicLanPrinterReport, printerQueue.LanPrinterReportHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", websocketConstant.TopicLanPrinterReport))
	}

	// 订阅供应商菜单更新消息
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicProviderMenuUpdate, takeoutQueue.TakeoutProviderMenuUpdateHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicProviderMenuUpdate))
	}

	// 订阅门店集成状态变更消息
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicStoreIntegrationState, takeoutQueue.HandleIntegrationStatus)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicStoreIntegrationState))
	}

	// 订阅供应商订单更新消息
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicProviderOrderUpdate, takeoutQueue.TakeoutProviderOrderUpdateHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicProviderOrderUpdate))
	}

	// 订阅 ERP 取消发票成功回调消息
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicErpCancelInvoiceCallback, erpQueue.ErpCancelInvoiceCallbackHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicErpCancelInvoiceCallback))
	}

	// 订阅 ERP Sales Invoice 异步回调消息（BMP → Main）
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicErpSalesInvoiceCallback, erpQueue.ErpSalesInvoiceCallbackHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicErpSalesInvoiceCallback))
	}

}

// Shutdown 优雅关闭 RocketMQ 消费者管理器
func Shutdown() error {
	if manager == nil {
		return nil
	}

	logger.Logger.Info("开始关闭 RocketMQ 消费者管理器")

	// 使用管理器的 Shutdown 方法优雅关闭
	ctx := context.Background()
	if err := manager.Shutdown(ctx); err != nil {
		logger.Logger.Error("关闭 RocketMQ 消费者管理器失败", zap.Error(err))
		return err
	}

	manager = nil
	logger.Logger.Info("RocketMQ 消费者管理器已关闭")
	return nil
}
