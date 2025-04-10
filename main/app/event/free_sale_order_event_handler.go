package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_free_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	freeSaleOrderEventHandler()
}

// freeSaleOrderEventHandler "免单"事件处理器
func freeSaleOrderEventHandler() {
	once_free_sale_order_event_handler.Do(func() {

		// 创建菜单-付款打印
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			saleOrder := payload.SaleBill.GetSaleOrder(payload.SaleOrderUuid)
			if saleOrder == nil {
				return
			}
			products := printer_model.Products{}
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsSendKitchen() {
					continue
				}
				products = append(products, printer_model.OrderProduct{
					OrderProductId:  saleOrderProduct.Uuid,
					ProductId:       saleOrderProduct.ProductPackageUuid,
					ProductName:     saleOrderProduct.MultiLanguageName.GetNames(),
					ProductAttr:     saleOrderProduct.GetAttributeName(),
					ProductAttrList: saleOrderProduct.GetAttributeNameList(),
					TotalNum:        saleOrderProduct.Num,
					IsBuffet:        saleOrderProduct.IsBuffet == 1,
					Remark:          saleOrderProduct.Remark,
				})
			}

			if len(products) > 0 {
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypePay,
					payload.SaleBillUuid,
					products,
				)
			}
		})

		// 创建结账单打印
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			_, err := printer.NewPrinterRepo(payload.Ctx).PrintingStatementOrder(
				constant.PrinterTemplateBilling,
				payload.SaleBill,
				payload.SaleOrderUuid,
				0,
				0,
			)
			if err != nil {
				fmt.Println("FreeSaleOrderEvent process, PrintingStatementOrder failed ", err)
			}
		})

		// 创建操作记录
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderFreeSale,
				Remark:        "免单",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeFreeSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:免单 %+v", payload), zap.Uint64("record", uuid))

			// 结账操作日志
			settleRecord := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderSettle,
				Remark:        "结账",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			checkoutSaleOrderPayload := event.CheckoutSaleOrderPayload{
				OrderPrice:    payload.OrderPrice,
				PayPrice:      payload.PayPrice,
				IsFree:        true,
				DiscountMoney: payload.DiscountMoney,
			}
			checkoutSaleOrderPayload.IsSplitOrder = payload.SaleBill.IsSplit()
			for i, saleOrder := range payload.SaleBill.SaleOrders {
				if saleOrder.Uuid == payload.SaleOrderUuid {
					checkoutSaleOrderPayload.Index = i + 1
				}
			}
			settleRecord.Data = checkoutSaleOrderPayload.ToJsonString()
			uuid, err = orderRecordRepo.CreateSaleOrderOperationRecord(settleRecord)
			if err != nil {
				logger.Logger.Error("SubscribeFreeSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:结账 %+v", payload), zap.Uint64("record", uuid))

		})
	})
}
