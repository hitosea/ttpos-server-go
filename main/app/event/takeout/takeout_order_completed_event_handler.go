package event

import (
	"context"
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_takeout_order_completed_event_handler sync.Once

// TakeoutOrderCompletedEventHandler 外卖订单完成事件处理器
func TakeoutOrderCompletedEventHandler() {
	once_takeout_order_completed_event_handler.Do(func() {
		// 订阅外卖订单完成事件
		event.GetDispatcher().Register(&takeoutOrderCompletedEventSubscriber{})
	})
}

// takeoutOrderCompletedEventSubscriber 外卖订单完成事件订阅者
type takeoutOrderCompletedEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderCompletedEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.completed"}
}

// Handle 处理事件
func (s *takeoutOrderCompletedEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderCompletedEvent, ok := domainEvent.(event.OrderCompletedEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderCompletedEvent.CompanyUuid,
			orderCompletedEvent.TakeoutOrderUuid,
			orderCompletedEvent.Platform,
			orderCompletedEvent.ShortOrderNumber,
			"completed",
			map[string]any{},
		)

		// 记录高峰期（订单完成时记录，直接使用事件中的订单数据，避免事务可见性问题）
		dbm := database.GetDBManager(config.DatabaseConf{})
		ctx := appContext.NewContext(
			appContext.WithContext(context.Background()),
		)

		takeoutSrv := takeoutService.NewTakeoutSrvImpl(dbm)
		if err := takeoutSrv.RecordTakeoutOrderPeakTime(ctx,
			orderCompletedEvent.AcceptedBy,
			orderCompletedEvent.CompletedTime,
			orderCompletedEvent.PlatformTotal,
			orderCompletedEvent.CompanyUuid,
			orderCompletedEvent.OrderCreateTime,
		); err != nil {
			logger.Logger.Error("记录外卖订单高峰期失败",
				zap.Uint64("orderUuid", orderCompletedEvent.OrderUuid),
				zap.String("takeoutOrderUuid", orderCompletedEvent.TakeoutOrderUuid),
				zap.Error(err))
		}
	})

	return nil
}
