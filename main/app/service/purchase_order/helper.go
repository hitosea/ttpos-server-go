package purchase_order

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// purchaseOrderHelper 采购订单辅助方法
type purchaseOrderHelper struct{}

// generateOrderNo 生成采购申请订单编号
// 格式：prefix+年月日+0000自增序列号
func (h *purchaseOrderHelper) generateOrderNo(db *gorm.DB, prefix string, timezone string) string {
	// 年月日部分
	datePart := utils.SetTimezone(timezone).Now().Format("20060102")

	// 生成自增序列号
	serialNo, err := h.generatePurchaseOrderSerialNo(db)
	if err != nil {
		return ""
	}

	// 组装订单编号：prefix+年月日+0000自增序列号
	orderNo := prefix + datePart + serialNo

	return orderNo
}

// generatePurchaseOrderSerialNo 生成采购申请自增序列号
func (h *purchaseOrderHelper) generatePurchaseOrderSerialNo(db *gorm.DB) (string, error) {
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)

	// 获取今天最新的采购申请
	latestOrder, err := purchaseOrderRepo.GetLatestOrderToday()
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", errors.WithMessage(errors.New("获取最新采购申请失败"), err.Error())
	}

	// 如果没有查询到今天的采购申请，则设置为0001
	if latestOrder == nil {
		return "0001", nil
	}

	// 从订单编号中提取序列号部分（最后4位）
	if len(latestOrder.OrderNo) >= 4 {
		lastSerialNo := latestOrder.OrderNo[len(latestOrder.OrderNo)-4:]
		serialNoNum, err := strconv.Atoi(lastSerialNo)
		if err != nil {
			// 如果解析失败，重新从0001开始
			return "0001", nil
		}
		// 序列号加1
		newSerialNoNum := serialNoNum + 1
		return fmt.Sprintf("%04d", newSerialNoNum), nil
	}

	// 如果订单编号格式不正确，重新从0001开始
	return "0001", nil
}

// generateReceiptNo 生成收货单号
// 格式：SHRK+年月日+0000自增序列号
func (h *purchaseOrderHelper) generateReceiptNo(db *gorm.DB, timezone string) string {
	// 固定前缀
	prefix := "SHRK"
	// 年月日部分
	datePart := utils.SetTimezone(timezone).Now().Format("20060102")

	// 生成自增序列号
	serialNo, err := h.generateReceiptOrderSerialNo(db)
	if err != nil {
		return ""
	}

	// 组装收货单号：SHRK+年月日+0000自增序列号
	receiptNo := prefix + datePart + serialNo

	return receiptNo
}

// generateReceiptOrderSerialNo 生成收货单自增序列号
func (h *purchaseOrderHelper) generateReceiptOrderSerialNo(db *gorm.DB) (string, error) {
	receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)

	// 获取今天最新的收货单
	latestReceipt, err := receiptOrderRepo.GetLatestReceiptToday()
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", errors.WithMessage(errors.New("获取最新收货单失败"), err.Error())
	}

	// 如果没有查询到今天的收货单，则设置为0001
	if latestReceipt == nil {
		return "0001", nil
	}

	// 从收货单号中提取序列号部分（最后4位）
	if len(latestReceipt.OrderNo) >= 4 {
		lastSerialNo := latestReceipt.OrderNo[len(latestReceipt.OrderNo)-4:]
		serialNoNum, err := strconv.Atoi(lastSerialNo)
		if err != nil {
			// 如果解析失败，重新从0001开始
			return "0001", nil
		}
		// 序列号加1
		newSerialNoNum := serialNoNum + 1
		return fmt.Sprintf("%04d", newSerialNoNum), nil
	}

	// 如果收货单号格式不正确，重新从0001开始
	return "0001", nil
}

