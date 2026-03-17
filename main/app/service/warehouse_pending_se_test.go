package service

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	if config.Database.TablePrefix == "" {
		config.Database.TablePrefix = "ttpos_"
	}
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}
}

// setupPendingSETestDB 创建 getPendingStockEntryConsumption 所需的全部 SQLite 内存表
func setupPendingSETestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
	})
	require.NoError(t, err)

	// sale_order
	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_sale_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			erp_stock_deducted INTEGER DEFAULT 0,
			erp_sales_invoice_name TEXT DEFAULT ''
		)
	`).Error)

	// material（供 Preload("Material") 关联）
	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_material (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			code TEXT DEFAULT '',
			name TEXT DEFAULT ''
		)
	`).Error)

	// sale_order_material
	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_sale_order_material (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			sale_order_uuid INTEGER DEFAULT 0,
			sale_bill_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			num REAL DEFAULT 0,
			staff_shift_log_uuid INTEGER DEFAULT 0,
			is_summarized INTEGER DEFAULT 0
		)
	`).Error)

	// takeout_order
	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_takeout_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			order_state INTEGER DEFAULT 0,
			erp_sync_status INTEGER DEFAULT 0,
			erp_stock_deducted INTEGER DEFAULT 0,
			erp_pos_invoice_resp TEXT DEFAULT '',
			staff_shift_log_uuid INTEGER DEFAULT 0
		)
	`).Error)

	// takeout_order_material
	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_takeout_order_material (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			takeout_order_uuid INTEGER DEFAULT 0,
			takeout_order_item_uuid INTEGER DEFAULT 0,
			takeout_order_item_modifier_uuid INTEGER DEFAULT 0,
			product_bom_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_name TEXT DEFAULT '',
			erp_code TEXT DEFAULT '',
			base_unit_uom TEXT DEFAULT '',
			warehouse_uuid INTEGER DEFAULT 0,
			num REAL DEFAULT 0,
			is_summarized INTEGER DEFAULT 0
		)
	`).Error)

	// stock_deduction_log
	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_stock_deduction_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			sale_order_uuid INTEGER DEFAULT 0,
			erp_code TEXT DEFAULT '',
			qty REAL DEFAULT 0,
			stock_entry_name TEXT DEFAULT ''
		)
	`).Error)

	return db
}

func newTestCtx() context.Context {
	return context.NewContext(context.WithCompanyUuid(99999))
}

// TestGetPendingStockEntryConsumption_Basic 验证基本聚合逻辑：
// 堂食和外卖订单的物料按 warehouseUuid+erpCode 汇总
func TestGetPendingStockEntryConsumption_Basic(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 插入一笔待扣减堂食订单
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 0, "ACC-SINV-001", 0,
	).Error)

	// 插入物料关联记录（Material.Code = "ITEM-A"）
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "ITEM-A", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 100, 3.5, 0,
	).Error)

	// 插入一笔待扣减外卖订单
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order (uuid, order_state, erp_sync_status, erp_stock_deducted, erp_pos_invoice_resp, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		2001, value_object.TakeoutOrderStateCompleted, constant.ErpSyncStatusSuccess, 0, `{"name":"SI-001"}`, 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order_material (uuid, takeout_order_uuid, erp_code, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		7001, 2001, "ITEM-A", 100, 2.0, 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	// warehouse 100, ITEM-A: 3.5 (堂食) + 2.0 (外卖) = 5.5
	assert.InDelta(t, 5.5, result[100]["ITEM-A"], 0.001)
}

// TestGetPendingStockEntryConsumption_MultiWarehouse 验证多仓库隔离
func TestGetPendingStockEntryConsumption_MultiWarehouse(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 堂食订单
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 0, "ACC-SINV-001", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "ITEM-B", 0,
	).Error)
	// 仓库 100: 4.0
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 100, 4.0, 0,
	).Error)
	// 仓库 200: 6.0
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6002, 1001, 5001, 200, 6.0, 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	assert.InDelta(t, 4.0, result[100]["ITEM-B"], 0.001)
	assert.InDelta(t, 6.0, result[200]["ITEM-B"], 0.001)
}

