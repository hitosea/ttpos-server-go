package purchase_order

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
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

// NewPurchaseOrderHelper 创建采购订单辅助方法
func NewPurchaseOrderHelper() *purchaseOrderHelper {
	return &purchaseOrderHelper{}
}

// generateOrderNo 生成采购申请/品牌采购订单编号
// 格式：prefix + yyyyMMddHHmmss + 序列号（4位）
// 例如：PR202504030915120001, TPHY202504030915120001
func (h *purchaseOrderHelper) generateOrderNo(
	saasDB *gorm.DB,
	companyUuid uint64,
	prefix string,
	numberType string,
	timezone string,
) (string, error) {
	// 获取秒级时间戳
	now := utils.SetTimezone(timezone).Now()
	timestamp := now.Format("20060102150405") // yyyyMMddHHmmss

	// 获取日期字符串（用于序列号表）
	dateStr := now.Format("2006-01-02") // YYYY-MM-DD

	// 从 ttpos_number_sequence 表获取下一个序列号
	seqRepo := repository.NewNumberSequenceRepo(saasDB)
	seq, err := seqRepo.GetNextSequence(companyUuid, numberType, dateStr)
	if err != nil {
		return "", errors.WithMessage(err, "获取序列号失败")
	}

	// 组装编号：prefix + timestamp + 序列号（4位）
	orderNo := fmt.Sprintf("%s%s%04d", prefix, timestamp, seq)
	return orderNo, nil
}

// generateReceiptNo 生成采购收货/品采收货编号
// 格式：prefix + yyyyMMddHHmmss + 序列号（4位）
// 例如：PRC202504030915120001, TPHY202504030915120001
func (h *purchaseOrderHelper) generateReceiptNo(
	saasDB *gorm.DB,
	companyUuid uint64,
	prefix string,
	numberType string,
	timezone string,
) (string, error) {
	// 获取秒级时间戳
	now := utils.SetTimezone(timezone).Now()
	timestamp := now.Format("20060102150405") // yyyyMMddHHmmss

	// 获取日期字符串（用于序列号表）
	dateStr := now.Format("2006-01-02") // YYYY-MM-DD

	// 从 ttpos_number_sequence 表获取下一个序列号
	seqRepo := repository.NewNumberSequenceRepo(saasDB)
	seq, err := seqRepo.GetNextSequence(companyUuid, numberType, dateStr)
	if err != nil {
		return "", errors.WithMessage(err, "获取序列号失败")
	}

	// 组装编号：prefix + timestamp + 序列号（4位）
	receiptNo := fmt.Sprintf("%s%s%04d", prefix, timestamp, seq)
	return receiptNo, nil
}

// translatePurchaseOrderLogRemark 仅翻译特定的采购订单日志备注
func translatePurchaseOrderLogRemark(lang, remark string) string {
	if remark == "品牌采购自动审批" {
		return i18n.Translate(lang, remark)
	}
	return remark
}

// createPurchaseOrderLog 创建采购订单操作日志
func (h *purchaseOrderHelper) createPurchaseOrderLog(
	db *gorm.DB,
	purchaseOrderUuid uint64,
	ctx context.Context,
	action, actionDesc string,
	oldStatus, newStatus int,
	remark string,
	content string,
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
		Content:           content,
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
	err = h.createPurchaseOrderLog(db, purchaseOrderUuid, ctx, "update_status", "更新采购申请状态", oldStatus, purchaseOrder.Status, "", "{}")
	if err != nil {
		return err
	}

	return nil
}

// HeadquarterUpdateInfo 总部更新信息结构
type HeadquarterUpdateInfo struct {
	DB            *gorm.DB                              // 总部数据库连接
	PurchaseOrder *model.PurchaseOrder                  // 总部采购订单
	ItemRepo      repository.IPurchaseOrderItemRepo     // 总部采购明细Repository
	ItemUnitRepo  repository.IPurchaseOrderItemUnitRepo // 总部采购明细单位Repository
	ItemsToUpdate []HeadquarterItemUpdate               // 需要更新的明细信息
}