// createPurchaseOrderLog 创建采购订单操作日志
func (h *purchaseOrderHelper) createPurchaseOrderLog(
	db *gorm.DB,
	purchaseOrderUuid uint64,
	ctx context.Context,
	action, actionDesc string,
	oldStatus, newStatus int,
	remark string,
) error {
	logRepo := repository.NewPurchaseOrderLogRepo(db)

	log := &model.PurchaseOrderLog{
		PurchaseOrderUuid: purchaseOrderUuid,
		OperatorUuid:      ctx.GetStaffUuid(),
		OperatorName:      ctx.GetStaff().RealName,
		Action:            action,
		ActionDesc:        actionDesc,
		OldStatus:         oldStatus,
		NewStatus:         newStatus,
		Remark:            remark,
	}

	return logRepo.Create(log)
}

// checkAndUpdatePurchaseOrderStatus 检查并更新采购单状态
func (h *purchaseOrderHelper) checkAndUpdatePurchaseOrderStatus(
	ctx context.Context,
	db *gorm.DB,
	purchaseOrderUuid uint64,
) error {
	purchaseOrderRepo := repository.NewPurchaseOrderRepo(db)
	purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(db)

	// 获取所有明细
	items, err := purchaseOrderItemRepo.GetByPurchaseOrderUuid(purchaseOrderUuid)
	if err != nil {
		return err
	}

	// 检查是否全部到货完成
	allCompleted := true
	partialReceived := false

	for _, item := range items {
		if item.ArrivalNum < item.Num {
			allCompleted = false
		}
		if item.ArrivalNum > 0 {
			partialReceived = true
		}
	}

	// 更新采购申请状态
	purchaseOrder, err := purchaseOrderRepo.GetByUuid(purchaseOrderUuid)
	if err != nil {
		return err
	}

	oldStatus := purchaseOrder.Status

	if allCompleted {
		purchaseOrder.Status = constant.PurchaseOrderStatusCompleted
		purchaseOrder.FinalReceiveTime = time.Now().Unix()
	} else if partialReceived && purchaseOrder.Status != constant.PurchaseOrderStatusCompleted {
		if purchaseOrder.FirstReceiveTime == 0 {
			purchaseOrder.FirstReceiveTime = time.Now().Unix()
		}
	}

	if oldStatus != purchaseOrder.Status {
		err = purchaseOrderRepo.Update(purchaseOrder)
		if err != nil {
			return err
		}
	}

	// 创建采购单操作日志
	err = h.createPurchaseOrderLog(db, purchaseOrderUuid, ctx, "update_status", "更新采购申请状态", oldStatus, purchaseOrder.Status, "")
	if err != nil {
		return err
	}

	return nil
}

// updateRelatedMaterialStock 更新规格/加料关联材料库存
func (h *purchaseOrderHelper) updateRelatedMaterialStock(db *gorm.DB, relatedMaterialUuids []uint64) error {
	// 如果材料UUID列表为空，直接返回
	if len(relatedMaterialUuids) == 0 {
		return nil
	}

	// 使用事务确保数据一致性
	return db.Transaction(func(tx *gorm.DB) error {
		// 构建复杂SQL查询来按成本卡更新产品BOM的库存数量
		sql := `
			UPDATE ttpos_product_bom AS pb 
			JOIN (
				SELECT 
					rm.related_uuid, 
					LEAST(IFNULL(
						FLOOR(
							MIN(
								wi.stock / rm.num
							)
						)
					, 0), 99999999) AS min_stock_num
				FROM ttpos_related_material AS rm
				JOIN ttpos_warehouse_item AS wi ON rm.material_uuid = wi.material_uuid
				JOIN ttpos_warehouse AS w ON wi.warehouse_uuid = w.uuid
				WHERE rm.uuid IN (?) 
				  AND rm.delete_time = 0 
				  AND rm.unit_uuid > 0
				  AND w.is_default = 1
				GROUP BY rm.related_uuid
			) AS sub ON pb.product_bom_card_uuid = sub.related_uuid
			SET pb.stock_num = sub.min_stock_num
			WHERE pb.product_bom_card_uuid IN (
				SELECT DISTINCT related_uuid 
				FROM ttpos_related_material 
				WHERE uuid IN (?) 
				AND delete_time = 0 
				AND unit_uuid > 0
			)
		`

		// 执行SQL更新
		err := tx.Exec(sql, relatedMaterialUuids, relatedMaterialUuids).Error
		if err != nil {
			return errors.WithMessage(errors.New("更新规格/加料关联材料库存失败"), err.Error())
		}

		return nil
	})
}

