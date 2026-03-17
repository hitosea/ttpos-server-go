package service

import (
	stderrors "errors"
	"strings"
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
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
	utils.InitIdGenerator()
}

// ─── buildProfitLossLog tests ───

func TestBuildProfitLossLog_NoDifference(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{WarehouseUuid: 100, OrderNo: "SR-001"}
	item := &model.StockReconciliationItem{
		CountedQuantity: decimal.NewFromFloat(10.0),
		BookedQuantity:  decimal.NewFromFloat(10.0),
	}
	assert.Nil(t, srv.buildProfitLossLog(sr, item), "No log when counted == booked")
}

func TestBuildProfitLossLog_ProfitIn(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{WarehouseUuid: 100, OrderNo: "SR-001"}
	item := &model.StockReconciliationItem{
		MaterialUuid:    5001,
		MaterialName:    "TestItem",
		CountedQuantity: decimal.NewFromFloat(15.0),
		BookedQuantity:  decimal.NewFromFloat(10.0),
		Material: &model.Material{
			Unit: &model.MaterialUnit{},
		},
	}
	log := srv.buildProfitLossLog(sr, item)
	require.NotNil(t, log)
	assert.Equal(t, constant.WarehouseInOutLogLogTypeIn, log.LogType, "Profit should be In")
	assert.Equal(t, constant.WarehouseInOutLogSceneProfitIn, log.Scene)
	assert.InDelta(t, 5.0, log.Num, 0.001)
	assert.Equal(t, uint64(100), log.WarehouseUuid)
	assert.Equal(t, "SR-001", log.OrderNo)
}

func TestBuildProfitLossLog_LossOut(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{WarehouseUuid: 200, OrderNo: "SR-002"}
	item := &model.StockReconciliationItem{
		MaterialUuid:    5002,
		MaterialName:    "LossItem",
		CountedQuantity: decimal.NewFromFloat(3.0),
		BookedQuantity:  decimal.NewFromFloat(10.0),
		Material: &model.Material{
			Unit: &model.MaterialUnit{},
		},
	}
	log := srv.buildProfitLossLog(sr, item)
	require.NotNil(t, log)
	assert.Equal(t, constant.WarehouseInOutLogLogTypeOut, log.LogType, "Loss should be Out")
	assert.Equal(t, constant.WarehouseInOutLogSceneLossOut, log.Scene)
	assert.InDelta(t, 7.0, log.Num, 0.001)
}

// ─── checkDisabledMaterials tests ───

func TestCheckDisabledMaterials_AllEnabled(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{
			{Material: &model.Material{Status: true, MultiLanguageName: model.MultiLanguageName{}}},
			{Material: &model.Material{Status: true, MultiLanguageName: model.MultiLanguageName{}}},
		},
	}
	result := srv.checkDisabledMaterials(sr)
	assert.Empty(t, result)
}

func TestCheckDisabledMaterials_SomeDisabled(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{
			{Material: &model.Material{Status: true, MultiLanguageName: model.MultiLanguageName{}}},
			{Material: &model.Material{Status: false, MultiLanguageName: model.MultiLanguageName{}}},
		},
	}
	result := srv.checkDisabledMaterials(sr)
	assert.Len(t, result, 1)
}

func TestCheckDisabledMaterials_SkipsDeleted(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{
			{
				Material: &model.Material{Status: false, MultiLanguageName: model.MultiLanguageName{}},
			},
			{
				Material: &model.Material{Status: false, MultiLanguageName: model.MultiLanguageName{}},
			},
		},
	}
	// Set DeleteTime on second item via BaseModel
	sr.StockReconciliationItems[1].DeleteTime = 100
	result := srv.checkDisabledMaterials(sr)
	assert.Len(t, result, 1, "Deleted items should be skipped")
}

// ─── upsertWarehouseItemStock tests ───

