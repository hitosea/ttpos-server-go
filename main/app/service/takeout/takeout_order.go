package takeout

import (
	"fmt"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	domainService "ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ITakeoutOrderSrv interface {
	// ProcessTakeoutOrderOutboundAndSales 处理外卖订单出库和销量
	ProcessTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64, acceptedBy uint64) error
	// RestoreTakeoutOrderOutboundAndSales 恢复外卖订单出库和销量（取消订单时调用）
	RestoreTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64) error
	// CreateProductionOrderForTakeout 为外卖订单创建送厨单
	CreateProductionOrderForTakeout(ctx context.Context, orderUuid uint64) error
	// PrintTakeoutOrder 打印外卖订单小票
	PrintTakeoutOrder(ctx context.Context, orderUuid uint64) (*resp.PrinterData, error)
}

// ToggleTakeoutStatus 切换指定平台外卖状态
func (s *takeoutSrv) ToggleTakeoutStatus(ctx context.Context, req request.ToggleTakeoutStatusRequest) (*response.TakeoutStatusResponse, error) {
	if req.Platform == "grab" {
		err := service.NewPaymentMethodSrv(s.dbm, s.settingSrv).SaveGrabPaymentMethod(ctx, ctx.GetDB())
		if err != nil {
			return nil, errors.WithMessage(err, "保存Grab支付方式失败")
		}
	} else if req.Platform == "lineman" {
		err := service.NewPaymentMethodSrv(s.dbm, s.settingSrv).SaveLineManPaymentMethod(ctx, ctx.GetDB())
		if err != nil {
			return nil, errors.WithMessage(err, "保存LINE MAN支付方式失败")
		}
	}
	return s.takeoutAppSrv.ToggleTakeoutStatus(ctx, req)
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
	staffShiftLogUuid := uint64(0)
	if acceptedBy > 0 {
		staffShiftLog, err := s.getCurrentStaffShiftLog(db, acceptedBy)
		if err != nil {
			logger.Logger.Warn("获取员工班次记录失败", zap.Uint64("staffUuid", acceptedBy), zap.Error(err))
		} else if staffShiftLog != nil {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
	}

	// 4. 创建出库单
	warehouseOutForms := model.NewWarehouseOutForm(decreaseStockList, false, order.Uuid, acceptedBy, staffShiftLogUuid, orderUuid)

	// 5. 在事务中创建出库单和更新销量
	return db.Transaction(func(tx *gorm.DB) error {
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
		materials, err := s.reduceTakeoutOrderStock(tx, order.Uuid)
		if err != nil {
			logger.Logger.Error("扣减外卖订单库存失败", zap.Uint64("orderUuid", order.Uuid), zap.Error(err))
			// 库存扣减失败不影响出库流程，只记录日志
		}

		// 5.4 汇总并保存外卖订单原料（使用实际扣减的材料数据）
		if len(materials) > 0 {
			if err := s.saveTakeoutOrderMaterialsFromMap(ctxTx, orderUuid, staffShiftLogUuid, materials); err != nil {
				return errors.WithMessage(err, "保存外卖订单原料失败")
			}
		}

		return nil
	})
}

// getCurrentStaffShiftLog 获取当前员工班次记录
// 如果 staffUuid 为空，则返回最新的正在当班的班次记录
func (s *takeoutSrv) getCurrentStaffShiftLog(db *gorm.DB, staffUuid uint64) (*model.StaffShiftLog, error) {
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
				return nil, nil
			}
			return nil, errors.WithMessage(err, "查询最新班次记录失败")
		}
		return &shiftLog, nil
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
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询员工班次记录失败")
	}
	return &shiftLog, nil
}

