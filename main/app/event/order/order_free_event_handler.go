package event

import (
	"fmt"
	"sort"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/controller"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/modules/printer"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	printer_request "ttpos-server-go/app/modules/printer/tyeps/request"
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
					printerConstant.PrinterProductTypePay,
					printer_model.Order{
						Uuid:                   payload.SaleBillUuid,
						SaleOrderUuid:          payload.SaleOrderUuid,
						OrderNo:                payload.SaleBill.OrderNo,
						MealNum:                payload.SaleBill.MealNum,
						IsTakeoutBill:          payload.SaleBill.IsTakeoutBill(),
						OrderSourceTakeoutText: payload.SaleBill.GetOrderSourceTakeoutText(),
						SerialNo:               payload.SaleBill.SerialNo,
						OrderRemark:            payload.SaleBill.GetLatestOrderRemarkRes(),
						DeskUuid:               payload.SaleBill.DeskUuid,
						Desk: &printer_model.OrderDesk{
							DeskNo: func() string {
								if payload.SaleBill.Desk == nil {
									return ""
								}
								return payload.SaleBill.Desk.DeskNo
							}(),
							RegionUuid: func() uint64 {
								if payload.SaleBill.Desk == nil {
									return 0
								}
								return payload.SaleBill.Desk.RegionUuid
							}(),
						},
						UpdateTime: payload.SaleBill.UpdateTime,
						FinishTime: payload.SaleBill.FinishTime,
						Products:   products,
					},
				)
			}
		})

		// 创建结账单打印
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			_, err := printer.NewPrinterRepo(payload.Ctx).PrintingStatementOrder(&printer_request.PrintingStatementOrderReq{
				PrintType:     printerConstant.PrinterTemplateBilling,
				SaleBill:      payload.SaleBill,
				SaleOrderUuid: payload.SaleOrderUuid,
			})
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
				logger.Logger.Error("SubscribeCheckoutSaleOrderEvent process, GetStoreSetting failed", zap.Error(err))
				return
			}
			//
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.SaleBill.Amount = 0
			err = repository.NewSaleOrderPeakTimeRepo(db).Record("inc", payload.SaleBill, 0.0, storeSetting.TimeZone)
			if err != nil {
				fmt.Println("SubscribeCheckoutSaleOrderEvent process, Record failed", payload, err)
				logger.Logger.Error("SubscribeCheckoutSaleOrderEvent process, Record failed", zap.Error(err))
			}
		})

		// 失效桌台缓存（桌台订单完成后）
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			// 如果订单未完成，不处理
			if !payload.SaleBill.IsFinish() {
				return
			}
			if adapter.IsObjectStorageCacheEnabled(payload.CompanyUuid) {
				// 如果是桌台订单，失效桌台缓存
				if payload.SaleBill.IsDeskSaleBill() {
					deskUuid := payload.SaleBill.DeskUuid
					deskUuid = persistence.GlobalObjectUuid // 还没有失效单个桌台缓存,全局失效桌台缓存
					if err := controller.GetDeskController().Invalidate(payload.Ctx, deskUuid); err != nil {
						logger.Logger.Error("SubscribeFreeSaleOrderEvent process, Invalidate desk cache failed", zap.Uint64("deskUuid", deskUuid), zap.Error(err))
					}
				}
			}
		})

		// 增加销量
		// 增加产品销量
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.Ctx.SetDB(db)
			HandleAddSalesVolumeForFree(payload)
		})
		// 增加材料销量
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.Ctx.SetDB(db)
			HandleAddMaterialSalesVolumeForFree(payload)
		})

		// 结账订单后，统计订单原料用量。
		event.NewSystemBus().SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.Ctx.SetDB(db)
			HandleRecordOrderMaterialUsage(payload.Ctx, db, payload.SaleBill, payload.SaleBillUuid, payload.SaleOrderUuid)
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

// 增加销量（免单）
// 注意：对 UUID 排序后再更新，保证所有并发事务获取锁的顺序一致，避免死锁
func HandleAddSalesVolumeForFree(payload event.FreeSaleOrderPayload) {
	ProductBoms, ProductPackages := GetSalesVolume(payload.SaleBill)

	// 提取并排序 ProductBom UUID，避免死锁
	bomUuids := make([]uint64, 0, len(ProductBoms))
	for uuid := range ProductBoms {
		bomUuids = append(bomUuids, uuid)
	}
	sort.Slice(bomUuids, func(i, j int) bool { return bomUuids[i] < bomUuids[j] })

	for _, productBomUuid := range bomUuids {
		saleNum := ProductBoms[productBomUuid]
		if err := repository.NewProductBomRepo(payload.Ctx.GetDB()).AddActualSaleNum(productBomUuid, saleNum); err != nil {
			logger.Logger.Error("HandleAddSalesVolumeForFree process, AddActualSaleNum failed", zap.Any("productBomUuid", productBomUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}

	// 提取并排序 ProductPackage UUID，避免死锁
	packageUuids := make([]uint64, 0, len(ProductPackages))
	for uuid := range ProductPackages {
		packageUuids = append(packageUuids, uuid)
	}
	sort.Slice(packageUuids, func(i, j int) bool { return packageUuids[i] < packageUuids[j] })

	for _, productPackageUuid := range packageUuids {
		saleNum := ProductPackages[productPackageUuid]
		if err := repository.NewProductPackageRepo(payload.Ctx.GetDB()).AddActualSaleNum(productPackageUuid, saleNum); err != nil {
			logger.Logger.Error("HandleAddSalesVolumeForFree process, AddActualSaleNum failed", zap.Any("productPackageUuid", productPackageUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}
}

// 增加材料销量（免单）
// 注意：对 UUID 排序后再更新，保证所有并发事务获取锁的顺序一致，避免死锁
func HandleAddMaterialSalesVolumeForFree(payload event.FreeSaleOrderPayload) {
	MaterialSalesVolume := GetMaterialSalesVolume(payload.CompanyUuid, payload.SaleOrderUuid)

	// 提取并排序 Material UUID，避免死锁
	materialUuids := make([]uint64, 0, len(MaterialSalesVolume))
	for uuid := range MaterialSalesVolume {
		materialUuids = append(materialUuids, uuid)
	}
	sort.Slice(materialUuids, func(i, j int) bool { return materialUuids[i] < materialUuids[j] })

	for _, materialUuid := range materialUuids {
		saleNum := MaterialSalesVolume[materialUuid]
		if err := repository.NewMaterialRepo(payload.Ctx.GetDB()).AddActualSaleNum(materialUuid, saleNum); err != nil {
			logger.Logger.Error("HandleAddMaterialSalesVolumeForFree process, AddActualSaleNum failed", zap.Any("materialUuid", materialUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}
}