func setupWarehouseItemTestDB(t *testing.T) *gorm.DB {
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

	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_warehouse_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_code TEXT DEFAULT '',
			stock REAL DEFAULT 0,
			reserved_stock REAL DEFAULT 0,
			valuation REAL DEFAULT 0
		)
	`).Error)

	return db
}

func TestUpsertWarehouseItemStock_Create(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	repo := repository.NewWarehouseItemRepo(db)
	item := &model.StockReconciliationItem{
		MaterialUuid:    5001,
		CountedQuantity: decimal.NewFromFloat(25.5),
		Material: &model.Material{
			Code: "ITEM-A",
		},
	}
	existingMaterials := map[uint64]struct{}{} // empty = material doesn't exist in warehouse

	err := srv.upsertWarehouseItemStock(repo, 100, item, existingMaterials)
	require.NoError(t, err)

	// Verify created
	var count int64
	db.Table("ttpos_warehouse_item").Where("warehouse_uuid = 100 AND material_uuid = 5001").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestUpsertWarehouseItemStock_Update(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	// Pre-insert existing warehouse item
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, material_code, stock, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		1, 100, 5001, "ITEM-A", 10.0, 0,
	).Error)

	repo := repository.NewWarehouseItemRepo(db)
	item := &model.StockReconciliationItem{
		MaterialUuid:    5001,
		CountedQuantity: decimal.NewFromFloat(30.0),
		Material: &model.Material{
			Code: "ITEM-A",
		},
	}
	existingMaterials := map[uint64]struct{}{5001: {}} // material exists

	err := srv.upsertWarehouseItemStock(repo, 100, item, existingMaterials)
	require.NoError(t, err)

	// Verify updated
	var stock float64
	db.Table("ttpos_warehouse_item").Where("warehouse_uuid = 100 AND material_uuid = 5001").Select("stock").Scan(&stock)
	assert.InDelta(t, 30.0, stock, 0.001)
}

// ─── getIsInventoryStatusException tests ───

func TestGetIsInventoryStatusException_BothZero(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	assert.False(t, srv.getIsInventoryStatusException(decimal.Zero, decimal.Zero),
		"Both zero should not be exception")
}

func TestGetIsInventoryStatusException_BookedZeroCountedNonZero(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	assert.True(t, srv.getIsInventoryStatusException(decimal.Zero, decimal.NewFromFloat(5.0)),
		"Booked=0, Counted>0 should be exception")
}

func TestGetIsInventoryStatusException_Within30Percent(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	// 10 vs 12 = 20% variance, within 30% threshold
	assert.False(t, srv.getIsInventoryStatusException(
		decimal.NewFromFloat(10.0), decimal.NewFromFloat(12.0)),
		"20% variance should not be exception")
}

func TestGetIsInventoryStatusException_Exactly30Percent(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	// 10 vs 13 = exactly 30%, GreaterThan(0.3) is false
	assert.False(t, srv.getIsInventoryStatusException(
		decimal.NewFromFloat(10.0), decimal.NewFromFloat(13.0)),
		"Exactly 30% should not be exception (threshold is >30%)")
}

func TestGetIsInventoryStatusException_Exceeds30Percent(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	// 10 vs 14 = 40% variance, exceeds 30%
	assert.True(t, srv.getIsInventoryStatusException(
		decimal.NewFromFloat(10.0), decimal.NewFromFloat(14.0)),
		"40% variance should be exception")
}

func TestGetIsInventoryStatusException_LossExceeds30Percent(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	// 10 vs 5 = 50% loss, exceeds 30%
	assert.True(t, srv.getIsInventoryStatusException(
		decimal.NewFromFloat(10.0), decimal.NewFromFloat(5.0)),
		"50% loss should be exception")
}

// ─── isZeroValuationRate tests ───

func TestIsZeroValuationRate_KeyMissing(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	m := map[string]float64{"OTHER": 10.0}
	assert.True(t, srv.isZeroValuationRate(m, "ITEM-A"),
		"Missing key means no bin record → treat as zero")
}

func TestIsZeroValuationRate_RateIsZero(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	m := map[string]float64{"ITEM-A": 0.0}
	assert.True(t, srv.isZeroValuationRate(m, "ITEM-A"),
		"Rate=0 should be zero valuation")
}

func TestIsZeroValuationRate_RatePositive(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	m := map[string]float64{"ITEM-A": 15.5}
	assert.False(t, srv.isZeroValuationRate(m, "ITEM-A"),
		"Positive rate should not be zero valuation")
}

func TestIsZeroValuationRate_EmptyMap(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	assert.True(t, srv.isZeroValuationRate(map[string]float64{}, "ITEM-A"),
		"Empty map means no bin data → zero valuation")
}

// ─── extractName tests ───

func TestExtractName_PlainText(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	errorMsg := "Item ITEM-001 is not available in warehouse"
	result := srv.extractName("Item", "is not available", errorMsg)
	assert.Equal(t, "ITEM-001", result)
}

func TestExtractName_WithCodeHashName(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	errorMsg := "Item MAT-001#鸡翅 is not available in warehouse"
	result := srv.extractName("Item", "is not available", errorMsg)
	assert.Equal(t, "鸡翅", result, "Should extract name part after #")
}

func TestExtractName_HTMLMode(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	errorMsg := "<b>TestItem</b>"
	result := srv.extractName("<b>", "</b>", errorMsg)
	assert.Equal(t, "TestItem", result)
}

func TestExtractName_NoMatch(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	result := srv.extractName("NotFound", "Boundary", "some random error message")
	assert.Equal(t, "", result, "No match should return empty string")
}

func TestExtractName_MultipleHashParts(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	// SplitN with 2 means only split at first #
	errorMsg := "Item CODE#Name#Extra is not in stock"
	result := srv.extractName("Item", "is not in stock", errorMsg)
	assert.Equal(t, "Name#Extra", result, "SplitN(2) should keep everything after first #")
}

// ─── getWarehouseMaterialSet tests ───

func TestGetWarehouseMaterialSet_Empty(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	m, err := srv.getWarehouseMaterialSet(db, 100)
	require.NoError(t, err)
	assert.Empty(t, m)
}

func TestGetWarehouseMaterialSet_WithItems(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	// Insert items for warehouse 100
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, delete_time) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
		1, 100, 5001, 0,
		2, 100, 5002, 0,
	).Error)
	// Insert item for different warehouse — should not appear
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, delete_time) VALUES (?, ?, ?, ?)",
		3, 200, 5003, 0,
	).Error)

	m, err := srv.getWarehouseMaterialSet(db, 100)
	require.NoError(t, err)
	assert.Len(t, m, 2)
	_, ok1 := m[5001]
	_, ok2 := m[5002]
	_, ok3 := m[5003]
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.False(t, ok3, "Item from warehouse 200 should not appear")
}

// ─── getWarehouseMaterialUuidMap tests ───

func TestGetWarehouseMaterialUuidMap_Empty(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	m, err := srv.getWarehouseMaterialUuidMap(db, 100)
	require.NoError(t, err)
	assert.Empty(t, m)
}

func TestGetWarehouseMaterialUuidMap_WithItems(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, delete_time) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
		1, 100, 5001, 0,
		2, 100, 5002, 0,
	).Error)

	m, err := srv.getWarehouseMaterialUuidMap(db, 100)
	require.NoError(t, err)
	assert.Len(t, m, 2)
	assert.True(t, m[5001])
	assert.True(t, m[5002])
}

// ─── getBookedQuantityMap tests ───

func TestGetBookedQuantityMap_Empty(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	m, err := srv.getBookedQuantityMap(db, 100)
	require.NoError(t, err)
	assert.Empty(t, m)
}

func TestGetBookedQuantityMap_WithItems(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, stock, delete_time) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
		1, 100, 5001, 25.5, 0,
		2, 100, 5002, 100.0, 0,
	).Error)

	m, err := srv.getBookedQuantityMap(db, 100)
	require.NoError(t, err)
	assert.Len(t, m, 2)
	assert.True(t, m[5001].Equal(decimal.NewFromFloat(25.5)))
	assert.True(t, m[5002].Equal(decimal.NewFromFloat(100.0)))
}

func TestGetBookedQuantityMap_IgnoresDeleted(t *testing.T) {
	t.Parallel()
	db := setupWarehouseItemTestDB(t)
	srv := &stockReconciliationSrv{}

	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, stock, delete_time) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
		1, 100, 5001, 25.5, 0,
		2, 100, 5002, 50.0, 1000, // soft-deleted
	).Error)

	m, err := srv.getBookedQuantityMap(db, 100)
	require.NoError(t, err)
	assert.Len(t, m, 1, "Deleted items should be excluded")
	assert.True(t, m[5001].Equal(decimal.NewFromFloat(25.5)))
}

// ─── updateStatusAndAnnotation tests ───

func setupStockReconciliationTestDB(t *testing.T) *gorm.DB {
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

	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_stock_reconciliation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			order_no TEXT DEFAULT '',
			erp_code TEXT DEFAULT '',
			type INTEGER DEFAULT 1,
			warehouse_uuid INTEGER DEFAULT 0,
			purpose INTEGER DEFAULT 1,
			status INTEGER DEFAULT 0,
			submit_time INTEGER DEFAULT 0,
			submitter_staff_uuid INTEGER DEFAULT 0
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_stock_reconciliation_annotation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			stock_reconciliation_uuid INTEGER DEFAULT 0,
			annotation_type INTEGER DEFAULT 0,
			content TEXT DEFAULT ''
		)
	`).Error)

	return db
}