// saveTakeoutOrderMaterialsFromMap 从材料map汇总并保存外卖订单原料到 ttpos_takeout_order_material 表
func (s *takeoutSrv) saveTakeoutOrderMaterialsFromMap(
	ctx context.Context,
	takeoutOrderUuid uint64,
	staffShiftLogUuid uint64,
	materials map[uint64]map[uint64]float64, // map[materialUuid]map[warehouseUuid]num
) error {
	if len(materials) == 0 {
		return nil
	}

	// 构建原料记录
	takeoutOrderMaterials := make([]*takeoutModel.TakeoutOrderMaterial, 0)
	for materialUuid, warehouseMap := range materials {
		for warehouseUuid, num := range warehouseMap {
			takeoutOrderMaterials = append(takeoutOrderMaterials, &takeoutModel.TakeoutOrderMaterial{
				TakeoutOrderUuid:  takeoutOrderUuid,
				MaterialUuid:      materialUuid,
				WarehouseUuid:     warehouseUuid,
				Num:               num,
				StaffShiftLogUuid: staffShiftLogUuid,
				IsSummarized:      0, // 初始为未统计
			})
		}
	}

	// 保存原料记录
	if len(takeoutOrderMaterials) > 0 {
		materialSrv := domainService.NewTakeoutOrderMaterialSrv(s.dbm)
		if err := materialSrv.SaveOrderMaterials(ctx, takeoutOrderMaterials); err != nil {
			return err
		}
	}

	return nil
}

// saveTakeoutOrderMaterials 从decreaseStockList汇总并保存外卖订单原料到 ttpos_takeout_order_material 表
// 注意：这个方法已弃用，请使用 saveTakeoutOrderMaterialsFromMap 确保数据一致性
func (s *takeoutSrv) saveTakeoutOrderMaterials(
	ctx context.Context,
	takeoutOrderUuid uint64,
	staffShiftLogUuid uint64,
	decreaseStockList []*model.Product,
) error {
	if len(decreaseStockList) == 0 {
		return nil
	}

	// 构建原料记录
	takeoutOrderMaterials := make([]*takeoutModel.TakeoutOrderMaterial, 0)
	for _, product := range decreaseStockList {
		for _, material := range product.ProductBomMaterials {
			takeoutOrderMaterials = append(takeoutOrderMaterials, &takeoutModel.TakeoutOrderMaterial{
				TakeoutOrderUuid:  takeoutOrderUuid,
				MaterialUuid:      material.MaterialUuid,
				WarehouseUuid:     material.WarehouseUuid,
				Num:               material.Num,
				StaffShiftLogUuid: staffShiftLogUuid,
				IsSummarized:      0, // 初始为未统计
			})
		}
	}

	// 保存原料记录
	if len(takeoutOrderMaterials) > 0 {
		materialSrv := domainService.NewTakeoutOrderMaterialSrv(s.dbm)
		if err := materialSrv.SaveOrderMaterials(ctx, takeoutOrderMaterials); err != nil {
			return err
		}
	}

	return nil
}