// HeadquarterUpdateInfo 总部更新信息结构
type HeadquarterUpdateInfo struct {
	DB            *gorm.DB                          // 总部数据库连接
	PurchaseOrder *model.PurchaseOrder              // 总部采购订单
	ItemRepo      repository.IPurchaseOrderItemRepo // 总部采购明细Repository
	ItemsToUpdate []HeadquarterItemUpdate           // 需要更新的明细信息
}

// HeadquarterItemUpdate 需要更新的总部明细信息
type HeadquarterItemUpdate struct {
	MaterialCode  string  // 物料编码
	NewArrivalNum float64 // 新的到货数量
}

// initHeadquarterInfo 初始化总部信息
func (h *purchaseOrderHelper) initHeadquarterInfo(
	ctx context.Context,
	dbm *database.DBManager,
	purchaseOrder *model.PurchaseOrder,
) (*HeadquarterUpdateInfo, error) {
	// 获取公司设置
	companySetting := ctx.GetCompanySetting()
	if companySetting.HeadquarterUuid == 0 {
		return nil, errors.New("总部UUID不能为空")
	}

	// 获取总部数据库连接
	headquarterDb := dbm.GetDB(companySetting.HeadquarterUuid)
	if headquarterDb == nil {
		return nil, errors.New("获取总部数据库失败")
	}

	// 获取总部采购订单
	headquarterPurchaseOrder, err := repository.NewPurchaseOrderRepo(headquarterDb).GetBySubUuid(purchaseOrder.Uuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取总部采购申请失败"), err.Error())
	}

	return &HeadquarterUpdateInfo{
		DB:            headquarterDb,
		PurchaseOrder: headquarterPurchaseOrder,
		ItemRepo:      repository.NewPurchaseOrderItemRepo(headquarterDb),
		ItemsToUpdate: make([]HeadquarterItemUpdate, 0),
	}, nil
}

// batchUpdateHeadquarterItems 批量更新总部采购申请明细
func (h *purchaseOrderHelper) batchUpdateHeadquarterItems(
	ctx context.Context,
	info *HeadquarterUpdateInfo,
) error {
	if len(info.ItemsToUpdate) == 0 {
		return nil
	}

	for _, itemUpdate := range info.ItemsToUpdate {
		// 获取总部采购申请明细
		headquarterItem, err := info.ItemRepo.GetByPurchaseOrderUuidAndMaterialCode(
			info.PurchaseOrder.Uuid,
			itemUpdate.MaterialCode,
		)
		if err != nil {
			// 如果找不到对应的明细，记录警告但不中断流程
			logger.Logger.Info("更新总部采购申请明细失败", zap.String("物料编码", itemUpdate.MaterialCode), zap.Error(err))
			continue
		}

		// 更新到货数量
		headquarterItem.ArrivalNum = itemUpdate.NewArrivalNum
		err = info.ItemRepo.Update(headquarterItem)
		if err != nil {
			return errors.WithMessage(
				errors.New(fmt.Sprintf(
					i18n.Translate(ctx.GetLanguage(), "更新总部采购申请明细失败，物料编码：%s"),
					itemUpdate.MaterialCode,
				)),
				err.Error(),
			)
		}
	}

	return nil
}