func TestUpdateStatusAndAnnotation_Success(t *testing.T) {
	t.Parallel()
	db := setupStockReconciliationTestDB(t)
	srv := &stockReconciliationSrv{}

	// Insert a stock reconciliation record
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, delete_time) VALUES (?, ?, ?)",
		9001, constant.StockReconciliationStatusSubmitted, 0,
	).Error)

	err := srv.updateStatusAndAnnotation(db, 9001, "审核通过")
	require.NoError(t, err)

	// Verify status updated to approved
	var status int
	db.Table("ttpos_stock_reconciliation").Where("uuid = 9001").Select("status").Scan(&status)
	assert.Equal(t, constant.StockReconciliationStatusApproved, status)

	// Verify annotation created
	var annotation model.StockReconciliationAnnotation
	db.Table("ttpos_stock_reconciliation_annotation").Where("stock_reconciliation_uuid = 9001").First(&annotation)
	assert.Equal(t, "审核通过", annotation.Content)
	assert.Equal(t, constant.StockReconciliationAnnotationTypeApprove, annotation.AnnotationType)
}

// ─── generateOrderNo tests ───

func setupNumberSequenceTestDB(t *testing.T) *gorm.DB {
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

	require.NoError(t, db.Exec(`
		CREATE TABLE ttpos_number_sequence (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			company_uuid INTEGER DEFAULT 0,
			type TEXT DEFAULT '',
			date TEXT DEFAULT '',
			sequence INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0
		)
	`).Error)

	return db
}

