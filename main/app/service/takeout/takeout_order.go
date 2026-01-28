package takeout

import (
	"fmt"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	printer "ttpos-server-go/app/modules/printer"
	printerConst "ttpos-server-go/app/modules/printer/constant"
	"ttpos-server-go/app/modules/printer/printer_model"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	domainService "ttpos-server-go/app/modules/takeout/domain/service"
	valueObject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ITakeoutOrderSrv interface {
	// ProcessTakeoutOrderOutboundAndSales 处理外卖订单出库和销量
	ProcessTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64, acceptedBy uint64) error
	// RestoreTakeoutOrderOutboundAndSales 恢复外卖订单出库和销量（取消订单时调用）
	RestoreTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64) error
	// ProcessOrderItemsStockAndSales 处理订单变动的库存和销量（订单更新时调用）
	// returnItems: 需要归还库存和减少销量的菜品（退菜）
	// kitchenItems: 需要扣减库存和增加销量的菜品（新增/加菜）
	ProcessOrderItemsStockAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64, changeResult *valueObject.OrderChangeResult) error
	// CreateProductionOrderForTakeout 为外卖订单创建送厨单
	CreateProductionOrderForTakeout(ctx context.Context, orderUuid uint64) error
	// UpdateProductionOrderForTakeout 增量更新外卖订单的生产单
	// 处理订单变动时的生产单同步：退菜商品标记为退菜状态，新增商品创建生产单商品
	UpdateProductionOrderForTakeout(ctx context.Context, orderUuid uint64, changeResult *valueObject.OrderChangeResult) error
	// PrintTakeoutOrder 打印外卖订单小票
	PrintTakeoutOrder(ctx context.Context, orderUuid uint64, printLang string, firstExecution int) (*resp.PrinterData, error)
	// 打印送厨单
	PrintProductionOrder(ctx context.Context, orderUuid uint64, printType int, productItems []req.PrintProductItem) (*resp.PrinterData, error)
	// PrintReturnOrder 打印退菜单（使用变更前的商品信息）
	// 注意：此方法使用 changeResult.ReturnItems 中的 OldItem 数据，包含变更前的数量
	PrintReturnOrder(ctx context.Context, orderUuid uint64, changeResult *valueObject.OrderChangeResult) error
	// RecordTakeoutOrderPeakTime 记录外卖订单高峰期
	// 自动根据订单状态判断是增加（inc）还是减少（dec）
	RecordTakeoutOrderPeakTime(ctx context.Context, orderUuid uint64, companyUuid uint64) error
}

// ProcessTakeoutOrderOutboundAndSales 处理外卖订单出库和销量
func (s *takeoutSrv) ProcessTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64, acceptedBy uint64) error {
	db := s.dbm.GetDB(companyUuid)

	// 设置上下文
	ctx.SetDB(db)
	ctx.SetCompanyUuid(companyUuid)

	// 1. 查询订单信息（包含商品列表）
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(orderUuid, orderRepo.WithTakeoutOrderItems(), orderRepo.WithTakeoutOrderItemModifiers())
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 2. 构建出库清单
	decreaseStockList, err := s.buildTakeoutOrderDecreaseStockList(ctx, order)
	if err != nil {
		return errors.WithMessage(err, "构建出库清单失败")
	}

	// 如果没有需要出库的商品，直接返回
	if len(decreaseStockList) == 0 {
		return nil
	}

	// 3. 获取员工班次记录
	staffShiftLogUuid, err := s.getCurrentStaffShiftLogUuid(db, acceptedBy)
	if err != nil {
		logger.Logger.Warn("获取员工班次记录失败", zap.Uint64("staffUuid", acceptedBy), zap.Error(err))
	}

	// 4. 创建出库单
	warehouseOutForms := model.NewWarehouseOutForm(decreaseStockList, false, 0, acceptedBy, staffShiftLogUuid, orderUuid)

	// 5. 在事务中创建出库单和更新销量
	err = db.Transaction(func(tx *gorm.DB) error {
		ctxTx := ctx.Copy()
		ctxTx.SetDB(tx)

		// 5.1 创建出库单
		if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormRecordAll(warehouseOutForms); err != nil {
			return errors.WithMessage(err, "创建出库单失败")
		}

		// 5.2 计算并更新销量
		if err := s.updateTakeoutOrderSalesVolume(ctxTx, order); err != nil {
			logger.Logger.Error("更新外卖订单销量失败", zap.Uint64("orderUuid", order.Uuid), zap.Error(err))
			// 销量更新失败不影响出库流程，只记录日志
		}

		// 5.3 触发库存扣减，并获取实际扣减的材料列表
		err := s.reduceTakeoutOrderStock(tx, companyUuid, order.Uuid)
		if err != nil {
			logger.Logger.Error("扣减外卖订单库存失败", zap.Uint64("orderUuid", order.Uuid), zap.Error(err))
			// 库存扣减失败不影响出库流程，只记录日志
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 同步外卖平台
	s.takeoutAppSrv.SyncMenuChanges(ctx, request.ExportMenuRequest{
		Platform:    order.Platform,
		CompanyUuid: companyUuid,
	})

	return nil
}

// getCurrentStaffShiftLog 获取当前员工班次记录
// 如果 staffUuid 为空，则返回最新的正在当班的班次记录
func (s *takeoutSrv) getCurrentStaffShiftLogUuid(db *gorm.DB, staffUuid uint64) (uint64, error) {
	shiftLogRepo := repository.NewShiftLogRepo(db)

	// 如果 staffUuid 为空，获取最新的正在当班的班次记录
	if staffUuid == 0 {
		shiftLog, err := shiftLogRepo.GetShiftLog(
			func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ?", constant.StaffNotHandedOver)
			},
			func(db *gorm.DB) *gorm.DB {
				return db.Order("uuid DESC")
			},
		)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, nil
			}
			return 0, errors.WithMessage(err, "查询最新班次记录失败")
		}
		return shiftLog.Uuid, nil
	}

	// 获取指定员工的当班记录
	shiftLog, err := shiftLogRepo.GetShiftLog(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("staff_uuid = ?", staffUuid)
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", constant.StaffNotHandedOver)
		},
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, errors.WithMessage(err, "查询员工班次记录失败")
	}
	return shiftLog.Uuid, nil
}

// buildTakeoutOrderDecreaseStockList 从外卖订单构建出库清单
// 优化策略：直接从 ttpos_takeout_order_material 表读取已保存的原料消耗记录
// 通过 product_bom_uuid 字段准确聚合原料到具体的 BOM
func (s *takeoutSrv) buildTakeoutOrderDecreaseStockList(ctx context.Context, order *takeoutModel.TakeoutOrder) ([]*model.Product, error) {
	db := ctx.GetDB()

	// 1. 构建 BOM 数量映射（获取 BOM UUID 和商品数量信息）
	bomQuantityMap, bomItemMap, err := domainService.NewTakeoutOrderSrv(s.dbm).BuildBomQuantityMap(ctx, order)
	if err != nil {
		return nil, errors.WithMessage(err, "构建BOM数量映射失败")
	}
	if len(bomQuantityMap) == 0 {
		return []*model.Product{}, nil
	}

	// 2. 从 ttpos_takeout_order_material 表查询已保存的原料消耗记录
	orderMaterials, err := persistence.NewTakeoutOrderMaterialRepo(db).GetByOrderUuid(order.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单原料消耗记录失败")
	}

	// 3. 按 BOM UUID 聚合原料消耗（使用 product_bom_uuid 字段）
	// 结构：map[bomUuid][]*ProductBomMaterials
	bomMaterialsMap := make(map[uint64][]*model.ProductBomMaterials)
	for _, orderMaterial := range orderMaterials {
		// 跳过被禁用的原料
		if orderMaterial.Material != nil && !orderMaterial.Material.Status {
			continue
		}
		bomUuid := orderMaterial.ProductBomUuid
		if bomUuid == 0 {
			logger.Logger.Warn("原料消耗记录缺少BOM UUID", zap.Uint64("materialUuid", orderMaterial.MaterialUuid), zap.Uint64("orderUuid", order.Uuid))
			continue
		}
		// 聚合原料到对应的 BOM
		if bomMaterialsMap[bomUuid] == nil {
			bomMaterialsMap[bomUuid] = make([]*model.ProductBomMaterials, 0)
		}
		bomMaterialsMap[bomUuid] = append(bomMaterialsMap[bomUuid], &model.ProductBomMaterials{
			MaterialUuid:  orderMaterial.MaterialUuid,
			WarehouseUuid: orderMaterial.WarehouseUuid,
			Num:           orderMaterial.Num,
		})
	}

	// 4. 构建出库清单
	decreaseStockList := make([]*model.Product, 0, len(bomQuantityMap))
	for bomUuid, quantity := range bomQuantityMap {
		productNum := float64(quantity)
		productBomMaterials := bomMaterialsMap[bomUuid]
		if productBomMaterials == nil {
			productBomMaterials = make([]*model.ProductBomMaterials, 0)
		}
		// 获取 PackageUuid
		packageUuid := uint64(0)
		SaleOrderProductUuid := uint64(0)
		if bomItem, ok := bomItemMap[bomUuid]; ok {
			if bomItem.TtposProductType > 0 {
				packageUuid = bomItem.TtposProductPackageUuid
			}
			SaleOrderProductUuid = bomItem.Uuid
		}
		// 添加到出库清单
		if productNum > 0 {
			decreaseStockList = append(decreaseStockList, &model.Product{
				TakeoutOrderUuid:     order.Uuid,
				SaleOrderProductUuid: SaleOrderProductUuid, // 外卖订单没有 SaleOrderProductUuid
				ProductBomUuid:       bomUuid,
				PackageUuid:          packageUuid,
				Num:                  productNum,
				ProductBomMaterials:  productBomMaterials,
			})
		}
	}

	return decreaseStockList, nil
}

