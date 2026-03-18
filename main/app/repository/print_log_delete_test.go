package repository

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func setupPrinterLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "ttpos_",
		},
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS ttpos_printer_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid BIGINT DEFAULT 0,
		printer_uuid BIGINT DEFAULT 0,
		cashier_device_id VARCHAR(255) DEFAULT '',
		related_type TINYINT DEFAULT 0,
		related_uuid BIGINT DEFAULT 0,
		data VARCHAR(255) DEFAULT '',
		type INT DEFAULT 0,
		data_type TINYINT DEFAULT 1,
		print_method TINYINT DEFAULT 1,
		num INT DEFAULT 0,
		status TINYINT DEFAULT 1,
		reason VARCHAR(255) DEFAULT '',
		printer_time INT DEFAULT 0,
		first_execution TINYINT DEFAULT 0,
		read_device_id VARCHAR(255) DEFAULT '',
		product_printer_uuid BIGINT DEFAULT 0,
		printer_type VARCHAR(50) DEFAULT '',
		printing_time INT DEFAULT 0,
		copies INT DEFAULT 0,
		print_speed TINYINT DEFAULT 2,
		create_time INT DEFAULT 0,
		update_time INT DEFAULT 0,
		delete_time INT DEFAULT 0
	)`)

	return db
}

func TestDelete7DaysAgo_PhysicalDelete(t *testing.T) {
	db := setupPrinterLogTestDB(t)

	now := time.Now()
	eightDaysAgo := now.Add(-8 * 24 * time.Hour).Unix()
	oneDayAgo := now.Add(-1 * 24 * time.Hour).Unix()

	// 插入：2条过期 + 1条未过期
	db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (1001, ?, 0)", eightDaysAgo)
	db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (1002, ?, 0)", eightDaysAgo)
	db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (1003, ?, 0)", oneDayAgo)

	repo := NewPrinterLogRepoImpl(db)
	if err := repo.Delete7DaysAgo(); err != nil {
		t.Fatalf("Delete7DaysAgo failed: %v", err)
	}

	// Raw SQL 查全量，验证物理删除
	var allRows []struct {
		Uuid       int64
		DeleteTime int64
	}
	db.Raw("SELECT uuid, delete_time FROM ttpos_printer_logs").Scan(&allRows)

	if len(allRows) != 1 || allRows[0].Uuid != 1003 {
		t.Errorf("expected only uuid=1003 remaining, got %+v", allRows)
	} else {
		t.Log("✅ 物理删除：过期记录已从数据库中永久删除")
	}
}

func TestDelete7DaysAgo_BatchDelete(t *testing.T) {
	db := setupPrinterLogTestDB(t)

	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour).Unix()
	oneDayAgo := time.Now().Add(-1 * 24 * time.Hour).Unix()

	// 插入 25000 条过期数据（超过单批 10000 的限制）+ 5 条未过期
	for i := 0; i < 25000; i++ {
		db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (?, ?, 0)", 10000+i, thirtyDaysAgo)
	}
	for i := 0; i < 5; i++ {
		db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (?, ?, 0)", 90000+i, oneDayAgo)
	}

	var totalBefore int64
	db.Raw("SELECT COUNT(*) FROM ttpos_printer_logs").Scan(&totalBefore)
	t.Logf("删除前: %d 条", totalBefore)

	repo := NewPrinterLogRepoImpl(db)
	if err := repo.Delete7DaysAgo(); err != nil {
		t.Fatalf("Delete7DaysAgo failed: %v", err)
	}

	var totalAfter int64
	db.Raw("SELECT COUNT(*) FROM ttpos_printer_logs").Scan(&totalAfter)
	t.Logf("删除后: %d 条", totalAfter)

	if totalAfter != 5 {
		t.Errorf("expected 5 remaining, got %d", totalAfter)
	} else {
		t.Logf("✅ 分批删除成功：25000 条过期数据已清理，保留 %d 条未过期", totalAfter)
	}
}

func TestDelete7DaysAgo_EmptyTable(t *testing.T) {
	db := setupPrinterLogTestDB(t)

	repo := NewPrinterLogRepoImpl(db)
	err := repo.Delete7DaysAgo()
	if err != nil {
		t.Errorf("empty table should not error, got: %v", err)
	} else {
		t.Log("✅ 空表执行无报错")
	}
}

func TestDelete7DaysAgo_NothingToDelete(t *testing.T) {
	db := setupPrinterLogTestDB(t)

	// 只有未过期数据
	oneDayAgo := time.Now().Add(-1 * 24 * time.Hour).Unix()
	for i := 0; i < 3; i++ {
		db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (?, ?, 0)", i+1, oneDayAgo)
	}

	repo := NewPrinterLogRepoImpl(db)
	if err := repo.Delete7DaysAgo(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	db.Raw("SELECT COUNT(*) FROM ttpos_printer_logs").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	} else {
		t.Log("✅ 无过期数据时不删除任何记录")
	}
}

func TestDelete7DaysAgo_BoundaryTime(t *testing.T) {
	db := setupPrinterLogTestDB(t)

	now := time.Now()
	exactly7DaysAgo := now.Add(-7 * 24 * time.Hour).Unix()
	justOver7Days := now.Add(-7*24*time.Hour - 1*time.Second).Unix()
	justUnder7Days := now.Add(-7*24*time.Hour + 1*time.Second).Unix()

	db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (1, ?, 0)", justOver7Days)
	db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (2, ?, 0)", exactly7DaysAgo)
	db.Exec("INSERT INTO ttpos_printer_logs (uuid, create_time, delete_time) VALUES (3, ?, 0)", justUnder7Days)

	repo := NewPrinterLogRepoImpl(db)
	if err := repo.Delete7DaysAgo(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var remaining []struct {
		Uuid       int64
		CreateTime int64
	}
	db.Raw("SELECT uuid, create_time FROM ttpos_printer_logs ORDER BY uuid").Scan(&remaining)

	t.Logf("阈值: %d (7天前)", exactly7DaysAgo)
	for _, r := range remaining {
		t.Logf("  保留: uuid=%d, create_time=%d", r.Uuid, r.CreateTime)
	}

	// create_time < threshold => uuid=1 被删，uuid=2 刚好等于不删，uuid=3 不删
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining (uuid=2,3), got %d: %+v", len(remaining), remaining)
	} else {
		fmt.Println("✅ 边界值正确：恰好7天的记录保留，超过7天的删除")
	}
}
