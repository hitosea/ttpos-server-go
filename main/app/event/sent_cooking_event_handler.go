package event

import (
	"fmt"
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

// init 自动注册"添加销售账单记录"事件处理器
func init() {
	// 只初始化一次
	sentCookingEventHandler()
}

// sentCookingEventHandler "送厨"事件处理器
func sentCookingEventHandler() {
	once_sent_cooking_event_handler.Do(func() {

		// 创建送厨单打印记录
		event.NewSystemBus().SubscribeSentCookingEvent(func(payload event.SentCookingPayload) {
			if len(payload.Products) == 0 {
				return
			}
			go func() {
				products := printer_model.Products{}
				copier.Copy(&products, payload.Products)
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypeKitchen,
					payload.SaleBillUuid,
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
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeSentCookingEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:送厨 %+v", payload), zap.Uint64("record", uuid))
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
	Materials := make(map[uint64]*model.Material)
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		warehouseOutFormItem.ReduceStock = constant.WarehouseOutFormItemReduceStockSuccess
		if warehouseOutFormItem.IsProductBom() {
			if ProductBoms[warehouseOutFormItem.ProductBomUuid] == nil {
				ProductBoms[warehouseOutFormItem.ProductBomUuid] = warehouseOutFormItem.ProductBom
			}
			ProductBoms[warehouseOutFormItem.ProductBomUuid].StockNum -= warehouseOutFormItem.Num      // 扣减库存
			ProductBoms[warehouseOutFormItem.ProductBomUuid].ActualSaleNum += warehouseOutFormItem.Num // 增加实际销量
			ProductBoms[warehouseOutFormItem.ProductBomUuid].ProductPackage.ActualSaleNum += warehouseOutFormItem.Num
		} else if warehouseOutFormItem.IsMaterial() {
			if Materials[warehouseOutFormItem.MaterialUuid] == nil {
				Materials[warehouseOutFormItem.MaterialUuid] = warehouseOutFormItem.Material
			}
			Materials[warehouseOutFormItem.MaterialUuid].StockNum -= warehouseOutFormItem.Num
			Materials[warehouseOutFormItem.MaterialUuid].ActualSaleNum += warehouseOutFormItem.Num
		}
	}

	ProductBomsList := make([]*model.ProductBom, 0)
	MaterialsList := make([]*model.Material, 0)
	for _, productBom := range ProductBoms {
		ProductBomsList = append(ProductBomsList, productBom)
	}
	for _, material := range Materials {
		MaterialsList = append(MaterialsList, material)
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
		if err := base.NewMaterialRepo(tx).UpdateMaterials(MaterialsList); err != nil {
			logger.Logger.Error("SubscribeSentCookingEvent process, UpdateMaterialStockNum failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
			return err
		}

		// 更新product_package的actual_sale_num字段
		for _, productBom := range ProductBomsList {
			if err := base.NewProductPackageRepo(tx).UpdateProductPackageActualSaleNum(productBom.ProductPackage); err != nil {
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