// recordErpStockInLog 记录ERP入库记录
func (h *purchaseOrderHelper) recordErpStockInLog(
	db *gorm.DB,
	receiptOrder *model.PurchaseReceiptOrder,
) error {
	// 使用事务确保数据一致性
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取仓库出入库日志Repository
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)
		warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
		materialRepo := repository.NewMaterialRepo(tx)

		// 获取目标仓库信息（通过ERP编码查找）
		targetWarehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(receiptOrder.TargetWarehouseErpCode)
		if err != nil {
			return errors.WithMessage(errors.New("获取目标仓库信息失败"), err.Error())
		}

		// 处理每个收货单明细
		for _, item := range receiptOrder.Items {
			actualNum := item.GetActualNum()
			if actualNum <= 0 {
				continue
			}

			// 查找或创建仓库商品库存记录
			warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(targetWarehouse.Uuid, item.MaterialUuid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 没有找到记录时创建新记录
					newWarehouseItem := &model.WarehouseItem{
						WarehouseUuid: targetWarehouse.Uuid,
						MaterialUuid:  item.MaterialUuid,
						MaterialCode:  item.MaterialCode,
						Stock:         0,
						ReservedStock: 0,
						Valuation:     1,
					}
					err = warehouseItemRepo.Create(newWarehouseItem)
					if err != nil {
						return errors.WithMessage(errors.New("创建仓库商品库存记录失败"), err.Error())
					}
					warehouseItem = newWarehouseItem
				} else {
					return errors.WithMessage(errors.New("查询仓库商品库存失败"), err.Error())
				}
			}
			err = warehouseItemRepo.AddStock(warehouseItem.Uuid, actualNum)
			if err != nil {
				return errors.WithMessage(errors.New("更新仓库商品库存失败"), err.Error())
			}

			// 更新规格/加料关联材料库存
			material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid, materialRepo.WithRelatedMaterialList())
			if err != nil {
				return errors.WithMessage(errors.New("获取物品信息失败"), err.Error())
			}
			relatedMaterialUuids := make([]uint64, 0)
			for _, relatedMaterial := range material.RelatedMaterialList {
				if relatedMaterial.IsDelete() {
					continue
				}
				if relatedMaterial.IsUsed == 0 {
					continue
				}
				relatedMaterialUuids = append(relatedMaterialUuids, relatedMaterial.Uuid)
			}
			err = h.updateRelatedMaterialStock(tx, relatedMaterialUuids)
			if err != nil {
				return errors.WithMessage(errors.New("更新规格/加料关联材料库存失败"), err.Error())
			}

			// 记录入库日志
			supplierUuid := func() uint64 {
				supplier, err := repository.NewSupplierRepo(tx).GetByErpCode(receiptOrder.GetSupplierErpCode())
				if err != nil {
					return 0
				}
				return supplier.Uuid
			}()

			warehouseLog := &model.WarehouseInOutLog{
				LogType:              0, // 入库
				Scene:                0, // 采购入库
				WarehouseUuid:        targetWarehouse.Uuid,
				MaterialUuid:         item.MaterialUuid,
				MaterialName:         item.MaterialName,
				MaterialBaseUnitUuid: item.BaseUnitUuid,
				MaterialBaseUnitName: item.BaseUnitName,
				Num:                  actualNum,
				Price:                item.Valuation,
				Amount: decimal.NewFromFloat(item.Valuation).
					Mul(decimal.NewFromFloat(actualNum)).
					InexactFloat64(),
				SupplierUuid:    supplierUuid,
				SupplierErpCode: receiptOrder.GetSupplierErpCode(),
				SupplierName:    receiptOrder.SupplierName,
				OrderNo:         receiptOrder.OrderNo,
				OtherOrgUuid:    supplierUuid,
				OtherOrgType:    0,
				OtherOrgName:    receiptOrder.SupplierName,
			}
			err = warehouseLogRepo.Create(warehouseLog)
			if err != nil {
				return errors.WithMessage(errors.New("记录入库日志失败"), err.Error())
			}
		}

		return nil
	})
}

