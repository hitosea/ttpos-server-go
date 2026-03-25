package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupDataManageTestDB(t *testing.T) *gorm.DB {
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
		CREATE TABLE ttpos_data_manage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			type INTEGER DEFAULT 0,
			data_uuid INTEGER DEFAULT 0,
			staff_uuid INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create data_manage table: %v", err)
	}

	return db
}

func insertDataManage(db *gorm.DB, uuid, dataUuid uint64, typ int) {
	db.Exec(`INSERT INTO ttpos_data_manage (uuid, data_uuid, type, delete_time) VALUES (?, ?, ?, 0)`,
		uuid, dataUuid, typ)
}

// TestCreateInBatches 验证批量插入后能正确查询
func TestCreateInBatches(t *testing.T) {
	db := setupDataManageTestDB(t)
	repo := NewDataManageRepoImpl(db)

	for i := uint64(1); i <= 250; i++ {
		insertDataManage(db, i, i+1000, 0)
	}

	uuids := repo.GetDataUuids(repo.WhereByType(0))
	if len(uuids) != 250 {
		t.Errorf("Expected 250 records, got %d", len(uuids))
	}
}

// TestWhereInDataUuids 验证按 data_uuid 列表过滤
func TestWhereInDataUuids(t *testing.T) {
	db := setupDataManageTestDB(t)
	repo := NewDataManageRepoImpl(db)

	for i := uint64(1); i <= 5; i++ {
		insertDataManage(db, i, i+100, 0)
	}

	list := repo.List(
		repo.WhereByType(0),
		repo.WhereInDataUuids([]uint64{101, 103}),
	)
	if len(list) != 2 {
		t.Errorf("Expected 2 records, got %d", len(list))
	}
}

// TestWhereInDataUuids_Empty 验证空切片不会匹配任何记录
func TestWhereInDataUuids_Empty(t *testing.T) {
	db := setupDataManageTestDB(t)
	repo := NewDataManageRepoImpl(db)

	insertDataManage(db, 1, 101, 0)

	list := repo.List(
		repo.WhereByType(0),
		repo.WhereInDataUuids([]uint64{}),
	)
	if len(list) != 0 {
		t.Errorf("Expected 0 records for empty uuids, got %d", len(list))
	}
}

// TestDeleteWithWhereInDataUuids 验证按 data_uuid 列表精确删除
func TestDeleteWithWhereInDataUuids(t *testing.T) {
	db := setupDataManageTestDB(t)
	repo := NewDataManageRepoImpl(db)

	for i := uint64(1); i <= 5; i++ {
		insertDataManage(db, i, i+100, 0)
	}

	err := repo.Delete(
		repo.WhereByType(0),
		repo.WhereInDataUuids([]uint64{102, 104}),
	)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	remaining := repo.GetDataUuids(repo.WhereByType(0))
	if len(remaining) != 3 {
		t.Errorf("Expected 3 remaining records, got %d", len(remaining))
	}

	for _, uid := range remaining {
		if uid == 102 || uid == 104 {
			t.Errorf("data_uuid %d should have been deleted", uid)
		}
	}
}

// TestGetDataUuids 验证按类型获取 data_uuid 列表
func TestGetDataUuids(t *testing.T) {
	db := setupDataManageTestDB(t)
	repo := NewDataManageRepoImpl(db)

	insertDataManage(db, 1, 101, 0)
	insertDataManage(db, 2, 102, 0)
	insertDataManage(db, 3, 201, 1) // 不同 type

	uuids := repo.GetDataUuids(repo.WhereByType(0))
	if len(uuids) != 2 {
		t.Errorf("Expected 2 uuids for type=0, got %d", len(uuids))
	}
}

// TestGetDataUuids_Empty 验证空表返回空切片
func TestGetDataUuids_Empty(t *testing.T) {
	db := setupDataManageTestDB(t)
	repo := NewDataManageRepoImpl(db)

	uuids := repo.GetDataUuids(repo.WhereByType(0))
	if len(uuids) != 0 {
		t.Errorf("Expected 0 uuids, got %d", len(uuids))
	}
}
