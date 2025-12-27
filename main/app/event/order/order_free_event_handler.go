package event

import (
	"fmt"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_free_sale_order_event_handler sync.Once

// freeSaleOrderEventHandler "免单"事件处理器
func FreeSaleOrderEventHandler() {
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
				// 构建备注信息（包含预设备注和自定义备注）
				orderItemRemarkList := saleOrderProduct.GetOrderItemRemark()
				remarkInfo := saleOrderProduct.BuildOrderItemRemarkInfo(orderItemRemarkList, saleOrderProduct.Remark)
				products = append(products, printer_model.OrderProduct{
					OrderProductId:        saleOrderProduct.Uuid,
					BatchTagUuid:          saleOrderProduct.BatchTagUuid,
					ProductId:             saleOrderProduct.ProductPackageUuid,
					ProductName:           saleOrderProduct.MultiLanguageName.GetNames(),
					ProductType:           saleOrderProduct.ProductType,
					ProductAttr:           saleOrderProduct.GetAttributeName(),
					ProductAttrList:       saleOrderProduct.GetAttributeNameList(),
					Attr:                  saleOrderProduct.GetPureAttributeName(),
					AttrList:              saleOrderProduct.GetPureAttributeNameList(),
					FlavorName:            saleOrderProduct.GetFlavorName(),
					ProductSauceNamesList: saleOrderProduct.GetSauceNamesList(),
					TotalNum:              saleOrderProduct.Num,
					NumType:               saleOrderProduct.NumType,
					IsBuffet:              saleOrderProduct.IsBuffet == 1,
					IsWrap:                saleOrderProduct.IsWrapProduct(),
					IsGift:                saleOrderProduct.IsGiftProduct(),
					Remark:                saleOrderProduct.Remark,
					RemarkLocale:          remarkInfo.Remark,
				})
			}

			if len(products) > 0 {
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypePay,
					payload.SaleBillUuid,
					payload.SaleOrderUuid,
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

		// 处理高峰时段
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			// 如果订单未完成，不处理
			if !payload.SaleBill.IsFinish() {
				return
			}
			// 获取门店设置
			setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
			storeSetting, err := setting.GetStoreSetting(payload.Ctx)
			if err != nil {
				logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetStoreSetting failed", zap.Error(err))
				return
			}
			//
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.SaleBill.Amount = 0
			err = repository.NewSaleOrderPeakTimeRepo(db).Record("inc", payload.SaleBill, 0.0, storeSetting.TimeZone)
			if err != nil {
				fmt.Println("SubscribeCheckoutSaleOrderEvent process, Record failed", payload, err)
				logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, Record failed", zap.Any("payload", payload), zap.Error(err))
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
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
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
			settleRecord.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err = orderRecordRepo.CreateSaleOrderOperationRecord(settleRecord)
			if err != nil {
				logger.Logger.Error("SubscribeFreeSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:结账 %+v", payload), zap.Uint64("record", uuid))

		})
	})
}