// updateTakeoutOrderSalesVolume 更新外卖订单销量
func (s *takeoutSrv) updateTakeoutOrderSalesVolume(ctx context.Context, order *takeoutModel.TakeoutOrder) error {
	db := ctx.GetDB()

	// 计算销量（调用 Domain Service 层方法）
	takeoutOrderSrv := domainService.NewTakeoutOrderSrv(s.dbm)
	productBoms, productPackages, err := takeoutOrderSrv.CalculateTakeoutOrderSalesVolume(order)
	if err != nil {
		return errors.WithMessage(err, "计算销量失败")
	}

	// 更新 BOM 销量
	productBomRepo := repository.NewProductBomRepo(db)
	for bomUuid, saleNum := range productBoms {
		if err := productBomRepo.AddActualSaleNum(bomUuid, saleNum); err != nil {
			logger.Logger.Error("更新BOM销量失败", zap.Uint64("bomUuid", bomUuid), zap.Float64("saleNum", saleNum), zap.Error(err))
			// 继续处理其他BOM，不中断流程
		}
	}

	// 更新 Package 销量
	productPackageRepo := repository.NewProductPackageRepo(db)
	for packageUuid, saleNum := range productPackages {
		if err := productPackageRepo.AddActualSaleNum(packageUuid, saleNum); err != nil {
			logger.Logger.Error("更新Package销量失败", zap.Uint64("packageUuid", packageUuid), zap.Float64("saleNum", saleNum), zap.Error(err))
			// 继续处理其他Package，不中断流程
		}
	}

	return nil
}