// reduceHeadquarterStockAndLog 减少总部库存并记录出入库日志
func (h *purchaseOrderHelper) reduceHeadquarterStockAndLog(
	ctx context.Context,
	subDb, headquarterDb *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
) error {
	lang := ctx.GetLanguage()

	// 使用事务确保数据一致性
	return headquarterDb.Transaction(func(tx *gorm.DB) error {
		// 获取仓库商品库存Repository
		warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)
		materialRepo := repository.NewMaterialRepo(tx)

		// 获取目标仓库
		targetWarehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(purchaseOrder.WarehouseErpCode)
		if err != nil {
			return errors.WithMessage(err, "获取总部出库仓库信息失败")
		}

		// 获取在途仓库
		transitWarehouse, _ := repository.NewWarehouseRepo(subDb).GetTransitWarehouse()

		// 处理每个采购明细
		var errMaterialsList []string
		updateMaterialsMap := make(map[uint64]model.PurchaseOrderItem)
		for _, item := range purchaseOrder.Items {
			// 计算实际出库数量（考虑单位转换率）
			actualNum := item.GetConversionRateNum()
			if actualNum <= 0 {
				continue
			}

			// 获取物料UUID
			material, err := materialRepo.GetMaterialByUuid(
				item.MaterialUuid,
				materialRepo.WithRelatedMaterialList(),
			)
			if err != nil {
				return errors.WithMessage(errors.New("获取物品信息失败"), err.Error())
			}
			item.Material = &material

			// 查找或创建仓库商品库存记录
			materialName := *language.JsonToLocaleResponse(item.MaterialName)
			warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(targetWarehouse.Uuid, item.MaterialUuid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					errMaterialsList = append(errMaterialsList, materialName.GetLocale(lang))
				} else {
					return errors.WithMessage(err, "查询仓库商品库存失败")
				}
				// 如果查询失败，跳过后续处理
				continue
			} else if warehouseItem.Stock < actualNum {
				errMaterialsList = append(errMaterialsList, materialName.GetLocale(lang))
			}

			// 添加到更新库存列表
			updateMaterialsMap[warehouseItem.Uuid] = item
		}
		if len(errMaterialsList) > 0 {
			return errors.NewWithCodeAndData(
				constant.CodeWarehouseStockNotEnough,
				map[string][]string{
					"material_names": errMaterialsList,
				},
				strings.Join(errMaterialsList, ", ")+" "+
					i18n.Translate(lang, "的物品库存不足")+"\n"+
					i18n.Translate(lang, "请补充库存"),
			)
		}

		// 获取供应商ID
		supplierUuid := func() uint64 {
			supplier, err := repository.NewSupplierRepo(tx).GetByErpCode(purchaseOrder.SupplierErpCode)
			if err != nil {
				return 0
			}
			return supplier.Uuid
		}()

		// 减少库存
		for warehouseItemUuid, item := range updateMaterialsMap {
			actualNum := item.GetConversionRateNum()

			// 更新规格/加料关联材料库存
			err = warehouseItemRepo.ReduceStock(warehouseItemUuid, actualNum)
			if err != nil {
				return errors.WithMessage(errors.New("减少总部库存失败"), err.Error())
			}

			// 更新规格/加料关联材料库存
			relatedMaterialUuids := make([]uint64, 0)
			for _, relatedMaterial := range item.Material.RelatedMaterialList {
				if relatedMaterial.IsDelete() || relatedMaterial.IsUsed == 0 {
					continue
				}
				relatedMaterialUuids = append(relatedMaterialUuids, relatedMaterial.Uuid)
			}
			err = h.updateRelatedMaterialStock(tx, relatedMaterialUuids)
			if err != nil {
				return errors.WithMessage(errors.New("更新规格/加料关联材料库存失败"), err.Error())
			}

			// 记录出库日志
			warehouseLog := &model.WarehouseInOutLog{
				LogType:              1, // 出库
				Scene:                2, // 发货出库
				WarehouseUuid:        targetWarehouse.Uuid,
				MaterialUuid:         item.MaterialUuid,
				MaterialName:         item.MaterialName,
				MaterialBaseUnitUuid: item.BaseUnitUuid,
				MaterialBaseUnitName: item.BaseUnitName,
				Num:                  actualNum,
				Price:                item.Valuation,
				Amount: decimal.NewFromFloat(item.Valuation).
					Mul(decimal.NewFromFloat(actualNum)).
					InexactFloat64(),
				OrderNo:         purchaseOrder.OrderNo,
				SupplierUuid:    supplierUuid,
				SupplierErpCode: purchaseOrder.SupplierErpCode,
				SupplierName:    purchaseOrder.SupplierName,
				OtherOrgUuid:    purchaseOrder.CompanyUuid,
				OtherOrgType:    1,
				OtherOrgName:    purchaseOrder.CompanyName,
			}
			err = warehouseLogRepo.Create(warehouseLog)
			if err != nil {
				return errors.WithMessage(errors.New("记录出库日志失败"), err.Error())
			}

			// 添加到子店的在途仓库
			if transitWarehouse != nil {
				material, err := repository.NewMaterialRepo(subDb).GetMaterialByErpCode(item.MaterialCode)
				if err != nil {
					continue
				}
				item.MaterialUuid = material.Uuid
				err = h.AddToTransitWarehouse(subDb, transitWarehouse, purchaseOrder, supplierUuid, &item, actualNum)
				if err != nil {
					return errors.WithMessage(errors.New("添加到在途仓库失败"), err.Error())
				}
			}
		}

		return nil
	})
}

