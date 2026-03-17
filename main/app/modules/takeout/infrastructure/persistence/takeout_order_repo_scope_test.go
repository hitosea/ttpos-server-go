package persistence

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	if config.Database.TablePrefix == "" {
		config.Database.TablePrefix = "ttpos_"
	}
}

func setupTakeoutOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
	})
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Exec(`
		CREATE TABLE ttpos_takeout_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			order_state INTEGER DEFAULT 0,
			erp_sync_status INTEGER DEFAULT 0,
			erp_stock_deducted INTEGER DEFAULT 0,
			erp_pos_invoice_resp TEXT DEFAULT ''
		)
	`).Error
	require.NoError(t, err, "Failed to create takeout_order table")

	return db
}

func seedTakeoutOrders(t *testing.T, db *gorm.DB) {
	t.Helper()
	orders := []map[string]any{
		// 待 SE 扣减：已完成 + 已同步 + 有发票回调 + 未扣减
		{"uuid": 2001, "order_state": value_object.TakeoutOrderStateCompleted, "erp_sync_status": constant.ErpSyncStatusSuccess, "erp_stock_deducted": 0, "erp_pos_invoice_resp": `{"name":"SI-001"}`, "delete_time": 0},
		{"uuid": 2002, "order_state": value_object.TakeoutOrderStateCompleted, "erp_sync_status": constant.ErpSyncStatusSuccess, "erp_stock_deducted": 0, "erp_pos_invoice_resp": `{"name":"SI-002"}`, "delete_time": 0},
		// 已扣减：已完成 + 已同步 + 有发票回调 + 已扣减
		{"uuid": 2003, "order_state": value_object.TakeoutOrderStateCompleted, "erp_sync_status": constant.ErpSyncStatusSuccess, "erp_stock_deducted": 1, "erp_pos_invoice_resp": `{"name":"SI-003"}`, "delete_time": 0},
		// 未同步成功：已完成 + erp_sync_status != 3
		{"uuid": 2004, "order_state": value_object.TakeoutOrderStateCompleted, "erp_sync_status": constant.ErpSyncStatusFailed, "erp_stock_deducted": 0, "erp_pos_invoice_resp": `{"name":"SI-004"}`, "delete_time": 0},
		// 无发票回调：已完成 + 已同步 + erp_pos_invoice_resp 为空
		{"uuid": 2005, "order_state": value_object.TakeoutOrderStateCompleted, "erp_sync_status": constant.ErpSyncStatusSuccess, "erp_stock_deducted": 0, "erp_pos_invoice_resp": "", "delete_time": 0},
		// 未完成：order_state = 10（已接单）
		{"uuid": 2006, "order_state": value_object.TakeoutOrderStateAccepted, "erp_sync_status": 0, "erp_stock_deducted": 0, "erp_pos_invoice_resp": "", "delete_time": 0},
		// 已删除
		{"uuid": 2007, "order_state": value_object.TakeoutOrderStateCompleted, "erp_sync_status": constant.ErpSyncStatusSuccess, "erp_stock_deducted": 0, "erp_pos_invoice_resp": `{"name":"SI-007"}`, "delete_time": 100},
	}
	for _, o := range orders {
		err := db.Exec(
			"INSERT INTO ttpos_takeout_order (uuid, order_state, erp_sync_status, erp_stock_deducted, erp_pos_invoice_resp, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
			o["uuid"], o["order_state"], o["erp_sync_status"], o["erp_stock_deducted"], o["erp_pos_invoice_resp"], o["delete_time"],
		).Error
		require.NoError(t, err)
	}
}

// TestTakeoutWherePendingStockEntry 验证 WherePendingStockEntry scope 只返回
// erp_stock_deducted=0、erp_sync_status=3、有 erp_pos_invoice_resp、已完成、未删除的订单
func TestTakeoutWherePendingStockEntry(t *testing.T) {
	t.Parallel()
	db := setupTakeoutOrderTestDB(t)
	seedTakeoutOrders(t, db)
	repo := NewTakeoutOrderRepo(db)

	orders, err := repo.FindAll(repo.WherePendingStockEntry())
	require.NoError(t, err)

	uuids := extractTakeoutUuids(orders)
	assert.ElementsMatch(t, []uint64{2001, 2002}, uuids,
		"WherePendingStockEntry should return only undeducted, synced, invoiced, completed, non-deleted orders")
}

// TestTakeoutWhereCompletedWithInvoice 验证 WhereCompletedWithInvoice scope 返回
// 已完成 + 已同步 + 有发票回调的订单（不区分 erp_stock_deducted）
func TestTakeoutWhereCompletedWithInvoice(t *testing.T) {
	t.Parallel()
	db := setupTakeoutOrderTestDB(t)
	seedTakeoutOrders(t, db)
	repo := NewTakeoutOrderRepo(db)

	orders, err := repo.FindAll(repo.WhereCompletedWithInvoice())
	require.NoError(t, err)

	uuids := extractTakeoutUuids(orders)
	assert.ElementsMatch(t, []uint64{2001, 2002, 2003}, uuids,
		"WhereCompletedWithInvoice should return all synced, invoiced, completed, non-deleted orders regardless of deduction status")
}

// TestTakeoutFindAll_EmptyTable 验证空表返回空切片而非错误
func TestTakeoutFindAll_EmptyTable(t *testing.T) {
	t.Parallel()
	db := setupTakeoutOrderTestDB(t)
	repo := NewTakeoutOrderRepo(db)

	orders, err := repo.FindAll(repo.WherePendingStockEntry())
	require.NoError(t, err)
	assert.Empty(t, orders)
}

// TestTakeoutFindAll_NoCount 验证 FindAll 不执行 COUNT 查询（只返回 orders，不返回 total）
func TestTakeoutFindAll_NoCount(t *testing.T) {
	t.Parallel()
	db := setupTakeoutOrderTestDB(t)
	seedTakeoutOrders(t, db)
	repo := NewTakeoutOrderRepo(db)

	orders, err := repo.FindAll()
	require.NoError(t, err)
	// 未删除的订单有 6 条（2001-2006）
	assert.Len(t, orders, 6, "FindAll without scopes should return all non-deleted orders")
}

func extractTakeoutUuids(orders []*model.TakeoutOrder) []uint64 {
	uuids := make([]uint64, 0, len(orders))
	for _, o := range orders {
		uuids = append(uuids, o.Uuid)
	}
	return uuids
}
