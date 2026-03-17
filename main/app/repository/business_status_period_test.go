package repository

import (
	"testing"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupPeriodTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	utils.InitIdGenerator()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	for _, ddl := range []string{
		`DROP TABLE IF EXISTS ttpos_business_status_period`,
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
			t.Fatalf("Failed to execute DDL: %v", err)
		}
	}

	return db
}

func TestGetOpenPeriod_Empty(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	period, err := repo.GetOpenPeriod()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if period != nil {
		t.Error("expected nil period when table is empty")
	}
}

func TestGetOpenPeriod_HasOpen(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	// 插入一条 end_time=0 的记录
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 0, 1000, 1000, 0)`)

	period, err := repo.GetOpenPeriod()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if period == nil {
		t.Fatal("expected non-nil period")
	}
	if period.Uuid != 1001 {
		t.Errorf("expected uuid 1001, got %d", period.Uuid)
	}
	if period.StartTime != 1000 {
		t.Errorf("expected start_time 1000, got %d", period.StartTime)
	}
}

func TestGetOpenPeriod_OnlyClosedPeriods(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	// 所有记录都已关闭（end_time > 0）
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 2000, 1000, 2000, 0)`)
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1002, 3000, 4000, 3000, 4000, 0)`)

	period, err := repo.GetOpenPeriod()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if period != nil {
		t.Error("expected nil period when all periods are closed")
	}
}

func TestGetOpenPeriod_IgnoresDeleted(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	// 软删除的未关闭记录不应返回
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 0, 1000, 1000, 9999)`)

	period, err := repo.GetOpenPeriod()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if period != nil {
		t.Error("expected nil period when open period is soft-deleted")
	}
}

func TestCreate(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	period := &model.BusinessStatusPeriod{
		StartTime: 5000,
	}
	if err := repo.Create(period); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证记录已写入
	var count int64
	db.Model(&model.BusinessStatusPeriod{}).Where("delete_time = 0").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}

	// 新创建的记录 end_time 应为 0
	open, _ := repo.GetOpenPeriod()
	if open == nil {
		t.Fatal("expected open period after create")
	}
	if open.StartTime != 5000 {
		t.Errorf("expected start_time 5000, got %d", open.StartTime)
	}
	if open.EndTime != 0 {
		t.Errorf("expected end_time 0, got %d", open.EndTime)
	}
}

func TestCloseOpenPeriod(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	// 插入一条未关闭的记录
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 0, 1000, 1000, 0)`)

	// 关闭
	if err := repo.CloseOpenPeriod(2000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证已关闭
	period, _ := repo.GetOpenPeriod()
	if period != nil {
		t.Error("expected no open period after close")
	}

	// 验证 end_time 已设置
	var closed model.BusinessStatusPeriod
	db.Where("uuid = 1001").First(&closed)
	if closed.EndTime != 2000 {
		t.Errorf("expected end_time 2000, got %d", closed.EndTime)
	}
}

func TestCloseOpenPeriod_OnlyClosesOpen(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	// 一条已关闭、一条未关闭
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 2000, 1000, 2000, 0)`)
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1002, 3000, 0, 3000, 3000, 0)`)

	if err := repo.CloseOpenPeriod(4000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 已关闭的记录不应被修改
	var first model.BusinessStatusPeriod
	db.Where("uuid = 1001").First(&first)
	if first.EndTime != 2000 {
		t.Errorf("closed period should remain unchanged, got end_time %d", first.EndTime)
	}

	// 未关闭的记录应被关闭
	var second model.BusinessStatusPeriod
	db.Where("uuid = 1002").First(&second)
	if second.EndTime != 4000 {
		t.Errorf("open period should be closed with end_time 4000, got %d", second.EndTime)
	}
}

// ─── GetBusinessStatus ──────────────────────────────────────────────

func TestGetBusinessStatus_Normal(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	status := repo.GetBusinessStatus()
	if status != constant.BusinessStatusNormal {
		t.Errorf("expected Normal(%d), got %d", constant.BusinessStatusNormal, status)
	}
}

func TestGetBusinessStatus_Test(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 0, 1000, 1000, 0)`)

	status := repo.GetBusinessStatus()
	if status != constant.BusinessStatusTest {
		t.Errorf("expected Test(%d), got %d", constant.BusinessStatusTest, status)
	}
}

func TestGetBusinessStatus_AllClosed(t *testing.T) {
	db := setupPeriodTestDB(t)
	repo := NewBusinessStatusPeriodRepo(db)

	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 2000, 1000, 2000, 0)`)

	status := repo.GetBusinessStatus()
	if status != constant.BusinessStatusNormal {
		t.Errorf("expected Normal(%d) when all periods closed, got %d", constant.BusinessStatusNormal, status)
	}
}
