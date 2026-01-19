package event

import (
	"context"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/takeout/domain/event"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	appContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
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

	// 创建数据库管理器
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(orderAcceptedEvent.CompanyUuid)

	// 创建上下文
	ctx := appContext.NewContext(
		appContext.WithContext(context.Background()),
		appContext.WithStaffUuid(orderAcceptedEvent.AcceptedBy),
	)
	ctx.SetCompanyUuid(orderAcceptedEvent.CompanyUuid)
	ctx.SetDB(db)
	// 创建 service（只传入必要的依赖，其他为 nil）
	takeoutSrv := takeoutService.NewTakeoutSrvImpl(dbm)

	// 异步处理，不阻塞事件分发
	utils.Go(func() {

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

		// 发送 CUSTOMER_CALL 通知（触发前端未处理提醒）
		sendCustomerCallWebSocketNotification(orderAcceptedEvent.CompanyUuid)

		// 成功后，推送到厨显端更新订单
		sendUpdateKitchenWebSocketNotification(orderAcceptedEvent.CompanyUuid)

		// 记录高峰期
		if err := recordTakeoutOrderPeakTime(db, orderAcceptedEvent.CompanyUuid, orderAcceptedEvent.OrderUuid, "inc"); err != nil {
			logger.Logger.Error("记录外卖订单高峰期失败",
				zap.Uint64("orderUuid", orderAcceptedEvent.OrderUuid),
				zap.String("takeoutOrderUuid", orderAcceptedEvent.TakeoutOrderUuid),
				zap.Error(err))
		}
	})

	// 异步打印外卖订单小票
	utils.Go(func() {
		takeoutSrv.PrintTakeoutOrder(ctx, orderAcceptedEvent.OrderUuid, "", 0)
	})

	// 异步打印送厨单
	utils.Go(func() {
		_, err := takeoutSrv.PrintProductionOrder(ctx, orderAcceptedEvent.OrderUuid, printerConstant.PrinterProductTypeKitchen, nil)
		if err != nil {
			logger.Logger.Error("打印送厨单失败",
				zap.String("takeoutOrderUuid", orderAcceptedEvent.TakeoutOrderUuid),
				zap.Error(err))
		}
	})

	return nil
}

// recordTakeoutOrderPeakTime 记录外卖订单高峰期
// recordType: "inc" - 增加, "dec" - 减少
func recordTakeoutOrderPeakTime(db *gorm.DB, companyUuid uint64, orderUuid uint64, recordType string) error {
	// 1. 查询外卖订单信息
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(orderUuid)
	if err != nil {
		return err
	}
	if order == nil {
		return nil
	}

	// 2. 构建 SaleBill
	saleBill := buildSaleBillFromTakeoutOrder(order, recordType)
	if saleBill == nil {
		return nil
	}

	// 3. 获取门店设置（时区）
	settingSrv := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	ctx := appContext.NewContext(
		appContext.WithContext(context.Background()),
		appContext.WithCompanyUuid(companyUuid),
	)
	ctx.SetDB(db)
	storeSetting, err := settingSrv.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Info("获取门店设置失败", zap.Error(err))
		return err
	}

	// 4. 记录高峰期
	peakTimeRepo := repository.NewSaleOrderPeakTimeRepo(db)
	return peakTimeRepo.Record(recordType, saleBill, utils.IfFloat64(recordType == "dec", order.EaterPayment, 0.0), storeSetting.TimeZone)
}

// buildSaleBillFromTakeoutOrder 从外卖订单构建 SaleBill
// recordType: "inc" - 接单时使用 AcceptedTime, "dec" - 取消时使用 RejectedTime
func buildSaleBillFromTakeoutOrder(order *takeoutModel.TakeoutOrder, recordType string) *model.SaleBill {
	saleBill := &model.SaleBill{
		Status:        constant.SaleBillStatusComplete, // 设置为已完成状态，IsFinish() 才能返回 true
		PaymentAmount: order.EaterPayment,              // 顾客实付金额（单位：元）
		CashierUuid:   0,                               // 默认值
		FinishTime:    0,                               // 默认值
	}

	// 根据 recordType 设置不同的时间和收银员
	if recordType == "inc" {
		// 接单时：使用接单时间和接单人
		if order.AcceptedTime > 0 {
			saleBill.FinishTime = order.AcceptedTime
			saleBill.CashierUuid = order.AcceptedBy
		} else {
			// 如果没有接单时间，使用订单时间
			saleBill.FinishTime = order.OrderTime
			saleBill.CashierUuid = order.AcceptedBy
		}
	} else if recordType == "dec" {
		// 取消时：使用取消时间和取消人
		if order.RejectedTime > 0 {
			saleBill.FinishTime = order.RejectedTime
			saleBill.CashierUuid = order.RejectedBy
		} else {
			// 如果没有取消时间，使用订单时间
			saleBill.FinishTime = order.OrderTime
			saleBill.CashierUuid = order.RejectedBy
		}
	}

	// 如果 FinishTime 为 0，无法记录高峰期
	if saleBill.FinishTime == 0 {
		return nil
	}

	return saleBill
}
