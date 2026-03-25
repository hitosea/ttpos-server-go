package event

import (
	"fmt"
	"sort"
	"sync"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/printer"
	printerConstant "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/config"
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

var once_cancel_sale_order_product_event_handler sync.Once

// cancelSaleOrderProductEventHandler "退菜"事件处理器
func CancelSaleOrderProductEventHandler() {
	once_cancel_sale_order_product_event_handler.Do(func() {
		event.NewSystemBus().SubscribeCancelSaleOrderProductEvent(func(payload event.CancelSaleOrderProductPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// 创建退菜操作记录
			utils.Go(func() {
				orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
				record := model.SaleOrderOperationRecord{
					Source:        payload.Source,
					Action:        constant.OrderRefundProduct,
					Remark:        "退菜",
					SaleBillUuid:  payload.SaleBillUuid,
					SaleOrderUuid: payload.SaleOrderUuid,
					OperatorUuid:  payload.GetOperatorUuid(),
				}
				record.Data = payload.ToJsonString()
				record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
				uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
				if err != nil {
					logger.Logger.Error("SubscribeCancelSaleOrderProductEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
					return
				}
				logger.Logger.Info(fmt.Sprintf("操作记录:退菜 %+v", payload), zap.Uint64("record", uuid))
			})

			// 创建退菜打印记录
			utils.Go(func() {
				products := printer_model.OrderProduct{}
				copier.Copy(&products, payload)
				// 获取订单信息
				billInfo, err := repository.NewOrderRepo(db).GetSaleBill(
					repository.CommonRepo.WhereByUuid(payload.SaleBillUuid),
					repository.CommonRepo.Preload(repository.WithPreload{Query: "Desk"}),
				)
				if err != nil {
					logger.Logger.Error("FinishMenuEventHandler process, GetSaleBillInfo failed", zap.Error(err))
					return
				}
				// 打印退菜
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					printerConstant.PrinterProductTypeBackFood,
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
						Products:   []printer_model.OrderProduct{products},
					},
				)
			})
		})
		event.NewSystemBus().SubscribeCancelSaleOrderProductEvent(func(payload event.CancelSaleOrderProductPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			AddStock(payload.Ctx, db, payload.SaleBillUuid)
		})
	})
}

func AddStock(payloadCtx context.Context, db *gorm.DB, saleBillUuid uint64) {
	// 加锁。防止多个取消退菜事件并发退还库存
	lock.NewSystemLock().LockUuid(saleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseFormItems, err := warehouseFormRepo.GetWarehouseFormItemNotProcessed(saleBillUuid)
	if err != nil {
		logger.Logger.Error("SubscribeCancelSaleOrderProductEvent process, GetWarehouseFormItemNotProcessed failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
		return
	}
	productBoms := make(map[uint64]*model.ProductBom)
	type StockNum struct {
		MaterialUuid  uint64
		WarehouseUuid uint64
		AddStockNum   float64
	}
	materials := make(map[uint64]*StockNum)
	for _, warehouseFormItem := range warehouseFormItems {
		if warehouseFormItem.IsProductBom() {
			productBoms[warehouseFormItem.ProductBomUuid] = warehouseFormItem.ProductBom
			productBoms[warehouseFormItem.ProductBomUuid].StockNum += warehouseFormItem.Num
		} else if warehouseFormItem.IsMaterial() {
			materials[warehouseFormItem.MaterialUuid] = &StockNum{
				MaterialUuid:  warehouseFormItem.MaterialUuid,
				WarehouseUuid: warehouseFormItem.Material.WarehouseUuid,
				AddStockNum:   warehouseFormItem.Num,
			}
		}
	}
	// 提取并排序 ProductBom UUID，避免死锁
	bomUuids := make([]uint64, 0, len(productBoms))
	for uuid := range productBoms {
		bomUuids = append(bomUuids, uuid)
	}
	sort.Slice(bomUuids, func(i, j int) bool { return bomUuids[i] < bomUuids[j] })

	productBomsList := make([]*model.ProductBom, 0, len(bomUuids))
	for _, uuid := range bomUuids {
		productBomsList = append(productBomsList, productBoms[uuid])
	}

	// 提取并排序 Material UUID，避免死锁
	materialUuids := make([]uint64, 0, len(materials))
	for uuid := range materials {
		materialUuids = append(materialUuids, uuid)
	}
	sort.Slice(materialUuids, func(i, j int) bool { return materialUuids[i] < materialUuids[j] })

	// 更新库存
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseFormItemRecordsAddStock(saleBillUuid); err != nil {
			return err
		}
		if err := repository.NewProductBomRepo(tx).UpdateProductBoms(productBomsList); err != nil {
			return err
		}
		// 按排序后的顺序更新仓库物品库存
		for _, materialUuid := range materialUuids {
			material := materials[materialUuid]
			if err := base.NewMaterialRepo(tx).UpdateMaterialsStockNum(material.MaterialUuid, material.WarehouseUuid, material.AddStockNum); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		logger.Logger.Error("SubscribeCancelSaleOrderProductEvent process, Transaction failed", zap.Any("saleBillUuid", saleBillUuid), zap.Error(err))
		return
	}
}
