package queue

import (
	"context"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/rocketmq"

	"github.com/apache/rocketmq-client-go/v2/rlog"
	"go.uber.org/zap"
)

const TAKEOUT = "takeout"
const MEMBER_ORDER_CANCEL = "member_order_cancel"

const (
	TopicItemChange       = "erp-item-change"
	TopicDocChange        = "erp-doc-change"
	TopicLanPrinterReport = "lan-printer-report"
)

var manager *rocketmq.Manager

func Init() {
	initMemberOrderCancel()

	manager = rocketmq.NewManager(logger.Logger)
	manager.RegisterConsumer(config.Rocketmq.GroupName, &config.Rocketmq)
	rlog.SetLogLevel(config.Rocketmq.LogLevel)

	service.Queue.RegisterManager(manager)

	err := manager.StartAll()
	if err != nil {
		logger.Logger.Error("启动 RocketMQ 消费者失败", zap.Error(err))
	}

	//订阅消息
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicItemChange, erpItemChangeHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicItemChange))
	}

	// 订阅 LAN 打印机上报消息
	err = manager.Subscribe(config.Rocketmq.GroupName, TopicLanPrinterReport, lanPrinterReportHandler)
	if err != nil {
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err), zap.String("topic", TopicLanPrinterReport))
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
