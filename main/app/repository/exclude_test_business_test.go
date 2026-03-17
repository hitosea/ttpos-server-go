package repository

import (
	"strings"
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

// ─── ExcludeTestBusinessSQL (pure function) ─────────────────────────

func TestExcludeTestBusinessSQL_NoAlias(t *testing.T) {
	sql := ExcludeTestBusinessSQL("")
	if !strings.Contains(sql, "NOT EXISTS") {
		t.Error("expected NOT EXISTS clause")
	}
	// 无别名时应使用派生表消除歧义
	if !strings.Contains(sql, "SELECT start_time, end_time FROM ttpos_business_status_period WHERE delete_time = 0") {
		t.Error("expected derived table for no-alias variant")
	}
	// 不应包含 tableAlias.create_time
	if strings.Contains(sql, ".create_time") {
		t.Error("no-alias variant should not contain qualified create_time")
	}
}

func TestExcludeTestBusinessSQL_WithAlias(t *testing.T) {
	sql := ExcludeTestBusinessSQL("sp")
	if !strings.Contains(sql, "NOT EXISTS") {
		t.Error("expected NOT EXISTS clause")
	}
	if !strings.Contains(sql, "sp.create_time") {
		t.Error("expected sp.create_time in SQL")
	}
	if !strings.Contains(sql, "bsp.delete_time = 0") {
		t.Error("expected delete_time filter")
	}
	if !strings.Contains(sql, "bsp.end_time = 0") {
		t.Error("expected open period check (end_time = 0)")
	}
}

func TestExcludeTestBusinessSQL_DifferentAliases(t *testing.T) {
	aliases := []string{"t", "sb", "to_order", "smp"}
	for _, alias := range aliases {
		sql := ExcludeTestBusinessSQL(alias)
		expected := alias + ".create_time"
		if !strings.Contains(sql, expected) {
			t.Errorf("alias %q: expected %q in SQL, got: %s", alias, expected, sql)
		}
	}
}

// ─── ExcludeTestBusinessByFieldSQL (pure function) ──────────────────

func TestExcludeTestBusinessByFieldSQL_CustomExpr(t *testing.T) {
	sql := ExcludeTestBusinessByFieldSQL("(date + hour * 3600)")
	if !strings.Contains(sql, "NOT EXISTS") {
		t.Error("expected NOT EXISTS clause")
	}
	if !strings.Contains(sql, "(date + hour * 3600) >= bsp.start_time") {
		t.Error("expected custom field expression in start_time comparison")
	}
	if !strings.Contains(sql, "(date + hour * 3600) <= bsp.end_time") {
		t.Error("expected custom field expression in end_time comparison")
	}
	if !strings.Contains(sql, "bsp.delete_time = 0") {
		t.Error("expected soft-delete filter")
	}
}

func TestExcludeTestBusinessByFieldSQL_SimpleField(t *testing.T) {
	sql := ExcludeTestBusinessByFieldSQL("t.create_time")
	if !strings.Contains(sql, "t.create_time >= bsp.start_time") {
		t.Error("expected t.create_time in start comparison")
	}
	// 应与 ExcludeTestBusinessSQL("t") 输出一致
	sqlFromAlias := ExcludeTestBusinessSQL("t")
	if sql != sqlFromAlias {
		t.Errorf("ExcludeTestBusinessByFieldSQL('t.create_time') should match ExcludeTestBusinessSQL('t')\ngot:  %s\nwant: %s", sql, sqlFromAlias)
	}
}

// ─── ExcludeTestBusinessByRelatedOrderSQL (pure function) ───────────

func TestExcludeTestBusinessByRelatedOrderSQL(t *testing.T) {
	sql := ExcludeTestBusinessByRelatedOrderSQL()
	if !strings.Contains(sql, "NOT EXISTS") {
		t.Error("expected NOT EXISTS clause")
	}
	if !strings.Contains(sql, "ttpos_sale_order so") {
		t.Error("expected join to sale_order")
	}
	if !strings.Contains(sql, "ttpos_sale_bill sb") {
		t.Error("expected join to sale_bill")
	}
	if !strings.Contains(sql, "sb.uuid = so.sale_bill_uuid") {
		t.Error("expected sale_order → sale_bill join condition")
	}
	if !strings.Contains(sql, "sb.create_time >= bsp.start_time") {
		t.Error("expected time range filter on sale_bill.create_time")
	}
	if !strings.Contains(sql, "so.uuid = related_order_uuid") {
		t.Error("expected correlation to related_order_uuid")
	}
	if !strings.Contains(sql, "bsp.delete_time = 0") {
		t.Error("expected soft-delete filter")
	}
}

// ─── WhereExcludeTestBusinessPeriod with table alias (DB) ───────────

func TestExcludeTestBusiness_WithTableAlias(t *testing.T) {
	db := setupExcludeTestDB(t)
	insertOrders(t, db, 100, 200, 300, 400)
	insertPeriod(t, db, 150, 250, 0) // 排除200

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("ttpos_sale_bill")
	var orders []testOrder
	if err := opt(db.Model(&testOrder{})).Find(&orders).Error; err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(orders) != 3 {
		t.Fatalf("expected 3 orders with alias, got %d", len(orders))
	}
	for _, o := range orders {
		if o.CreateTime == 200 {
			t.Error("order with create_time=200 should be excluded with alias")
		}
	}
}

// ─── WhereExcludeTestBusinessByRelatedOrder (DB) ────────────────────

// testReturnOrder 简化的退单模型
type testReturnOrder struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	RelatedOrderUuid uint64 `gorm:"column:related_order_uuid"`
}

