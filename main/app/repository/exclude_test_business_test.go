package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testOrder 简化的订单模型，用于验证过滤逻辑
type testOrder struct {
	ID         uint  `gorm:"primaryKey;autoIncrement"`
	CreateTime int64 `gorm:"column:create_time"`
}

func (testOrder) TableName() string {
	return "ttpos_sale_bill"
}

// setupExcludeTestDB 创建内存数据库并建表
func setupExcludeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	for _, ddl := range []string{
		`CREATE TABLE ttpos_sale_bill (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			create_time INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_business_status_period (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			start_time INTEGER DEFAULT 0,
			end_time INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("ddl: %v", err)
		}
	}
	return db
}

// insertOrders 批量插入订单
func insertOrders(t *testing.T, db *gorm.DB, createTimes ...int64) {
	t.Helper()
	for _, ct := range createTimes {
		if err := db.Exec("INSERT INTO ttpos_sale_bill (create_time) VALUES (?)", ct).Error; err != nil {
			t.Fatalf("insert order: %v", err)
		}
	}
}

// insertPeriod 插入测试营业时段
func insertPeriod(t *testing.T, db *gorm.DB, startTime, endTime, deleteTime int64) {
	t.Helper()
	if err := db.Exec(
		"INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time) VALUES (0, ?, ?, 0, 0, ?)",
		startTime, endTime, deleteTime,
	).Error; err != nil {
		t.Fatalf("insert period: %v", err)
	}
}

// queryOrderIDs 执行查询并返回剩余订单ID列表
func queryOrderIDs(t *testing.T, db *gorm.DB) []uint {
	t.Helper()
	var orders []testOrder
	if err := db.Find(&orders).Error; err != nil {
		t.Fatalf("query orders: %v", err)
	}
	ids := make([]uint, 0, len(orders))
	for _, o := range orders {
		ids = append(ids, o.ID)
	}
	return ids
}

// ─── WhereExcludeTestBusinessPeriod (DBOption) ─────────────────────────

func TestExcludeTestBusiness_NoPeriods(t *testing.T) {
	db := setupExcludeTestDB(t)
	insertOrders(t, db, 100, 200, 300)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 3 {
		t.Fatalf("expected 3 orders, got %d", len(ids))
	}
}

func TestExcludeTestBusiness_ClosedPeriod(t *testing.T) {
	db := setupExcludeTestDB(t)
	// 订单: 100, 200, 300, 400
	insertOrders(t, db, 100, 200, 300, 400)
	// 测试时段: [150, 250] → 排除 create_time 在 150~250 的订单(200)
	insertPeriod(t, db, 150, 250, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 3 {
		t.Fatalf("expected 3 orders, got %d", len(ids))
	}
	for _, id := range ids {
		if id == 2 { // id=2 is create_time=200, should be excluded
			t.Fatal("order with create_time=200 should be excluded")
		}
	}
}

func TestExcludeTestBusiness_OpenPeriod(t *testing.T) {
	db := setupExcludeTestDB(t)
	// 订单: 100, 200, 300
	insertOrders(t, db, 100, 200, 300)
	// 开放时段: [150, 0) → 排除 create_time >= 150 的订单(200, 300)
	insertPeriod(t, db, 150, 0, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 1 {
		t.Fatalf("expected 1 order, got %d", len(ids))
	}
	if ids[0] != 1 {
		t.Fatalf("expected order id=1, got %d", ids[0])
	}
}

func TestExcludeTestBusiness_MultiplePeriods(t *testing.T) {
	db := setupExcludeTestDB(t)
	// 订单: 100, 200, 300, 400, 500
	insertOrders(t, db, 100, 200, 300, 400, 500)
	// 时段1: [150, 250] → 排除200
	insertPeriod(t, db, 150, 250, 0)
	// 时段2: [350, 450] → 排除400
	insertPeriod(t, db, 350, 450, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 3 {
		t.Fatalf("expected 3 orders, got %d", len(ids))
	}
	for _, id := range ids {
		if id == 2 || id == 4 {
			t.Fatalf("order id=%d should be excluded", id)
		}
	}
}

func TestExcludeTestBusiness_DeletedPeriodIgnored(t *testing.T) {
	db := setupExcludeTestDB(t)
	// 订单: 100, 200, 300
	insertOrders(t, db, 100, 200, 300)
	// 已删除的时段: [50, 350] delete_time=999 → 应被忽略，全部订单保留
	insertPeriod(t, db, 50, 350, 999)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 3 {
		t.Fatalf("expected 3 orders (deleted period ignored), got %d", len(ids))
	}
}

func TestExcludeTestBusiness_BoundaryExact(t *testing.T) {
	db := setupExcludeTestDB(t)
	// 订单 create_time 正好等于 start_time 和 end_time
	insertOrders(t, db, 100, 200, 300)
	// 时段: [100, 300] → 全部订单的 create_time 都在范围内，应全部排除
	insertPeriod(t, db, 100, 300, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 0 {
		t.Fatalf("expected 0 orders (all within boundary), got %d", len(ids))
	}
}

// ─── applyExcludeTestBusiness (takeout helper) ─────────────────────────

// testTakeoutOrder 简化的外卖订单模型
type testTakeoutOrder struct {
	ID         uint  `gorm:"primaryKey;autoIncrement"`
	CreateTime int64 `gorm:"column:create_time"`
}

func (testTakeoutOrder) TableName() string {
	return "ttpos_takeout_order"
}

func setupTakeoutExcludeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	for _, ddl := range []string{
		`CREATE TABLE ttpos_takeout_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			create_time INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_business_status_period (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			start_time INTEGER DEFAULT 0,
			end_time INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("ddl: %v", err)
		}
	}
	return db
}

