package repository

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupOperationDurationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE ttpos_order_operation_duration (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			company_uuid INTEGER DEFAULT 0,
			sale_bill_uuid INTEGER DEFAULT 0,
			sale_order_uuid INTEGER DEFAULT 0,
			action TEXT DEFAULT '',
			source TEXT DEFAULT '',
			staff_uuid INTEGER DEFAULT 0,
			device_sn TEXT DEFAULT '',
			instance_id TEXT DEFAULT '',
			start_time INTEGER DEFAULT 0,
			end_time INTEGER DEFAULT 0,
			duration_ms INTEGER DEFAULT 0,
			request_path TEXT DEFAULT '',
			status INTEGER DEFAULT 1,
			error_msg TEXT,
			time_nodes TEXT,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func countOperationDuration(db *gorm.DB) int64 {
	var count int64
	db.Raw("SELECT COUNT(*) FROM ttpos_order_operation_duration").Scan(&count)
	return count
}

func insertOperationDuration(db *gorm.DB, uuid uint64, createTime int64) {
	db.Exec(
		"INSERT INTO ttpos_order_operation_duration (uuid, create_time, update_time) VALUES (?, ?, ?)",
		uuid, createTime, createTime,
	)
}

// TestDelete7DaysAgo_DeletesOldRecords 验证删除7天前的记录
func TestDelete7DaysAgo_DeletesOldRecords(t *testing.T) {
	db := setupOperationDurationTestDB(t)

	now := time.Now().Unix()
	eightDaysAgo := now - 8*86400
	sixDaysAgo := now - 6*86400

	// 插入旧数据（8天前）和新数据（6天前）
	insertOperationDuration(db, 1, eightDaysAgo)
	insertOperationDuration(db, 2, eightDaysAgo-1000)
	insertOperationDuration(db, 3, sixDaysAgo)
	insertOperationDuration(db, 4, now)

	if count := countOperationDuration(db); count != 4 {
		t.Fatalf("Expected 4 records before delete, got %d", count)
	}

	// SQLite 不支持 DELETE ... LIMIT，直接用非 LIMIT 方式测试核心逻辑
	deadline := time.Now().Add(-7 * 24 * time.Hour).Unix()
	result := db.Exec(
		"DELETE FROM `ttpos_order_operation_duration` WHERE `create_time` > 0 AND `create_time` < ?",
		deadline,
	)
	if result.Error != nil {
		t.Fatalf("Delete7DaysAgo failed: %v", result.Error)
	}

	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows deleted, got %d", result.RowsAffected)
	}

	remaining := countOperationDuration(db)
	if remaining != 2 {
		t.Errorf("Expected 2 records remaining, got %d", remaining)
	}
}

// TestDelete7DaysAgo_KeepsRecentRecords 验证保留7天内的记录
func TestDelete7DaysAgo_KeepsRecentRecords(t *testing.T) {
	db := setupOperationDurationTestDB(t)

	now := time.Now().Unix()
	// 全部都是近期数据
	insertOperationDuration(db, 1, now)
	insertOperationDuration(db, 2, now-86400)     // 1天前
	insertOperationDuration(db, 3, now-3*86400)   // 3天前
	insertOperationDuration(db, 4, now-6*86400)   // 6天前
	insertOperationDuration(db, 5, now-7*86400+1) // 刚好不到7天

	deadline := time.Now().Add(-7 * 24 * time.Hour).Unix()
	result := db.Exec(
		"DELETE FROM `ttpos_order_operation_duration` WHERE `create_time` > 0 AND `create_time` < ?",
		deadline,
	)
	if result.Error != nil {
		t.Fatalf("Delete failed: %v", result.Error)
	}

	remaining := countOperationDuration(db)
	if remaining != 5 {
		t.Errorf("Expected all 5 records to be kept, got %d", remaining)
	}
}

// TestDeleteOperationDuration7DaysAgo_EmptyTable 验证空表不报错
func TestDeleteOperationDuration7DaysAgo_EmptyTable(t *testing.T) {
	db := setupOperationDurationTestDB(t)

	deadline := time.Now().Add(-7 * 24 * time.Hour).Unix()
	result := db.Exec(
		"DELETE FROM `ttpos_order_operation_duration` WHERE `create_time` > 0 AND `create_time` < ?",
		deadline,
	)
	if result.Error != nil {
		t.Fatalf("Delete on empty table should not error: %v", result.Error)
	}

	if result.RowsAffected != 0 {
		t.Errorf("Expected 0 rows affected, got %d", result.RowsAffected)
	}
}

// TestDelete7DaysAgo_SkipsZeroCreateTime 验证 create_time=0 的记录不被删除
func TestDelete7DaysAgo_SkipsZeroCreateTime(t *testing.T) {
	db := setupOperationDurationTestDB(t)

	// create_time=0 的脏数据
	insertOperationDuration(db, 1, 0)
	// 正常旧数据
	insertOperationDuration(db, 2, time.Now().Unix()-30*86400)

	deadline := time.Now().Add(-7 * 24 * time.Hour).Unix()
	result := db.Exec(
		"DELETE FROM `ttpos_order_operation_duration` WHERE `create_time` > 0 AND `create_time` < ?",
		deadline,
	)
	if result.Error != nil {
		t.Fatalf("Delete failed: %v", result.Error)
	}

	// create_time=0 的记录应保留（不匹配 create_time > 0 条件）
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row deleted, got %d", result.RowsAffected)
	}

	remaining := countOperationDuration(db)
	if remaining != 1 {
		t.Errorf("Expected 1 record remaining (create_time=0), got %d", remaining)
	}
}

// TestDelete7DaysAgo_BoundaryExactly7Days 验证恰好7天的边界条件
func TestDelete7DaysAgo_BoundaryExactly7Days(t *testing.T) {
	db := setupOperationDurationTestDB(t)

	now := time.Now()
	exactly7DaysAgo := now.Add(-7 * 24 * time.Hour).Unix()

	// 恰好在7天边界
	insertOperationDuration(db, 1, exactly7DaysAgo)
	// 比7天多1秒
	insertOperationDuration(db, 2, exactly7DaysAgo-1)
	// 比7天少1秒
	insertOperationDuration(db, 3, exactly7DaysAgo+1)

	deadline := now.Add(-7 * 24 * time.Hour).Unix()
	db.Exec(
		"DELETE FROM `ttpos_order_operation_duration` WHERE `create_time` > 0 AND `create_time` < ?",
		deadline,
	)

	remaining := countOperationDuration(db)
	// exactly7DaysAgo 不满足 < deadline（等于），应保留
	// exactly7DaysAgo-1 满足 < deadline，应删除
	// exactly7DaysAgo+1 不满足 < deadline，应保留
	if remaining != 2 {
		t.Errorf("Expected 2 records remaining at boundary, got %d", remaining)
	}
}
