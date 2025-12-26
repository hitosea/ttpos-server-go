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

var once_takeout_order_cancel_event_handler sync.Once

// TakeoutOrderCancelEventHandler 外卖订单取消事件处理器
func TakeoutOrderCancelEventHandler() {
	once_takeout_order_cancel_event_handler.Do(func() {
		// 订阅外卖订单取消事件
		event.GetDispatcher().Register(&takeoutOrderCancelEventSubscriber{})
	})
}

// takeoutOrderCancelEventSubscriber 外卖订单取消事件订阅者
type takeoutOrderCancelEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderCancelEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.cancelled"}
}

// Handle 处理事件
func (s *takeoutOrderCancelEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderCancelEvent, ok := domainEvent.(event.OrderCancelEvent)
	if !ok {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 创建数据库管理器
		dbm := database.GetDBManager(config.DatabaseConf{})

		// 创建 service（只传入必要的依赖，其他为 nil）
		takeoutSrv := takeoutService.NewTakeoutSrv(dbm, nil, nil, nil, nil, nil)

		// 创建上下文
		ctx := appContext.NewContext(
			appContext.WithContext(context.Background()),
		)
		ctx.SetCompanyUuid(orderCancelEvent.CompanyUuid)

		// 调用 service 恢复出库和销量
		if err := takeoutSrv.RestoreTakeoutOrderOutboundAndSales(ctx, orderCancelEvent.OrderUuid, orderCancelEvent.CompanyUuid); err != nil {
			logger.Logger.Error("恢复外卖订单出库和销量失败",
				zap.Uint64("orderUuid", orderCancelEvent.OrderUuid),
				zap.String("takeoutOrderUuid", orderCancelEvent.TakeoutOrderUuid),
				zap.Error(err))
		}

		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderCancelEvent.CompanyUuid,
			orderCancelEvent.TakeoutOrderUuid,
			orderCancelEvent.Platform,
			orderCancelEvent.ShortOrderNumber,
			"cancelled",
			map[string]any{},
		)

		// 成功后，推送到厨显端更新订单
		sendUpdateKitchenWebSocketNotification(orderCancelEvent.CompanyUuid)
	})

	return nil
}
