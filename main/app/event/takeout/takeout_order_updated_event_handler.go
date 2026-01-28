package event

import (
	"context"
	"sync"
	"ttpos-server-go/app/dto/req"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
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

	// 检查是否有菜品变动
	if orderUpdatedEvent.HasItemChange() {
		changeResult := orderUpdatedEvent.ChangeResult

		logger.Logger.Info("订单菜品变动，开始处理",
			zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
			zap.String("platform", orderUpdatedEvent.Platform),
			zap.Int("returnItemCount", changeResult.GetReturnItemCount()),
			zap.Int("kitchenItemCount", changeResult.GetKitchenItemCount()))

		// Step 1: 先打印退菜单（同步执行，确保顺序）
		if changeResult.GetReturnItemCount() > 0 {
			if err := takeoutSrv.PrintReturnOrder(ctx, orderUpdatedEvent.OrderUuid, changeResult); err != nil {
				logger.Logger.Error("打印退菜单失败",
					zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
					zap.String("platform", orderUpdatedEvent.Platform),
					zap.Int("returnItemCount", changeResult.GetReturnItemCount()),
					zap.Error(err))
			}
		}

		// Step 2: 再打印送厨单
		if changeResult.GetKitchenItemCount() > 0 {
			s.printKitchenReceipt(ctx, takeoutSrv, orderUpdatedEvent, changeResult)
		}

		// Step 3: 同步更新生产单（厨显端数据来源）
		// - 退菜商品：标记为退菜状态
		// - 新增商品：创建新的生产单商品
		if err := takeoutSrv.UpdateProductionOrderForTakeout(ctx, orderUpdatedEvent.OrderUuid, changeResult); err != nil {
			logger.Logger.Error("同步更新生产单失败",
				zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
				zap.String("platform", orderUpdatedEvent.Platform),
				zap.Error(err))
		}

		// Step 4: 处理库存和销量同步（异步执行）
		if err := takeoutSrv.ProcessOrderItemsStockAndSales(
			ctx,
			orderUpdatedEvent.OrderUuid,
			orderUpdatedEvent.CompanyUuid,
			changeResult,
		); err != nil {
			logger.Logger.Error("处理订单变动库存销量失败",
				zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
				zap.String("platform", orderUpdatedEvent.Platform),
				zap.Error(err))
		}
	}

	// 异步重新打印外卖订单小票（无论是否有变动都打印）
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

	// 发送 WebSocket 通知
	sendTakeoutOrderWebSocketNotification(
		orderUpdatedEvent.CompanyUuid,
		orderUpdatedEvent.TakeoutOrderUuid,
		orderUpdatedEvent.Platform,
		orderUpdatedEvent.ShortOrderNumber,
		"updated",
		map[string]any{},
	)

	// 成功后，推送到厨显端更新订单
	sendUpdateKitchenWebSocketNotification(orderUpdatedEvent.CompanyUuid)

	return nil
}

// printKitchenReceipt 打印送厨单
// 根据门店配置：一菜一单/整单打印
func (s *takeoutOrderUpdatedEventSubscriber) printKitchenReceipt(
	ctx appContext.Context,
	takeoutSrv takeoutService.ITakeoutSrv,
	orderUpdatedEvent event.OrderUpdatedEvent,
	changeResult *value_object.OrderChangeResult,
) {

	// 构建要打印的菜品列表
	productItems := make([]req.PrintProductItem, 0, len(changeResult.KitchenItems))
	for _, item := range changeResult.AllChanges {
		if item.NewItem != nil && item.NewItem.TtposProductPackageUuid > 0 {
			productItems = append(productItems, req.PrintProductItem{
				ProductUuid: item.NewItem.TtposProductPackageUuid,
				Num:         &item.NewItem.Quantity,
			})
		} else if item.OldItem != nil && item.OldItem.TtposProductPackageUuid > 0 {
			productItems = append(productItems, req.PrintProductItem{
				ProductUuid: item.OldItem.TtposProductPackageUuid,
				Num:         &item.OldItem.Quantity,
			})
		}
	}
	if len(productItems) > 0 {
		// 打印送厨单
		_, err := takeoutSrv.PrintProductionOrder(
			ctx,
			orderUpdatedEvent.OrderUuid,
			printerConstant.PrinterProductTypeKitchen, // 1 送厨打印
			productItems,
		)
		if err != nil {
			logger.Logger.Error("打印送厨单失败",
				zap.Uint64("orderUuid", orderUpdatedEvent.OrderUuid),
				zap.String("platform", orderUpdatedEvent.Platform),
				zap.Int("itemCount", len(productItems)),
				zap.Error(err))
		}
	}
}