// extractName 从错误信息中提取供应商名称
func (h *purchaseOrderHelper) extractName(name, after, errorMsg string) string {
	// 转义正则表达式中的特殊字符
	escapedName := regexp.QuoteMeta(name)
	escapedAfter := regexp.QuoteMeta(after)

	// 使用正则表达式匹配，支持HTML标签和普通文本
	var re *regexp.Regexp
	if strings.Contains(name, "<") || strings.Contains(after, "<") {
		// HTML标签模式：不要求空格分隔
		re = regexp.MustCompile(escapedName + `(.+?)` + escapedAfter)
	} else {
		// 普通文本模式：要求空格分隔
		re = regexp.MustCompile(escapedName + `\s+(.+?)\s+` + escapedAfter)
	}

	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		supplierInfo := strings.TrimSpace(matches[1])
		// 如果包含编码#名称格式，提取名称部分
		if strings.Contains(supplierInfo, "#") {
			parts := strings.SplitN(supplierInfo, "#", 2)
			if len(parts) == 2 {
				return parts[1] // 返回供应商名称
			}
		}
		return supplierInfo
	}
	return ""
}

// handleErpError 处理ERP错误
func (h *purchaseOrderHelper) handleErpError(ctx context.Context, err error, purchaseOrder *model.PurchaseOrder) error {
	// 检查供应商状态
	if strings.Contains(err.Error(), "Supplier") && strings.Contains(err.Error(), "is disabled") {
		// 提取供应商名称
		// supplierName := h.extractName("Supplier", "is disabled", err.Error())
		// if supplierName != "" {
		// 	return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, fmt.Sprintf(
		// 		i18n.Translate(ctx.GetLanguage(), "ERP中供应商 %s 已禁用，请修改供应商状态"), supplierName),
		// 	)
		// }
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "供应商已禁用，请修改供应商状态")
	}
	// 检查物品状态
	if strings.Contains(err.Error(), "Item") && strings.Contains(err.Error(), "is disabled") {
		// 提取物品名称
		itemName := h.extractName("Item", "is disabled", err.Error())
		if itemName != "" {
			return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, fmt.Sprintf(
				i18n.Translate(ctx.GetLanguage(), "物品 %s 已禁用，请修改物品状态"), itemName),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "有物品已禁用，请修改物品状态")
	}
	// 检查仓库状态
	if strings.Contains(err.Error(), "Warehouse") && strings.Contains(err.Error(), "is disabled") {
		// 提取仓库名称
		warehouseName := h.extractName("Warehouse", "is disabled", err.Error())
		if warehouseName != "" {
			return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, fmt.Sprintf("仓库 %s 已禁用，请修改仓库状态", warehouseName))
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "仓库已禁用，请修改仓库状态")
	}
	// 检查采购数量
	if strings.Contains(err.Error(), "cannot be less than minimum order qty") {
		itemName := h.extractName("Item", ":Ordered qty", err.Error())
		if itemName != "" {
			num := h.extractName("minimum order qty", "(defined in Item).", err.Error())
			if num != "" {
				return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
					itemName+" "+i18n.Translate(ctx.GetLanguage(), "采购数量不能小于ERP中设置的最小采购数量")+" "+num,
				)
			}
			return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
				itemName+" "+i18n.Translate(ctx.GetLanguage(), "采购数量不能小于ERP中设置的最小采购数量"),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "采购数量不能小于ERP中设置的最小采购数量")
	}
	// before Transaction Date
	if strings.Contains(err.Error(), "before Transaction Date") {
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, i18n.Translate(ctx.GetLanguage(), "期望到货日期不能小于今天"))
	}
	// 检查单位是否为整数
	if strings.Contains(err.Error(), "Must be Whole Number") {
		itemName := h.extractName("UOM <strong>", "</strong>", err.Error())
		if itemName != "" {
			return errors.NewWithCode(
				constant.CodePurchaseOrderSupplierDisabled,
				fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "单位 %s 只能使用整数"), itemName),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "单位只能使用整数")
	}
	// 检查发货库存不足
	if strings.Contains(err.Error(), "创建发货单失败") && strings.Contains(err.Error(), "NegativeStockError") {
		itemName := h.extractName("Item ", "</a> needed in", err.Error())
		if itemName != "" {
			return errors.NewWithCode(
				constant.CodePurchaseOrderSupplierDisabled,
				fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 库存不足，请补充库存"), itemName),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "物品库存不足，请补充库存")
	}
	// 创建采购单失败
	if purchaseOrder != nil && strings.Contains(err.Error(), "调用erp接口返回空") {
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
			i18n.Translate(ctx.GetLanguage(), "审批失败")+": "+purchaseOrder.OrderNo,
		)
	}
	return err
}

