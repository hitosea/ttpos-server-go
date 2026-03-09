package service

import (
	"fmt"
	"regexp"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	vo "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IErpStockEntrySrv Stock Entry 合并扣减服务
type IErpStockEntrySrv interface {
	// TriggerStockEntryDeduction 触发 Stock Entry 合并扣减
	// 查询 erp_stock_deducted=0 的已结账订单，从预计算的原材料表按 erp_code 合并提交 Stock Entry
	TriggerStockEntryDeduction(ctx cc.Context, companyUuid uint64) error
	// GenerateStocktakeSnapshot 生成盘点快照
	// 查询未通过 Stock Entry 扣减的订单，从预计算的原材料表按 item_code 合并后保存到快照表
	GenerateStocktakeSnapshot(db *gorm.DB, stockReconciliationUuid uint64, warehouseErpCode string) error
}

type erpStockEntrySrv struct {
	dbm *database.DBManager
}

func NewErpStockEntrySrv(dbm *database.DBManager) IErpStockEntrySrv {
	return &erpStockEntrySrv{dbm: dbm}
}

// StockDeductionItem 合并用的中间结构
type StockDeductionItem struct {
	ItemCode string
	ItemName string
	Qty      float64
}

// orderItemDetail 订单维度的商品明细（用于扣减日志回写）
type orderItemDetail struct {
	OrderUuid uint64
	ErpCode   string
	Qty       float64
	OrderType string // "sale_order" or "takeout"
}

const (
	orderTypeSaleOrder = "sale_order"
	orderTypeTakeout   = "takeout"
)

// TriggerStockEntryDeduction 查询未扣减库存的已结账 ERP 订单，从预计算的原材料表合并后提交 Stock Entry
// 策略：先尝试整体提交，失败后解析错误提取有问题的 item_code，排除后再整体提交
func (s *erpStockEntrySrv) TriggerStockEntryDeduction(ctx cc.Context, companyUuid uint64) error {
	// 分布式锁：同一商家同一时间只允许一个 Stock Entry 扣减流程运行
	lockKey := fmt.Sprintf("stock_entry_deduction_%d", companyUuid)
	sysLock := lock.NewSystemLock()
	sysLock.LockUuidString(lockKey)
	defer sysLock.UnlockUuidString(lockKey)

	db := s.dbm.GetDB(companyUuid)
	if db == nil {
		return fmt.Errorf("无法获取商家数据库: %d", companyUuid)
	}

	companySetting := ctx.GetCompanySetting()
	if companySetting.ErpnextSiteCode == "" {
		return nil // 未开启 ERP
	}

	// === 堂食订单：查询 uuid 列表（轻量，无 Preload） ===
	saleOrderRepo := repository.NewSaleOrderRepo(db)
	orders, err := saleOrderRepo.GetSaleOrderList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_stock_deducted = ?", 0).
				Where("erp_sales_invoice_name != ''").
				Where("status = ?", constant.SaleOrderStatusFinish).
				Where("delete_time = 0").
				Select("uuid") // 只需要 uuid
		},
	)
	if err != nil {
		return fmt.Errorf("查询未出库订单失败: %w", err)
	}

	var orderUuids []uint64
	orderTypeMap := make(map[uint64]string) // uuid → "sale_order" or "takeout"
	for _, order := range orders {
		orderUuids = append(orderUuids, order.Uuid)
		orderTypeMap[order.Uuid] = orderTypeSaleOrder
	}

	logger.Logger.Info("Stock Entry合并扣减",
		zap.Uint64("company_uuid", companyUuid),
		zap.Int("order_count", len(orders)),
	)

	// 查询已有的扣减记录，构建已扣减集合
	deductionLogRepo := repository.NewStockDeductionLogRepo(db)
	existingLogs, err := deductionLogRepo.GetByOrderUuids(orderUuids)
	if err != nil {
		return fmt.Errorf("查询扣减日志失败: %w", err)
	}
	deductedSet := make(map[string]bool) // key: "orderUuid:erpCode"
	for _, log := range existingLogs {
		deductedSet[fmt.Sprintf("%d:%s", log.SaleOrderUuid, log.ErpCode)] = true
	}

	// 收集待扣减的商品明细（过滤已扣减的）
	var allDetails []orderItemDetail                       // 订单维度明细（用于日志回写）
	mergeMap := make(map[string]*StockDeductionItem)       // 按 item_code 合并
	orderRequiredItems := make(map[uint64]map[string]bool) // 每个订单需要哪些 item_code

	// === 堂食订单：从 sale_order_material 表读取预计算的原材料 ===
	if len(orderUuids) > 0 {
		saleOrderMaterials, err := repository.NewSaleOrderMaterialRepo(db).GetBySaleOrderUuids(orderUuids)
		if err != nil {
			return fmt.Errorf("查询堂食订单原材料失败: %w", err)
		}

		for _, m := range saleOrderMaterials {
			erpCode := ""
			if m.Material != nil {
				erpCode = m.Material.Code
			}
			if erpCode == "" {
				continue
			}

			if orderRequiredItems[m.SaleOrderUuid] == nil {
				orderRequiredItems[m.SaleOrderUuid] = make(map[string]bool)
			}
			orderRequiredItems[m.SaleOrderUuid][erpCode] = true

			key := fmt.Sprintf("%d:%s", m.SaleOrderUuid, erpCode)
			if deductedSet[key] {
				continue // 已扣减，跳过
			}

			allDetails = append(allDetails, orderItemDetail{
				OrderUuid: m.SaleOrderUuid,
				ErpCode:   erpCode,
				Qty:       m.Num,
				OrderType: orderTypeSaleOrder,
			})

			if existing, ok := mergeMap[erpCode]; ok {
				existing.Qty += m.Num
			} else {
				mergeMap[erpCode] = &StockDeductionItem{
					ItemCode: erpCode,
					Qty:      m.Num,
				}
			}
		}

		// 没有原材料的订单也需要标记
		for _, orderUuid := range orderUuids {
			if orderRequiredItems[orderUuid] == nil {
				orderRequiredItems[orderUuid] = map[string]bool{"__none__": true}
			}
		}
	}

	// === 外卖订单部分 ===
	takeoutRepo := persistence.NewTakeoutOrderRepo(db)
	takeoutOrders, _, err := takeoutRepo.GetList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_stock_deducted = ?", 0).
				Where("erp_pos_invoice_resp != ''").
				Where("order_state = ?", vo.TakeoutOrderStateCompleted).
				Where("delete_time = 0")
		},
		takeoutRepo.WithTakeoutOrderMaterials(),
	)
	if err != nil {
		logger.Logger.Error("查询外卖未出库订单失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Error(err),
		)
		// 不阻塞堂食订单扣减，继续执行
	}

	var takeoutOrderUuids []uint64
	for _, to := range takeoutOrders {
		takeoutOrderUuids = append(takeoutOrderUuids, to.Uuid)
		orderTypeMap[to.Uuid] = orderTypeTakeout
	}

	if len(takeoutOrders) > 0 {
		logger.Logger.Info("Stock Entry外卖订单",
			zap.Uint64("company_uuid", companyUuid),
			zap.Int("takeout_order_count", len(takeoutOrders)),
		)

		// 查询外卖订单的已有扣减记录
		takeoutExistingLogs, err := deductionLogRepo.GetByOrderUuids(takeoutOrderUuids)
		if err != nil {
			logger.Logger.Error("查询外卖扣减日志失败",
				zap.Uint64("company_uuid", companyUuid),
				zap.Error(err),
			)
		}
		for _, log := range takeoutExistingLogs {
			deductedSet[fmt.Sprintf("%d:%s", log.SaleOrderUuid, log.ErpCode)] = true
		}

		for _, to := range takeoutOrders {
			requiredItems := make(map[string]bool)

			for _, m := range to.TakeoutOrderMaterials {
				if m.ErpCode == "" {
					continue
				}
				requiredItems[m.ErpCode] = true

				key := fmt.Sprintf("%d:%s", to.Uuid, m.ErpCode)
				if deductedSet[key] {
					continue
				}

				allDetails = append(allDetails, orderItemDetail{
					OrderUuid: to.Uuid,
					ErpCode:   m.ErpCode,
					Qty:       m.Num,
					OrderType: orderTypeTakeout,
				})

				if existing, ok := mergeMap[m.ErpCode]; ok {
					existing.Qty += m.Num
				} else {
					mergeMap[m.ErpCode] = &StockDeductionItem{
						ItemCode: m.ErpCode,
						Qty:      m.Num,
					}
				}
			}

			if len(requiredItems) == 0 {
				requiredItems["__none__"] = true
			}
			orderRequiredItems[to.Uuid] = requiredItems
		}
	}

	// 合并所有订单 UUID（堂食 + 外卖）
	allOrderUuids := append(orderUuids, takeoutOrderUuids...)
	if len(allOrderUuids) == 0 {
		return nil
	}

	if len(mergeMap) == 0 {
		// 所有商品都已扣减或无需扣减，标记订单
		s.markFullyDeductedOrders(db, allOrderUuids, orderRequiredItems, deductedSet, orderTypeMap)
		return nil
	}

	// 构建合并的 Stock Entry 请求
	items := make([]*stock.StockEntryItem, 0, len(mergeMap))
	for _, item := range mergeMap {
		items = append(items, &stock.StockEntryItem{
			ItemCode: item.ItemCode,
			ItemName: item.ItemName,
			Qty:      item.Qty,
		})
	}

	erpSrv := erp.NewIErpSrv(s.dbm)

	// 第一步：尝试整体提交
	submitReq := &stock.SubmitStockEntryReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
		Items:       items,
		Remarks:     fmt.Sprintf("TTPOS合并扣减, sale_orders=%d, takeout_orders=%d, items=%d", len(orders), len(takeoutOrders), len(mergeMap)),
	}

	resp, err := erpSrv.SubmitStockEntry(ctx, companySetting, submitReq)
	if err == nil {
		// 整体提交成功
		logger.Logger.Info("Stock Entry合并扣减成功",
			zap.Uint64("company_uuid", companyUuid),
			zap.String("stock_entry_name", resp.StockEntryName),
			zap.Int("order_count", len(orders)+len(takeoutOrders)),
		)

		// 写入所有扣减日志
		if err := s.writeDeductionLogs(db, companyUuid, allDetails, resp.StockEntryName); err != nil {
			return fmt.Errorf("Stock Entry扣减成功但写入日志失败: %w", err)
		}

		// 所有待扣减项都成功，更新 deductedSet 后标记完成的订单
		for _, d := range allDetails {
			deductedSet[fmt.Sprintf("%d:%s", d.OrderUuid, d.ErpCode)] = true
		}
		s.markFullyDeductedOrders(db, allOrderUuids, orderRequiredItems, deductedSet, orderTypeMap)
		return nil
	}

	// 第二步：整体提交失败，循环排除有问题的 item 后重试（ERPNext 每次只返回第一个错误的 item）
	excludedSet := make(map[string]bool)
	lastErr := err
	const maxRetry = 5

	for retry := 0; retry < maxRetry; retry++ {
		newCodes := parseStockEntryErrorItemCodes(lastErr.Error())
		if len(newCodes) == 0 {
			return fmt.Errorf("Stock Entry合并扣减失败且无法解析失败item: %w", lastErr)
		}

		for _, code := range newCodes {
			excludedSet[code] = true
		}

		logger.Logger.Warn("Stock Entry排除有问题item后重试",
			zap.Uint64("company_uuid", companyUuid),
			zap.Int("retry", retry+1),
			zap.Strings("new_excluded", newCodes),
			zap.Int("total_excluded", len(excludedSet)),
			zap.Error(lastErr),
		)

		// 构建排除后的 items
		retryItems := make([]*stock.StockEntryItem, 0)
		for _, item := range mergeMap {
			if excludedSet[item.ItemCode] {
				continue
			}
			retryItems = append(retryItems, &stock.StockEntryItem{
				ItemCode: item.ItemCode,
				ItemName: item.ItemName,
				Qty:      item.Qty,
			})
		}

		if len(retryItems) == 0 {
			logger.Logger.Warn("Stock Entry所有item均被排除，跳过本次扣减",
				zap.Uint64("company_uuid", companyUuid),
				zap.Int("excluded_count", len(excludedSet)),
			)
			return fmt.Errorf("Stock Entry所有item均被排除: excluded=%d", len(excludedSet))
		}

		excludedCodes := make([]string, 0, len(excludedSet))
		for code := range excludedSet {
			excludedCodes = append(excludedCodes, code)
		}

		retryReq := &stock.SubmitStockEntryReq{
			CompanyAbbr: companySetting.ErpnextCompanyAbbr,
			Branch:      companySetting.ErpnextBranchName,
			Items:       retryItems,
			Remarks:     fmt.Sprintf("TTPOS合并扣减(排除%d项), sale_orders=%d, takeout_orders=%d, items=%d", len(excludedSet), len(orders), len(takeoutOrders), len(retryItems)),
		}

		retryResp, retryErr := erpSrv.SubmitStockEntry(ctx, companySetting, retryReq)
		if retryErr != nil {
			lastErr = retryErr
			continue // 继续排除下一个有问题的 item
		}

		// 重试成功
		logger.Logger.Info("Stock Entry排除后重试成功",
			zap.Uint64("company_uuid", companyUuid),
			zap.String("stock_entry_name", retryResp.StockEntryName),
			zap.Int("excluded_count", len(excludedSet)),
		)

		var successDetails []orderItemDetail
		for _, d := range allDetails {
			if !excludedSet[d.ErpCode] {
				successDetails = append(successDetails, d)
			}
		}
		if err := s.writeDeductionLogs(db, companyUuid, successDetails, retryResp.StockEntryName); err != nil {
			return fmt.Errorf("Stock Entry排除后扣减成功但写入日志失败: %w", err)
		}

		for _, d := range successDetails {
			deductedSet[fmt.Sprintf("%d:%s", d.OrderUuid, d.ErpCode)] = true
		}

		s.markFullyDeductedOrders(db, allOrderUuids, orderRequiredItems, deductedSet, orderTypeMap)

		if len(excludedSet) > 0 {
			return fmt.Errorf("Stock Entry部分item被排除: %v (成功=%d, 排除=%d)",
				excludedCodes, len(retryItems), len(excludedSet))
		}

		return nil
	}

	return fmt.Errorf("Stock Entry重试%d次后仍失败: %w (排除=%v)", maxRetry, lastErr, excludedSet)
}