// reduceTakeoutOrderStock 扣减外卖订单库存（从出入库记录表读取数据，与结账单保持一致）
func (s *takeoutSrv) reduceTakeoutOrderStock(db *gorm.DB, companyUuid uint64, takeoutOrderUuid uint64) error {
	// 加锁，防止并发扣减库存
	lockKey := fmt.Sprintf("takeout_order_stock:%d:%d", companyUuid, takeoutOrderUuid)
	systemLock := lock.NewSystemLock()
	systemLock.LockUuidString(lockKey)
	defer systemLock.UnlockUuidString(lockKey)

	// Step 1: 从出入库记录表（ttpos_warehouse_out_form_item）查询未减库存的出库记录
	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItemNotProcessedByTakeoutOrderUuid(takeoutOrderUuid)
	if err != nil {
		logger.Logger.Error("reduceTakeoutOrderStock, GetWarehouseOutFormItemNotProcessedByTakeoutOrderUuid failed", zap.Uint64("takeoutOrderUuid", takeoutOrderUuid), zap.Error(err))
		return err
	}
	if len(warehouseOutFormItems) == 0 {
		return nil
	}

	// Step 2: 按类型分组处理（BOM 和 原料）
	ProductBoms := make(map[uint64]*model.ProductBom)
	type StockNum struct {
		MaterialUuid   uint64
		WarehouseUuid  uint64
		ReduceStockNum float64
	}
	Materials := make(map[uint64]*StockNum)

	for _, warehouseOutFormItem := range warehouseOutFormItems {
		if warehouseOutFormItem.IsProductBom() {
			if ProductBoms[warehouseOutFormItem.ProductBomUuid] == nil {
				ProductBoms[warehouseOutFormItem.ProductBomUuid] = warehouseOutFormItem.ProductBom
			}
			ProductBoms[warehouseOutFormItem.ProductBomUuid].StockNum -= warehouseOutFormItem.Num // 扣减库存
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
	for _, productBom := range ProductBoms {
		ProductBomsList = append(ProductBomsList, productBom)
	}

	MaterialSalesVolume := make(map[uint64]float64) // 材料销量 map[材料UUID]销量
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		MaterialSalesVolume[warehouseOutFormItem.MaterialUuid] = decimal.NewFromFloat(MaterialSalesVolume[warehouseOutFormItem.MaterialUuid]).Add(decimal.NewFromFloat(warehouseOutFormItem.Num)).Round(4).InexactFloat64()
	}

	// Step 3: 在事务中更新库存
	err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新出库单明细状态（标记为已减库存）
		if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormItemRecordsReduceStockByTakeoutOrderUuid(takeoutOrderUuid); err != nil {
			logger.Logger.Error("reduceTakeoutOrderStock, UpdateWarehouseOutFormItemRecordsReduceStock failed", zap.Uint64("takeoutOrderUuid", takeoutOrderUuid), zap.Error(err))
			return err
		}

		// 更新 BOM 库存
		if err := repository.NewProductBomRepo(tx).UpdateProductBoms(ProductBomsList); err != nil {
			logger.Logger.Error("reduceTakeoutOrderStock, UpdateProductBoms failed", zap.Uint64("takeoutOrderUuid", takeoutOrderUuid), zap.Error(err))
			return err
		}

		// 更新材料库存
		for _, material := range Materials {
			if err := base.NewMaterialRepo(tx).UpdateMaterialsStockNum(material.MaterialUuid, material.WarehouseUuid, -material.ReduceStockNum); err != nil {
				logger.Logger.Error("reduceTakeoutOrderStock, UpdateMaterialsStockNum failed", zap.Uint64("takeoutOrderUuid", takeoutOrderUuid), zap.Error(err))
				return err
			}
		}

		// 通过sale_order_uuid查询出库单明细中有效的出库材料,然后统计每个材料的销量
		for materialUuid, saleNum := range MaterialSalesVolume {
			if err := repository.NewMaterialRepo(tx).AddActualSaleNum(materialUuid, saleNum); err != nil {
				logger.Logger.Error("HandleAddMaterialSalesVolume process, AddActualSaleNum failed", zap.Any("materialUuid", materialUuid), zap.Any("saleNum", saleNum), zap.Error(err))
				continue
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// RestoreTakeoutOrderOutboundAndSales 恢复外卖订单出库和销量（取消订单时调用）
func (s *takeoutSrv) RestoreTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)

	// 设置上下文
	ctx.SetDB(db)
	ctx.SetCompanyUuid(companyUuid)

	// 1. 查询订单信息（包含商品列表）
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(orderUuid, orderRepo.WithTakeoutOrderItems(), orderRepo.WithTakeoutOrderItemModifiers())
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 2. 查询该订单的所有出库单（通过 takeout_order_uuid）
	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItem(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("takeout_order_uuid = ? AND revoke_time = ?", orderUuid, 0)
		},
	)
	if err != nil {
		return errors.WithMessage(err, "查询出库单明细失败")
	}

	if len(warehouseOutFormItems) == 0 {
		// 没有出库记录，直接返回
		return nil
	}

	// 3. 获取出库单UUID列表，用于撤销出库单
	warehouseOutFormUuids := make(map[uint64]bool)
	for _, item := range warehouseOutFormItems {
		warehouseOutFormUuids[item.WarehouseOutFormUuid] = true
	}

	// 4. 查询出库单记录
	warehouseOutForms := make([]*model.WarehouseOutForm, 0)
	for formUuid := range warehouseOutFormUuids {
		form, err := warehouseFormRepo.GetWarehouseOutForms(
			func(db *gorm.DB) *gorm.DB {
				return db.Where("uuid = ?", formUuid)
			},
			repository.CommonRepo.Preload(
				repository.WithPreload{
					Query: "WarehouseOutFormItems",
				},
			),
		)
		if err != nil {
			logger.Logger.Warn("查询出库单失败", zap.Uint64("formUuid", formUuid), zap.Error(err))
			continue
		}
		if len(form) > 0 {
			warehouseOutForms = append(warehouseOutForms, form[0])
		}
	}

	// 5. 在事务中撤销出库单、恢复库存、减少销量
	err = db.Transaction(func(tx *gorm.DB) error {
		ctxTx := ctx.Copy()
		ctxTx.SetDB(tx)

		// 5.1 撤销出库单
		for _, form := range warehouseOutForms {
			form.RevokeForm()
			if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormRecord(*form); err != nil {
				return errors.WithMessage(err, "撤销出库单失败")
			}
			// 撤销出库单明细
			for _, item := range form.WarehouseOutFormItems {
				if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormItemRecord(*item); err != nil {
					return errors.WithMessage(err, "撤销出库单明细失败")
				}
			}
		}

		// 5.2 恢复库存
		if err := s.restoreTakeoutOrderStock(tx, companyUuid, warehouseOutFormItems); err != nil {
			logger.Logger.Error("恢复外卖订单库存失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			// 库存恢复失败不影响流程，只记录日志
		}

		// 5.3 减少销量
		if err := s.reduceTakeoutOrderSalesVolume(ctxTx, order); err != nil {
			logger.Logger.Error("减少外卖订单销量失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			// 销量减少失败不影响流程，只记录日志
		}

		// 5.4 标记外卖订单原料为已汇总，避免被日终统计重复处理
		takeoutOrderMaterialRepo := persistence.NewTakeoutOrderMaterialRepo(tx)
		if err := takeoutOrderMaterialRepo.MarkTakeoutOrderMaterialsAsSummarized(orderUuid); err != nil {
			logger.Logger.Error("标记外卖订单原料为已汇总失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			// 标记失败不影响流程，只记录日志
		}

		// 5.5 删除关联的生产订单（ttpos_production_order 和 ttpos_production_order_product）
		productionRepo := repository.NewProductionRepo(tx)
		if err := productionRepo.DeleteProductionOrderByTakeoutOrderUuid(orderUuid); err != nil {
			logger.Logger.Error("删除外卖订单关联的生产订单失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			// 删除失败不影响流程，只记录日志
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 同步外卖平台
	s.takeoutAppSrv.SyncMenuChanges(ctx, request.ExportMenuRequest{
		Platform:    order.Platform,
		CompanyUuid: companyUuid,
	})

	return nil
}

// restoreTakeoutOrderStock 恢复外卖订单库存
func (s *takeoutSrv) restoreTakeoutOrderStock(db *gorm.DB, companyUuid uint64, warehouseOutFormItems []*model.WarehouseOutFormItem) error {
	// 按类型分组处理
	productBoms := make(map[uint64]*model.ProductBom)
	materials := make(map[uint64]map[uint64]float64) // map[materialUuid]map[warehouseUuid]restoreStockNum

	for _, item := range warehouseOutFormItems {
		// 只处理已减库存的出库单明细
		if item.ReduceStock != constant.WarehouseOutFormItemReduceStockSuccess {
			continue
		}

		if item.IsProductBom() {
			// 查询 BOM 信息
			if productBoms[item.ProductBomUuid] == nil {
				productBomRepo := repository.NewProductBomRepo(db)
				bom, err := productBomRepo.GetFlavorProductBomByUuid(companyUuid, item.ProductBomUuid)
				if err != nil {
					logger.Logger.Error("查询BOM信息失败", zap.Uint64("bomUuid", item.ProductBomUuid), zap.Error(err))
					continue
				}
				productBoms[item.ProductBomUuid] = bom
			}

			// 恢复 BOM 库存
			if productBoms[item.ProductBomUuid] != nil {
				productBoms[item.ProductBomUuid].StockNum += item.Num
			}
		} else if item.IsMaterial() {
			// 恢复材料库存
			if materials[item.MaterialUuid] == nil {
				materials[item.MaterialUuid] = make(map[uint64]float64)
			}
			materials[item.MaterialUuid][item.WarehouseUuid] += item.Num
		}
	}

	// 更新库存
	// 更新 BOM 库存
	productBomList := make([]*model.ProductBom, 0, len(productBoms))
	for _, bom := range productBoms {
		productBomList = append(productBomList, bom)
	}
	if len(productBomList) > 0 {
		if err := repository.NewProductBomRepo(db).UpdateProductBoms(productBomList); err != nil {
			return errors.WithMessage(err, "恢复BOM库存失败")
		}
	}

	// 恢复材料库存
	for materialUuid, warehouseMap := range materials {
		for warehouseUuid, restoreStockNum := range warehouseMap {
			if err := base.NewMaterialRepo(db).UpdateMaterialsStockNum(materialUuid, warehouseUuid, restoreStockNum); err != nil {
				return errors.WithMessage(err, "恢复材料库存失败")
			}
		}
	}

	// 通过sale_order_uuid查询出库单明细中有效的出库材料,然后统计每个材料的销量
	materialSalesVolumes := make(map[uint64]float64)
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		materialSalesVolumes[warehouseOutFormItem.MaterialUuid] = decimal.NewFromFloat(materialSalesVolumes[warehouseOutFormItem.MaterialUuid]).Add(decimal.NewFromFloat(warehouseOutFormItem.Num)).Round(4).InexactFloat64()
	}
	for materialUuid, saleNum := range materialSalesVolumes {
		if err := repository.NewMaterialRepo(db).AddActualSaleNum(materialUuid, -saleNum); err != nil {
			logger.Logger.Error("restoreTakeoutOrderStock process, AddActualSaleNum failed", zap.Any("materialUuid", materialUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}

	return nil
}

// reduceTakeoutOrderSalesVolume 减少外卖订单销量
func (s *takeoutSrv) reduceTakeoutOrderSalesVolume(ctx context.Context, order *takeoutModel.TakeoutOrder) error {
	db := ctx.GetDB()

	// 计算销量（调用 Domain Service 层方法）
	takeoutOrderSrv := domainService.NewTakeoutOrderSrv(s.dbm)
	productBoms, productPackages, err := takeoutOrderSrv.CalculateTakeoutOrderSalesVolume(order)
	if err != nil {
		return errors.WithMessage(err, "计算销量失败")
	}

	// 减少 BOM 销量
	productBomRepo := repository.NewProductBomRepo(db)
	for bomUuid, saleNum := range productBoms {
		if err := productBomRepo.SubActualSaleNum(bomUuid, saleNum); err != nil {
			logger.Logger.Error("减少BOM销量失败", zap.Uint64("bomUuid", bomUuid), zap.Float64("saleNum", saleNum), zap.Error(err))
			// 继续处理其他BOM，不中断流程
		}
	}

	// 减少 Package 销量
	productPackageRepo := repository.NewProductPackageRepo(db)
	for packageUuid, saleNum := range productPackages {
		if err := productPackageRepo.SubActualSaleNum(packageUuid, saleNum); err != nil {
			logger.Logger.Error("减少Package销量失败", zap.Uint64("packageUuid", packageUuid), zap.Float64("saleNum", saleNum), zap.Error(err))
			// 继续处理其他Package，不中断流程
		}
	}

	return nil
}

// ProcessOrderItemsStockAndSales 处理订单变动的库存和销量
// 退菜项：归还库存 + 减少销量
// 送厨项：扣减库存 + 增加销量
func (s *takeoutSrv) ProcessOrderItemsStockAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64, changeResult *valueObject.OrderChangeResult) error {
	if changeResult == nil || !changeResult.HasChange {
		return nil
	}

	db := s.dbm.GetDB(companyUuid)
	ctx.SetDB(db)
	ctx.SetCompanyUuid(companyUuid)

	// 查询订单信息
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(orderUuid, orderRepo.WithTakeoutOrderItems(), orderRepo.WithTakeoutOrderItemModifiers())
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 在事务中处理库存和销量变动
	err = db.Transaction(func(tx *gorm.DB) error {
		ctxTx := ctx.Copy()
		ctxTx.SetDB(tx)

		// 1. 处理退菜项：归还库存 + 减少销量
		if len(changeResult.ReturnItems) > 0 {
			if err := s.processReturnItemsStock(ctxTx, order, changeResult.ReturnItems); err != nil {
				logger.Logger.Error("处理退菜项库存失败",
					zap.Uint64("orderUuid", orderUuid),
					zap.Error(err))
				// 不中断流程，继续处理其他项
			}
		}

		// 2. 处理送厨项：扣减库存 + 增加销量
		if len(changeResult.KitchenItems) > 0 {
			if err := s.processKitchenItemsStock(ctxTx, order, changeResult.KitchenItems); err != nil {
				logger.Logger.Error("处理送厨项库存失败",
					zap.Uint64("orderUuid", orderUuid),
					zap.Error(err))
				// 不中断流程
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 同步外卖平台库存
	s.takeoutAppSrv.SyncMenuChanges(ctx, request.ExportMenuRequest{
		Platform:    order.Platform,
		CompanyUuid: companyUuid,
	})

	return nil
}

// processReturnItemsStock 处理退菜项的库存归还和销量减少
func (s *takeoutSrv) processReturnItemsStock(ctx context.Context, order *takeoutModel.TakeoutOrder, returnItems []valueObject.ItemChange) error {
	db := ctx.GetDB()

	// 收集需要处理的 BOM UUID 和对应的数量
	bomQuantityMap := make(map[uint64]int) // map[bomUuid]quantity
	for _, item := range returnItems {
		if item.OldItem != nil && item.OldItem.TtposProductPackageUuid > 0 {
			// 对于退菜，使用原数量
			bomQuantityMap[item.OldItem.TtposProductPackageUuid] += item.OldQuantity
		}
	}

	if len(bomQuantityMap) == 0 {
		return nil
	}

	// 查询该订单的出库单明细
	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItem(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("takeout_order_uuid = ? AND revoke_time = ?", order.Uuid, 0)
		},
	)
	if err != nil {
		return errors.WithMessage(err, "查询出库单明细失败")
	}

	// 按 BOM UUID 筛选需要撤销的出库单明细
	itemsToRevoke := make([]*model.WarehouseOutFormItem, 0)
	for _, outItem := range warehouseOutFormItems {
		if _, exists := bomQuantityMap[outItem.ProductBomUuid]; exists {
			itemsToRevoke = append(itemsToRevoke, outItem)
		}
	}

	if len(itemsToRevoke) == 0 {
		return nil
	}

	// 撤销出库单明细
	for _, item := range itemsToRevoke {
		item.RevokeTime = time.Now().Unix()
		if err := warehouseFormRepo.UpdateWarehouseOutFormItemRecord(*item); err != nil {
			logger.Logger.Error("撤销出库单明细失败",
				zap.Uint64("itemUuid", item.Uuid),
				zap.Error(err))
		}
	}

	// 恢复库存
	if err := s.restoreTakeoutOrderStock(db, ctx.GetCompanyUuid(), itemsToRevoke); err != nil {
		logger.Logger.Error("恢复退菜项库存失败",
			zap.Uint64("orderUuid", order.Uuid),
			zap.Error(err))
	}

	// 减少销量
	productBomRepo := repository.NewProductBomRepo(db)
	productPackageRepo := repository.NewProductPackageRepo(db)
	for bomUuid, quantity := range bomQuantityMap {
		saleNum := float64(quantity)
		// 减少 BOM 销量
		if err := productBomRepo.SubActualSaleNum(bomUuid, saleNum); err != nil {
			logger.Logger.Error("减少退菜项BOM销量失败",
				zap.Uint64("bomUuid", bomUuid),
				zap.Float64("saleNum", saleNum),
				zap.Error(err))
		}
		// 减少 Package 销量（如果是套餐）
		for _, item := range returnItems {
			if item.OldItem != nil && item.OldItem.TtposProductPackageUuid == bomUuid && item.OldItem.TtposProductType > 0 {
				if err := productPackageRepo.SubActualSaleNum(bomUuid, saleNum); err != nil {
					logger.Logger.Error("减少退菜项Package销量失败",
						zap.Uint64("packageUuid", bomUuid),
						zap.Float64("saleNum", saleNum),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

// processKitchenItemsStock 处理送厨项的库存扣减和销量增加
func (s *takeoutSrv) processKitchenItemsStock(ctx context.Context, order *takeoutModel.TakeoutOrder, kitchenItems []valueObject.ItemChange) error {
	db := ctx.GetDB()

	// 收集需要处理的商品和数量
	decreaseStockList := make([]*model.Product, 0)
	bomQuantityMap := make(map[uint64]int) // map[bomUuid]quantity

	for _, item := range kitchenItems {
		if item.NewItem == nil || item.NewItem.TtposProductPackageUuid == 0 {
			continue
		}

		// 计算需要处理的数量（新增或增量）
		quantity := item.NewQuantity
		if item.ChangeType == valueObject.ChangeTypeQuantity {
			// 数量变动只处理增量部分
			quantity = item.NewQuantity - item.OldQuantity
		}

		if quantity <= 0 {
			continue
		}

		bomUuid := item.NewItem.TtposProductPackageUuid
		bomQuantityMap[bomUuid] += quantity

		// 构建出库清单项
		packageUuid := uint64(0)
		if item.NewItem.TtposProductType > 0 {
			packageUuid = item.NewItem.TtposProductPackageUuid
		}

		decreaseStockList = append(decreaseStockList, &model.Product{
			TakeoutOrderUuid:     order.Uuid,
			SaleOrderProductUuid: item.NewItem.Uuid,
			ProductBomUuid:       bomUuid,
			PackageUuid:          packageUuid,
			Num:                  float64(quantity),
		})
	}

	if len(decreaseStockList) == 0 {
		return nil
	}

	// 获取员工班次记录
	staffShiftLogUuid, err := s.getCurrentStaffShiftLogUuid(db, 0)
	if err != nil {
		logger.Logger.Warn("获取员工班次记录失败", zap.Error(err))
	}

	// 创建出库单
	warehouseOutForms := model.NewWarehouseOutForm(decreaseStockList, false, 0, 0, staffShiftLogUuid, order.Uuid)
	if err := repository.NewWarehouseFormRepo(db).CreateWarehouseOutFormRecordAll(warehouseOutForms); err != nil {
		logger.Logger.Error("创建送厨项出库单失败",
			zap.Uint64("orderUuid", order.Uuid),
			zap.Error(err))
	}

	// 扣减库存
	if err := s.reduceTakeoutOrderStock(db, ctx.GetCompanyUuid(), order.Uuid); err != nil {
		logger.Logger.Error("扣减送厨项库存失败",
			zap.Uint64("orderUuid", order.Uuid),
			zap.Error(err))
	}

	// 增加销量
	productBomRepo := repository.NewProductBomRepo(db)
	productPackageRepo := repository.NewProductPackageRepo(db)
	for bomUuid, quantity := range bomQuantityMap {
		saleNum := float64(quantity)
		// 增加 BOM 销量
		if err := productBomRepo.AddActualSaleNum(bomUuid, saleNum); err != nil {
			logger.Logger.Error("增加送厨项BOM销量失败",
				zap.Uint64("bomUuid", bomUuid),
				zap.Float64("saleNum", saleNum),
				zap.Error(err))
		}
		// 增加 Package 销量（如果是套餐）
		for _, item := range kitchenItems {
			if item.NewItem != nil && item.NewItem.TtposProductPackageUuid == bomUuid && item.NewItem.TtposProductType > 0 {
				if err := productPackageRepo.AddActualSaleNum(bomUuid, saleNum); err != nil {
					logger.Logger.Error("增加送厨项Package销量失败",
						zap.Uint64("packageUuid", bomUuid),
						zap.Float64("saleNum", saleNum),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

// CreateProductionOrderForTakeout 为外卖订单创建送厨单
func (s *takeoutSrv) CreateProductionOrderForTakeout(ctx context.Context, orderUuid uint64) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 查询外卖订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(orderUuid, orderRepo.WithTakeoutOrderItems(), orderRepo.WithTakeoutOrderItemModifiers())
	if err != nil {
		return errors.WithMessage(err, "查询外卖订单失败")
	}
	if order == nil {
		return errors.New("外卖订单不存在")
	}

	// 生成 ProductionOrder UUID
	productionOrderUuid, err := utils.GetID()
	if err != nil {
		return errors.WithMessage(err, "生成送厨单UUID失败")
	}

	// 初始化 Repo
	productionRepo := repository.NewProductionRepo(db)
	productPackageRepo := repository.NewProductPackageRepo(db)
	bomMappingRepo := persistence.NewTakeoutBomMappingRepo(db)

	// 1. 收集所有需要查询的 UUID
	var (
		normalProductPackageUuids = make(map[uint64]bool) // 普通商品的 ProductPackage UUID（用于查询分类）
		packageGroupItemUuids     = make([]uint64, 0)     // 套餐商品的 groupItem UUID（用于查询 RelatedUuid）
	)

	// 遍历所有商品，收集需要查询的 UUID
	for _, takeoutItem := range order.TakeoutOrderItems {
		if takeoutItem.TtposProductType == 0 {
			// 普通商品：收集 ProductPackage UUID（仅用于查询分类）
			if takeoutItem.TtposProductPackageUuid > 0 {
				normalProductPackageUuids[takeoutItem.TtposProductPackageUuid] = true
			}
		} else {
			// 套餐商品：收集 groupItem UUID（用于查询子商品 UUID）
			for _, modifier := range takeoutItem.TakeoutOrderItemModifiers {
				if modifier.IsMapped == 1 && modifier.TtposModifierType == "commodity" && modifier.TtposModifierUuid > 0 {
					packageGroupItemUuids = append(packageGroupItemUuids, modifier.TtposModifierUuid)
				}
			}
		}
	}

	// 2. 批量查询所有需要的数据
	// 2.1 批量查询普通商品的 ProductPackage（预加载 ProductCategory）
	normalPackageUuidList := make([]uint64, 0, len(normalProductPackageUuids))
	for uuid := range normalProductPackageUuids {
		normalPackageUuidList = append(normalPackageUuidList, uuid)
	}
	normalProductPackages, err := productPackageRepo.GetProductPackageList(
		repository.CommonRepo.WhereInUuids(normalPackageUuidList),
		productPackageRepo.WithProductCategory(),
	)
	if err != nil {
		logger.Logger.Error("批量查询普通商品套餐失败", zap.Error(err))
		return errors.WithMessage(err, "批量查询普通商品套餐失败")
	}
	normalProductPackageMap := make(map[uint64]*model.ProductPackage)
	for _, pkg := range normalProductPackages {
		normalProductPackageMap[pkg.Uuid] = pkg
	}

	// 2.2 批量查询套餐商品的 BOM 映射（仅需要 RelatedUuid）
	var groupItemBomMapping map[uint64]persistence.GroupItemBomMapping
	var packageSubProductPackageUuids = make(map[uint64]bool)
	if len(packageGroupItemUuids) > 0 {
		groupItemBomMapping, err = bomMappingRepo.GetGroupItemBomMapping(packageGroupItemUuids)
		if err != nil {
			logger.Logger.Error("批量查询套餐商品BOM映射失败", zap.Error(err))
			return errors.WithMessage(err, "批量查询套餐商品BOM映射失败")
		}
		// 收集套餐子商品的 ProductPackage UUID（用于查询分类）
		for _, mapping := range groupItemBomMapping {
			if mapping.RelatedUuid > 0 {
				packageSubProductPackageUuids[mapping.RelatedUuid] = true
			}
		}
	}

	// 2.3 批量查询套餐子商品的 ProductPackage（预加载 ProductCategory）
	packageSubPackageUuidList := make([]uint64, 0, len(packageSubProductPackageUuids))
	for uuid := range packageSubProductPackageUuids {
		packageSubPackageUuidList = append(packageSubPackageUuidList, uuid)
	}
	packageSubProductPackages, err := productPackageRepo.GetProductPackageList(
		repository.CommonRepo.WhereInUuids(packageSubPackageUuidList),
		productPackageRepo.WithProductCategory(),
	)
	if err != nil {
		logger.Logger.Error("批量查询套餐子商品套餐失败", zap.Error(err))
		return errors.WithMessage(err, "批量查询套餐子商品套餐失败")
	}
	packageSubProductPackageMap := make(map[uint64]*model.ProductPackage)
	for _, pkg := range packageSubProductPackages {
		packageSubProductPackageMap[pkg.Uuid] = pkg
	}

	// 3. 创建 ProductionOrderProduct 列表
	productionOrderProducts := make([]*model.ProductionOrderProduct, 0)

	for _, takeoutItem := range order.TakeoutOrderItems {
		if takeoutItem.TtposProductType == 0 {
			// 普通商品处理
			productPackage, ok := normalProductPackageMap[takeoutItem.TtposProductPackageUuid]
			if !ok {
				logger.Logger.Warn("未找到普通商品套餐", zap.Uint64("productPackageUuid", takeoutItem.TtposProductPackageUuid))
				continue
			}

			// 生成 ProductionOrderProduct UUID
			productUuid, err := utils.GetID()
			if err != nil {
				logger.Logger.Error("生成送厨商品UUID失败", zap.Error(err))
				continue
			}

			// 获取商品信息
			productPackageUuid := productPackage.Uuid
			var firstCategoryUuid uint64
			if productPackage.ProductCategory.Uuid > 0 {
				firstCategoryUuid = productPackage.ProductCategory.GetFirstCategoryUuid()
			}

			// 处理 modifiers，直接从 modifier 中获取已保存的名称
			var productBomUuid uint64
			var productBomName string
			var attributeNames []string

			for _, modifier := range takeoutItem.TakeoutOrderItemModifiers {
				if modifier.IsMapped == 0 || modifier.TtposModifierUuid == 0 {
					continue
				}

				switch modifier.TtposModifierType {
				case string(valueObject.ModifierTypeFlavor):
					// 规格：直接使用已保存的 TtposFlavorName
					productBomUuid = modifier.TtposModifierUuid
					productBomName = modifier.TtposModifierName
				case string(valueObject.ModifierTypeAttr), string(valueObject.ModifierTypeSauce):
					// 属性和加料：直接使用已保存的 TtposModifierName
					if modifier.TtposModifierName != "" {
						attributeNames = append(attributeNames, modifier.TtposModifierName)
					}
				}
			}

			// 合并多个属性的名称
			var attributeNamesStr string
			if len(attributeNames) > 0 {
				// 假设 attributeNames 已经是多语言JSON格式，需要合并
				localeResponses := make([]dto.LocaleResponse, 0, len(attributeNames))
				for _, attrName := range attributeNames {
					localeResp := language.JsonToLocaleResponse(attrName)
					if localeResp != nil && !localeResp.IsNull() {
						localeResponses = append(localeResponses, *localeResp)
					}
				}
				if len(localeResponses) > 0 {
					mergedLocale := language.MergeLocaleResponses(localeResponses, ";")
					attributeNamesStr = mergedLocale.ToJson()
				}
			}

			// 使用商品名称（优先使用 TTPOS 商品名称）
			itemName := takeoutItem.TtposItemName
			if itemName == "" {
				itemName = takeoutItem.ItemName
			}

			// 创建 ProductionOrderProduct
			productionOrderProduct := &model.ProductionOrderProduct{
				BaseModel:             model.BaseModel{Uuid: productUuid, CreateTime: currentTime, UpdateTime: currentTime},
				Name:                  itemName,
				Num:                   float64(takeoutItem.Quantity),
				InitNum:               float64(takeoutItem.Quantity),
				FlavorName:            productBomName,
				ProductBomUuid:        productBomUuid,
				ProductAttributeNames: attributeNamesStr,
				ProductSaucesNames:    "",
				Status:                constant.ProductionOrderProductStatusCooking,
				Remark:                "",
				HasMaterial:           0,
				ProductPackageUuid:    productPackageUuid,
				ProductionOrderUuid:   productionOrderUuid,
				TakeoutOrderUuid:      order.Uuid,
				TakeoutOrderItemUuid:  takeoutItem.Uuid,
				FirstCategoryUuid:     firstCategoryUuid,
			}
			productionOrderProducts = append(productionOrderProducts, productionOrderProduct)
		} else {
			// 套餐商品处理
			// 收集所有 commodity modifiers
			commodityModifiers := make([]*takeoutModel.TakeoutOrderItemModifier, 0)
			for i := range takeoutItem.TakeoutOrderItemModifiers {
				modifier := &takeoutItem.TakeoutOrderItemModifiers[i]
				if modifier.IsMapped == 1 && modifier.TtposModifierType == "commodity" && modifier.TtposModifierUuid > 0 {
					commodityModifiers = append(commodityModifiers, modifier)
				}
			}

			if len(commodityModifiers) == 0 {
				continue
			}

			// 为每个 commodity modifier 创建一个 ProductionOrderProduct
			for _, commodityModifier := range commodityModifiers {
				groupItemUuid := commodityModifier.TtposModifierUuid
				mapping, ok := groupItemBomMapping[groupItemUuid]
				if !ok {
					logger.Logger.Warn("未找到套餐商品BOM映射", zap.Uint64("groupItemUuid", groupItemUuid))
					continue
				}

				// 获取子商品的 ProductPackage（仅用于获取分类）
				subProductPackage, ok := packageSubProductPackageMap[mapping.RelatedUuid]
				if !ok {
					logger.Logger.Error("未找到子商品套餐", zap.Uint64("relatedUuid", mapping.RelatedUuid))
					continue
				}

				productPackageUuid := subProductPackage.Uuid
				var firstCategoryUuid uint64
				if subProductPackage.ProductCategory.Uuid > 0 {
					firstCategoryUuid = subProductPackage.ProductCategory.GetFirstCategoryUuid()
				}

				// 直接从 modifier 中获取已保存的数据
				productBomUuid := commodityModifier.TtposFlavorProductBomUuid // 规格UUID
				productBomName := commodityModifier.TtposFlavorName           // 规格名称
				itemName := commodityModifier.TtposModifierName               // 商品名称
				productNum := float64(commodityModifier.Quantity)             // 数量（已在创建订单时设置为 groupItem.Num * takeoutItem.Quantity）

				// 如果没有 TTPOS 商品名称，回退到平台名称
				if itemName == "" {
					itemName = commodityModifier.ModifierName
				}

				// 生成 ProductionOrderProduct UUID
				productUuid, err := utils.GetID()
				if err != nil {
					logger.Logger.Error("生成送厨商品UUID失败", zap.Error(err))
					continue
				}

				// 创建 ProductionOrderProduct
				productionOrderProduct := &model.ProductionOrderProduct{
					BaseModel:             model.BaseModel{Uuid: productUuid, CreateTime: currentTime, UpdateTime: currentTime},
					Name:                  itemName,
					Num:                   productNum,
					InitNum:               productNum,
					FlavorName:            productBomName,
					ProductBomUuid:        productBomUuid,
					ProductAttributeNames: productBomName,
					ProductSaucesNames:    "",
					Status:                constant.ProductionOrderProductStatusCooking,
					Remark:                "",
					HasMaterial:           0,
					ProductPackageUuid:    productPackageUuid,
					ProductionOrderUuid:   productionOrderUuid,
					TakeoutOrderUuid:      order.Uuid,
					TakeoutOrderItemUuid:  takeoutItem.Uuid,
					FirstCategoryUuid:     firstCategoryUuid,
				}
				productionOrderProducts = append(productionOrderProducts, productionOrderProduct)
			}
		}
	}

	// 4. 创建 ProductionOrder
	productionOrder := &model.ProductionOrder{
		BaseModel:               model.BaseModel{Uuid: productionOrderUuid, CreateTime: currentTime, UpdateTime: currentTime},
		TakeoutOrderUuid:        order.Uuid,
		Source:                  order.Platform,
		ProductionOrderProducts: productionOrderProducts,
	}

	// 5. 保存到数据库
	if err := productionRepo.CreateProductionOrder(productionOrder); err != nil {
		return errors.WithMessage(err, "创建送厨单失败")
	}

	return nil
}

// UpdateProductionOrderForTakeout 增量更新外卖订单的生产单
// 处理订单变动时的生产单同步：
// - 退菜商品：标记为退菜状态 (Status = 3)
// - 新增商品：创建新的生产单商品
// - 数量变更：更新生产单商品数量
func (s *takeoutSrv) UpdateProductionOrderForTakeout(ctx context.Context, orderUuid uint64, changeResult *valueObject.OrderChangeResult) error {
	if changeResult == nil || !changeResult.HasChange {
		return nil
	}

	db := ctx.GetDB()
	productionRepo := repository.NewProductionRepo(db)

	// 1. 查询现有的生产单
	productionOrder, err := productionRepo.GetProductionOrderByTakeoutOrderUuid(orderUuid)
	if err != nil {
		logger.Logger.Error("查询生产单失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
		return errors.WithMessage(err, "查询生产单失败")
	}
	if productionOrder == nil || productionOrder.Uuid == 0 {
		logger.Logger.Warn("生产单不存在，跳过更新", zap.Uint64("orderUuid", orderUuid))
		return nil
	}

	// 2. 处理送厨商品（新增或数量/属性变更）
	for _, item := range changeResult.KitchenItems {
		switch item.ChangeType {
		case valueObject.ChangeTypeAdded:
			// 新增商品：创建生产单商品（使用完整字段）
			if item.NewItem == nil || item.NewItem.Uuid == 0 {
				logger.Logger.Warn("新增商品缺少信息，跳过创建生产单商品",
					zap.String("platformItemId", item.PlatformItemId))
				continue
			}

			// 使用辅助方法创建完整的生产单商品
			productionOrderProduct, err := s.buildProductionOrderProductFromTakeoutItem(
				ctx, item.NewItem.Uuid, productionOrder.Uuid, orderUuid,
			)
			if err != nil {
				logger.Logger.Error("构建生产单商品失败",
					zap.Uint64("takeoutOrderItemUuid", item.NewItem.Uuid),
					zap.Error(err))
				continue
			}
			if productionOrderProduct == nil {
				logger.Logger.Warn("新增商品未映射，跳过创建生产单商品",
					zap.String("platformItemId", item.PlatformItemId))
				continue
			}

			if err := productionRepo.CreateProductionOrderProduct(productionOrderProduct); err != nil {
				logger.Logger.Error("创建生产单商品失败",
					zap.Uint64("takeoutOrderItemUuid", item.NewItem.Uuid),
					zap.Error(err))
			}

		case valueObject.ChangeTypeQuantity, valueObject.ChangeTypeAttribute:
			// 数量变更或属性变更（加料等）：退旧增新
			// 1. 旧商品标记退菜状态
			if item.OldItem != nil && item.OldItem.Uuid > 0 {
				if err := productionRepo.UpdateProductionOrderProductNumByTakeoutItemUuid(
					item.OldItem.Uuid,
					0,
				); err != nil {
					logger.Logger.Error("变更商品标记退菜失败",
						zap.Uint64("takeoutOrderItemUuid", item.OldItem.Uuid),
						zap.String("changeType", item.ChangeType.String()),
						zap.Error(err))
				}
			}

			productionOrderProduct, err := s.buildProductionOrderProductFromTakeoutItem(
				ctx, item.OldItem.Uuid, productionOrder.Uuid, orderUuid,
			)
			if err != nil {
				logger.Logger.Error("构建生产单商品失败",
					zap.Uint64("takeoutOrderItemUuid", item.NewItem.Uuid),
					zap.String("changeType", item.ChangeType.String()),
					zap.Error(err))
				continue
			}
			if productionOrderProduct == nil {
				continue
			}

			if err := productionRepo.CreateProductionOrderProduct(productionOrderProduct); err != nil {
				logger.Logger.Error("变更商品创建生产单商品失败",
					zap.Uint64("takeoutOrderItemUuid", item.NewItem.Uuid),
					zap.String("changeType", item.ChangeType.String()),
					zap.Error(err))
			}

		case valueObject.ChangeTypeRemoved:
			// 数量变更或属性变更（加料等）：退旧增新
			// 1. 旧商品标记退菜状态
			if item.OldItem != nil && item.OldItem.Uuid > 0 {
				if err := productionRepo.UpdateProductionOrderProductNumByTakeoutItemUuid(
					item.OldItem.Uuid,
					0,
				); err != nil {
					logger.Logger.Error("变更商品标记退菜失败",
						zap.Uint64("takeoutOrderItemUuid", item.OldItem.Uuid),
						zap.String("changeType", item.ChangeType.String()),
						zap.Error(err))
				}
			}
		}
	}

	// 2. 处理送厨商品（新增或数量/属性变更）
	for _, item := range changeResult.ReturnItems {
		switch item.ChangeType {
		case valueObject.ChangeTypeRemoved:
			// 数量变更或属性变更（加料等）：退旧增新
			// 1. 旧商品标记退菜状态
			if item.OldItem != nil && item.OldItem.Uuid > 0 {
				if err := productionRepo.UpdateProductionOrderProductNumByTakeoutItemUuid(
					item.OldItem.Uuid,
					0,
				); err != nil {
					logger.Logger.Error("变更商品标记退菜失败",
						zap.Uint64("takeoutOrderItemUuid", item.OldItem.Uuid),
						zap.String("changeType", item.ChangeType.String()),
						zap.Error(err))
				}
			}
		}
	}
	return nil
}

// buildProductionOrderProductFromTakeoutItem 从外卖订单商品构建生产单商品
// 用于增量更新生产单时创建新的生产单商品
func (s *takeoutSrv) buildProductionOrderProductFromTakeoutItem(
	ctx context.Context,
	takeoutItemUuid uint64,
	productionOrderUuid uint64,
	takeoutOrderUuid uint64,
) (*model.ProductionOrderProduct, error) {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 1. 查询外卖订单商品（包含 modifiers）
	takeoutItemRepo := persistence.NewTakeoutOrderItemRepo(db)
	takeoutItem, err := takeoutItemRepo.GetByUuid(takeoutItemUuid, takeoutItemRepo.WithModifiers())
	if err != nil {
		return nil, errors.WithMessage(err, "查询外卖订单商品失败")
	}
	if takeoutItem == nil {
		return nil, errors.New("外卖订单商品不存在")
	}

	// 如果没有映射的商品套餐，返回 nil
	if takeoutItem.TtposProductPackageUuid == 0 {
		return nil, nil
	}

	// 2. 生成 UUID
	productUuid, err := utils.GetID()
	if err != nil {
		return nil, errors.WithMessage(err, "生成生产单商品UUID失败")
	}

	// 3. 查询 ProductPackage 获取分类信息
	productPackageRepo := repository.NewProductPackageRepo(db)
	productPackages, err := productPackageRepo.GetProductPackageList(
		repository.CommonRepo.WhereInUuids([]uint64{takeoutItem.TtposProductPackageUuid}),
		productPackageRepo.WithProductCategory(),
	)
	if err != nil {
		logger.Logger.Error("查询商品套餐失败", zap.Error(err))
	}

	var firstCategoryUuid uint64
	if len(productPackages) > 0 && productPackages[0].ProductCategory.Uuid > 0 {
		firstCategoryUuid = productPackages[0].ProductCategory.GetFirstCategoryUuid()
	}

	// 4. 处理 modifiers，提取规格和属性
	var productBomUuid uint64
	var productBomName string
	var attributeNames []string

	for _, modifier := range takeoutItem.TakeoutOrderItemModifiers {
		if modifier.IsMapped == 0 || modifier.TtposModifierUuid == 0 {
			continue
		}

		switch modifier.TtposModifierType {
		case string(valueObject.ModifierTypeFlavor):
			// 规格
			productBomUuid = modifier.TtposModifierUuid
			productBomName = modifier.TtposModifierName
		case string(valueObject.ModifierTypeAttr), string(valueObject.ModifierTypeSauce):
			// 属性和加料
			if modifier.TtposModifierName != "" {
				attributeNames = append(attributeNames, modifier.TtposModifierName)
			}
		}
	}

	// 5. 合并多个属性的名称
	var attributeNamesStr string
	if len(attributeNames) > 0 {
		localeResponses := make([]dto.LocaleResponse, 0, len(attributeNames))
		for _, attrName := range attributeNames {
			localeResp := language.JsonToLocaleResponse(attrName)
			if localeResp != nil && !localeResp.IsNull() {
				localeResponses = append(localeResponses, *localeResp)
			}
		}
		if len(localeResponses) > 0 {
			mergedLocale := language.MergeLocaleResponses(localeResponses, ";")
			attributeNamesStr = mergedLocale.ToJson()
		}
	}

	// 6. 使用商品名称（优先使用 TTPOS 商品名称）
	itemName := takeoutItem.TtposItemName
	if itemName == "" {
		itemName = takeoutItem.ItemName
	}

	// 7. 构建 ProductionOrderProduct
	productionOrderProduct := &model.ProductionOrderProduct{
		BaseModel:             model.BaseModel{Uuid: productUuid, CreateTime: currentTime, UpdateTime: currentTime},
		Name:                  itemName,
		Num:                   float64(takeoutItem.Quantity),
		InitNum:               float64(takeoutItem.Quantity),
		FlavorName:            productBomName,
		ProductBomUuid:        productBomUuid,
		ProductAttributeNames: attributeNamesStr,
		ProductSaucesNames:    "",
		Status:                constant.ProductionOrderProductStatusCooking,
		Remark:                "",
		HasMaterial:           0,
		ProductPackageUuid:    takeoutItem.TtposProductPackageUuid,
		ProductionOrderUuid:   productionOrderUuid,
		TakeoutOrderUuid:      takeoutOrderUuid,
		TakeoutOrderItemUuid:  takeoutItem.Uuid,
		FirstCategoryUuid:     firstCategoryUuid,
	}

	return productionOrderProduct, nil
}

// PrintTakeoutOrder 打印外卖订单小票
func (s *takeoutSrv) PrintTakeoutOrder(ctx context.Context, orderUuid uint64, printLang string, firstExecution int) (*resp.PrinterData, error) {
	// 1. 从领域层获取订单数据
	order, err := domainService.NewTakeoutOrderSrv(s.dbm).GetOrderForPrint(ctx, orderUuid)
	if err != nil {
		return nil, err
	}

	//打印退单联
	if order.OrderState == valueObject.TakeoutOrderStateCanceled {
		receiptData, err := printer.NewPrinterRepo(ctx, printLang).PrintingPlatformTakeoutReceipt(
			order,
			printerConst.TakeoutReceiptTypeRefund,
			firstExecution,
		)
		if err != nil {
			return nil, err
		}
		return receiptData, nil
	} else {
		// 异步打印顾客联
		utils.Go(func() {
			_, err := printer.NewPrinterRepo(ctx, printLang).PrintingPlatformTakeoutReceipt(
				order,
				printerConst.TakeoutReceiptTypeCustomer,
				0,
			)
			if err != nil {
				logger.Logger.Error("打印顾客联失败", zap.Error(err))
			}
		})

		//打印商家联
		receiptData, err := printer.NewPrinterRepo(ctx, printLang).PrintingPlatformTakeoutReceipt(
			order,
			printerConst.TakeoutReceiptTypeMerchant,
			firstExecution,
		)
		if err != nil {
			logger.Logger.Error("打印商家联失败", zap.Error(err))
			return nil, err
		}
		return receiptData, nil
	}
}

// PrintProductionOrder 打印送厨单
func (s *takeoutSrv) PrintProductionOrder(ctx context.Context, orderUuid uint64, printType int, productItems []req.PrintProductItem) (*resp.PrinterData, error) {
	db := ctx.GetDB()
	if db == nil {
		return nil, errors.New("数据库连接失败")
	}

	// 1. 从领域层获取订单数据
	order, err := domainService.NewTakeoutOrderSrv(s.dbm).GetOrderForPrint(ctx, orderUuid)
	if err != nil {
		return nil, err
	}

	productItemsMap := make(map[uint64]bool)
	productItemsBomMap := make(map[uint64]bool)
	for _, productItem := range productItems {
		productItemsMap[productItem.ProductUuid] = true
		productItemsBomMap[productItem.ProductBomUuid] = true
	}

	// 6. 打印送厨单
	// 异步打印
	utils.Go(func() {
		// 1. 准备打印商品数据
		// 直接从外卖订单商品转换为打印模型
		products := make([]printer_model.OrderProduct, 0, len(order.TakeoutOrderItems))

		for _, item := range order.TakeoutOrderItems {
			// 使用 TTPOS 标准名称（商家联打印）
			itemName := item.TtposItemName
			if itemName == "" {
				itemName = item.ItemName // 回退到平台名称
			}

			if len(productItemsMap) > 0 && !item.IsPackage() && !productItemsMap[item.TtposProductPackageUuid] {
				continue
			}

			// 构建商品规格、属性、加料的完整信息
			attrList := make([]dto.LocaleResponse, 0)
			saucesList := make([]dto.LocaleResponse, 0)
			flavorName := dto.LocaleResponse{}
			flavorNameList := make([]dto.LocaleResponse, 0)
			subProducts := make([]printer_model.OrderProduct, 0)
			// 遍历修饰符，分类处理
			for _, modifier := range item.TakeoutOrderItemModifiers {
				if modifier.IsMapped == 0 {
					continue // 跳过未映射的修饰符
				}

				modifierName := modifier.TtposModifierName
				if modifierName == "" {
					modifierName = modifier.ModifierName
				}

				switch modifier.TtposModifierType {
				case string(valueObject.ModifierTypeFlavor):
					// 规格
					attrList = append(attrList, *language.JsonToLocaleResponse(modifierName))
					flavorName = *language.JsonToLocaleResponse(modifier.TtposFlavorName)
					flavorNameList = append(flavorNameList, *language.JsonToLocaleResponse(modifier.TtposFlavorName))
				case string(valueObject.ModifierTypeAttr):
					// 属性
					attrList = append(attrList, *language.JsonToLocaleResponse(modifierName))
				case string(valueObject.ModifierTypeSauce):
					// 加料
					saucesList = append(saucesList, *language.JsonToLocaleResponse(modifierName))
				case string(valueObject.ModifierTypeCommodity):
					if len(productItemsBomMap) > 0 {
						if !productItemsMap[modifier.TtposProductPackageUuid] || !productItemsBomMap[modifier.TtposFlavorProductBomUuid] {
							continue
						}
					}
					subProducts = append(subProducts, printer_model.OrderProduct{
						OrderProductId:  modifier.Uuid,
						ProductId:       modifier.TtposProductPackageUuid,
						ProductName:     *language.JsonToLocaleResponse(modifierName),                                   // 商品名称
						FlavorName:      *language.JsonToLocaleResponse(modifier.TtposFlavorName),                       // 商品规格
						Attr:            *language.JsonToLocaleResponse(modifier.TtposFlavorName),                       // 商品属性
						ProductAttrList: []dto.LocaleResponse{*language.JsonToLocaleResponse(modifier.TtposFlavorName)}, // 规格+属性列表
						TotalNum:        float64(modifier.Quantity),                                                     // 商品数量
						ProductPrice:    modifier.Price,                                                                 // 商品价格
						ProductType:     uint8(item.TtposProductType),                                                   // 商品类型
						IsWrap:          order.IsTakeawayOrder(),                                                        // 是否打包
						TotalPrice:      utils.Round(modifier.Price*float64(modifier.Quantity), 2),                      // 商品总价格
					})
					continue
				}
			}

			product := printer_model.OrderProduct{
				OrderProductId:        item.Uuid,
				ProductId:             item.TtposProductPackageUuid,
				ProductName:           *language.JsonToLocaleResponse(itemName),           // 商品名称
				ProductType:           uint8(item.TtposProductType),                       // 商品类型
				FlavorName:            flavorName,                                         // 商品规格
				Attr:                  language.MergeLocaleResponses(flavorNameList, ","), // 商品属性
				ProductAttrList:       attrList,                                           // 规格+属性列表
				ProductSauceNamesList: saucesList,                                         // 加料列表
				TotalNum:              float64(item.Quantity),                             // 商品数量
				ProductPrice:          item.Price,                                         // 商品价格
				TotalPrice:            item.GetTotalPrice(),                               // 商品总价格
				Remark:                item.Specifications,                                // 商品备注
				IsWrap:                order.IsTakeawayOrder(),                            // 是否打包
				SubProducts:           subProducts,                                        // 套餐子商品列表
			}
			products = append(products, product)
		}

		// 2. 构建打印数据
		printOrder := printer_model.Order{
			Uuid:                   order.Uuid,                   // 使用外卖订单 UUID
			SaleOrderUuid:          0,                            // 外卖订单没有 SaleOrder
			OrderNo:                order.PlatformOrderId,        // 订单号
			MealNum:                1,                            // 外卖订单默认1人
			OrderSourceTakeoutText: order.GetSpacePlatformName(), // 显示平台名称（grab/lineman）
			SerialNo:               order.ShortOrderNumber,       // 外卖订单流水号
			OrderRemark:            nil,                          // 外卖订单备注（目前无此信息）
			DeskUuid:               0,                            // 外卖订单无桌台
			Desk:                   nil,                          // 外卖订单无桌台信息
			UpdateTime:             int64(order.UpdateTime),      // 订单更新时间
			FinishTime:             time.Now().Unix(),            // 订单完成时间
			IsTakeout:              true,                         // 标记为第三方外卖平台订单
			Products:               products,                     // 商品列表
		}

		// 3. 执行打印
		printerRepo := printer.NewPrinterRepo(ctx, "")
		printerRepo.SetFinishedTime(time.Now().Unix())
		// 送厨打印
		if printType == printerConst.PrinterProductTypeKitchen {
			printerRepo.PrintingDishes(printerConst.PrinterProductTypePay, printOrder)
		}
		printerRepo.PrintingDishes(printType, printOrder)
	})

	return nil, nil
}

// PrintReturnOrder 打印退菜单（使用变更前的商品信息）
// 此方法使用 changeResult.ReturnItems 中的 OldItem 数据，确保打印的是变更前的数量
func (s *takeoutSrv) PrintReturnOrder(ctx context.Context, orderUuid uint64, changeResult *valueObject.OrderChangeResult) error {
	if changeResult == nil || len(changeResult.ReturnItems) == 0 {
		return nil
	}

	db := ctx.GetDB()
	if db == nil {
		return errors.New("数据库连接失败")
	}

	// 1. 获取订单基本信息（用于打印头信息）
	order, err := domainService.NewTakeoutOrderSrv(s.dbm).GetOrderForPrint(ctx, orderUuid)
	if err != nil {
		return err
	}

	// 2. 构建退菜商品列表（使用 OldItem 中的数量）
	products := make([]printer_model.OrderProduct, 0, len(changeResult.ReturnItems))

	for _, item := range changeResult.ReturnItems {
		if item.OldItem == nil || item.OldItem.TtposProductPackageUuid == 0 {
			continue
		}

		// 使用 OldItem 中的信息
		itemName := item.OldItem.ItemName
		quantity := item.OldQuantity

		// 构建修饰符信息
		attrList := make([]dto.LocaleResponse, 0)
		saucesList := make([]dto.LocaleResponse, 0)
		flavorName := dto.LocaleResponse{}

		for _, modifier := range item.OldItem.Modifiers {
			modifierName := modifier.ModifierName
			if modifierName == "" {
				continue
			}

			localeResp := language.JsonToLocaleResponse(modifierName)
			if localeResp == nil || localeResp.IsNull() {
				continue
			}

			switch modifier.TtposModifierType {
			case string(valueObject.ModifierTypeFlavor):
				flavorName = *localeResp
			case string(valueObject.ModifierTypeAttr):
				attrList = append(attrList, *localeResp)
			case string(valueObject.ModifierTypeSauce):
				saucesList = append(saucesList, *localeResp)
			}
		}

		product := printer_model.OrderProduct{
			OrderProductId:        item.OldItem.Uuid,
			ProductId:             item.OldItem.TtposProductPackageUuid,
			ProductName:           *language.JsonToLocaleResponse(itemName),
			ProductType:           uint8(item.OldItem.TtposProductType),
			FlavorName:            flavorName,
			ProductAttrList:       attrList,
			ProductSauceNamesList: saucesList,
			TotalNum:              float64(quantity), // 使用变更前的数量
			ProductPrice:          item.OldItem.Price,
			TotalPrice:            item.OldItem.Price * float64(quantity),
			IsWrap:                order.IsTakeawayOrder(),
		}
		products = append(products, product)
	}

	if len(products) == 0 {
		logger.Logger.Warn("退菜单打印：没有有效的商品",
			zap.Uint64("orderUuid", orderUuid))
		return nil
	}

	// 3. 构建打印数据
	printOrder := printer_model.Order{
		Uuid:                   order.Uuid,
		SaleOrderUuid:          0,
		OrderNo:                order.PlatformOrderId,
		MealNum:                1,
		OrderSourceTakeoutText: order.GetSpacePlatformName(),
		SerialNo:               order.ShortOrderNumber,
		OrderRemark:            nil,
		DeskUuid:               0,
		Desk:                   nil,
		UpdateTime:             int64(order.UpdateTime),
		FinishTime:             time.Now().Unix(),
		IsTakeout:              true,
		Products:               products,
	}

	// 4. 异步执行打印
	utils.Go(func() {
		printerRepo := printer.NewPrinterRepo(ctx, "")
		printerRepo.SetFinishedTime(time.Now().Unix())
		printerRepo.PrintingDishes(printerConst.PrinterProductTypeBackFood, printOrder)
	})

	return nil
}

// RecordTakeoutOrderPeakTime 记录外卖订单高峰期
// 自动根据订单状态判断是增加（inc）还是减少（dec）
// 判断规则：
//   - order.AcceptedTime > 0 && order.OrderState == 10 (已接单) → inc
//   - order.AcceptedTime > 0 && order.OrderState == 60 (已取消) → dec
//   - 其他情况不记录
func (s *takeoutSrv) RecordTakeoutOrderPeakTime(ctx context.Context, orderUuid uint64, companyUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)

	// 设置上下文
	ctx.SetDB(db)
	ctx.SetCompanyUuid(companyUuid)

	// 1. 查询外卖订单信息
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(orderUuid)
	if err != nil {
		return err
	}
	if order == nil {
		return nil
	}

	// 2. 自动判断操作类型
	recordType := determineRecordType(order)
	if recordType == "" {
		// 不符合记录条件，直接返回
		return nil
	}

	// 3. 构建 SaleBill
	saleBill := buildSaleBillFromTakeoutOrder(order, recordType)
	if saleBill == nil {
		return nil
	}

	// 4. 获取门店设置（时区）
	settingSrv := setting.NewSrv(s.dbm, cache.Global)
	storeSetting, err := settingSrv.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Info("获取门店设置失败", zap.Error(err))
		return err
	}

	// 5. 记录高峰期
	peakTimeRepo := repository.NewSaleOrderPeakTimeRepo(db)
	refundMoney := utils.IfFloat64(recordType == "dec", order.PlatformTotal, 0.0)
	return peakTimeRepo.Record(recordType, saleBill, refundMoney, storeSetting.TimeZone)
}

// determineRecordType 根据订单状态判断记录类型
// 返回: "inc" - 增加, "dec" - 减少, "" - 不记录
func determineRecordType(order *takeoutModel.TakeoutOrder) string {
	// 必须要有接单人和接单时间
	if order.AcceptedBy <= 0 || order.AcceptedTime <= 0 {
		return ""
	}

	// 判断订单状态
	if order.OrderState == valueObject.TakeoutOrderStateAccepted {
		// 已接单配餐中 → inc
		return "inc"
	} else if order.OrderState == valueObject.TakeoutOrderStateCanceled {
		// 已取消 → dec
		return "dec"
	}

	// 其他状态不记录
	return ""
}

// buildSaleBillFromTakeoutOrder 从外卖订单构建 SaleBill
// recordType: "inc" - 接单时使用 AcceptedTime, "dec" - 取消时使用 RejectedTime
func buildSaleBillFromTakeoutOrder(order *takeoutModel.TakeoutOrder, recordType string) *model.SaleBill {
	saleBill := &model.SaleBill{
		Status:        constant.SaleBillStatusComplete, // 设置为已完成状态，IsFinish() 才能返回 true
		PaymentAmount: order.PlatformTotal,             // 顾客实付金额（单位：元）
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
		rejectedBy := order.RejectedBy
		if rejectedBy == 0 {
			rejectedBy = order.AcceptedBy
		}
		if order.RejectedTime > 0 {
			saleBill.FinishTime = order.RejectedTime
			saleBill.CashierUuid = rejectedBy
		} else {
			// 如果没有取消时间，使用订单时间
			saleBill.FinishTime = order.OrderTime
			saleBill.CashierUuid = rejectedBy
		}
	}

	// 如果 FinishTime 为 0，无法记录高峰期
	if saleBill.FinishTime == 0 {
		return nil
	}

	return saleBill
}
