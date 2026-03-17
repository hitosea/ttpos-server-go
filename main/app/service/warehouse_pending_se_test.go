package service

import (
	"testing"
	"ttpos-server-go/app/constant"
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