func TestGenerateOrderNo_Format(t *testing.T) {
	t.Parallel()
	db := setupNumberSequenceTestDB(t)
	srv := &stockReconciliationSrv{}

	orderNo, err := srv.generateOrderNo(db, 8267304538112000, "Asia/Shanghai")
	require.NoError(t, err)

	// Should start with "ST"
	assert.True(t, strings.HasPrefix(orderNo, "ST"), "Order number should start with ST")
	// ST + 14-digit timestamp + 4-digit sequence = 20 chars
	assert.Equal(t, 20, len(orderNo), "Order number should be 20 characters: ST(2) + timestamp(14) + seq(4)")
	// Should end with 0001 (first sequence)
	assert.True(t, strings.HasSuffix(orderNo, "0001"), "First order number should end with 0001")
}

func TestGenerateOrderNo_Incrementing(t *testing.T) {
	t.Parallel()
	db := setupNumberSequenceTestDB(t)
	srv := &stockReconciliationSrv{}

	no1, err := srv.generateOrderNo(db, 8267304538112000, "Asia/Shanghai")
	require.NoError(t, err)
	no2, err := srv.generateOrderNo(db, 8267304538112000, "Asia/Shanghai")
	require.NoError(t, err)

	// Same prefix (ST + same day timestamp up to seconds may differ), different sequence
	assert.True(t, strings.HasSuffix(no1, "0001"))
	assert.True(t, strings.HasSuffix(no2, "0002"), "Second call should increment sequence")
}

// ─── checkDisabledMaterials edge cases ───

func TestCheckDisabledMaterials_EmptyItems(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{},
	}
	result := srv.checkDisabledMaterials(sr)
	assert.Empty(t, result, "Empty item list should return empty result")
}

func TestCheckDisabledMaterials_NilItems(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	sr := &model.StockReconciliation{}
	result := srv.checkDisabledMaterials(sr)
	assert.Empty(t, result, "Nil item list should return empty result")
}

// ─── processStockReconciliationItems tests ───