func (testReturnOrder) TableName() string {
	return "ttpos_return_order"
}

func setupRelatedOrderTestDB(t *testing.T) *gorm.DB {
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
		`CREATE TABLE ttpos_return_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			related_order_uuid INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_sale_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			sale_bill_uuid INTEGER DEFAULT 0
		)`,
		`CREATE TABLE ttpos_sale_bill (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
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

func TestWhereExcludeTestBusinessByRelatedOrder_NoPeriods(t *testing.T) {
	db := setupRelatedOrderTestDB(t)
	// sale_bill(uuid=1001, create_time=100) → sale_order(uuid=2001) → return_order
	db.Exec("INSERT INTO ttpos_sale_bill (uuid, create_time) VALUES (1001, 100)")
	db.Exec("INSERT INTO ttpos_sale_order (uuid, sale_bill_uuid) VALUES (2001, 1001)")
	db.Exec("INSERT INTO ttpos_return_order (related_order_uuid) VALUES (2001)")

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessByRelatedOrder()
	var orders []testReturnOrder
	if err := opt(db.Model(&testReturnOrder{})).Find(&orders).Error; err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 return order (no periods), got %d", len(orders))
	}
}

func TestWhereExcludeTestBusinessByRelatedOrder_ExcludesMatching(t *testing.T) {
	db := setupRelatedOrderTestDB(t)
	// 订单1: sale_bill create_time=100 → 不在测试时段
	db.Exec("INSERT INTO ttpos_sale_bill (uuid, create_time) VALUES (1001, 100)")
	db.Exec("INSERT INTO ttpos_sale_order (uuid, sale_bill_uuid) VALUES (2001, 1001)")
	db.Exec("INSERT INTO ttpos_return_order (related_order_uuid) VALUES (2001)")

	// 订单2: sale_bill create_time=200 → 在测试时段 [150, 250]
	db.Exec("INSERT INTO ttpos_sale_bill (uuid, create_time) VALUES (1002, 200)")
	db.Exec("INSERT INTO ttpos_sale_order (uuid, sale_bill_uuid) VALUES (2002, 1002)")
	db.Exec("INSERT INTO ttpos_return_order (related_order_uuid) VALUES (2002)")

	// 测试时段 [150, 250]
	insertPeriod(t, db, 150, 250, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessByRelatedOrder()
	var orders []testReturnOrder
	if err := opt(db.Model(&testReturnOrder{})).Find(&orders).Error; err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 return order, got %d", len(orders))
	}
	if orders[0].RelatedOrderUuid != 2001 {
		t.Errorf("expected related_order_uuid=2001, got %d", orders[0].RelatedOrderUuid)
	}
}

func TestWhereExcludeTestBusinessByRelatedOrder_OpenPeriod(t *testing.T) {
	db := setupRelatedOrderTestDB(t)
	// 订单1: create_time=100
	db.Exec("INSERT INTO ttpos_sale_bill (uuid, create_time) VALUES (1001, 100)")
	db.Exec("INSERT INTO ttpos_sale_order (uuid, sale_bill_uuid) VALUES (2001, 1001)")
	db.Exec("INSERT INTO ttpos_return_order (related_order_uuid) VALUES (2001)")

	// 订单2: create_time=300
	db.Exec("INSERT INTO ttpos_sale_bill (uuid, create_time) VALUES (1002, 300)")
	db.Exec("INSERT INTO ttpos_sale_order (uuid, sale_bill_uuid) VALUES (2002, 1002)")
	db.Exec("INSERT INTO ttpos_return_order (related_order_uuid) VALUES (2002)")

	// 开放测试时段 [200, 0)
	insertPeriod(t, db, 200, 0, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessByRelatedOrder()
	var orders []testReturnOrder
	if err := opt(db.Model(&testReturnOrder{})).Find(&orders).Error; err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 return order, got %d", len(orders))
	}
	if orders[0].RelatedOrderUuid != 2001 {
		t.Errorf("expected related_order_uuid=2001, got %d", orders[0].RelatedOrderUuid)
	}
}

func TestWhereExcludeTestBusinessByRelatedOrder_DeletedPeriodIgnored(t *testing.T) {
	db := setupRelatedOrderTestDB(t)
	db.Exec("INSERT INTO ttpos_sale_bill (uuid, create_time) VALUES (1001, 200)")
	db.Exec("INSERT INTO ttpos_sale_order (uuid, sale_bill_uuid) VALUES (2001, 1001)")
	db.Exec("INSERT INTO ttpos_return_order (related_order_uuid) VALUES (2001)")

	// 已删除时段覆盖订单时间，但应被忽略
	insertPeriod(t, db, 100, 300, 999)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessByRelatedOrder()
	var orders []testReturnOrder
	if err := opt(db.Model(&testReturnOrder{})).Find(&orders).Error; err != nil {
		t.Fatalf("query error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 return order (deleted period), got %d", len(orders))
	}
}

// ─── Additional edge cases ──────────────────────────────────────────

func TestExcludeTestBusiness_OverlappingPeriods(t *testing.T) {
	db := setupExcludeTestDB(t)
	insertOrders(t, db, 100, 200, 300, 400, 500)
	// 重叠时段: [150, 350] 和 [250, 450]
	insertPeriod(t, db, 150, 350, 0)
	insertPeriod(t, db, 250, 450, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	// 200,300 在时段1; 300,400 在时段2 → 排除 200,300,400 → 剩 100,500
	if len(ids) != 2 {
		t.Fatalf("expected 2 orders with overlapping periods, got %d", len(ids))
	}
	for _, id := range ids {
		if id != 1 && id != 5 {
			t.Errorf("unexpected order id=%d", id)
		}
	}
}

func TestExcludeTestBusiness_AllOrdersExcluded(t *testing.T) {
	db := setupExcludeTestDB(t)
	insertOrders(t, db, 100, 200, 300)
	// 开放时段从0开始，排除所有订单
	insertPeriod(t, db, 0, 0, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 0 {
		t.Fatalf("expected 0 orders (all excluded by open period from 0), got %d", len(ids))
	}
}

func TestExcludeTestBusiness_MixedDeletedAndActive(t *testing.T) {
	db := setupExcludeTestDB(t)
	insertOrders(t, db, 100, 200, 300, 400, 500)
	// 已删除时段: [50, 250] → 应忽略
	insertPeriod(t, db, 50, 250, 999)
	// 活跃时段: [350, 450] → 排除400
	insertPeriod(t, db, 350, 450, 0)

	repo := &commonRepo{}
	opt := repo.WhereExcludeTestBusinessPeriod("")
	ids := queryOrderIDs(t, opt(db.Model(&testOrder{})))

	if len(ids) != 4 {
		t.Fatalf("expected 4 orders (deleted period ignored, only 400 excluded), got %d", len(ids))
	}
	for _, id := range ids {
		if id == 4 {
			t.Fatal("order id=4 should be excluded by active period")
		}
	}
}

func TestApplyExcludeTestBusiness_OverlappingPeriods(t *testing.T) {
	db := setupTakeoutExcludeTestDB(t)
	insertTakeoutOrders(t, db, 100, 200, 300, 400, 500)
	insertPeriod(t, db, 150, 350, 0)
	insertPeriod(t, db, 250, 450, 0)

	query := applyExcludeTestBusiness(db.Model(&testTakeoutOrder{}), "")
	ids := queryTakeoutOrderIDs(t, query)

	if len(ids) != 2 {
		t.Fatalf("expected 2 orders with overlapping periods, got %d", len(ids))
	}
}

func TestApplyExcludeTestBusiness_DeletedPeriodIgnored(t *testing.T) {
	db := setupTakeoutExcludeTestDB(t)
	insertTakeoutOrders(t, db, 100, 200, 300)
	insertPeriod(t, db, 50, 350, 999) // 已删除

	query := applyExcludeTestBusiness(db.Model(&testTakeoutOrder{}), "")
	ids := queryTakeoutOrderIDs(t, query)

	if len(ids) != 3 {
		t.Fatalf("expected 3 orders (deleted period ignored), got %d", len(ids))
	}
}
