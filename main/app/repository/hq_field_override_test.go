package repository

import (
	"sync"
	"testing"

	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var overrideTestOnce sync.Once

func setupOverrideTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	overrideTestOnce.Do(func() {
		utils.InitIdGenerator()
		if logger.Logger == nil {
			logger.Logger, _ = zap.NewDevelopment()
		}
	})

	config.Database.TablePrefix = "ttpos_"

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE ttpos_hq_field_override (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			entity_type TEXT DEFAULT '',
			entity_uuid INTEGER DEFAULT 0,
			field_type TEXT DEFAULT ''
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestHqFieldOverride_MarkOverridden_Creates(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	err := repo.MarkOverridden(1001, "product", "dine_shelf")
	if err != nil {
		t.Fatalf("MarkOverridden failed: %v", err)
	}

	var count int64
	db.Model(&model.HqFieldOverride{}).Where("entity_uuid = 1001 AND field_type = 'dine_shelf' AND delete_time = 0").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestHqFieldOverride_MarkOverridden_Idempotent(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	_ = repo.MarkOverridden(1001, "product", "dine_shelf")

	var count int64
	db.Model(&model.HqFieldOverride{}).Where("entity_uuid = 1001 AND field_type = 'dine_shelf' AND delete_time = 0").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row (idempotent), got %d", count)
	}
}

func TestHqFieldOverride_IsOverridden_Lifecycle(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	// initially false
	if repo.IsOverridden(1001, "dine_shelf") {
		t.Error("expected false before marking")
	}

	// mark → true
	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	if !repo.IsOverridden(1001, "dine_shelf") {
		t.Error("expected true after marking")
	}

	// clear → false
	_ = repo.ClearOverride(1001, "dine_shelf")
	if repo.IsOverridden(1001, "dine_shelf") {
		t.Error("expected false after clearing")
	}
}

func TestHqFieldOverride_ClearOverride_SoftDelete(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	_ = repo.ClearOverride(1001, "dine_shelf")

	var override model.HqFieldOverride
	db.Where("entity_uuid = 1001 AND field_type = 'dine_shelf'").First(&override)
	if override.DeleteTime == 0 {
		t.Error("expected delete_time > 0 after soft delete")
	}
}

func TestHqFieldOverride_ClearOverride_NotAffectOtherFields(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	_ = repo.MarkOverridden(1001, "product_takeout", "takeout_shelf")

	_ = repo.ClearOverride(1001, "dine_shelf")

	if repo.IsOverridden(1001, "dine_shelf") {
		t.Error("dine_shelf should be cleared")
	}
	if !repo.IsOverridden(1001, "takeout_shelf") {
		t.Error("takeout_shelf should still exist")
	}
}

func TestHqFieldOverride_ClearOverridesByEntityUuids_Batch(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	_ = repo.MarkOverridden(1002, "product", "dine_shelf")
	_ = repo.MarkOverridden(1003, "product", "dine_shelf")

	err := repo.ClearOverridesByEntityUuids([]uint64{1001, 1002}, "dine_shelf")
	if err != nil {
		t.Fatalf("ClearOverridesByEntityUuids failed: %v", err)
	}

	if repo.IsOverridden(1001, "dine_shelf") {
		t.Error("1001 should be cleared")
	}
	if repo.IsOverridden(1002, "dine_shelf") {
		t.Error("1002 should be cleared")
	}
	if !repo.IsOverridden(1003, "dine_shelf") {
		t.Error("1003 should still exist")
	}
}

func TestHqFieldOverride_ClearOverridesByEntityUuids_EmptySlice(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	err := repo.ClearOverridesByEntityUuids([]uint64{}, "dine_shelf")
	if err != nil {
		t.Errorf("expected nil error for empty slice, got %v", err)
	}
}

func TestHqFieldOverride_BatchCheckOverridden_Map(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	_ = repo.MarkOverridden(1003, "product", "dine_shelf")

	result, err := repo.BatchCheckOverridden([]uint64{1001, 1002, 1003}, "dine_shelf")
	if err != nil {
		t.Fatalf("BatchCheckOverridden failed: %v", err)
	}

	if !result[1001] {
		t.Error("expected 1001 to be overridden")
	}
	if result[1002] {
		t.Error("expected 1002 to NOT be overridden")
	}
	if !result[1003] {
		t.Error("expected 1003 to be overridden")
	}
}

func TestHqFieldOverride_BatchCheckOverridden_IgnoresSoftDeleted(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	_ = repo.MarkOverridden(1001, "product", "dine_shelf")
	_ = repo.ClearOverride(1001, "dine_shelf")

	result, err := repo.BatchCheckOverridden([]uint64{1001}, "dine_shelf")
	if err != nil {
		t.Fatalf("BatchCheckOverridden failed: %v", err)
	}
	if result[1001] {
		t.Error("expected soft-deleted override to be ignored")
	}
}

func TestHqFieldOverride_BatchCheckOverridden_EmptySlice(t *testing.T) {
	db := setupOverrideTestDB(t)
	repo := NewHqFieldOverrideRepo(db)

	result, err := repo.BatchCheckOverridden([]uint64{}, "dine_shelf")
	if err != nil {
		t.Fatalf("BatchCheckOverridden failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}
