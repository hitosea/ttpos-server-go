package repository

import (
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	if config.Database.TablePrefix == "" {
		config.Database.TablePrefix = "ttpos_"
	}
}

func setupSaleOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Exec(`
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
	`).Error
	require.NoError(t, err, "Failed to create sale_order table")

	return db
}

func seedSaleOrders(t *testing.T, db *gorm.DB) {
	t.Helper()
	orders := []map[string]any{
		// 待 SE 扣减：已结账 + 有 SI + 未扣减
		{"uuid": 1001, "status": constant.SaleOrderStatusFinish, "erp_stock_deducted": 0, "erp_sales_invoice_name": "ACC-SINV-001", "delete_time": 0},
		{"uuid": 1002, "status": constant.SaleOrderStatusFinish, "erp_stock_deducted": 0, "erp_sales_invoice_name": "ACC-SINV-002", "delete_time": 0},
		// 已扣减：已结账 + 有 SI + 已扣减
		{"uuid": 1003, "status": constant.SaleOrderStatusFinish, "erp_stock_deducted": 1, "erp_sales_invoice_name": "ACC-SINV-003", "delete_time": 0},
		// 未同步：已结账 + 无 SI
		{"uuid": 1004, "status": constant.SaleOrderStatusFinish, "erp_stock_deducted": 0, "erp_sales_invoice_name": "", "delete_time": 0},
		// 未结账
		{"uuid": 1005, "status": 0, "erp_stock_deducted": 0, "erp_sales_invoice_name": "", "delete_time": 0},
		// 已删除
		{"uuid": 1006, "status": constant.SaleOrderStatusFinish, "erp_stock_deducted": 0, "erp_sales_invoice_name": "ACC-SINV-006", "delete_time": 100},
	}
	for _, o := range orders {
		err := db.Exec("INSERT INTO ttpos_sale_order (uuid, status, erp_stock_deducted, erp_sales_invoice_name, delete_time) VALUES (?, ?, ?, ?, ?)",
			o["uuid"], o["status"], o["erp_stock_deducted"], o["erp_sales_invoice_name"], o["delete_time"]).Error
		require.NoError(t, err)
	}
}

func TestWherePendingStockEntry(t *testing.T) {
	t.Parallel()
	db := setupSaleOrderTestDB(t)
	seedSaleOrders(t, db)
	repo := NewSaleOrderRepo(db)

	orders, err := repo.GetSaleOrderList(repo.WherePendingStockEntry())
	require.NoError(t, err)

	// 应该只返回 1001、1002（已结账 + 有 SI + 未扣减 + 未删除）
	uuids := extractUuids(orders)
	assert.ElementsMatch(t, []uint64{1001, 1002}, uuids,
		"WherePendingStockEntry should return only undeducted, invoiced, finished, non-deleted orders")
}

func TestWhereSettledWithInvoice(t *testing.T) {
	t.Parallel()
	db := setupSaleOrderTestDB(t)
	seedSaleOrders(t, db)
	repo := NewSaleOrderRepo(db)

	orders, err := repo.GetSaleOrderList(repo.WhereSettledWithInvoice())
	require.NoError(t, err)

	// 应该返回 1001、1002、1003（已结账 + 有 SI + 未删除，不区分 erp_stock_deducted）
	uuids := extractUuids(orders)
	assert.ElementsMatch(t, []uint64{1001, 1002, 1003}, uuids,
		"WhereSettledWithInvoice should return all invoiced, finished, non-deleted orders regardless of deduction status")
}

func TestWherePendingStockEntry_EmptyTable(t *testing.T) {
	t.Parallel()
	db := setupSaleOrderTestDB(t)
	repo := NewSaleOrderRepo(db)

	orders, err := repo.GetSaleOrderList(repo.WherePendingStockEntry())
	require.NoError(t, err)
	assert.Empty(t, orders)
}

func extractUuids(orders []model.SaleOrder) []uint64 {
	uuids := make([]uint64, 0, len(orders))
	for _, o := range orders {
		uuids = append(uuids, o.Uuid)
	}
	return uuids
}