// HeadquarterItemUpdate 需要更新的总部明细信息
type HeadquarterItemUpdate struct {
	MaterialCode  string                                   // 物料编码
	NewArrivalNum float64                                  // 新的到货数量
	UnitList      []req.PurchaseReceiptItemMaterialUnitReq // 单位列表
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
		ItemUnitRepo:  repository.NewPurchaseOrderItemUnitRepo(headquarterDb),
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
			logger.Logger.Error("更新总部采购申请明细失败", zap.String("物料编码", itemUpdate.MaterialCode), zap.Error(err))
			continue
		}

		// 更新总部采购申请明细单位到货数量
		for _, unit := range itemUpdate.UnitList {
			headquarterItemUnit, err := info.ItemUnitRepo.GetByUuid(unit.Uuid)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					return errors.WithMessage(errors.New("获取总部采购申请明细单位失败"), err.Error())
				}
			}
			if headquarterItemUnit != nil {
				headquarterItemUnit.ArrivalNum = headquarterItemUnit.ArrivalNum + unit.Num
				err = info.ItemUnitRepo.Update(*headquarterItemUnit)
				if err != nil {
					return errors.WithMessage(errors.New("更新总部采购申请明细单位到货数量失败"), err.Error())
				}
			}
		}

		// 更新到货数量
		headquarterItem.ArrivalNum = itemUpdate.NewArrivalNum
		headquarterItem.SetNil()
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

		// 获取目标仓库信息（通过ERP编码查找）
		targetWarehouse, err := repository.NewWarehouseRepo(tx).GetByErpCode(receiptOrder.TargetWarehouseErpCode)
		if err != nil {
			logger.Logger.Error("recordErpStockInLog-GetByErpCode", zap.Any("targetWarehouseErpCode", receiptOrder.TargetWarehouseErpCode), zap.Any("err", err))
			return errors.WithMessage(errors.New("获取目标仓库信息失败"), err.Error())
		}

		// 处理每个收货单明细
		for _, item := range receiptOrder.Items {
			actualNum := item.GetUnitsTotalConversionRateNum()
			if actualNum <= 0 {
				continue
			}

			// 查找或创建仓库商品库存记录
			warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterialOrCreate(targetWarehouse.Uuid, item.MaterialUuid, item.MaterialCode, item.Valuation)
			if err != nil {
				logger.Logger.Error("MoveStockToTargetWarehouse-GetByWarehouseAndMaterialOrCreate", zap.Any("err", err))
				return errors.WithMessage(errors.New("查询仓库商品库存失败"), err.Error())
			}
			err = warehouseItemRepo.AddStock(warehouseItem.Uuid, actualNum)
			if err != nil {
				logger.Logger.Error("recordErpStockInLog-AddStock", zap.Any("warehouseItemUuid", warehouseItem.Uuid), zap.Any("actualNum", actualNum), zap.Any("err", err))
				return errors.WithMessage(errors.New("更新仓库商品库存失败"), err.Error())
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
				LogType:              constant.WarehouseInOutLogLogTypeIn,     // 入库
				Scene:                constant.WarehouseInOutLogScenePurchase, // 采购入库
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
				logger.Logger.Error("recordErpStockInLog-Create", zap.Any("warehouseLog", warehouseLog), zap.Any("err", err))
				return errors.WithMessage(errors.New("记录入库日志失败"), err.Error())
			}
		}

		return nil
	})
}