// TestGetPendingStockEntryConsumption_DeductionLogOffset 验证 stock_deduction_log 已扣减量按比例扣除
func TestGetPendingStockEntryConsumption_DeductionLogOffset(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 堂食订单
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 0, "ACC-SINV-001", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "ITEM-C", 0,
	).Error)
	// 仓库 100: 4.0, 仓库 200: 6.0, 总计 10.0
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 100, 4.0, 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6002, 1001, 5001, 200, 6.0, 0,
	).Error)

	// 已通过 SE 扣减了 5.0（按比例分配：仓库100扣2.0，仓库200扣3.0）
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_deduction_log (uuid, sale_order_uuid, erp_code, qty, stock_entry_name, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		8001, 1001, "ITEM-C", 5.0, "SE-001", 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	// 仓库100: 4.0 - 5.0*(4.0/10.0) = 4.0 - 2.0 = 2.0
	// 仓库200: 6.0 - 5.0*(6.0/10.0) = 6.0 - 3.0 = 3.0
	assert.InDelta(t, 2.0, result[100]["ITEM-C"], 0.001)
	assert.InDelta(t, 3.0, result[200]["ITEM-C"], 0.001)
}

// TestGetPendingStockEntryConsumption_FullyDeducted 验证已完全扣减时条目被清除
func TestGetPendingStockEntryConsumption_FullyDeducted(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 0, "ACC-SINV-001", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "ITEM-D", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 100, 10.0, 0,
	).Error)

	// 已扣减量 >= 待扣减量
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_deduction_log (uuid, sale_order_uuid, erp_code, qty, stock_entry_name, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		8001, 1001, "ITEM-D", 10.0, "SE-001", 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	// 完全扣减后，该条目应被删除，仓库 map 可能为空
	if wh, ok := result[100]; ok {
		assert.Zero(t, wh["ITEM-D"], "Fully deducted item should be 0 or removed")
	}
}

// TestGetPendingStockEntryConsumption_EmptyDB 验证空数据库返回空 map
func TestGetPendingStockEntryConsumption_EmptyDB(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)
	assert.Empty(t, result)
}

// TestGetPendingStockEntryConsumption_SkipsNonPendingOrders 验证已扣减/未结账/已删除的订单被排除
func TestGetPendingStockEntryConsumption_SkipsNonPendingOrders(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "ITEM-E", 0,
	).Error)

	// 已扣减的堂食订单（erp_stock_deducted=1）→ 不应出现
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 1, "ACC-SINV-001", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 100, 10.0, 0,
	).Error)

	// 未结账的堂食订单（status=0）→ 不应出现
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1002, 0, 0, "", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6002, 1002, 5001, 100, 10.0, 0,
	).Error)

	// 已删除的外卖订单 → 不应出现
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order (uuid, order_state, erp_sync_status, erp_stock_deducted, erp_pos_invoice_resp, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		2001, value_object.TakeoutOrderStateCompleted, constant.ErpSyncStatusSuccess, 0, `{"name":"SI-001"}`, 100,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order_material (uuid, takeout_order_uuid, erp_code, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		7001, 2001, "ITEM-E", 100, 10.0, 0,
	).Error)

	// 未同步的外卖订单（erp_sync_status=4）→ 不应出现
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order (uuid, order_state, erp_sync_status, erp_stock_deducted, erp_pos_invoice_resp, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		2002, value_object.TakeoutOrderStateCompleted, constant.ErpSyncStatusFailed, 0, `{"name":"SI-002"}`, 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order_material (uuid, takeout_order_uuid, erp_code, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		7002, 2002, "ITEM-E", 100, 10.0, 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)
	assert.Empty(t, result, "No pending orders should be found when all orders are excluded by scope conditions")
}

// ─── collectSaleOrderConsumption edge cases ───

