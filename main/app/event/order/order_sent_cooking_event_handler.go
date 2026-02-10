package event

import (
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_sent_cooking_event_handler sync.Once

// sentCookingEventHandler "送厨"事件处理器
func SentCookingEventHandler() {
	once_sent_cooking_event_handler.Do(func() {

		// 创建送厨单打印记录
		event.NewSystemBus().SubscribeSentCookingEvent(func(payload event.SentCookingPayload) {
			if len(payload.Products) == 0 {
				return
			}

			// 分批商品不打印送厨
			products := make([]event.OrderProduct, 0)
			batchProduct := make([]event.OrderProduct, 0)
			for _, unCookingSaleOrderProduct := range payload.Products {
				if unCookingSaleOrderProduct.IsBatch {
					// 如果是分批送厨的商品,此时商品处于预送厨状态,如果是分批前置模式,且开启合并打印模式,则需要打印送厨单. 将这些预送厨商品根据分批类型进行分组,然后每个组进行一次打印送厨单
					batchProduct = append(batchProduct, unCookingSaleOrderProduct)
					continue
				}
				products = append(products, unCookingSaleOrderProduct)
			}
			payload.Products = products

			// 如果有分批商品，处理合并打印逻辑
			if len(batchProduct) > 0 && payload.BatchMode == constant.BatchCookingModePre && payload.BatchPrintMode == constant.BatchPrintModeMerge {
				utils.Go(func() {
					// handleBatchProductsMergePrint(payload, batchProduct)  // 暂时不在“送厨”事件中触发合并打印逻辑, 改为在“预送厨”事件中触发合并打印逻辑
				})
			}

			utils.Go(func() {
				products := printer_model.Products{}
				copier.Copy(&products, payload.Products)

				// 获取订单信息
				db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
				billInfo, err := repository.NewOrderRepo(db).GetSaleBill(
					repository.CommonRepo.WhereByUuid(payload.SaleBillUuid),
					repository.CommonRepo.Preload(repository.WithPreload{Query: "Desk"}),
				)
				if err != nil {
					logger.Logger.Error("FinishMenuEventHandler process, GetSaleBillInfo failed", zap.Error(err))
					return
				}
				// 打印送厨
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					printerConstant.PrinterProductTypeKitchen,
					printer_model.Order{
						Uuid:                   payload.SaleBillUuid,
						SaleOrderUuid:          payload.SaleOrderUuid,
						OrderNo:                billInfo.OrderNo,
						MealNum:                billInfo.MealNum,
						IsTakeoutBill:          billInfo.IsTakeoutBill(),
						OrderSourceTakeoutText: billInfo.GetOrderSourceTakeoutText(),
						SerialNo:               billInfo.SerialNo,
						OrderRemark:            billInfo.GetLatestOrderRemarkRes(),
						DeskUuid:               billInfo.DeskUuid,
						Desk: &printer_model.OrderDesk{
							DeskNo: func() string {
								if billInfo.Desk == nil {
									return ""
								}
								return billInfo.Desk.DeskNo
							}(),
							RegionUuid: func() uint64 {
								if billInfo.Desk == nil {
									return 0
								}
								return billInfo.Desk.RegionUuid
							}(),
						},
						UpdateTime: billInfo.UpdateTime,
						FinishTime: billInfo.FinishTime,
						Products:   products,
					},
				)
			})
		})

		// 创建操作记录
		event.NewSystemBus().SubscribeSentCookingEvent(func(payload event.SentCookingPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderSendKitchen,
				Remark:        "送厨",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				H5OrderUuid:   payload.H5OrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			_, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeSentCookingEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
		})

		// 扣减库存
		event.NewSystemBus().SubscribeSentCookingEvent(func(payload event.SentCookingPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			ReduceStock(payload.Ctx, db, payload.SaleBillUuid)
		})
	})
}