// reduceHeadquarterStockAndLog 减少总部库存并记录出入库日志
// itemDefaultWarehouseMap: 物品编码 -> 本地仓库UUID的映射（从 ERP item_defaults 查询得到），新流程时使用
func (h *purchaseOrderHelper) reduceHeadquarterStockAndLog(
	ctx context.Context,
	subDb, headquarterDb *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
	itemDefaultWarehouseMap map[string]uint64,
) error {
	lang := ctx.GetLanguage()

	// 使用事务确保数据一致性
	return headquarterDb.Transaction(func(tx *gorm.DB) error {
		// 获取仓库商品库存Repository
		warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)
		materialRepo := repository.NewMaterialRepo(tx)

		// 获取在途仓库
		transitWarehouse, _ := repository.NewWarehouseRepo(subDb).GetTransitWarehouse()

		// 处理每个采购明细
		type stockReduceItem struct {
			warehouseItemUuid uint64
			warehouseUuid     uint64
			item              model.PurchaseOrderItem
		}
		var errMaterialsList []string
		var reduceItems []stockReduceItem

		for _, item := range purchaseOrder.Items {
			// 直采物品（由供应商直接发货）不扣减总部库存
			if item.DeliveredBySupplier == 1 {
				continue
			}

			actualNum := item.GetUnitsTotalConversionRateNum()
			if actualNum <= 0 {
				continue
			}

			// 获取物料信息
			material, err := materialRepo.GetMaterialByUuid(
				item.MaterialUuid,
				materialRepo.WithRelatedMaterialList(),
			)
			if err != nil {
				return errors.WithMessage(errors.New("获取物品信息失败"), err.Error())
			}
			item.Material = &material

			// 确定出库仓库：使用 ERP item_defaults 中物品的默认仓库
			var targetWarehouseUuid uint64
			if whUuid, ok := itemDefaultWarehouseMap[item.MaterialCode]; ok && whUuid != 0 {
				targetWarehouseUuid = whUuid
			} else {
				// 物品在 ERP 未配置默认仓库，跳过库存扣减（ERP侧会走保底方案）
				logger.Logger.Warn("reduceHeadquarterStockAndLog-物品未配置默认仓库，跳过库存扣减",
					zap.Uint64("material_uuid", item.MaterialUuid),
					zap.String("material_code", item.MaterialCode),
					zap.Uint64("company_uuid", purchaseOrder.CompanyUuid),
				)
				continue
			}

			// 查找仓库商品库存记录
			materialName := *language.JsonToLocaleResponse(item.MaterialName)
			warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(targetWarehouseUuid, item.MaterialUuid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					errMaterialsList = append(errMaterialsList, materialName.GetLocale(lang))
				} else {
					logger.Logger.Error("reduceHeadquarterStockAndLog-GetByWarehouseAndMaterial", zap.Any("targetWarehouseUuid", targetWarehouseUuid), zap.Any("itemMaterialUuid", item.MaterialUuid), zap.Any("err", err))
					return errors.WithMessage(errors.New("查询仓库商品库存失败"), err.Error())
				}
				continue
			} else if warehouseItem.Stock < actualNum {
				errMaterialsList = append(errMaterialsList, materialName.GetLocale(lang))
			}

			reduceItems = append(reduceItems, stockReduceItem{
				warehouseItemUuid: warehouseItem.Uuid,
				warehouseUuid:     targetWarehouseUuid,
				item:              item,
			})
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
		for _, reduceItem := range reduceItems {
			item := reduceItem.item
			actualNum := item.GetUnitsTotalConversionRateNum()

			// 减少仓库物品库存
			err := warehouseItemRepo.ReduceStock(reduceItem.warehouseItemUuid, actualNum)
			if err != nil {
				logger.Logger.Error("reduceHeadquarterStockAndLog-ReduceStock", zap.Any("warehouseItemUuid", reduceItem.warehouseItemUuid), zap.Any("actualNum", actualNum), zap.Any("err", err))
				return errors.WithMessage(errors.New("减少总部库存失败"), err.Error())
			}

			// 记录出库日志
			warehouseLog := &model.WarehouseInOutLog{
				LogType:              constant.WarehouseInOutLogLogTypeOut,    // 出库
				Scene:                constant.WarehouseInOutLogSceneDelivery, // 发货出库
				WarehouseUuid:        reduceItem.warehouseUuid,
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
				logger.Logger.Error("reduceHeadquarterStockAndLog-Create", zap.Any("warehouseLog", warehouseLog), zap.Any("err", err))
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
					logger.Logger.Error("reduceHeadquarterStockAndLog-AddToTransitWarehouse", zap.Any("transitWarehouse", transitWarehouse), zap.Any("purchaseOrder", purchaseOrder), zap.Any("supplierUuid", supplierUuid), zap.Any("item", item), zap.Any("actualNum", actualNum), zap.Any("err", err))
					return errors.WithMessage(errors.New("添加到在途仓库失败"), err.Error())
				}
			}
		}

		return nil
	})
}

// extractName 从错误信息中提取名称
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

