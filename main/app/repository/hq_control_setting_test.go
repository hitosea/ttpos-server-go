package repository

import (
	"sync"
	"testing"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var controlSettingTestOnce sync.Once

func setupControlSettingTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	controlSettingTestOnce.Do(func() {
		utils.InitIdGenerator()
		if logger.Logger == nil {
			logger.Logger, _ = zap.NewDevelopment()
		}
	})

	config.Database.TablePrefix = "ttpos_"

	// 每次创建新缓存实例，隔离测试
	cache.Global = cache.NewCache(cache.GoCache, cache.Config{})
	t.Cleanup(func() { cache.Global = nil })

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
		CREATE TABLE ttpos_hq_control_setting (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			field_type TEXT DEFAULT '' UNIQUE,
			control_mode INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestHqControlSetting_GetControlMap_EmptyDB(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	result := repo.GetControlMap(100)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestHqControlSetting_Upsert_Insert(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	err := repo.Upsert("dine_shelf", 1)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	var count int64
	db.Raw("SELECT COUNT(*) FROM ttpos_hq_control_setting WHERE field_type = 'dine_shelf'").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestHqControlSetting_Upsert_Update(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	_ = repo.Upsert("dine_shelf", 0)
	_ = repo.Upsert("dine_shelf", 1)

	var count int64
	db.Raw("SELECT COUNT(*) FROM ttpos_hq_control_setting WHERE field_type = 'dine_shelf'").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row after upsert, got %d", count)
	}

	// 验证值已更新 — 失效缓存后重新读取
	repo.InvalidateCache(100)
	controlMap := repo.GetControlMap(100)
	if controlMap["dine_shelf"] != 1 {
		t.Errorf("expected control_mode=1, got %d", controlMap["dine_shelf"])
	}
}

func TestHqControlSetting_GetControlMap_ReturnsAll(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	_ = repo.Upsert("dine_shelf", 1)
	_ = repo.Upsert("takeout_shelf", 0)
	_ = repo.Upsert("negative_stock", 1)

	repo.InvalidateCache(100)
	result := repo.GetControlMap(100)

	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}
	if result["dine_shelf"] != 1 {
		t.Errorf("dine_shelf: expected 1, got %d", result["dine_shelf"])
	}
	if result["takeout_shelf"] != 0 {
		t.Errorf("takeout_shelf: expected 0, got %d", result["takeout_shelf"])
	}
	if result["negative_stock"] != 1 {
		t.Errorf("negative_stock: expected 1, got %d", result["negative_stock"])
	}
}

func TestHqControlSetting_IsUnifiedControl_True(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	_ = repo.Upsert("dine_shelf", 1)
	repo.InvalidateCache(100)

	if !repo.IsUnifiedControl(100, "dine_shelf") {
		t.Error("expected IsUnifiedControl to return true for mode=1")
	}
}

func TestHqControlSetting_IsUnifiedControl_False(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	_ = repo.Upsert("dine_shelf", 0)
	repo.InvalidateCache(100)

	if repo.IsUnifiedControl(100, "dine_shelf") {
		t.Error("expected IsUnifiedControl to return false for mode=0")
	}
}

// P0: 默认值是核心业务规则 — negative_stock 默认统一，其他默认分开
func TestHqControlSetting_IsUnifiedControl_DefaultValues(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	tests := []struct {
		fieldType string
		expected  bool
	}{
		{constant.HqFieldDineShelf, false},
		{constant.HqFieldTakeoutShelf, false},
		{constant.HqFieldTakeoutPrice, false},
		{constant.HqFieldNegativeStock, true}, // 唯一默认统一的字段
	}

	for _, tt := range tests {
		t.Run(tt.fieldType, func(t *testing.T) {
			result := repo.IsUnifiedControl(100, tt.fieldType)
			if result != tt.expected {
				t.Errorf("IsUnifiedControl(%q): expected %v, got %v", tt.fieldType, tt.expected, result)
			}
		})
	}
}

func TestHqControlSetting_CacheHitAfterGet(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	_ = repo.Upsert("dine_shelf", 1)
	repo.InvalidateCache(100)

	// 第一次调用，填充缓存
	m1 := repo.GetControlMap(100)
	if m1["dine_shelf"] != 1 {
		t.Fatalf("initial get: expected 1, got %d", m1["dine_shelf"])
	}

	// 直接修改 DB（绕过 repo）
	db.Exec("UPDATE ttpos_hq_control_setting SET control_mode = 0 WHERE field_type = 'dine_shelf'")

	// 第二次调用应该读缓存，还是旧值
	m2 := repo.GetControlMap(100)
	if m2["dine_shelf"] != 1 {
		t.Errorf("cache hit: expected 1 (cached), got %d", m2["dine_shelf"])
	}
}

func TestHqControlSetting_InvalidateCache(t *testing.T) {
	db := setupControlSettingTestDB(t)
	repo := NewHqControlSettingRepo(db)

	_ = repo.Upsert("dine_shelf", 1)
	repo.InvalidateCache(100)

	// 填充缓存
	_ = repo.GetControlMap(100)

	// 直接修改 DB
	db.Exec("UPDATE ttpos_hq_control_setting SET control_mode = 0 WHERE field_type = 'dine_shelf'")

	// 失效缓存后读到新值
	repo.InvalidateCache(100)
	m := repo.GetControlMap(100)
	if m["dine_shelf"] != 0 {
		t.Errorf("after invalidation: expected 0 (fresh from DB), got %d", m["dine_shelf"])
	}
}

func TestHqControlSetting_GetHqControlMap_NilDB(t *testing.T) {
	_ = setupControlSettingTestDB(t) // 初始化 cache.Global

	// 清空缓存确保 miss
	cache.Global.Del("hq_control_setting:999")

	result := GetHqControlMap(nil, 999)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map for nil DB + cache miss, got %d entries", len(result))
	}
}