func ReduceStock(payloadCtx context.Context, db *gorm.DB, saleBillUuid uint64) {
	// 加锁。防止多个送厨事件并发扣减库存
	lock.NewSystemLock().LockUuid(saleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
	// 扣减库存
	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItemNotProcessed(saleBillUuid)
	if err != nil {
		logger.Logger.Error("SubscribeSentCookingEvent process, GetWarehouseOutFormItemNotProcessed failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
		return
	}

	ProductBoms := make(map[uint64]*model.ProductBom)
	ProducttPackages := make(map[uint64]*model.ProductPackage)
	type StockNum struct {
		MaterialUuid   uint64
		WarehouseUuid  uint64
		ReduceStockNum float64
	}
	Materials := make(map[uint64]*StockNum)
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		warehouseOutFormItem.ReduceStock = constant.WarehouseOutFormItemReduceStockSuccess
		if warehouseOutFormItem.IsProductBom() { // ProductBom包含规格和小料
			if ProductBoms[warehouseOutFormItem.ProductBomUuid] == nil {
				ProductBoms[warehouseOutFormItem.ProductBomUuid] = warehouseOutFormItem.ProductBom
			}
			ProductBoms[warehouseOutFormItem.ProductBomUuid].StockNum -= warehouseOutFormItem.Num // 扣减库存
			productPackageUuid := ProductBoms[warehouseOutFormItem.ProductBomUuid].ProductPackageUuid
			// 只有规格才给商品包增加销量
			if ProductBoms[warehouseOutFormItem.ProductBomUuid].IsFlavor() /* 商品规格 */ || ProductBoms[warehouseOutFormItem.ProductBomUuid].IsPackageFlavor() /* 套餐规格 */ {
				if ProducttPackages[productPackageUuid] == nil {
					ProducttPackages[productPackageUuid] = &ProductBoms[warehouseOutFormItem.ProductBomUuid].ProductPackage
				}
			}
		} else if warehouseOutFormItem.IsMaterial() {
			if Materials[warehouseOutFormItem.MaterialUuid] == nil {
				Materials[warehouseOutFormItem.MaterialUuid] = &StockNum{
					MaterialUuid:   warehouseOutFormItem.MaterialUuid,
					WarehouseUuid:  warehouseOutFormItem.Material.WarehouseUuid,
					ReduceStockNum: 0,
				}
			}
			Materials[warehouseOutFormItem.MaterialUuid].ReduceStockNum += warehouseOutFormItem.Num
		}
	}

	ProductBomsList := make([]*model.ProductBom, 0)
	ProductPackagesList := make([]*model.ProductPackage, 0)
	for _, productBom := range ProductBoms {
		ProductBomsList = append(ProductBomsList, productBom)
	}
	for _, productPackage := range ProducttPackages {
		ProductPackagesList = append(ProductPackagesList, productPackage)
	}
	// 更新库存
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid); err != nil {
			logger.Logger.Error("SubscribeSentCookingEvent process, UpdateWarehouseOutFormItemRecordsReduceStock failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
			return err
		}
		if err := repository.NewProductBomRepo(tx).UpdateProductBoms(ProductBomsList); err != nil {
			logger.Logger.Error("SubscribeSentCookingEvent process, UpdateProductBomStockNum failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
			return err
		}
		for _, material := range Materials {
			if err := base.NewMaterialRepo(tx).UpdateMaterialsStockNum(material.MaterialUuid, material.WarehouseUuid, -material.ReduceStockNum); err != nil {
				logger.Logger.Error("SubscribeSentCookingEvent process, UpdateMaterialsStockNum failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
				return err
			}
		}

		// 更新product_package的actual_sale_num字段
		for _, productPackage := range ProductPackagesList {
			if err := base.NewProductPackageRepo(tx).UpdateProductPackageActualSaleNum(*productPackage); err != nil {
				logger.Logger.Error("SubscribeSentCookingEvent process, UpdateProductPackageActualSaleNum failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
				return err
			}
		}

		return nil
	}); err != nil {
		logger.Logger.Error("SubscribeSentCookingEvent process, Transaction failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
		return
	}

	// 同步外卖平台
	payloadCtxCopy := payloadCtx.Copy()
	payloadCtxCopy.SetDB(db)
	utils.Go(func() {
		dbm := database.GetDBManager(config.Database)
		takeoutSrv := takeoutService.NewTakeoutSrv(dbm, cache.Global, nil, nil, nil, nil)
		// 同步Grab外卖平台
		_, err := takeoutSrv.SyncMenuChanges(payloadCtxCopy, value_object.TakeoutPlatformGrab)
		if err != nil {
			logger.Logger.Error("同步Grab外卖平台失败", zap.Error(err))
		}
		// 同步Lineman外卖平台
		_, err = takeoutSrv.SyncMenuChanges(payloadCtxCopy, value_object.TakeoutPlatformLineman)
		if err != nil {
			logger.Logger.Error("同步Lineman外卖平台失败", zap.Error(err))
		}
	})
}

// handleBatchProductsMergePrint 处理分批商品的合并打印逻辑
func handleBatchProductsMergePrint(payload event.SentCookingPayload, batchProduct []event.OrderProduct) {
	// 按 BatchTagUuid 分组
	typeGroups := make(map[uint64][]event.OrderProduct)
	for _, product := range batchProduct {
		batchTagUuid := product.BatchTagUuid
		if batchTagUuid == 0 {
			// 如果没有分批类型，跳过不打印
			continue
		}
		if typeGroups[batchTagUuid] == nil {
			typeGroups[batchTagUuid] = []event.OrderProduct{product}
		} else {
			typeGroups[batchTagUuid] = append(typeGroups[batchTagUuid], product)
		}
	}

	// 每个组进行一次打印送厨单
	for _, products := range typeGroups {
		if len(products) == 0 {
			continue
		}

		// 转换为打印模型
		printProducts := printer_model.Products{}
		copier.Copy(&printProducts, products)
		if len(printProducts) == 0 {
			continue
		}

		// 标记显示为延迟送厨
		for i := range printProducts {
			product := &printProducts[i]
			product.ShowDelayTag = true
		}

		// 使用第一个商品的销售订单UUID（从 payload 中获取）
		saleOrderUuid := payload.SaleOrderUuid
		if saleOrderUuid > 0 {
			// 获取订单信息
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			billInfo, err := repository.NewOrderRepo(db).GetSaleBill(
				repository.CommonRepo.WhereByUuid(payload.SaleBillUuid),
				repository.CommonRepo.Preload(repository.WithPreload{Query: "Desk"}),
			)
			if err != nil {
				logger.Logger.Error("handleBatchProductsMergePrint process, GetSaleBill failed", zap.Error(err))
				return
			}
			// 打印送厨
			printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
				printerConstant.PrinterProductTypeKitchen,
				printer_model.Order{
					Uuid:                   payload.SaleBillUuid,
					SaleOrderUuid:          saleOrderUuid,
					OrderNo:                billInfo.OrderNo,
					MealNum:                billInfo.MealNum,
					IsTakeoutBill:          billInfo.IsTakeoutBill(),
					OrderSourceTakeoutText: billInfo.GetOrderSourceTakeoutText(),
					SerialNo:               billInfo.SerialNo,
					OrderRemark:            billInfo.GetLatestOrderRemarkRes(),
					DeskUuid:               billInfo.DeskUuid,
					Desk: &printer_model.OrderDesk{
						DeskNo: func() string {
							if billInfo.Desk == nil {
								return ""
							}
							return billInfo.Desk.DeskNo
						}(),
						RegionUuid: func() uint64 {
							if billInfo.Desk == nil {
								return 0
							}
							return billInfo.Desk.RegionUuid
						}(),
					},
					UpdateTime: billInfo.UpdateTime,
					FinishTime: billInfo.FinishTime,
					Products:   printProducts,
				},
			)
		} else {
			logger.Logger.Warn("打印商品列表不为空但未找到销售订单UUID", zap.Uint64("saleBillUuid", payload.SaleBillUuid), zap.Int("productCount", len(printProducts)))
		}
	}
}