func setupProcessItemsTestDB(t *testing.T) *gorm.DB {
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

	for _, ddl := range []string{
		`CREATE TABLE ttpos_warehouse_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_code TEXT DEFAULT '',
			stock REAL DEFAULT 0,
			reserved_stock REAL DEFAULT 0,
			valuation REAL DEFAULT 0
		)`,
		`CREATE TABLE ttpos_warehouse_in_out_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			log_type INTEGER DEFAULT 0,
			scene INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_name TEXT DEFAULT '',
			material_base_unit_uuid INTEGER DEFAULT 0,
			material_base_unit_name TEXT DEFAULT '',
			num REAL DEFAULT 0,
			price REAL DEFAULT 0,
			amount REAL DEFAULT 0,
			supplier_uuid INTEGER DEFAULT 0,
			supplier_erp_code TEXT DEFAULT '',
			supplier_name TEXT DEFAULT '',
			order_no TEXT DEFAULT '',
			other_org_uuid INTEGER DEFAULT 0,
			other_org_type INTEGER DEFAULT 0,
			other_org_name TEXT DEFAULT '',
			opening_hours TEXT DEFAULT ''
		)`,
		`CREATE TABLE ttpos_stock_reconciliation_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			stock_reconciliation_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_name TEXT DEFAULT '',
			counted_quantity REAL DEFAULT 0,
			booked_quantity REAL DEFAULT 0,
			material_code TEXT DEFAULT ''
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func TestProcessStockReconciliationItems_HappyPath(t *testing.T) {
	t.Parallel()
	db := setupProcessItemsTestDB(t)
	srv := &stockReconciliationSrv{}

	// Pre-insert existing warehouse item for material 5001
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse_item (uuid, warehouse_uuid, material_uuid, material_code, stock, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		1, 100, 5001, "MAT-A", 10.0, 0,
	).Error)

	sr := &model.StockReconciliation{
		WarehouseUuid: 100,
		OrderNo:       "SR-001",
		StockReconciliationItems: []*model.StockReconciliationItem{
			{
				MaterialUuid:    5001,
				MaterialName:    "ItemA",
				CountedQuantity: decimal.NewFromFloat(15.0),
				BookedQuantity:  decimal.NewFromFloat(10.0),
				Material: &model.Material{
					Code: "MAT-A",
					Unit: &model.MaterialUnit{},
				},
			},
			{
				MaterialUuid:    5002,
				MaterialName:    "ItemB",
				CountedQuantity: decimal.NewFromFloat(20.0),
				BookedQuantity:  decimal.NewFromFloat(20.0), // no difference → no log
				Material: &model.Material{
					Code: "MAT-B",
					Unit: &model.MaterialUnit{},
				},
			},
		},
	}
	existingMaterials := map[uint64]struct{}{5001: {}}

	err := srv.processStockReconciliationItems(db, sr, existingMaterials)
	require.NoError(t, err)

	// Verify warehouse item 5001 was updated to 15.0
	var stock float64
	db.Table("ttpos_warehouse_item").Where("warehouse_uuid = 100 AND material_uuid = 5001").Select("stock").Scan(&stock)
	assert.InDelta(t, 15.0, stock, 0.001)

	// Verify warehouse item 5002 was created (new)
	var count int64
	db.Table("ttpos_warehouse_item").Where("warehouse_uuid = 100 AND material_uuid = 5002").Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify one profit log was created (only item 5001 has difference)
	var logCount int64
	db.Table("ttpos_warehouse_in_out_log").Count(&logCount)
	assert.Equal(t, int64(1), logCount, "Only one in/out log should be created (ItemA has profit)")
}

func TestProcessStockReconciliationItems_SkipsDeletedItems(t *testing.T) {
	t.Parallel()
	db := setupProcessItemsTestDB(t)
	srv := &stockReconciliationSrv{}

	sr := &model.StockReconciliation{
		WarehouseUuid: 100,
		OrderNo:       "SR-002",
		StockReconciliationItems: []*model.StockReconciliationItem{
			{
				MaterialUuid:    5001,
				MaterialName:    "DeletedItem",
				CountedQuantity: decimal.NewFromFloat(10.0),
				BookedQuantity:  decimal.NewFromFloat(5.0),
				Material: &model.Material{
					Code: "MAT-A",
					Unit: &model.MaterialUnit{},
				},
			},
		},
	}
	sr.StockReconciliationItems[0].DeleteTime = 100 // soft-deleted

	err := srv.processStockReconciliationItems(db, sr, map[uint64]struct{}{})
	require.NoError(t, err)

	// No warehouse items should be created
	var count int64
	db.Table("ttpos_warehouse_item").Count(&count)
	assert.Equal(t, int64(0), count, "Deleted items should be skipped entirely")

	// No logs should be created
	var logCount int64
	db.Table("ttpos_warehouse_in_out_log").Count(&logCount)
	assert.Equal(t, int64(0), logCount)
}

// ─── helper: create test context ───

func newTestContext(version, lang string) context.Context {
	ctx := context.NewContext(
		context.WithLanguage(lang),
	)
	ctx.SetVersion(version)
	return ctx
}

// ─── getIsInventoryStatusException tests ───

func TestGetIsInventoryStatusException(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}

	tests := []struct {
		name     string
		booked   float64
		counted  float64
		expected bool
	}{
		{"both zero → no exception", 0, 0, false},
		{"booked zero counted nonzero → exception", 0, 5, true},
		{"equal → no exception", 100, 100, false},
		{"within 30% → no exception", 100, 125, false},
		{"exactly 30% → no exception", 100, 130, false},
		{"over 30% → exception", 100, 131, true},
		{"loss within 30% → no exception", 100, 75, false},
		{"loss over 30% → exception", 100, 60, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := srv.getIsInventoryStatusException(
				decimal.NewFromFloat(tt.booked),
				decimal.NewFromFloat(tt.counted),
			)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ─── handleErpApproveError tests ───

func TestHandleErpApproveError_DisabledWarehouse_V2100(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")
	sr := &model.StockReconciliation{}

	err := srv.handleErpApproveError(ctx, sr, stderrors.New("Disabled Warehouse WH-001"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "仓库状态已关闭")
}

func TestHandleErpApproveError_DisabledWarehouse_OldVersion(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.9.0", "zh")
	sr := &model.StockReconciliation{}

	// Old version should NOT trigger the Disabled Warehouse branch
	err := srv.handleErpApproveError(ctx, sr, stderrors.New("Disabled Warehouse WH-001"))
	require.Error(t, err)
	// Falls through to extractName which returns "" → generic error
	assert.Contains(t, err.Error(), errMsgApproveReconcilation)
}

func TestHandleErpApproveError_ItemDisabled_WithMatch(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{
			{
				Material: &model.Material{
					Code: "MAT-001",
					MultiLanguageName: model.MultiLanguageName{
						ZhName: "鸡翅",
					},
				},
			},
		},
	}

	err := srv.handleErpApproveError(ctx, sr, stderrors.New("Item MAT-001 is disabled"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "鸡翅")
}

func TestHandleErpApproveError_ItemDisabled_NoMatch(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{
			{
				Material: &model.Material{
					Code:              "OTHER",
					MultiLanguageName: model.MultiLanguageName{},
				},
			},
		},
	}

	err := srv.handleErpApproveError(ctx, sr, stderrors.New("Item MAT-999 is disabled"))
	require.Error(t, err)
	// itemName stays as "MAT-999" since no match in sr items
	assert.Contains(t, err.Error(), "MAT-999")
}

func TestHandleErpApproveError_GenericError(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")
	sr := &model.StockReconciliation{}

	err := srv.handleErpApproveError(ctx, sr, stderrors.New("some random ERP error"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), errMsgApproveReconcilation)
}

func TestHandleErpApproveError_ItemDisabled_OldVersion(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.9.0", "zh")
	sr := &model.StockReconciliation{
		StockReconciliationItems: []*model.StockReconciliationItem{
			{
				Material: &model.Material{
					Code:              "MAT-001",
					MultiLanguageName: model.MultiLanguageName{ZhName: "鸡翅"},
				},
			},
		},
	}

	err := srv.handleErpApproveError(ctx, sr, stderrors.New("Item MAT-001 is disabled"))
	require.Error(t, err)
	// Old version uses i18n.Translate which may return format string in test env
	assert.Contains(t, err.Error(), "状态已关闭")
}

// ─── loadAndValidateStockReconciliation tests ───

func setupLoadValidateTestDB(t *testing.T) *gorm.DB {
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

	for _, ddl := range []string{
		`CREATE TABLE ttpos_stock_reconciliation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			order_no TEXT DEFAULT '',
			erp_code TEXT DEFAULT '',
			type INTEGER DEFAULT 1,
			warehouse_uuid INTEGER DEFAULT 0,
			purpose INTEGER DEFAULT 1,
			status INTEGER DEFAULT 0,
			submit_time INTEGER DEFAULT 0,
			submitter_staff_uuid INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_warehouse (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			multi_language_name_uuid INTEGER DEFAULT 0,
			type TEXT DEFAULT '',
			code TEXT DEFAULT '',
			status INTEGER DEFAULT 1,
			contact TEXT DEFAULT '',
			phone TEXT DEFAULT '',
			address TEXT DEFAULT '',
			is_default INTEGER DEFAULT 0,
			erp_code TEXT DEFAULT '',
			headquarter_uuid INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_stock_reconciliation_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			stock_reconciliation_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_name TEXT DEFAULT '',
			counted_quantity REAL DEFAULT 0,
			booked_quantity REAL DEFAULT 0
		)`,
		`CREATE TABLE ttpos_material (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			code TEXT DEFAULT '',
			multi_language_name_uuid INTEGER DEFAULT 0,
			unit_uuid INTEGER DEFAULT 0,
			status INTEGER DEFAULT 1,
			barcode_value TEXT DEFAULT '',
			internal_code TEXT DEFAULT '',
			category_uuid INTEGER DEFAULT 0,
			supplier_uuid INTEGER DEFAULT 0,
			image_uuid INTEGER DEFAULT 0,
			image_name TEXT DEFAULT '',
			purchase_unit_uuid INTEGER DEFAULT 0,
			cost_unit_uuid INTEGER DEFAULT 0,
			default_sales_unit_uuid INTEGER DEFAULT 0,
			price REAL DEFAULT 0,
			stock_num REAL DEFAULT 0,
			actual_sale_num REAL DEFAULT 0,
			headquarter_uuid INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			init_stock REAL DEFAULT 0,
			allow_substore_visible INTEGER DEFAULT 0,
			origin_country_code TEXT DEFAULT '',
			allow_negative_stock INTEGER DEFAULT 0,
			delivered_by_supplier INTEGER DEFAULT 0,
			supplier_erp_code TEXT DEFAULT '',
			specification TEXT DEFAULT ''
		)`,
		`CREATE TABLE ttpos_multi_language_name (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			zh_name TEXT DEFAULT '',
			en_name TEXT DEFAULT '',
			zh_tw_name TEXT DEFAULT '',
			th_name TEXT DEFAULT '',
			my_name TEXT DEFAULT '',
			ja_name TEXT DEFAULT '',
			ko_name TEXT DEFAULT '',
			tr_name TEXT DEFAULT '',
			sv_name TEXT DEFAULT '',
			not_overwrite INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_material_unit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			unit_uuid INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			conversion_rate REAL DEFAULT 0,
			is_default INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_stock_reconciliation_annotation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			stock_reconciliation_uuid INTEGER DEFAULT 0,
			annotation_type INTEGER DEFAULT 0,
			content TEXT DEFAULT ''
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func TestLoadAndValidateStockReconciliation_NotFound(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")

	result, err := srv.loadAndValidateStockReconciliation(ctx, db, 99999)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "盘点单不存在")
}

func TestLoadAndValidateStockReconciliation_WrongStatus(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")

	// Insert a record with status=0 (saved, not submitted)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, warehouse_uuid, delete_time) VALUES (?, ?, ?, ?)",
		1001, constant.StockReconciliationStatusSaved, 100, 0,
	).Error)

	result, err := srv.loadAndValidateStockReconciliation(ctx, db, 1001)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "当前状态不允许审核")
}

