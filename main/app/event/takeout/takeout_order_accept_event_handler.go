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

var once_takeout_order_accept_event_handler sync.Once

// TakeoutOrderAcceptEventHandler 外卖订单接单事件处理器
func TakeoutOrderAcceptEventHandler() {
	once_takeout_order_accept_event_handler.Do(func() {
		// 订阅外卖订单接单事件
		event.GetDispatcher().Register(&takeoutOrderAcceptEventSubscriber{})
	})
}

// takeoutOrderAcceptEventSubscriber 外卖订单接单事件订阅者
type takeoutOrderAcceptEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderAcceptEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.accepted"}
}

// Handle 处理事件
func (s *takeoutOrderAcceptEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderAcceptedEvent, ok := domainEvent.(event.OrderAcceptedEvent)
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
			appContext.WithStaffUuid(orderAcceptedEvent.AcceptedBy),
		)
		ctx.SetCompanyUuid(orderAcceptedEvent.CompanyUuid)

		// 调用 service 处理出库和销量（包含原料汇总）
		if err := takeoutSrv.ProcessTakeoutOrderOutboundAndSales(ctx, orderAcceptedEvent.OrderUuid, orderAcceptedEvent.CompanyUuid, orderAcceptedEvent.AcceptedBy); err != nil {
			logger.Logger.Error("处理外卖订单出库和销量失败",
				zap.Uint64("orderUuid", orderAcceptedEvent.OrderUuid),
				zap.String("takeoutOrderUuid", orderAcceptedEvent.TakeoutOrderUuid),
				zap.Error(err))
		}

		// 创建送厨单
		if err := takeoutSrv.CreateProductionOrderForTakeout(ctx, orderAcceptedEvent.OrderUuid); err != nil {
			logger.Logger.Error("创建送厨单失败",
				zap.String("takeoutOrderUuid", orderAcceptedEvent.TakeoutOrderUuid),
				zap.Error(err))
		}

		// 发送 WebSocket 通知
		sendTakeoutOrderWebSocketNotification(
			orderAcceptedEvent.CompanyUuid,
			orderAcceptedEvent.TakeoutOrderUuid,
			orderAcceptedEvent.Platform,
			orderAcceptedEvent.ShortOrderNumber,
			"accepted",
			map[string]any{},
		)

		// 成功后，推送到厨显端更新订单
		sendUpdateKitchenWebSocketNotification(orderAcceptedEvent.CompanyUuid)
	})

	return nil
}