func insertTakeoutOrders(t *testing.T, db *gorm.DB, createTimes ...int64) {
	t.Helper()
	for _, ct := range createTimes {
		if err := db.Exec("INSERT INTO ttpos_takeout_order (create_time) VALUES (?)", ct).Error; err != nil {
			t.Fatalf("insert takeout order: %v", err)
		}
	}
}

func queryTakeoutOrderIDs(t *testing.T, db *gorm.DB) []uint {
	t.Helper()
	var orders []testTakeoutOrder
	if err := db.Find(&orders).Error; err != nil {
		t.Fatalf("query orders: %v", err)
	}
	ids := make([]uint, 0, len(orders))
	for _, o := range orders {
		ids = append(ids, o.ID)
	}
	return ids
}

func TestApplyExcludeTestBusiness_NoAlias(t *testing.T) {
	db := setupTakeoutExcludeTestDB(t)
	insertTakeoutOrders(t, db, 100, 200, 300)
	insertPeriod(t, db, 150, 250, 0) // 排除 200

	query := applyExcludeTestBusiness(db.Model(&testTakeoutOrder{}), "")
	ids := queryTakeoutOrderIDs(t, query)

	if len(ids) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(ids))
	}
}

func TestApplyExcludeTestBusiness_WithAlias(t *testing.T) {
	db := setupTakeoutExcludeTestDB(t)
	insertTakeoutOrders(t, db, 100, 200, 300)
	insertPeriod(t, db, 150, 250, 0) // 排除 200

	query := applyExcludeTestBusiness(
		db.Table("ttpos_takeout_order AS t").Select("t.*"),
		"t",
	)
	ids := queryTakeoutOrderIDs(t, query)

	if len(ids) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(ids))
	}
}

func TestApplyExcludeTestBusiness_OpenPeriod(t *testing.T) {
	db := setupTakeoutExcludeTestDB(t)
	insertTakeoutOrders(t, db, 100, 200, 300, 400)
	// 开放时段: [250, 0) → 排除 300, 400
	insertPeriod(t, db, 250, 0, 0)

	query := applyExcludeTestBusiness(db.Model(&testTakeoutOrder{}), "")
	ids := queryTakeoutOrderIDs(t, query)

	if len(ids) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(ids))
	}
	for _, id := range ids {
		if id == 3 || id == 4 {
			t.Fatalf("order id=%d should be excluded (open period)", id)
		}
	}
}

func TestApplyExcludeTestBusiness_NoPeriods(t *testing.T) {
	db := setupTakeoutExcludeTestDB(t)
	insertTakeoutOrders(t, db, 100, 200, 300)

	query := applyExcludeTestBusiness(db.Model(&testTakeoutOrder{}), "")
	ids := queryTakeoutOrderIDs(t, query)

	if len(ids) != 3 {
		t.Fatalf("expected 3 orders (no periods), got %d", len(ids))
	}
}