// extractOverReceiptInfo 从超收错误信息中提取物品编码和超额数量
// 错误消息格式: "This document is over limit by <strong>Qty</strong> <strong>5.0</strong> for item <strong>BE02003</strong>..."
func (h *purchaseOrderHelper) extractOverReceiptInfo(errorMsg string) (itemCode string, overQty string) {
	// 提取物品编码: for item <strong>BE02003</strong>
	itemCodeRe := regexp.MustCompile(`for item\s*<strong>([^<]+)</strong>`)
	itemCodeMatches := itemCodeRe.FindStringSubmatch(errorMsg)
	if len(itemCodeMatches) > 1 {
		itemCode = strings.TrimSpace(itemCodeMatches[1])
	}

	// 提取超额数量: <strong>Qty</strong> <strong>5.0</strong>
	qtyRe := regexp.MustCompile(`<strong>Qty</strong>\s*<strong>([^<]+)</strong>`)
	qtyMatches := qtyRe.FindStringSubmatch(errorMsg)
	if len(qtyMatches) > 1 {
		overQty = strings.TrimSpace(qtyMatches[1])
	}

	return itemCode, overQty
}

// getItemNameByCode 根据物品编码获取物品名称
func (h *purchaseOrderHelper) getItemNameByCode(ctx context.Context, itemCode string) string {
	db := ctx.GetDB()
	if db == nil {
		return ""
	}

	materialRepo := repository.NewMaterialRepo(db)
	material, err := materialRepo.GetMaterialByCode(itemCode, materialRepo.WithMultiLanguageName())
	if err != nil {
		return ""
	}

	// 优先使用多语言名称
	if name := material.MultiLanguageName.GetNameByLangWithFallback(ctx.GetLanguage()); name != "" {
		return name
	}

	return material.Name
}