// buildTakeoutOrderDecreaseStockList 从外卖订单构建出库清单
func (s *takeoutSrv) buildTakeoutOrderDecreaseStockList(ctx context.Context, order *takeoutModel.TakeoutOrder) ([]*model.Product, error) {
	db := ctx.GetDB()

	// 1. 构建 BOM 数量映射
	orderService := domainService.NewTakeoutOrderSrv(s.dbm)
	bomQuantityMap, _, err := orderService.BuildBomQuantityMap(ctx, order)
	if err != nil {
		return nil, errors.WithMessage(err, "构建BOM数量映射失败")
	}

	if len(bomQuantityMap) == 0 {
		return []*model.Product{}, nil
	}

	// 2. 批量查询 BOM 信息（包含原材料信息）
	bomUuids := make([]uint64, 0, len(bomQuantityMap))
	for bomUuid := range bomQuantityMap {
		bomUuids = append(bomUuids, bomUuid)
	}

	productBomRepo := repository.NewProductBomRepo(db)
	productBoms, err := productBomRepo.GetProductBoms(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("uuid IN ?", bomUuids)
		},
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "FlavorMaterials",
				Args: []interface{}{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
			repository.WithPreload{
				Query: "FlavorMaterials.Material.WarehouseItems",
			},
			repository.WithPreload{
				Query: "ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			repository.WithPreload{
				Query: "ProductSauce.SauceMaterials",
				Args: []interface{}{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
			repository.WithPreload{
				Query: "ProductSauce.SauceMaterials.Material.WarehouseItems",
			},
			repository.WithPreload{
				Query: "ProductSauce.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询BOM信息失败")
	}

	// 3. 构建 BOM UUID -> ProductBom 映射
	bomMap := make(map[uint64]*model.ProductBom)
	for _, bom := range productBoms {
		bomMap[bom.Uuid] = bom
	}

	// 4. 构建出库清单
	decreaseStockList := make([]*model.Product, 0)
	for bomUuid, quantity := range bomQuantityMap {
		bom, ok := bomMap[bomUuid]
		if !ok {
			logger.Logger.Warn("BOM不存在", zap.Uint64("bomUuid", bomUuid))
			continue
		}

		productNum := float64(quantity)
		productBomMaterials := make([]*model.ProductBomMaterials, 0)

		// 4.1 处理规格商品的原材料
		if bom.IsFlavor() {
			// 优先使用成本卡的原材料
			var flavorMaterials []*model.RelatedMaterial
			if bom.HasProductBomCard() && bom.ProductBomCard != nil {
				flavorMaterials = bom.ProductBomCard.RelatedMaterials
			} else {
				flavorMaterials = bom.FlavorMaterials
			}

			// 遍历原材料
			for _, material := range flavorMaterials {
				if material.IsDelete() || material.Material == nil {
					continue
				}
				// 如果材料被禁用，则跳过
				if !material.Material.Status {
					continue
				}
				if num := material.GetDecreaseNum(productNum); num > 0 {
					productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
						MaterialUuid:  material.MaterialUuid,
						WarehouseUuid: material.Material.WarehouseUuid,
						Num:           num,
						SaleOrderUuid: 0, // 外卖订单没有 SaleOrderUuid
					})
				}
			}
		}

		// 4.2 处理小料的原材料
		if bom.IsSauce() {
			// 优先使用成本卡的原材料
			var sauceMaterials []*model.RelatedMaterial
			if bom.ProductSauce.HasProductBomCard() && bom.ProductSauce.ProductBomCard != nil {
				sauceMaterials = bom.ProductSauce.ProductBomCard.RelatedMaterials
			} else {
				sauceMaterials = bom.ProductSauce.SauceMaterials
			}

			// 遍历原材料
			for _, material := range sauceMaterials {
				if material.Material == nil {
					continue
				}
				// 如果材料被禁用，则跳过
				if !material.Material.Status {
					continue
				}
				if num := material.GetDecreaseNum(productNum); num > 0 {
					productBomMaterials = append(productBomMaterials, &model.ProductBomMaterials{
						MaterialUuid:  material.MaterialUuid,
						WarehouseUuid: material.Material.WarehouseUuid,
						Num:           num,
						SaleOrderUuid: 0, // 外卖订单没有 SaleOrderUuid
					})
				}
			}
		}

		// 4.3 添加到出库清单
		if productNum > 0 {
			decreaseStockList = append(decreaseStockList, &model.Product{
				TakeoutOrderUuid:    order.Uuid,
				ProductBomUuid:      bomUuid,
				PackageUuid:         bom.ProductPackageUuid,
				Num:                 productNum,
				ProductBomMaterials: productBomMaterials,
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

// reduceTakeoutOrderStock 扣减外卖订单库存，并返回实际扣减的材料列表
func (s *takeoutSrv) reduceTakeoutOrderStock(db *gorm.DB, takeoutOrderUuid uint64) (map[uint64]map[uint64]float64, error) {
	// 加锁，防止并发扣减库存
	lockKey := fmt.Sprintf("takeout_order_stock:%d", takeoutOrderUuid)
	systemLock := lock.NewSystemLock()
	systemLock.LockUuidString(lockKey)
	defer systemLock.UnlockUuidString(lockKey)

	// 获取未减库存的出库单明细
	warehouseFormRepo := repository.NewWarehouseFormRepo(db)
	warehouseOutFormItems, err := warehouseFormRepo.GetWarehouseOutFormItem(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("takeout_order_uuid = ?", takeoutOrderUuid)
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Where("reduce_stock = ?", constant.WarehouseOutFormItemReduceStockNotProcessed)
		},
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询出库单明细失败")
	}

	if len(warehouseOutFormItems) == 0 {
		return nil, nil
	}

	// 按类型分组处理
	productBoms := make(map[uint64]*model.ProductBom)
	materials := make(map[uint64]map[uint64]float64) // map[materialUuid]map[warehouseUuid]reduceStockNum

	for _, item := range warehouseOutFormItems {
		item.ReduceStock = constant.WarehouseOutFormItemReduceStockSuccess

		if item.IsProductBom() {
			// 查询 BOM 信息
			if productBoms[item.ProductBomUuid] == nil {
				productBomRepo := repository.NewProductBomRepo(db)
				bom, err := productBomRepo.GetFlavorProductBomByUuid(item.ProductBomUuid)
				if err != nil {
					logger.Logger.Error("查询BOM信息失败", zap.Uint64("bomUuid", item.ProductBomUuid), zap.Error(err))
					continue
				}
				productBoms[item.ProductBomUuid] = bom
			}

			// 扣减 BOM 库存
			if productBoms[item.ProductBomUuid] != nil {
				productBoms[item.ProductBomUuid].StockNum -= item.Num
			}
		} else if item.IsMaterial() {
			// 扣减材料库存
			if materials[item.MaterialUuid] == nil {
				materials[item.MaterialUuid] = make(map[uint64]float64)
			}
			materials[item.MaterialUuid][item.WarehouseUuid] += item.Num
		}
	}

	// 在事务中更新库存
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新出库单明细状态
		if err := repository.NewWarehouseFormRepo(tx).UpdateWarehouseOutFormItemRecordsReduceStockByTakeoutOrderUuid(takeoutOrderUuid); err != nil {
			return errors.WithMessage(err, "更新出库单明细状态失败")
		}

		// 更新 BOM 库存
		productBomList := make([]*model.ProductBom, 0, len(productBoms))
		for _, bom := range productBoms {
			productBomList = append(productBomList, bom)
		}
		if len(productBomList) > 0 {
			if err := repository.NewProductBomRepo(tx).UpdateProductBoms(productBomList); err != nil {
				return errors.WithMessage(err, "更新BOM库存失败")
			}
		}

		// 更新材料库存
		for materialUuid, warehouseMap := range materials {
			for warehouseUuid, reduceStockNum := range warehouseMap {
				if err := base.NewMaterialRepo(tx).UpdateMaterialsStockNum(materialUuid, warehouseUuid, -reduceStockNum); err != nil {
					return errors.WithMessage(err, "更新材料库存失败")
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 返回实际扣减的材料列表（map[materialUuid]map[warehouseUuid]reduceStockNum）
	return materials, nil
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
			return db.Where("takeout_order_uuid = ?", orderUuid)
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Where("revoke_time = ?", 0) // 只查询未撤销的出库单
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
	return db.Transaction(func(tx *gorm.DB) error {
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
		if err := s.restoreTakeoutOrderStock(tx, warehouseOutFormItems); err != nil {
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
}

// restoreTakeoutOrderStock 恢复外卖订单库存
func (s *takeoutSrv) restoreTakeoutOrderStock(db *gorm.DB, warehouseOutFormItems []*model.WarehouseOutFormItem) error {
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
				bom, err := productBomRepo.GetFlavorProductBomByUuid(item.ProductBomUuid)
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
	productBomRepo := repository.NewProductBomRepo(db)
	productPackageAttributeRepo := repository.NewProductPackageAttributeRepo(db)
	bomMappingRepo := persistence.NewTakeoutBomMappingRepo(db)

	// 1. 收集所有需要查询的 UUID
	var (
		normalProductPackageUuids = make(map[uint64]bool) // 普通商品的 ProductPackage UUID
		packageGroupItemUuids     = make([]uint64, 0)     // 套餐商品的 groupItem UUID
		flavorBomUuids            = make(map[uint64]bool) // flavor 类型的 BOM UUID
		sauceBomUuids             = make(map[uint64]bool) // sauce 类型的 BOM UUID
		attrPackageAttributeUuids = make(map[uint64]bool) // attr 类型的 ProductPackageAttribute UUID
	)

	// 遍历所有商品，收集需要查询的 UUID
	for _, takeoutItem := range order.TakeoutOrderItems {
		if takeoutItem.TtposProductType == 0 {
			// 普通商品：收集 ProductPackage UUID
			if takeoutItem.TtposProductUuid > 0 {
				normalProductPackageUuids[takeoutItem.TtposProductUuid] = true
			}
			// 收集 modifiers 相关的 UUID
			for _, modifier := range takeoutItem.TakeoutOrderItemModifiers {
				if modifier.IsMapped == 0 || modifier.TtposModifierUuid == 0 {
					continue
				}
				switch modifier.TtposModifierType {
				case "flavor":
					flavorBomUuids[modifier.TtposModifierUuid] = true
				case "sauce":
					sauceBomUuids[modifier.TtposModifierUuid] = true
				case "attr":
					attrPackageAttributeUuids[modifier.TtposModifierUuid] = true
				}
			}
		} else {
			// 套餐商品：收集 groupItem UUID
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

	// 2.2 批量查询套餐商品的 BOM 映射
	var groupItemBomMapping map[uint64]persistence.GroupItemBomMapping
	var packageSubProductPackageUuids = make(map[uint64]bool)
	var packageBomUuids = make(map[uint64]bool)
	if len(packageGroupItemUuids) > 0 {
		groupItemBomMapping, err = bomMappingRepo.GetGroupItemBomMapping(packageGroupItemUuids)
		if err != nil {
			logger.Logger.Error("批量查询套餐商品BOM映射失败", zap.Error(err))
			return errors.WithMessage(err, "批量查询套餐商品BOM映射失败")
		}
		// 收集套餐子商品的 ProductPackage UUID 和 BOM UUID
		for _, mapping := range groupItemBomMapping {
			if mapping.RelatedUuid > 0 {
				packageSubProductPackageUuids[mapping.RelatedUuid] = true
			}
			if mapping.ProductBomUuid > 0 {
				packageBomUuids[mapping.ProductBomUuid] = true
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

	// 2.4 批量查询所有 BOM（flavor + sauce + package）
	allBomUuids := make([]uint64, 0)
	for uuid := range flavorBomUuids {
		allBomUuids = append(allBomUuids, uuid)
	}
	for uuid := range sauceBomUuids {
		allBomUuids = append(allBomUuids, uuid)
	}
	for uuid := range packageBomUuids {
		allBomUuids = append(allBomUuids, uuid)
	}
	// 去重
	bomUuidMap := make(map[uint64]bool)
	uniqueBomUuids := make([]uint64, 0)
	for _, uuid := range allBomUuids {
		if !bomUuidMap[uuid] {
			bomUuidMap[uuid] = true
			uniqueBomUuids = append(uniqueBomUuids, uuid)
		}
	}
	// 批量查询 flavor BOM（需要预加载 ProductFlavor.MultiLanguageName）
	flavorBomUuidList := make([]uint64, 0)
	for uuid := range flavorBomUuids {
		flavorBomUuidList = append(flavorBomUuidList, uuid)
	}
	flavorBoms, err := productBomRepo.GetProductBoms(
		repository.CommonRepo.WhereInUuids(flavorBomUuidList),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductFlavor.MultiLanguageName",
		}),
	)
	if err != nil {
		logger.Logger.Error("批量查询规格BOM失败", zap.Error(err))
		return errors.WithMessage(err, "批量查询规格BOM失败")
	}
	flavorBomMap := make(map[uint64]*model.ProductBom)
	for _, bom := range flavorBoms {
		flavorBomMap[bom.Uuid] = bom
	}

	// 批量查询 sauce BOM（需要预加载 ProductSauce.MultiLanguageName）
	sauceBomUuidList := make([]uint64, 0)
	for uuid := range sauceBomUuids {
		sauceBomUuidList = append(sauceBomUuidList, uuid)
	}
	sauceBoms, err := productBomRepo.GetProductBoms(
		repository.CommonRepo.WhereInUuids(sauceBomUuidList),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductSauce.MultiLanguageName",
		}),
	)
	if err != nil {
		logger.Logger.Error("批量查询加料BOM失败", zap.Error(err))
		return errors.WithMessage(err, "批量查询加料BOM失败")
	}
	sauceBomMap := make(map[uint64]*model.ProductBom)
	for _, bom := range sauceBoms {
		sauceBomMap[bom.Uuid] = bom
	}

	// 批量查询 package BOM（不需要预加载）
	packageBomUuidList := make([]uint64, 0)
	for uuid := range packageBomUuids {
		packageBomUuidList = append(packageBomUuidList, uuid)
	}
	packageBoms, err := productBomRepo.GetProductBomsByUuids(packageBomUuidList)
	if err != nil {
		logger.Logger.Error("批量查询套餐BOM失败", zap.Error(err))
		return errors.WithMessage(err, "批量查询套餐BOM失败")
	}
	packageBomMap := make(map[uint64]*model.ProductBom)
	for _, bom := range packageBoms {
		packageBomMap[bom.Uuid] = bom
	}

	// 2.5 批量查询 ProductPackageAttribute
	attrPackageAttributeUuidList := make([]uint64, 0, len(attrPackageAttributeUuids))
	for uuid := range attrPackageAttributeUuids {
		attrPackageAttributeUuidList = append(attrPackageAttributeUuidList, uuid)
	}
	productPackageAttributes, err := productPackageAttributeRepo.GetProductPackageAttributesByUuids(attrPackageAttributeUuidList)
	if err != nil {
		logger.Logger.Error("批量查询商品属性失败", zap.Error(err))
		return errors.WithMessage(err, "批量查询商品属性失败")
	}
	productPackageAttributeMap := make(map[uint64]*model.ProductPackageAttribute)
	for _, attr := range productPackageAttributes {
		productPackageAttributeMap[attr.Uuid] = attr
	}

	// 3. 创建 ProductionOrderProduct 列表
	productionOrderProducts := make([]*model.ProductionOrderProduct, 0)

	for _, takeoutItem := range order.TakeoutOrderItems {
		if takeoutItem.TtposProductType == 0 {
			// 普通商品处理
			productPackage, ok := normalProductPackageMap[takeoutItem.TtposProductUuid]
			if !ok {
				logger.Logger.Warn("未找到普通商品套餐", zap.Uint64("productPackageUuid", takeoutItem.TtposProductUuid))
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
			productPackageName := productPackage.Name
			var firstCategoryUuid uint64
			if productPackage.ProductCategory.Uuid > 0 {
				firstCategoryUuid = productPackage.ProductCategory.GetFirstCategoryUuid()
			}

			// 处理 modifiers，提取属性名称
			var productBomUuid uint64
			var productBomName string
			var attributeMultiLanguageNames []*model.MultiLanguageName

			for _, modifier := range takeoutItem.TakeoutOrderItemModifiers {
				if modifier.IsMapped == 0 || modifier.TtposModifierUuid == 0 {
					continue
				}

				switch modifier.TtposModifierType {
				case "flavor":
					productBomUuid = modifier.TtposModifierUuid
					if bom, ok := flavorBomMap[productBomUuid]; ok && bom != nil {
						productBomName = bom.ProductFlavor.MultiLanguageName.ToJson()
						attributeMultiLanguageNames = append(attributeMultiLanguageNames, &bom.ProductFlavor.MultiLanguageName)
					}
				case "attr":
					if attr, ok := productPackageAttributeMap[modifier.TtposModifierUuid]; ok && attr != nil {
						attributeMultiLanguageNames = append(attributeMultiLanguageNames, &attr.Attribute.MultiLanguageName)
					}
				case "sauce":
					if bom, ok := sauceBomMap[modifier.TtposModifierUuid]; ok && bom != nil {
						attributeMultiLanguageNames = append(attributeMultiLanguageNames, &bom.ProductSauce.MultiLanguageName)
					}
				}
			}

			// 合并多个属性的多语言名称
			var attributeNames string
			if len(attributeMultiLanguageNames) > 0 {
				localeResponses := make([]dto.LocaleResponse, 0, len(attributeMultiLanguageNames))
				for _, attrName := range attributeMultiLanguageNames {
					localeResp := attrName.GetNames()
					if !localeResp.IsNull() {
						localeResponses = append(localeResponses, localeResp)
					}
				}
				if len(localeResponses) > 0 {
					mergedLocale := language.MergeLocaleResponses(localeResponses, ";")
					attributeNames = mergedLocale.ToJson()
				}
			}

			// 使用商品名称（优先使用 TTPOS 商品名称，否则使用平台商品名称）
			itemName := takeoutItem.ItemName
			if productPackageName != "" {
				itemName = productPackageName
			}

			// 创建 ProductionOrderProduct
			productionOrderProduct := &model.ProductionOrderProduct{
				BaseModel:             model.BaseModel{Uuid: productUuid, CreateTime: currentTime, UpdateTime: currentTime},
				Name:                  itemName,
				Num:                   float64(takeoutItem.Quantity),
				InitNum:               float64(takeoutItem.Quantity),
				FlavorName:            productBomName,
				ProductBomUuid:        productBomUuid,
				ProductAttributeNames: attributeNames,
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

				// 获取子商品的 ProductPackage
				subProductPackage, ok := packageSubProductPackageMap[mapping.RelatedUuid]
				if !ok {
					logger.Logger.Error("未找到子商品套餐", zap.Uint64("relatedUuid", mapping.RelatedUuid))
					continue
				}

				productPackageUuid := subProductPackage.Uuid
				productPackageName := subProductPackage.Name
				var firstCategoryUuid uint64
				if subProductPackage.ProductCategory.Uuid > 0 {
					firstCategoryUuid = subProductPackage.ProductCategory.GetFirstCategoryUuid()
				}

				// 获取 BOM 信息
				productBomUuid := mapping.ProductBomUuid
				var productBomName string
				if bom, ok := packageBomMap[productBomUuid]; ok && bom != nil {
					productBomName = bom.Name
				}

				// 生成 ProductionOrderProduct UUID
				productUuid, err := utils.GetID()
				if err != nil {
					logger.Logger.Error("生成送厨商品UUID失败", zap.Error(err))
					continue
				}

				// 使用商品名称（优先使用 TTPOS 商品名称，否则使用平台商品名称）
				itemName := takeoutItem.ItemName
				if productPackageName != "" {
					itemName = productPackageName
				}

				// 套餐子商品的数量 = groupItem.Num * takeoutItem.Quantity
				productNum := mapping.Num * float64(takeoutItem.Quantity)

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

// PrintTakeoutOrder 打印外卖订单小票
func (s *takeoutSrv) PrintTakeoutOrder(ctx context.Context, orderUuid uint64) (*resp.PrinterData, error) {
	return nil, errors.New("功能未实现")

	// // 1. 从领域层获取订单数据
	// takeoutOrderSrv := domainService.NewTakeoutOrderSrv(s.dbm)
	// order, err := takeoutOrderSrv.GetOrderForPrint(ctx, orderUuid)
	// if err != nil {
	// 	return nil, err
	// }

	// // 2. 构建打印内容（简化版本，返回订单信息）
	// // TODO: 完整实现需要调用 printer 模块创建打印任务
	// printData := map[string]interface{}{
	// 	"message": "打印小票",
	// 	"order": map[string]interface{}{
	// 		"uuid":               order.Uuid,
	// 		"platform":           order.Platform,
	// 		"short_order_number": order.ShortOrderNumber,
	// 		"order_state":        order.OrderState,
	// 		"currency_symbol":    order.CurrencySymbol,
	// 		"subtotal":           order.Subtotal,
	// 		"delivery_fee":       order.DeliveryFee,
	// 		"eater_payment":      order.EaterPayment,
	// 		"order_time":         order.OrderTime,
	// 	},
	// }

	// // 3. 添加商品信息
	// items := make([]map[string]interface{}, 0)
	// for _, item := range order.TakeoutOrderItems {
	// 	itemData := map[string]interface{}{
	// 		"name":     item.ItemName,
	// 		"quantity": item.Quantity,
	// 		"price":    item.Price,
	// 	}

	// 	// 添加修饰符信息
	// 	if len(item.TakeoutOrderItemModifiers) > 0 {
	// 		modifiers := make([]string, 0)
	// 		for _, modifier := range item.TakeoutOrderItemModifiers {
	// 			if modifier.ModifierName != "" {
	// 				modifiers = append(modifiers, modifier.ModifierName)
	// 			}
	// 		}
	// 		if len(modifiers) > 0 {
	// 			itemData["modifiers"] = modifiers
	// 		}
	// 	}

	// 	items = append(items, itemData)
	// }
	// printData["order"].(map[string]interface{})["items"] = items

	// // 4. 转换为 JSON 字符串
	// jsonData, err := json.Marshal(printData)
	// if err != nil {
	// 	return nil, errors.New("序列化打印数据失败")
	// }

	// // 5. 构建标准的 PrinterData 响应
	// printerData := &resp.PrinterData{
	// 	Data:        string(jsonData),
	// 	PrintMethod: 2, // 2: 图片打印
	// }

	// return printerData, nil
}
