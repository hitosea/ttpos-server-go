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

var once_takeout_order_updated_event_handler sync.Once

// TakeoutOrderUpdatedEventHandler 外卖订单更新事件处理器
func TakeoutOrderUpdatedEventHandler() {
	once_takeout_order_updated_event_handler.Do(func() {
		// 订阅外卖订单更新事件
		event.GetDispatcher().Register(&takeoutOrderUpdatedEventSubscriber{})
	})
}

// takeoutOrderUpdatedEventSubscriber 外卖订单更新事件订阅者
type takeoutOrderUpdatedEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderUpdatedEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.updated"}
}

// Handle 处理事件
func (s *takeoutOrderUpdatedEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	orderUpdatedEvent, ok := domainEvent.(event.OrderUpdatedEvent)
	if !ok {
		return nil
	}

	// 创建数据库管理器
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(orderUpdatedEvent.CompanyUuid)

	// 创建上下文
	ctx := appContext.NewContext(
		appContext.WithContext(context.Background()),
	)
	ctx.SetCompanyUuid(orderUpdatedEvent.CompanyUuid)
	ctx.SetDB(db)

	// 创建 service
	takeoutSrv := takeoutService.NewTakeoutSrvImpl(dbm)

	// 异步重新打印外卖订单小票
	utils.Go(func() {
		_, err := takeoutSrv.PrintTakeoutOrder(ctx, orderUpdatedEvent.OrderUuid, "", 0)
		if err != nil {
			logger.Logger.Error("订单更新后重新打印外卖订单小票失败",
				zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
				zap.String("takeoutOrderUuid", orderUpdatedEvent.TakeoutOrderUuid),
				zap.String("platform", orderUpdatedEvent.Platform),
				zap.Error(err))
		}
	})

	// 异步重新打印送厨单
	// utils.Go(func() {
	// 	_, err := takeoutSrv.PrintProductionOrder(ctx, orderUpdatedEvent.OrderUuid, printerConstant.PrinterProductTypeKitchen, nil)
	// 	if err != nil {
	// 		logger.Logger.Error("订单更新后重新打印送厨单失败",
	// 			zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
	// 			zap.String("takeoutOrderUuid", orderUpdatedEvent.TakeoutOrderUuid),
	// 			zap.String("platform", orderUpdatedEvent.Platform),
	// 			zap.Error(err))
	// 	}
	// })

	// 发送 WebSocket 通知
	sendTakeoutOrderWebSocketNotification(
		orderUpdatedEvent.CompanyUuid,
		orderUpdatedEvent.TakeoutOrderUuid,
		orderUpdatedEvent.Platform,
		orderUpdatedEvent.ShortOrderNumber,
		"updated",
		map[string]any{},
	)

	return nil
}