// TestGetPendingStockEntryConsumption_SaleOrderSkipInvalidMaterial 验证堂食中无效物料被跳过
// 覆盖 collectSaleOrderConsumption 的 continue 分支 (materialCode=="" || WarehouseUuid==0)
func TestGetPendingStockEntryConsumption_SaleOrderSkipInvalidMaterial(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 待扣减堂食订单
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 0, "ACC-SINV-001", 0,
	).Error)

	// 物料 5001 有 Code，但 WarehouseUuid=0 → 应跳过
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "ITEM-VALID", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 0, 10.0, 0,
	).Error)

	// 物料 5002 无关联 Material（material_uuid 不存在）→ materialCode="" → 应跳过
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6002, 1001, 99999, 100, 5.0, 0,
	).Error)

	// 物料 5003 有效 → 应包含
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5003, "ITEM-OK", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6003, 1001, 5003, 200, 3.0, 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	// 只有 ITEM-OK 在仓库 200 应该出现
	assert.Empty(t, result[0], "WarehouseUuid=0 materials should be skipped")
	assert.InDelta(t, 3.0, result[200]["ITEM-OK"], 0.001)
	_, hasInvalid := result[100][""]
	assert.False(t, hasInvalid, "Materials with nil Material record should be skipped")
}

// ─── collectTakeoutOrderConsumption edge cases ───

// TestGetPendingStockEntryConsumption_TakeoutOnlyWarehouse 验证外卖订单创建新仓库的 map 条目
// 目的：覆盖 collectTakeoutOrderConsumption 中 result[m.WarehouseUuid] == nil 的 map 初始化分支
func TestGetPendingStockEntryConsumption_TakeoutOnlyWarehouse(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 外卖订单的物料归属仓库 300 (堂食中没有出现过此仓库)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order (uuid, order_state, erp_sync_status, erp_stock_deducted, erp_pos_invoice_resp, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		2001, value_object.TakeoutOrderStateCompleted, constant.ErpSyncStatusSuccess, 0, `{"name":"SI-001"}`, 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order_material (uuid, takeout_order_uuid, erp_code, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		7001, 2001, "ITEM-X", 300, 5.5, 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	// 仓库 300 应该被新建 map 并包含 ITEM-X
	assert.InDelta(t, 5.5, result[300]["ITEM-X"], 0.001, "Takeout-only warehouse should have map entry initialized")
}

// TestGetPendingStockEntryConsumption_SkipEmptyErpCode 验证 ErpCode 为空或 WarehouseUuid=0 的物料被跳过
func TestGetPendingStockEntryConsumption_SkipEmptyErpCode(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 外卖订单
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order (uuid, order_state, erp_sync_status, erp_stock_deducted, erp_pos_invoice_resp, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		2001, value_object.TakeoutOrderStateCompleted, constant.ErpSyncStatusSuccess, 0, `{"name":"SI-001"}`, 0,
	).Error)
	// 物料 ErpCode 为空 → 应跳过
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order_material (uuid, takeout_order_uuid, erp_code, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		7001, 2001, "", 100, 10.0, 0,
	).Error)
	// 物料 WarehouseUuid=0 → 应跳过
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_takeout_order_material (uuid, takeout_order_uuid, erp_code, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		7002, 2001, "ITEM-Y", 0, 10.0, 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)
	assert.Empty(t, result, "Materials with empty ErpCode or WarehouseUuid=0 should be skipped")
}

// ─── buildSIModeOpts tests ───

func TestBuildSIModeOpts_NotSIMode(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)
	srv := &warehouseSrv{}

	// CompanySetting with ErpInvoiceMode=0 → not SI mode → early return nil
	ctx := context.NewContext(
		context.WithCompanyUuid(99999),
		context.WithCompanySetting(model.CompanySetting{ErpInvoiceMode: 0}),
	)

	consumption, opts, err := srv.buildSIModeOpts(ctx, db)
	require.NoError(t, err)
	assert.Nil(t, consumption, "Should return nil when not SI mode")
	assert.Nil(t, opts, "Should return nil opts when not SI mode")
}