// handleErpError 处理ERP错误
func (h *purchaseOrderHelper) handleErpError(ctx context.Context, err error, purchaseOrder *model.PurchaseOrder) error {
	// 记录日志，方便排查问题
	if purchaseOrder != nil {
		logger.Logger.Error("handleErpError", zap.Any("purchaseOrder", purchaseOrder), zap.Any("err", err))
	} else {
		logger.Logger.Error("handleErpError", zap.Any("err", err))
	}
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
	// 检查超收限额错误
	// 错误消息格式: "This document is over limit by <strong>Qty</strong> <strong>5.0</strong> for item <strong>BE02003</strong>..."
	if strings.Contains(err.Error(), "Over Receipt/Delivery Allowance") {
		itemCode, overQty := h.extractOverReceiptInfo(err.Error())
		if itemCode != "" {
			// 尝试根据物品编码获取物品名称
			itemName := h.getItemNameByCode(ctx, itemCode)
			if itemName == "" {
				itemName = itemCode // 如果获取不到名称，使用编码
			}
			return errors.NewWithCode(
				constant.CodePurchaseOrderSupplierDisabled,
				fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 数量超出限额 %s，如需超额收货，请设置物品超收限额比例"), itemName, overQty),
			)
		}
		return errors.NewWithCode(
			constant.CodePurchaseOrderSupplierDisabled,
			i18n.Translate(ctx.GetLanguage(), "收货数量超出限额，如需超额收货，请设置物品超收限额比例"),
		)
	}
	// 未知错误
	return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
		i18n.Translate(ctx.GetLanguage(), "操作失败")+": "+purchaseOrder.OrderNo,
	)
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
	warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
	// 获取在途仓库库存
	warehouseItem, err := warehouseItemRepo.GetTransitWarehouseItemByWarehouseAndMaterial(
		transitWarehouse.Uuid,
		item.MaterialUuid,
		item.MaterialCode,
		item.Valuation,
	)
	if err != nil {
		return errors.WithMessage(errors.New("查询在途仓库库存失败"), err.Error())
	}
	//
	err = warehouseItemRepo.AddStock(warehouseItem.Uuid, actualNum)
	if err != nil {
		return errors.WithMessage(errors.New("增加在途仓库库存失败"), err.Error())
	}
	// 记录在途仓出库日志
	warehouseLog := &model.WarehouseInOutLog{
		LogType:              constant.WarehouseInOutLogLogTypeIn,      // 入库
		Scene:                constant.WarehouseInOutLogSceneTransitIn, // 在途入库
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

// checkPurchaseOrderNeedUpdate 检查采购申请是否需要更新
// 通过对比现有数据和请求数据，判断是否有变动，避免不必要的数据库操作
func (h *purchaseOrderHelper) checkPurchaseOrderNeedUpdate(
	purchaseOrder *model.PurchaseOrder,
	req *req.PurchaseOrderUpdateReq,
	itemRepo repository.IPurchaseOrderItemRepo,
) (bool, error) {
	needUpdate := false

	// 1. 检查基本信息是否变动
	if purchaseOrder.Status != constant.PurchaseOrderStatusPending &&
		purchaseOrder.Status != constant.PurchaseOrderStatusHeadquarterPending {
		// 设置期望到货时间，如果为空则默认为2035-12-31
		expectArrivalTime := req.ExpectedDeliveryTime
		if expectArrivalTime == 0 {
			expectArrivalTime = 2082672000 // 2035-12-31的时间戳
		}

		if purchaseOrder.SupplierName != req.SupplierName ||
			purchaseOrder.SupplierErpCode != req.SupplierErpCode ||
			purchaseOrder.ExpectArrivalTime != expectArrivalTime ||
			purchaseOrder.WarehouseErpCode != req.WarehouseErpCode {
			needUpdate = true
		}
	}

	// 2. 检查明细是否变动（物料数量或明细列表）
	if purchaseOrder.Num != float64(len(req.Items)) {
		needUpdate = true
	} else {
		// 查询现有明细
		existingItems, err := itemRepo.GetByPurchaseOrderUuid(
			req.Uuid,
			itemRepo.WithPreloadUnits(),
		)
		if err != nil {
			return false, errors.WithMessage(errors.New("查询现有明细失败"), err.Error())
		}

		// 构建现有明细映射：MaterialUuid -> Item
		existingItemMap := make(map[uint64]*model.PurchaseOrderItem)
		for i := range existingItems {
			existingItemMap[existingItems[i].MaterialUuid] = &existingItems[i]
		}

		// 构建请求明细映射：MaterialUuid -> Units
		type unitInfo struct {
			Uuid uint64
			Num  float64
		}
		reqItemMap := make(map[uint64][]unitInfo)
		for _, item := range req.Items {
			units := make([]unitInfo, len(item.UnitList))
			for i, unit := range item.UnitList {
				units[i] = unitInfo{
					Uuid: unit.Uuid,
					Num:  unit.Num,
				}
			}
			reqItemMap[item.MaterialUuid] = units
		}

		// 检查物料列表是否一致
		if len(existingItemMap) != len(reqItemMap) {
			needUpdate = true
		} else {
			// 逐个对比物料和单位
			for materialUuid, reqUnits := range reqItemMap {
				existingItem, exists := existingItemMap[materialUuid]
				if !exists {
					needUpdate = true
					break
				}

				// 对比单位列表
				if len(existingItem.Units) != len(reqUnits) {
					needUpdate = true
					break
				}

				// 构建现有单位映射：UnitUuid -> Num
				existingUnitMap := make(map[uint64]float64)
				for _, unit := range existingItem.Units {
					existingUnitMap[unit.UnitUuid] = unit.Num
				}

				// 检查每个单位的数量是否一致
				for _, reqUnit := range reqUnits {
					if existingNum, ok := existingUnitMap[reqUnit.Uuid]; !ok || existingNum != reqUnit.Num {
						needUpdate = true
						break
					}
				}

				if needUpdate {
					break
				}
			}
		}
	}

	return needUpdate, nil
}

// checkCompanyShopNeedSync 检查子店铺采购申请是否需要同步
// 通过对比子店铺现有数据和总部当前数据，判断是否有变动，避免不必要的同步操作
func (h *purchaseOrderHelper) checkCompanyShopNeedSync(
	companyOrder *model.PurchaseOrder,
	currentItems []model.PurchaseOrderItem,
	reqItems []req.PurchaseOrderItemUpdateReq,
) bool {
	// 构建现有明细的映射：MaterialCode -> Item
	existingItemMap := make(map[string]*model.PurchaseOrderItem)
	for i := range companyOrder.Items {
		existingItemMap[companyOrder.Items[i].MaterialCode] = &companyOrder.Items[i]
	}

	// 构建当前明细的映射：MaterialUuid -> MaterialCode 和 MaterialCode -> Item
	materialCodeMap := make(map[uint64]string)
	currentItemMap := make(map[string]*model.PurchaseOrderItem)
	for i := range currentItems {
		materialCodeMap[currentItems[i].MaterialUuid] = currentItems[i].MaterialCode
		currentItemMap[currentItems[i].MaterialCode] = &currentItems[i]
	}

	// 构建请求中的物品映射：MaterialCode -> MaterialUuid
	reqMaterialCodes := make(map[string]uint64)
	for _, item := range reqItems {
		if materialCode, ok := materialCodeMap[item.MaterialUuid]; ok {
			reqMaterialCodes[materialCode] = item.MaterialUuid
		}
	}

	// 1. 检查物料数量是否一致
	if len(existingItemMap) != len(reqMaterialCodes) {
		return true
	}

	// 2. 检查每个物料是否存在且数量、单位一致
	for materialCode, materialUuid := range reqMaterialCodes {
		existingItem, existsInCompany := existingItemMap[materialCode]
		currentItem, existsInCurrent := currentItemMap[materialCode]

		// 子店铺没有这个物料，需要同步
		if !existsInCompany {
			return true
		}

		// 总部没有这个物料（理论上不应该发生），需要同步
		if !existsInCurrent {
			return true
		}

		// 检查数量是否一致
		if existingItem.Num != currentItem.Num {
			return true
		}

		// 检查单位列表是否一致
		if len(existingItem.Units) != len(currentItem.Units) {
			return true
		}

		// 构建单位映射：ErpnextUom -> Num
		existingUnitMap := make(map[string]float64)
		for _, unit := range existingItem.Units {
			existingUnitMap[unit.ErpnextUom] = unit.Num
		}

		currentUnitMap := make(map[string]float64)
		for _, unit := range currentItem.Units {
			currentUnitMap[unit.ErpnextUom] = unit.Num
		}

		// 检查每个单位的数量是否一致
		for erpnextUom, existingNum := range existingUnitMap {
			if currentNum, ok := currentUnitMap[erpnextUom]; !ok || currentNum != existingNum {
				return true
			}
		}

		// 避免未使用变量警告
		_ = materialUuid
	}

	// 没有变动
	return false
}

// 品牌采购：批量查询限购配置（包含最大和最小限购数量，避免 N+1 查询问题）
func (h *purchaseOrderHelper) getQuotaLimitMap(
	ctx context.Context,
	dbm *database.DBManager,
	purchaseOrder *model.PurchaseOrder,
) map[string]repository.QuotaLimitConfig {
	companySetting := ctx.GetCompanySetting()
	var quotaLimitMap map[string]repository.QuotaLimitConfig
	if purchaseOrder.IsHeadquarterPurchase() && len(purchaseOrder.Items) > 0 {
		// 提取所有物品编码
		materialCodes := make([]string, 0, len(purchaseOrder.Items))
		for _, item := range purchaseOrder.Items {
			materialCodes = append(materialCodes, item.MaterialCode)
		}
		// 批量查询限购配置
		headquarterUuid := companySetting.HeadquarterUuid
		if headquarterUuid > 0 {
			headquarterDb := dbm.GetDB(headquarterUuid)
			schemeRepo := repository.NewPurchaseLimitSchemeRepo(headquarterDb)
			quotaLimits, err := schemeRepo.GetQuotaLimitConfigBatchByMaterialCodes(
				companySetting.CompanyUuid,
				materialCodes,
				utils.SetTimezone(companySetting.GetTimezone()).CurrentWeekday(),
			)
			if err != nil {
				logger.Logger.Error("批量查询限购配置失败", zap.Error(err))
				return quotaLimitMap
			}
			quotaLimitMap = quotaLimits
		}
	}
	return quotaLimitMap
}

// 获取禁止采购的物品编码列表
func (h *purchaseOrderHelper) getDisallowedPurchaseMaterials(
	ctx context.Context,
	dbm *database.DBManager,
	purchaseOrder *model.PurchaseOrder,
) []string {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	if headquarterUuid == 0 {
		return nil
	}
	if purchaseOrder.IsHeadquarterPurchase() && len(purchaseOrder.Items) > 0 {
		// 提取所有物品编码
		materialCodes := make([]string, 0, len(purchaseOrder.Items))
		for _, item := range purchaseOrder.Items {
			materialCodes = append(materialCodes, item.MaterialCode)
		}
		// 查询禁止采购的物品
		headquarterDb := dbm.GetDB(headquarterUuid)
		schemeRepo := repository.NewPurchaseLimitSchemeRepo(headquarterDb)
		disallowedCodes, err := schemeRepo.GetDisallowedPurchaseMaterialCodes(
			companySetting.CompanyUuid,
			materialCodes,
			utils.SetTimezone(companySetting.GetTimezone()).CurrentWeekday(),
		)
		if err != nil {
			logger.Logger.Error("获取禁止采购物品失败", zap.Error(err))
			return nil
		}
		return disallowedCodes
	}
	return nil
}

// 获取当天最小的每日申请次数限制
func (h *purchaseOrderHelper) getMinDailyLimit(
	ctx context.Context,
	dbm *database.DBManager,
	purchaseOrder *model.PurchaseOrder,
) int {
	companySetting := ctx.GetCompanySetting()
	headquarterUuid := companySetting.HeadquarterUuid
	if headquarterUuid == 0 {
		return -1
	}
	if purchaseOrder.IsHeadquarterPurchase() && len(purchaseOrder.Items) > 0 {
		// 提取所有物品编码
		materialCodes := make([]string, 0, len(purchaseOrder.Items))
		for _, item := range purchaseOrder.Items {
			materialCodes = append(materialCodes, item.MaterialCode)
		}
		// 批量查询限购配置
		headquarterUuid := companySetting.HeadquarterUuid
		if headquarterUuid > 0 {
			headquarterDb := dbm.GetDB(headquarterUuid)
			schemeRepo := repository.NewPurchaseLimitSchemeRepo(headquarterDb)
			minDailyLimit, err := schemeRepo.GetMinDailyLimit(
				companySetting.CompanyUuid,
				utils.SetTimezone(companySetting.GetTimezone()).CurrentWeekday(),
			)
			if err != nil {
				logger.Logger.Error("获取当天最小的每日申请次数限制失败", zap.Error(err))
				return -1
			}
			return minDailyLimit
		}
	}
	return -1
}

// getLastPurchaseQtyByMaterialCode 查询上次品牌采购数量，转换为默认销售单位，返回 map[MaterialCode]qty
func (h *purchaseOrderHelper) getLastPurchaseQtyByMaterialCode(
	db *gorm.DB,
	purchaseOrder *model.PurchaseOrder,
) map[string]float64 {
	result := make(map[string]float64)

	// 收集物品UUID和Code映射
	materialUuids := make([]uint64, 0, len(purchaseOrder.Items))
	uuidToCode := make(map[uint64]string)
	for _, item := range purchaseOrder.Items {
		if item.MaterialCode != "" {
			materialUuids = append(materialUuids, item.MaterialUuid)
			uuidToCode[item.MaterialUuid] = item.MaterialCode
		}
	}
	if len(materialUuids) == 0 {
		return result
	}

	// 查询上次完成品牌采购的基准单位数量
	purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(db)
	baseQtyMap, err := purchaseOrderItemRepo.GetLastCompletedBrandPurchaseBaseQty(materialUuids)
	if err != nil {
		logger.Logger.Warn("查询上次品牌采购数量失败", zap.Error(err))
		return result
	}

	// 查询物品的默认销售单位转换率
	materialRepo := repository.NewMaterialRepo(db)
	materials, err := materialRepo.GetMaterialByUuids(materialUuids, materialRepo.WithNotBaseUnitList())
	if err != nil {
		logger.Logger.Warn("查询物品默认销售单位失败", zap.Error(err))
		return result
	}
	conversionRateMap := make(map[uint64]float64)
	for _, m := range materials {
		if m.DefaultSalesUnitUuid > 0 {
			for _, unit := range m.NotBaseUnitList {
				if unit.Uuid == m.DefaultSalesUnitUuid && unit.ConversionRate != 0 {
					conversionRateMap[m.Uuid] = unit.ConversionRate
				}
			}
		}
	}

	// 转换为默认销售单位并按MaterialCode建立映射
	for materialUuid, baseQty := range baseQtyMap {
		code, ok := uuidToCode[materialUuid]
		if !ok || baseQty == 0 {
			continue
		}
		if rate, hasRate := conversionRateMap[materialUuid]; hasRate {
			result[code] = decimal.NewFromFloat(baseQty).Div(decimal.NewFromFloat(rate)).Round(4).InexactFloat64()
		} else {
			result[code] = decimal.NewFromFloat(baseQty).Round(4).InexactFloat64()
		}
	}
	return result
}
