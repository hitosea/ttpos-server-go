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

var once_cancel_sale_order_product_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	cancelSaleOrderProductEventHandler()
}

// cancelSaleOrderProductEventHandler "退菜"事件处理器
func cancelSaleOrderProductEventHandler() {
	once_cancel_sale_order_product_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCancelSaleOrderProductEvent(func(payload event.CancelSaleOrderProductPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// 创建退菜操作记录
			go func() {
				orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
				record := model.SaleBillOperationRecord{
					Source:        payload.Source,
					Action:        constant.OrderRefundProduct,
					Remark:        "退菜",
					SaleBillUuid:  payload.SaleBillUuid,
					SaleOrderUuid: payload.SaleOrderUuid,
					OperatorUuid:  payload.GetOperatorUuid(),
				}
				record.Data = payload.ToJsonString()
				uuid, err := orderRecordRepo.CreateSaleBillOperationRecord(record)
				if err != nil {
					logger.Logger.Error("SubscribeCancelSaleOrderProductEvent process, CreateSaleBillOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
					return
				}
				logger.Logger.Info(fmt.Sprintf("操作记录:退菜 %+v", payload), zap.Uint64("record", uuid))
			}()

			// 创建退菜打印记录
			go func() {
				products := printer_model.OrderProduct{}
				copier.Copy(&products, payload)
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypeBackFood,
					payload.SaleBillUuid,
					[]printer_model.OrderProduct{products},
				)
			}()
		})
		event.NewSystemBus().SubscribeCancelSaleOrderProductEvent(func(payload event.CancelSaleOrderProductPayload) {
			fmt.Println("SubscribeCancelSaleOrderProductEvent payload", utils.ToJsonString(payload))
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			AddStock(db, payload.SaleBillUuid)
		})
	})
}

func AddStock(db *gorm.DB, saleBillUuid uint64) {
	// 加锁。防止多个取消退菜事件并发退还库存
	lock.NewSystemLock().LockUuid(saleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseFormItems, err := warehouseFormRepo.GetWarehouseFormItemNotProcessed(saleBillUuid)
	if err != nil {
		logger.Logger.Info("SubscribeCancelSaleOrderProductEvent process, GetWarehouseFormItemNotProcessed failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
		return
	}
	fmt.Println("warehouseFormItems", utils.ToJsonString(warehouseFormItems))
	productBoms := make(map[uint64]*model.ProductBom)
	materials := make(map[uint64]*model.Material)
	for _, warehouseFormItem := range warehouseFormItems {
		if warehouseFormItem.IsProductBom() {
			productBoms[warehouseFormItem.ProductBomUuid] = warehouseFormItem.ProductBom
			productBoms[warehouseFormItem.ProductBomUuid].StockNum += warehouseFormItem.Num
		} else if warehouseFormItem.IsMaterial() {
			materials[warehouseFormItem.MaterialUuid] = warehouseFormItem.Material
			materials[warehouseFormItem.MaterialUuid].StockNum += warehouseFormItem.Num
		}
	}
	productBomsList := make([]*model.ProductBom, 0)
	materialsList := make([]*model.Material, 0)
	for _, productBom := range productBoms {
		productBomsList = append(productBomsList, productBom)
	}
	for _, material := range materials {
		materialsList = append(materialsList, material)
	}
	// 更新库存
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseFormItemRecordsAddStock(saleBillUuid); err != nil {
			return err
		}
		if err := repository.NewProductBomRepo(tx).UpdateProductBoms(productBomsList); err != nil {
			return err
		}
		if err := base.NewMaterialRepo(db).UpdateMaterials(materialsList); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Logger.Info("SubscribeCancelSaleOrderProductEvent process, Transaction failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
		return
	}
}