func TestLoadAndValidateStockReconciliation_DisabledWarehouse(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")

	// Insert warehouse with status=0 (disabled)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse (uuid, status, delete_time) VALUES (?, ?, ?)",
		100, 0, 0,
	).Error)
	// Insert submitted stock reconciliation
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, warehouse_uuid, delete_time) VALUES (?, ?, ?, ?)",
		1001, constant.StockReconciliationStatusSubmitted, 100, 0,
	).Error)

	result, err := srv.loadAndValidateStockReconciliation(ctx, db, 1001)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "仓库状态已关闭")
}

func TestLoadAndValidateStockReconciliation_HappyPath(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")

	// Insert active warehouse
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse (uuid, status, delete_time) VALUES (?, ?, ?)",
		100, 1, 0,
	).Error)
	// Insert submitted stock reconciliation
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, warehouse_uuid, order_no, erp_code, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		1001, constant.StockReconciliationStatusSubmitted, 100, "SR-001", "ERP-001", 0,
	).Error)

	result, err := srv.loadAndValidateStockReconciliation(ctx, db, 1001)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint64(1001), result.Uuid)
	assert.Equal(t, "SR-001", result.OrderNo)
	assert.Equal(t, constant.StockReconciliationStatusSubmitted, result.Status)
}

// ─── approveStockReconciliationInERP tests ───

