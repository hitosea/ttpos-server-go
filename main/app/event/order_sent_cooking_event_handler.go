package event

import (
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/config"
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
func sentCookingEventHandler() {
	once_sent_cooking_event_handler.Do(func() {

		// 创建送厨单打印记录
		event.NewSystemBus().SubscribeSentCookingEvent(func(payload event.SentCookingPayload) {
			if len(payload.Products) == 0 {
				return
			}

			// 分批商品不打印送厨
			products := make([]event.OrderProduct, 0)
			for _, unCookingSaleOrderProduct := range payload.Products {
				if unCookingSaleOrderProduct.IsBatch {
					continue
				}
				products = append(products, unCookingSaleOrderProduct)
			}
			payload.Products = products

			go func() {
				products := printer_model.Products{}
				copier.Copy(&products, payload.Products)
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypeKitchen,
					payload.SaleBillUuid,
					payload.SaleOrderUuid,
					products,
				)
			}()
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
			ReduceStock(db, payload.SaleBillUuid)
		})
	})
}

func ReduceStock(db *gorm.DB, saleBillUuid uint64) {
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
}