// stockEntryErrorItemCodeRegex 匹配 ERPNext 库存不足错误中的 item_code
// 匹配模式：Item Code: <strong>WPR3685375438618625</strong>
var stockEntryErrorItemCodeRegex = regexp.MustCompile(`Item Code: <strong>([^<]+)</strong>`)

// stockEntryNotStockItemRegex 匹配 ERPNext "is not a stock Item" 错误中的 item_code
// 匹配模式：SP3700735403493377_01 is not a stock Item
var stockEntryNotStockItemRegex = regexp.MustCompile(`(\S+) is not a stock Item`)

// parseStockEntryErrorItemCodes 从 ERPNext 错误信息中提取有问题的 item_code 列表
func parseStockEntryErrorItemCodes(errMsg string) []string {
	seen := make(map[string]bool)
	var codes []string

	// 匹配库存不足: Item Code: <strong>XXX</strong>
	for _, m := range stockEntryErrorItemCodeRegex.FindAllStringSubmatch(errMsg, -1) {
		code := m[1]
		if !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}

	// 匹配非库存物料: XXX is not a stock Item
	for _, m := range stockEntryNotStockItemRegex.FindAllStringSubmatch(errMsg, -1) {
		code := m[1]
		if !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}

	return codes
}

// writeDeductionLogs 写入扣减日志
func (s *erpStockEntrySrv) writeDeductionLogs(db *gorm.DB, companyUuid uint64, details []orderItemDetail, stockEntryName string) error {
	if len(details) == 0 {
		return nil
	}

	logs := make([]*model.StockDeductionLog, 0, len(details))
	for _, d := range details {
		logs = append(logs, &model.StockDeductionLog{
			SaleOrderUuid:  d.OrderUuid,
			ErpCode:        d.ErpCode,
			Qty:            d.Qty,
			StockEntryName: stockEntryName,
		})
	}

	deductionLogRepo := repository.NewStockDeductionLogRepo(db)
	if err := deductionLogRepo.BatchCreate(logs); err != nil {
		logger.Logger.Error("写入扣减日志失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Int("count", len(logs)),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// markFullyDeductedOrders 检查并标记所有 item 都已扣减完成的订单
func (s *erpStockEntrySrv) markFullyDeductedOrders(db *gorm.DB, orderUuids []uint64, orderRequiredItems map[uint64]map[string]bool, deductedSet map[string]bool, orderTypeMap map[uint64]string) {
	var saleOrderDeducted []uint64
	var takeoutOrderDeducted []uint64

	for _, orderUuid := range orderUuids {
		requiredItems := orderRequiredItems[orderUuid]
		allDone := true
		for erpCode := range requiredItems {
			if erpCode == "__none__" {
				continue // 无需扣减的占位符
			}
			if !deductedSet[fmt.Sprintf("%d:%s", orderUuid, erpCode)] {
				allDone = false
				break
			}
		}
		if !allDone {
			continue
		}

		switch orderTypeMap[orderUuid] {
		case orderTypeTakeout:
			takeoutOrderDeducted = append(takeoutOrderDeducted, orderUuid)
		default:
			saleOrderDeducted = append(saleOrderDeducted, orderUuid)
		}
	}

	if len(saleOrderDeducted) > 0 {
		s.markOrdersDeducted(db, saleOrderDeducted)
	}
	if len(takeoutOrderDeducted) > 0 {
		s.markTakeoutOrdersDeducted(db, takeoutOrderDeducted)
	}
}

// markOrdersDeducted 批量更新堂食订单 erp_stock_deducted=1
func (s *erpStockEntrySrv) markOrdersDeducted(db *gorm.DB, orderUuids []uint64) {
	if len(orderUuids) == 0 {
		return
	}
	db.Model(&model.SaleOrder{}).
		Where("uuid IN ?", orderUuids).
		Update("erp_stock_deducted", 1)
}

// markTakeoutOrdersDeducted 批量更新外卖订单 erp_stock_deducted=1
func (s *erpStockEntrySrv) markTakeoutOrdersDeducted(db *gorm.DB, orderUuids []uint64) {
	if len(orderUuids) == 0 {
		return
	}
	db.Model(&takeoutModel.TakeoutOrder{}).
		Where("uuid IN ?", orderUuids).
		Update("erp_stock_deducted", 1)
}

// GenerateStocktakeSnapshot 生成盘点快照
// 查询 erp_stock_deducted=0 且 SI 已创建的已结账订单
// 从预计算的原材料表读取，按 item_code + warehouseErpCode 合并，保存到 stocktake_snapshot 表
func (s *erpStockEntrySrv) GenerateStocktakeSnapshot(db *gorm.DB, stockReconciliationUuid uint64, warehouseErpCode string) error {
	// === 堂食订单：查询 uuid 列表（轻量） ===
	saleOrderRepo := repository.NewSaleOrderRepo(db)
	orders, err := saleOrderRepo.GetSaleOrderList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_stock_deducted = ?", 0).
				Where("erp_sales_invoice_name != ''").
				Where("status = ?", constant.SaleOrderStatusFinish).
				Where("delete_time = 0").
				Select("uuid")
		},
	)
	if err != nil {
		return fmt.Errorf("查询未出库订单失败: %w", err)
	}

	var orderUuids []uint64
	for _, order := range orders {
		orderUuids = append(orderUuids, order.Uuid)
	}

	// 按 item_code 合并
	type snapshotItem struct {
		Qty        float64
		OrderCount int
	}
	mergeMap := make(map[string]*snapshotItem) // key: itemCode
	orderSet := make(map[uint64]bool)

	// === 堂食订单：从 sale_order_material 表读取预计算的原材料 ===
	if len(orderUuids) > 0 {
		saleOrderMaterials, err := repository.NewSaleOrderMaterialRepo(db).GetBySaleOrderUuids(orderUuids)
		if err != nil {
			return fmt.Errorf("查询堂食订单原材料失败: %w", err)
		}

		for _, m := range saleOrderMaterials {
			erpCode := ""
			if m.Material != nil {
				erpCode = m.Material.Code
			}
			if erpCode == "" {
				continue
			}
			orderSet[m.SaleOrderUuid] = true
			if existing, ok := mergeMap[erpCode]; ok {
				existing.Qty += m.Num
			} else {
				mergeMap[erpCode] = &snapshotItem{
					Qty: m.Num,
				}
			}
		}
	}

	// === 外卖订单部分 ===
	takeoutRepo := persistence.NewTakeoutOrderRepo(db)
	takeoutOrders, _, err := takeoutRepo.GetList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_stock_deducted = ?", 0).
				Where("erp_pos_invoice_resp != ''").
				Where("order_state = ?", vo.TakeoutOrderStateCompleted).
				Where("delete_time = 0")
		},
		takeoutRepo.WithTakeoutOrderMaterials(),
	)
	if err != nil {
		logger.Logger.Error("查询外卖未出库订单失败（快照）", zap.Error(err))
	}

	for _, to := range takeoutOrders {
		orderSet[to.Uuid] = true
		for _, m := range to.TakeoutOrderMaterials {
			if m.ErpCode == "" {
				continue
			}
			if existing, ok := mergeMap[m.ErpCode]; ok {
				existing.Qty += m.Num
			} else {
				mergeMap[m.ErpCode] = &snapshotItem{
					Qty: m.Num,
				}
			}
		}
	}

	if len(mergeMap) == 0 {
		return nil
	}

	// 扣除已部分扣减的数量（部分item已通过Stock Entry扣减但订单整体尚未标记为erp_stock_deducted=1）
	allSnapshotOrderUuids := make([]uint64, 0, len(orderSet))
	for uuid := range orderSet {
		allSnapshotOrderUuids = append(allSnapshotOrderUuids, uuid)
	}
	if len(allSnapshotOrderUuids) > 0 {
		existingLogs, logErr := repository.NewStockDeductionLogRepo(db).GetByOrderUuids(allSnapshotOrderUuids)
		if logErr != nil {
			logger.Logger.Error("查询已扣减日志失败（快照）", zap.Error(logErr))
		}
		for _, log := range existingLogs {
			if item, ok := mergeMap[log.ErpCode]; ok {
				item.Qty -= log.Qty
				if item.Qty <= 0 {
					delete(mergeMap, log.ErpCode)
				}
			}
		}
	}

	if len(mergeMap) == 0 {
		return nil
	}

	// 保存快照
	snapshots := make([]model.StocktakeSnapshot, 0, len(mergeMap))
	for itemCode, item := range mergeMap {
		snapshots = append(snapshots, model.StocktakeSnapshot{
			StockReconciliationUuid: stockReconciliationUuid,
			ItemCode:                itemCode,
			WarehouseErpCode:        warehouseErpCode,
			PendingQty:              item.Qty,
			OrderCount:              len(orderSet),
		})
	}

	if err := db.CreateInBatches(&snapshots, 100).Error; err != nil {
		return fmt.Errorf("保存盘点快照失败: %w", err)
	}

	logger.Logger.Info("盘点快照生成成功",
		zap.Uint64("stock_reconciliation_uuid", stockReconciliationUuid),
		zap.Int("snapshot_count", len(snapshots)),
		zap.Int("order_count", len(orderSet)),
	)

	return nil
}