func TestApproveStockReconciliationInERP_NoErp(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")
	// Company with IsEnableErp=0 → IsOpenErp() returns false
	ctx.SetCompany(model.Company{IsEnableErp: 0})
	sr := &model.StockReconciliation{ErpCode: "SR-ERP-001"}

	err := srv.approveStockReconciliationInERP(ctx, sr)
	assert.NoError(t, err, "Should return nil when ERP is not enabled")
}

func TestApproveStockReconciliationInERP_EmptyErpCode(t *testing.T) {
	t.Parallel()
	srv := &stockReconciliationSrv{}
	ctx := newTestContext("2.10.0", "zh")
	ctx.SetCompany(model.Company{IsEnableErp: 1})
	sr := &model.StockReconciliation{ErpCode: ""} // empty erp code

	err := srv.approveStockReconciliationInERP(ctx, sr)
	assert.NoError(t, err, "Should return nil when ErpCode is empty")
}

// ─── ApproveStockReconciliation tests ───

// setupApproveTestDB extends setupLoadValidateTestDB with tables needed for
// the full approve flow (warehouse_item, warehouse_in_out_log).
func setupApproveTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupLoadValidateTestDB(t)
	for _, ddl := range []string{
		`CREATE TABLE ttpos_warehouse_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_code TEXT DEFAULT '',
			stock REAL DEFAULT 0,
			reserved_stock REAL DEFAULT 0,
			valuation REAL DEFAULT 0
		)`,
		`CREATE TABLE ttpos_warehouse_in_out_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			log_type INTEGER DEFAULT 0,
			scene INTEGER DEFAULT 0,
			warehouse_uuid INTEGER DEFAULT 0,
			material_uuid INTEGER DEFAULT 0,
			material_name TEXT DEFAULT '',
			material_base_unit_uuid INTEGER DEFAULT 0,
			material_base_unit_name TEXT DEFAULT '',
			num REAL DEFAULT 0,
			price REAL DEFAULT 0,
			amount REAL DEFAULT 0,
			supplier_uuid INTEGER DEFAULT 0,
			supplier_erp_code TEXT DEFAULT '',
			supplier_name TEXT DEFAULT '',
			order_no TEXT DEFAULT '',
			other_org_uuid INTEGER DEFAULT 0,
			other_org_type INTEGER DEFAULT 0,
			other_org_name TEXT DEFAULT '',
			opening_hours TEXT DEFAULT ''
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func TestApproveStockReconciliation_NotFound(t *testing.T) {
	t.Parallel()
	db := setupApproveTestDB(t)
	srv := newSrvWithLock()
	ctx := newTestContext("2.10.0", "zh")
	ctx.SetDB(db)

	resp, err := srv.ApproveStockReconciliation(ctx, req.StockReconciliationApproveReq{Uuid: 99999})
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "盘点单不存在")
}

func TestApproveStockReconciliation_DisabledMaterial(t *testing.T) {
	t.Parallel()
	db := setupApproveTestDB(t)
	srv := newSrvWithLock()
	ctx := newTestContext("2.10.0", "zh")
	ctx.SetDB(db)

	// Active warehouse
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse (uuid, status, delete_time) VALUES (?, ?, ?)", 100, 1, 0,
	).Error)
	// Multi-language name
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_multi_language_name (uuid, zh_name, en_name) VALUES (?, ?, ?)", 5001, "测试物品", "TestItem",
	).Error)
	// Disabled material (status=0)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, status, multi_language_name_uuid, unit_uuid, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		3001, "MAT-001", 0, 5001, 0, 0,
	).Error)
	// Submitted stock reconciliation
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, warehouse_uuid, delete_time) VALUES (?, ?, ?, ?)",
		1001, constant.StockReconciliationStatusSubmitted, 100, 0,
	).Error)
	// Reconciliation item referencing the disabled material
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation_item (uuid, stock_reconciliation_uuid, material_uuid, material_name, counted_quantity, booked_quantity, delete_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		2001, 1001, 3001, "测试物品", 10.0, 8.0, 0,
	).Error)

	resp, err := srv.ApproveStockReconciliation(ctx, req.StockReconciliationApproveReq{Uuid: 1001})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "请修改物品状态")
	assert.NotEmpty(t, resp, "Should return disabled material names")
}

func TestApproveStockReconciliation_Success_NoErp(t *testing.T) {
	t.Parallel()
	db := setupApproveTestDB(t)
	srv := newSrvWithLock()
	ctx := newTestContext("2.10.0", "zh")
	ctx.SetDB(db)
	ctx.SetCompany(model.Company{IsEnableErp: 0}) // skip ERP call

	// Active warehouse
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_warehouse (uuid, status, delete_time) VALUES (?, ?, ?)", 100, 1, 0,
	).Error)
	// Multi-language name + unit
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_multi_language_name (uuid, zh_name) VALUES (?, ?)", 5001, "物品A",
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material_unit (uuid, material_uuid, name, is_default) VALUES (?, ?, ?, ?)", 6001, 3001, "kg", 1,
	).Error)
	// Active material (status=1)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_material (uuid, code, status, multi_language_name_uuid, unit_uuid, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		3001, "MAT-001", 1, 5001, 6001, 0,
	).Error)
	// Submitted stock reconciliation (no erp_code → approveStockReconciliationInERP skips)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, warehouse_uuid, order_no, erp_code, delete_time) VALUES (?, ?, ?, ?, ?, ?)",
		1001, constant.StockReconciliationStatusSubmitted, 100, "SR-001", "", 0,
	).Error)
	// Reconciliation item
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation_item (uuid, stock_reconciliation_uuid, material_uuid, material_name, counted_quantity, booked_quantity, delete_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		2001, 1001, 3001, "物品A", 15.0, 10.0, 0,
	).Error)

	resp, err := srv.ApproveStockReconciliation(ctx, req.StockReconciliationApproveReq{
		Uuid:       1001,
		Annotation: "审核通过",
	})
	require.NoError(t, err)
	assert.Nil(t, resp, "Should return nil on success")

	// Verify status updated to approved
	var status int
	db.Table("ttpos_stock_reconciliation").Where("uuid = 1001").Select("status").Scan(&status)
	assert.Equal(t, constant.StockReconciliationStatusApproved, status)

	// Verify annotation created
	var annotation model.StockReconciliationAnnotation
	db.Table("ttpos_stock_reconciliation_annotation").Where("stock_reconciliation_uuid = 1001").First(&annotation)
	assert.Equal(t, "审核通过", annotation.Content)
	assert.Equal(t, constant.StockReconciliationAnnotationTypeApprove, annotation.AnnotationType)

	// Verify warehouse item created (material was not in warehouse before)
	var warehouseItem model.WarehouseItem
	db.Table("ttpos_warehouse_item").Where("warehouse_uuid = 100 AND material_uuid = 3001").First(&warehouseItem)
	assert.Equal(t, 15.0, warehouseItem.Stock, "Stock should be set to counted quantity")
}

// ─── RejectStockReconciliation tests ───

func newSrvWithLock() *stockReconciliationSrv {
	return &stockReconciliationSrv{
		lock: lock.InitLocalLock(),
	}
}

func TestRejectStockReconciliation_NotFound(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := newSrvWithLock()

	ctx := newTestContext("2.10.0", "zh")
	ctx.SetDB(db)

	err := srv.RejectStockReconciliation(ctx, req.StockReconciliationRejectReq{
		Uuid:       99999,
		Annotation: "不合格",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "盘点单不存在")
}

func TestRejectStockReconciliation_WrongStatus(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := newSrvWithLock()

	// Insert record with status=saved (not submitted)
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, delete_time) VALUES (?, ?, ?)",
		9001, constant.StockReconciliationStatusSaved, 0,
	).Error)

	ctx := newTestContext("2.10.0", "zh")
	ctx.SetDB(db)

	err := srv.RejectStockReconciliation(ctx, req.StockReconciliationRejectReq{
		Uuid:       9001,
		Annotation: "不合格",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "盘点单状态不允许驳回")
}

func TestRejectStockReconciliation_Success(t *testing.T) {
	t.Parallel()
	db := setupLoadValidateTestDB(t)
	srv := newSrvWithLock()

	// Insert a submitted stock reconciliation
	require.NoError(t, db.Exec(
		"INSERT INTO ttpos_stock_reconciliation (uuid, status, delete_time) VALUES (?, ?, ?)",
		9001, constant.StockReconciliationStatusSubmitted, 0,
	).Error)

	ctx := newTestContext("2.10.0", "zh")
	ctx.SetDB(db)

	err := srv.RejectStockReconciliation(ctx, req.StockReconciliationRejectReq{
		Uuid:       9001,
		Annotation: "数量有误，请重新核对",
	})
	require.NoError(t, err)

	// Verify status updated to rejected
	var status int
	db.Table("ttpos_stock_reconciliation").Where("uuid = 9001").Select("status").Scan(&status)
	assert.Equal(t, constant.StockReconciliationStatusRejected, status)

	// Verify annotation created
	var annotation model.StockReconciliationAnnotation
	db.Table("ttpos_stock_reconciliation_annotation").Where("stock_reconciliation_uuid = 9001").First(&annotation)
	assert.Equal(t, "数量有误，请重新核对", annotation.Content)
	assert.Equal(t, constant.StockReconciliationAnnotationTypeReject, annotation.AnnotationType)
}