// addToTransitWarehouse 添加到在途仓库
func (h *purchaseOrderHelper) AddToTransitWarehouse(
	tx *gorm.DB,
	transitWarehouse *model.Warehouse,
	purchaseOrder *model.PurchaseOrder,
	supplierUuid uint64,
	item *model.PurchaseOrderItem,
	actualNum float64,
) error {
	// 添加到本店的在途仓库
	if transitWarehouse == nil {
		return nil
	}
	// 获取在途仓库出入库日志Repository
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)
	// 获取在途仓库库存
	warehouseItem, err := repository.NewWarehouseItemRepo(tx).GetByWarehouseAndMaterial(
		transitWarehouse.Uuid,
		item.MaterialUuid,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有找到记录时创建新记录
			newWarehouseItem := &model.WarehouseItem{
				WarehouseUuid: transitWarehouse.Uuid,
				MaterialUuid:  item.MaterialUuid,
				MaterialCode:  item.MaterialCode,
				Stock:         0,
				Valuation:     item.Valuation,
			}
			err = repository.NewWarehouseItemRepo(tx).Create(newWarehouseItem)
			if err != nil {
				return errors.WithMessage(errors.New("创建仓库商品库存记录失败"), err.Error())
			}
			warehouseItem = newWarehouseItem
		} else {
			return errors.WithMessage(errors.New("查询在途仓库库存失败"), err.Error())
		}
	}
	//
	err = repository.NewWarehouseItemRepo(tx).AddStock(warehouseItem.Uuid, actualNum)
	if err != nil {
		return errors.WithMessage(errors.New("增加在途仓库库存失败"), err.Error())
	}
	// 记录在途仓出库日志
	warehouseLog := &model.WarehouseInOutLog{
		LogType:              0,  // 入库
		Scene:                20, // 在途入库
		WarehouseUuid:        transitWarehouse.Uuid,
		MaterialUuid:         item.MaterialUuid,
		MaterialName:         item.MaterialName,
		MaterialBaseUnitUuid: item.BaseUnitUuid,
		MaterialBaseUnitName: item.BaseUnitName,
		Num:                  actualNum,
		Price:                item.Valuation,
		Amount: decimal.NewFromFloat(item.Valuation).
			Mul(decimal.NewFromFloat(actualNum)).
			InexactFloat64(),
		SupplierUuid:    supplierUuid,
		SupplierErpCode: purchaseOrder.SupplierErpCode,
		SupplierName:    purchaseOrder.SupplierName,
		OrderNo:         purchaseOrder.OrderNo,
		OtherOrgUuid:    purchaseOrder.CompanyUuid,
		OtherOrgType:    0,
		OtherOrgName:    purchaseOrder.CompanyName,
	}
	err = warehouseLogRepo.Create(warehouseLog)
	if err != nil {
		return errors.WithMessage(errors.New("记录入库日志失败"), err.Error())
	}

	return nil
}