func TestBuildSIModeOpts_SIMode(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)
	srv := &warehouseSrv{}

	// CompanySetting with ErpInvoiceMode=1 → SI mode
	ctx := context.NewContext(
		context.WithCompanyUuid(99999),
		context.WithCompanySetting(model.CompanySetting{ErpInvoiceMode: 1}),
	)

	// With empty DB, should still return empty consumption + 2 extra opts
	consumption, opts, err := srv.buildSIModeOpts(ctx, db)
	require.NoError(t, err)
	assert.NotNil(t, consumption, "Should return non-nil consumption map")
	assert.Len(t, opts, 2, "Should return 2 extra DB options for SI mode")

	// Invoke the returned opts to cover closure bodies (L1028, L1031)
	if len(opts) == 2 {
		mockDB := db.Table("ttpos_sale_order")
		result1 := opts[0](mockDB)
		assert.NotNil(t, result1, "First opt should return a valid *gorm.DB")
		result2 := opts[1](mockDB)
		assert.NotNil(t, result2, "Second opt should return a valid *gorm.DB")
	}
}

// TestOffsetDeductedConsumption_SkipsCodeWithoutDeduction 验证 offsetDeductedConsumption
// 在遇到没有扣减记录的物品编码时跳过（deducted <= 0 → continue）
func TestOffsetDeductedConsumption_SkipsCodeWithoutDeduction(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)

	// 堂食订单含两种物料
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
		1001, constant.SaleOrderStatusFinish, 0, "ACC-SINV-001", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5001, "HAS-DEDUCTION", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, delete_time) VALUES (?, ?, ?)", 5002, "NO-DEDUCTION", 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6001, 1001, 5001, 100, 10.0, 0,
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_sale_order_material (uuid, sale_order_uuid, material_uuid, warehouse_uuid, num, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		6002, 1001, 5002, 100, 8.0, 0,
	).Error)

	// 只有 HAS-DEDUCTION 有扣减记录，NO-DEDUCTION 没有 → 触发 continue 分支
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_deduction_log (uuid, sale_order_uuid, erp_code, qty, stock_entry_name, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		8001, 1001, "HAS-DEDUCTION", 3.0, "SE-001", 0,
	).Error)

	srv := &warehouseSrv{}
	result, err := srv.getPendingStockEntryConsumption(newTestCtx(), db)
	require.NoError(t, err)

	// NO-DEDUCTION 应保持原值 8.0（deducted=0 → continue，不扣减）
	assert.InDelta(t, 8.0, result[100]["NO-DEDUCTION"], 0.001, "Code without deduction should remain unchanged")
	// HAS-DEDUCTION: 10.0 - 3.0 = 7.0
	assert.InDelta(t, 7.0, result[100]["HAS-DEDUCTION"], 0.001, "Code with deduction should be reduced")
}

// TestQueryDeductedByCode_EmptyOrderUuids 验证空 orderUuids 直接返回空 map
func TestQueryDeductedByCode_EmptyOrderUuids(t *testing.T) {
	t.Parallel()
	db := setupPendingSETestDB(t)
	srv := &warehouseSrv{}

	result := srv.queryDeductedByCode(newTestCtx(), db, []uint64{})
	assert.Empty(t, result, "Empty orderUuids should return empty map immediately")
}

// TestBuildSIModeOpts_DBError 验证 getPendingStockEntryConsumption 出错时正确透传
func TestBuildSIModeOpts_DBError(t *testing.T) {
	t.Parallel()
	// 创建一个没有 sale_order 表的 DB，触发 collectSaleOrderConsumption 查询失败
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
	})
	require.NoError(t, err)

	srv := &warehouseSrv{}
	ctx := context.NewContext(
		context.WithCompanyUuid(99999),
		context.WithCompanySetting(model.CompanySetting{ErpInvoiceMode: 1}),
	)

	consumption, opts, resultErr := srv.buildSIModeOpts(ctx, db)
	assert.Error(t, resultErr, "Should return error when DB query fails")
	assert.Nil(t, consumption)
	assert.Nil(t, opts)
}

// TestQueryDeductedByCode_DBError 验证查询失败时优雅降级返回空 map
func TestQueryDeductedByCode_DBError(t *testing.T) {
	t.Parallel()
	// DB without stock_deduction_log table → query will fail
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
	})
	require.NoError(t, err)

	srv := &warehouseSrv{}
	result := srv.queryDeductedByCode(newTestCtx(), db, []uint64{1001, 1002})
	assert.Empty(t, result, "DB error should return empty map (graceful fallback)")
}
