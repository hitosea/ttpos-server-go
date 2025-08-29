package queue

import (
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
	TopicItemChange = "erp-item-change"
	TopicDocChange  = "erp-doc-change"
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
		logger.Logger.Error("订阅 RocketMQ 主题失败", zap.Error(err))
	}

}
