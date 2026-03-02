package event

import (
	"context"
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_takeout_order_peak_time_event_handler sync.Once

// TakeoutOrderPeakTimeEventHandler 外卖订单高峰期记录事件处理器
func TakeoutOrderPeakTimeEventHandler() {
	once_takeout_order_peak_time_event_handler.Do(func() {
		// 订阅外卖订单高峰期记录事件
		event.GetDispatcher().Register(&takeoutOrderPeakTimeEventSubscriber{})
	})
}

// takeoutOrderPeakTimeEventSubscriber 外卖订单高峰期记录事件订阅者
type takeoutOrderPeakTimeEventSubscriber struct{}

// SubscribedEvents 返回订阅的事件类型
func (s *takeoutOrderPeakTimeEventSubscriber) SubscribedEvents() []string {
	return []string{"takeout.order.peak_time.record"}
}

// Handle 处理事件
func (s *takeoutOrderPeakTimeEventSubscriber) Handle(domainEvent event.DomainEvent) error {
	// 类型断言
	peakTimeEvent, ok := domainEvent.(event.OrderPeakTimeRecordEvent)
	if !ok {
		return nil
	}

	// 如果没有订单，直接返回
	if len(peakTimeEvent.OrderUuids) == 0 {
		return nil
	}

	// 异步处理，不阻塞事件分发
	utils.Go(func() {
		// 创建数据库管理器
		dbm := database.GetDBManager(config.DatabaseConf{})
		db := dbm.GetDB(peakTimeEvent.CompanyUuid)

		// 创建上下文
		ctx := appContext.NewContext(
			appContext.WithContext(context.Background()),
		)
		ctx.SetCompanyUuid(peakTimeEvent.CompanyUuid)
		ctx.SetDB(db)

		// 创建 service
		takeoutSrv := takeoutService.NewTakeoutSrvImpl(dbm)

		// 批量记录高峰期：先查询订单获取所需字段，再调用记录方法
		orderRepo := persistence.NewTakeoutOrderRepo(db)
		for _, orderUuid := range peakTimeEvent.OrderUuids {
			order, err := orderRepo.GetByUuid(orderUuid)
			if err != nil || order == nil {
				logger.Logger.Error("查询外卖订单失败",
					zap.Uint64("orderUuid", orderUuid),
					zap.Error(err))
				continue
			}
			if err := takeoutSrv.RecordTakeoutOrderPeakTime(ctx,
				order.AcceptedBy,
				order.CompletedTime,
				order.PlatformTotal,
				peakTimeEvent.CompanyUuid,
			); err != nil {
				logger.Logger.Error("记录外卖订单高峰期失败",
					zap.Uint64("orderUuid", orderUuid),
					zap.Uint64("companyUuid", peakTimeEvent.CompanyUuid),
					zap.Error(err))
			}
		}

		logger.Logger.Debug("批量记录高峰期完成", zap.Int("count", len(peakTimeEvent.OrderUuids)))
	})

	return nil
}
