package service

import (
	"fmt"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IErpStockEntrySrv Stock Entry 合并扣减服务
type IErpStockEntrySrv interface {
	// TriggerStockEntryDeduction 触发 Stock Entry 合并扣减
	// 查询 erp_stock_deducted=0 的已结账订单，按 (erp_code) 合并后提交 Stock Entry
	TriggerStockEntryDeduction(ctx cc.Context, companyUuid uint64) error
	// GenerateStocktakeSnapshot 生成盘点快照
	// 查询未通过 Stock Entry 扣减的订单，按 item_code 合并后保存到快照表
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

// TriggerStockEntryDeduction 查询未扣减库存的已结账 ERP 订单，合并后提交 Stock Entry
func (s *erpStockEntrySrv) TriggerStockEntryDeduction(ctx cc.Context, companyUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)
	if db == nil {
		return fmt.Errorf("无法获取商家数据库: %d", companyUuid)
	}

	companySetting := ctx.GetCompanySetting()
	if companySetting.ErpnextSiteCode == "" {
		return nil // 未开启 ERP
	}

	// 查询 erp_stock_deducted=0 且 erp_sales_invoice_name != '' 的已结账订单
	saleOrderRepo := repository.NewSaleOrderRepo(db)
	orders, err := saleOrderRepo.GetSaleOrderList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_stock_deducted = ?", 0).
				Where("erp_sales_invoice_name != ''").
				Where("status = ?", constant.SaleOrderStatusFinish).
				Where("delete_time = 0")
		},
	)
	if err != nil {
		return fmt.Errorf("查询未出库订单失败: %w", err)
	}
	if len(orders) == 0 {
		return nil
	}

	logger.Logger.Info("Stock Entry合并扣减",
		zap.Uint64("company_uuid", companyUuid),
		zap.Int("order_count", len(orders)),
	)

	// 收集所有订单的物品明细，按 item_code 合并
	// 仓库由 BMP 侧根据 branch 自动解析默认仓库
	mergeMap := make(map[string]*StockDeductionItem) // key: itemCode
	var orderUuids []uint64

	for _, order := range orders {
		orderUuids = append(orderUuids, order.Uuid)

		products := s.getOrderProducts(db, order.Uuid)
		for _, p := range products {
			if p.ErpCode == "" {
				continue
			}
			if existing, ok := mergeMap[p.ErpCode]; ok {
				existing.Qty += p.Num
			} else {
				mergeMap[p.ErpCode] = &StockDeductionItem{
					ItemCode: p.ErpCode,
					Qty:      p.Num,
				}
			}
		}
	}

	if len(mergeMap) == 0 {
		// 没有可扣减的物品，直接标记所有订单为已扣减
		s.markOrdersDeducted(db, orderUuids)
		return nil
	}

	// 构建 Stock Entry 请求
	items := make([]*stock.StockEntryItem, 0, len(mergeMap))
	for _, item := range mergeMap {
		items = append(items, &stock.StockEntryItem{
			ItemCode: item.ItemCode,
			ItemName: item.ItemName,
			Qty:      item.Qty,
		})
	}

	erpSrv := erp.NewIErpSrv(s.dbm)
	submitReq := &stock.SubmitStockEntryReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
		Items:       items,
		Remarks:     fmt.Sprintf("TTPOS合并扣减, orders=%d", len(orders)),
	}

	resp, err := erpSrv.SubmitStockEntry(ctx, companySetting, submitReq)
	if err != nil {
		logger.Logger.Error("Stock Entry合并扣减失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Error(err),
		)
		return fmt.Errorf("Stock Entry提交失败: %w", err)
	}

	logger.Logger.Info("Stock Entry合并扣减成功",
		zap.Uint64("company_uuid", companyUuid),
		zap.String("stock_entry_name", resp.StockEntryName),
		zap.Int("order_count", len(orders)),
	)

	// 更新订单 erp_stock_deducted=1
	s.markOrdersDeducted(db, orderUuids)

	return nil
}

// getOrderProducts 获取订单的商品和物品列表（用于 Stock Entry 合并）
func (s *erpStockEntrySrv) getOrderProducts(db *gorm.DB, saleOrderUuid uint64) []model.SaleOrderProduct {
	var products []model.SaleOrderProduct
	db.Where("sale_order_uuid = ? AND delete_time = 0 AND erp_code != ''", saleOrderUuid).
		Find(&products)
	return products
}

// markOrdersDeducted 批量更新订单 erp_stock_deducted=1
func (s *erpStockEntrySrv) markOrdersDeducted(db *gorm.DB, orderUuids []uint64) {
	if len(orderUuids) == 0 {
		return
	}
	db.Model(&model.SaleOrder{}).
		Where("uuid IN ?", orderUuids).
		Update("erp_stock_deducted", 1)
}

// GenerateStocktakeSnapshot 生成盘点快照
// 查询 erp_stock_deducted=0 且 erp_sales_invoice_name != ” 的已结账订单
// 按 item_code + warehouseErpCode 合并，保存到 stocktake_snapshot 表
func (s *erpStockEntrySrv) GenerateStocktakeSnapshot(db *gorm.DB, stockReconciliationUuid uint64, warehouseErpCode string) error {
	// 查询未扣减库存的已结账订单
	saleOrderRepo := repository.NewSaleOrderRepo(db)
	orders, err := saleOrderRepo.GetSaleOrderList(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("erp_stock_deducted = ?", 0).
				Where("erp_sales_invoice_name != ''").
				Where("status = ?", constant.SaleOrderStatusFinish).
				Where("delete_time = 0")
		},
	)
	if err != nil {
		return fmt.Errorf("查询未出库订单失败: %w", err)
	}
	if len(orders) == 0 {
		return nil
	}

	// 按 item_code 合并
	type snapshotItem struct {
		Qty        float64
		OrderCount int
	}
	mergeMap := make(map[string]*snapshotItem) // key: itemCode
	orderSet := make(map[uint64]bool)

	for _, order := range orders {
		orderSet[order.Uuid] = true
		products := s.getOrderProducts(db, order.Uuid)
		for _, p := range products {
			if p.ErpCode == "" {
				continue
			}
			if existing, ok := mergeMap[p.ErpCode]; ok {
				existing.Qty += p.Num
			} else {
				mergeMap[p.ErpCode] = &snapshotItem{
					Qty: p.Num,
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
