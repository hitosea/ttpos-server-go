package service

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDeskMapService_GetAreaListWithStatus 测试获取区域列表及状态
//
// 任务: story-admin-desktop-table-map Phase 4.1
// 需求: R1.1-R1.6
//
// @version v2.10.0
func TestDeskMapService_GetAreaListWithStatus(t *testing.T) {
	dbm := setupTestDB(t)
	defer cleanupTestData(t, dbm)

	// 场景1: 无区域时返回空列表
	srv := NewDeskMapSrv(dbm)
	result, err := srv.GetAreaListWithStatus(1)
	if err != nil {
		t.Fatalf("GetAreaListWithStatus failed: %v", err)
	}
	if len(result.List) != 0 {
		t.Errorf("Expected empty list, got %d items", len(result.List))
	}

	// 场景2: 创建区域但无布局
	area1 := createTestArea(t, dbm, "测试区域1", 1)
	createTestDesks(t, dbm, area1.Uuid, 5)

	result, err = srv.GetAreaListWithStatus(1)
	if err != nil {
		t.Fatalf("GetAreaListWithStatus failed: %v", err)
	}
	if len(result.List) != 1 {
		t.Fatalf("Expected 1 area, got %d", len(result.List))
	}
	if result.List[0].AreaUuid != area1.Uuid {
		t.Errorf("Expected area_uuid %d, got %d", area1.Uuid, result.List[0].AreaUuid)
	}
	if result.List[0].AreaName != "测试区域1" {
		t.Errorf("Expected area_name '测试区域1', got '%s'", result.List[0].AreaName)
	}
	if result.List[0].DeskCount != 5 {
		t.Errorf("Expected desk_count 5, got %d", result.List[0].DeskCount)
	}
	if result.List[0].LayoutStatus != "unset" {
		t.Errorf("Expected layout_status 'unset', got '%s'", result.List[0].LayoutStatus)
	}

	// 场景3: 创建布局后状态应为 set
	createTestLayout(t, dbm, area1.Uuid)

	result, err = srv.GetAreaListWithStatus(1)
	if err != nil {
		t.Fatalf("GetAreaListWithStatus failed: %v", err)
	}
	if result.List[0].LayoutStatus != "set" {
		t.Errorf("Expected layout_status 'set', got '%s'", result.List[0].LayoutStatus)
	}

	// 场景4: 多个区域按 sort 排序
	area2 := createTestArea(t, dbm, "测试区域2", 0) // sort=0 应排在前面
	createTestDesks(t, dbm, area2.Uuid, 3)

	result, err = srv.GetAreaListWithStatus(1)
	if err != nil {
		t.Fatalf("GetAreaListWithStatus failed: %v", err)
	}
	if len(result.List) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.List))
	}
	// 验证排序: sort=0 的区域应该在前面
	if result.List[0].AreaName != "测试区域2" {
		t.Errorf("Expected first area '测试区域2', got '%s'", result.List[0].AreaName)
	}
	if result.List[1].AreaName != "测试区域1" {
		t.Errorf("Expected second area '测试区域1', got '%s'", result.List[1].AreaName)
	}
}

// TestDeskMapService_GetLayoutDetail 测试获取布局详情
//
// 任务: story-admin-desktop-table-map Phase 4.1
// 需求: R1.1-R1.6
//
// @version v2.10.0
func TestDeskMapService_GetLayoutDetail(t *testing.T) {
	dbm := setupTestDB(t)
	defer cleanupTestData(t, dbm)

	srv := NewDeskMapSrv(dbm)

	// 场景1: 区域不存在 - 应返回错误
	_, err := srv.GetLayoutDetail(1, 99999999)
	if err == nil {
		t.Error("Expected error for non-existent area, got nil")
	}

	// 场景2: 区域存在但无布局 - 应返回空布局
	area1 := createTestArea(t, dbm, "测试区域1", 1)
	desk1 := createTestDeskWithDetails(t, dbm, area1.Uuid, "A01", 1)
	desk2 := createTestDeskWithDetails(t, dbm, area1.Uuid, "A02", 2)

	result, err := srv.GetLayoutDetail(1, area1.Uuid)
	if err != nil {
		t.Fatalf("GetLayoutDetail failed: %v", err)
	}

	// 验证区域信息
	if result.Area.AreaUuid != area1.Uuid {
		t.Errorf("Expected area_uuid %d, got %d", area1.Uuid, result.Area.AreaUuid)
	}
	if result.Area.AreaName != "测试区域1" {
		t.Errorf("Expected area_name '测试区域1', got '%s'", result.Area.AreaName)
	}

	// 验证桌台列表
	if len(result.Desks) != 2 {
		t.Fatalf("Expected 2 desks, got %d", len(result.Desks))
	}
	if result.Desks[0].DeskUuid != desk1.Uuid {
		t.Errorf("Expected desk_uuid %d, got %d", desk1.Uuid, result.Desks[0].DeskUuid)
	}
	if result.Desks[0].DeskName != "A01" {
		t.Errorf("Expected desk_name 'A01', got '%s'", result.Desks[0].DeskName)
	}

	// 验证布局为空
	if result.Layout.Desks != nil && len(result.Layout.Desks) != 0 {
		t.Error("Expected empty layout, got desks")
	}

	// 场景3: 区域存在且有布局 - 应返回完整的布局数据
	layoutData := map[string]interface{}{
		"desks": []map[string]interface{}{
			{
				"desk_uuid": desk1.Uuid,
				"shape":     "circle",
				"capacity":  4,
				"x":         100.0,
				"y":         200.0,
				"width":     80.0,
				"height":    80.0,
				"rotation":  0.0,
			},
			{
				"desk_uuid": desk2.Uuid,
				"shape":     "rect",
				"capacity":  6,
				"x":         300.0,
				"y":         200.0,
				"width":     100.0,
				"height":    60.0,
				"rotation":  0.0,
			},
		},
	}
	createTestLayoutWithData(t, dbm, area1.Uuid, layoutData)

	result, err = srv.GetLayoutDetail(1, area1.Uuid)
	if err != nil {
		t.Fatalf("GetLayoutDetail failed: %v", err)
	}

	// 验证布局数据
	if result.Layout.Desks == nil || len(result.Layout.Desks) != 2 {
		t.Fatalf("Expected 2 desks in layout, got %v", result.Layout.Desks)
	}

	// 验证第一个桌台布局
	desk1Layout := result.Layout.Desks[0]
	if desk1Layout.DeskUuid != desk1.Uuid {
		t.Errorf("Expected desk_uuid %d, got %d", desk1.Uuid, desk1Layout.DeskUuid)
	}
	if desk1Layout.Shape != "circle" {
		t.Errorf("Expected shape 'circle', got '%s'", desk1Layout.Shape)
	}
	if desk1Layout.X != 100.0 {
		t.Errorf("Expected x 100.0, got %v", desk1Layout.X)
	}
	if desk1Layout.RangeMin != 4 {
		t.Errorf("Expected range_min 4, got %d", desk1Layout.RangeMin)
	}
	if desk1Layout.RangeMax != 4 {
		t.Errorf("Expected range_max 4, got %d", desk1Layout.RangeMax)
	}

	// 验证桌台的 selected 字段
	if !result.Desks[0].Selected {
		t.Error("Expected desk to be selected in layout")
	}
}

// TestDeskMapService_SaveLayout 测试保存布局
//
// 任务: story-admin-desktop-table-map Phase 4.1
// 需求: R1.1-R1.6
//
// @version v2.10.0
func TestDeskMapService_SaveLayout(t *testing.T) {
	dbm := setupTestDB(t)
	defer cleanupTestData(t, dbm)

	srv := NewDeskMapSrv(dbm)

	// 场景1: 区域不存在 - 应返回错误
	invalidReq := req.DeskMapSaveLayoutReq{
		AreaUuid: 99999999,
		Desks: []req.DeskMapLayoutDeskReq{
			{
				DeskUuid: 123,
				Shape:    "circle",
				RangeMin: 4,
				RangeMax: 4,
				X:        100,
				Y:        200,
			},
		},
	}

	err := srv.SaveLayout(1, invalidReq)
	if err == nil {
		t.Error("Expected error for non-existent area, got nil")
	}

	// 场景2: 首次保存（创建）- 应成功创建新布局
	area := createTestArea(t, dbm, "测试区域", 1)
	desk1 := createTestDeskWithDetails(t, dbm, area.Uuid, "A01", 1)
	desk2 := createTestDeskWithDetails(t, dbm, area.Uuid, "A02", 2)

	saveReq := req.DeskMapSaveLayoutReq{
		AreaUuid: area.Uuid,
		Desks: []req.DeskMapLayoutDeskReq{
			{
				DeskUuid: desk1.Uuid,
				Shape:    "circle",
				RangeMin: 4,
				RangeMax: 4,
				X:        100.0,
				Y:        200.0,
				Width:    80.0,
				Height:   80.0,
			},
		},
	}

	err = srv.SaveLayout(1, saveReq)
	if err != nil {
		t.Fatalf("SaveLayout (create) failed: %v", err)
	}

	// 验证布局已创建
	db := dbm.GetDB(1)
	var layout model.DeskMapLayout
	err = db.Where("area_uuid = ?", area.Uuid).First(&layout).Error
	if err != nil {
		t.Fatalf("Failed to find created layout: %v", err)
	}
	// 验证 JSON 格式正确
	var savedLayout resp.DeskMapLayoutData
	err = json.Unmarshal([]byte(layout.LayoutJson), &savedLayout)
	if err != nil {
		t.Errorf("Failed to unmarshal saved layout: %v", err)
	}
	if len(savedLayout.Desks) != 1 {
		t.Errorf("Expected 1 desk in saved layout, got %d", len(savedLayout.Desks))
	}

	// 场景3: 再次保存（更新）- 应成功更新现有布局
	saveReq2 := req.DeskMapSaveLayoutReq{
		AreaUuid: area.Uuid,
		Desks: []req.DeskMapLayoutDeskReq{
			{
				DeskUuid: desk1.Uuid,
				Shape:    "circle",
				RangeMin: 4,
				RangeMax: 4,
				X:        100.0,
				Y:        200.0,
				Width:    80.0,
				Height:   80.0,
			},
			{
				DeskUuid: desk2.Uuid,
				Shape:    "rect",
				RangeMin: 6,
				RangeMax: 6,
				X:        300.0,
				Y:        400.0,
				Width:    100.0,
				Height:   60.0,
			},
		},
	}

	err = srv.SaveLayout(1, saveReq2)
	if err != nil {
		t.Fatalf("SaveLayout (update) failed: %v", err)
	}

	// 验证布局已更新
	var updatedLayout model.DeskMapLayout
	err = db.Where("area_uuid = ?", area.Uuid).First(&updatedLayout).Error
	if err != nil {
		t.Fatalf("Failed to find updated layout: %v", err)
	}
	// 验证更新后的 JSON 格式
	var updatedLayoutData resp.DeskMapLayoutData
	err = json.Unmarshal([]byte(updatedLayout.LayoutJson), &updatedLayoutData)
	if err != nil {
		t.Errorf("Failed to unmarshal updated layout: %v", err)
	}
	if len(updatedLayoutData.Desks) != 2 {
		t.Errorf("Expected 2 desks in updated layout, got %d", len(updatedLayoutData.Desks))
	}

	// 场景4: 验证错误 - 形状错误应返回错误
	invalidShapeReq := req.DeskMapSaveLayoutReq{
		AreaUuid: area.Uuid,
		Desks: []req.DeskMapLayoutDeskReq{
			{
				DeskUuid: desk1.Uuid,
				Shape:    "invalid",
				RangeMin: 4,
				RangeMax: 4,
			},
		},
	}

	err = srv.SaveLayout(1, invalidShapeReq)
	if err == nil {
		t.Error("Expected error for invalid shape, got nil")
	}

	// 场景5: 空布局 - 应返回错误（布局数据不能为空）
	emptyReq := req.DeskMapSaveLayoutReq{
		AreaUuid: area.Uuid,
		Desks:    []req.DeskMapLayoutDeskReq{},
	}

	err = srv.SaveLayout(1, emptyReq)
	if err == nil {
		t.Error("Expected error for empty desks, got nil")
	}
}

// 测试辅助函数

// setupTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupTestDB(t *testing.T) *database.DBManager {
	t.Helper()

	// 创建 SQLite 内存数据库，禁用外键约束以简化测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 手动创建表结构（SQLite 兼容版本）
	// 由于 model 中使用了 MySQL 特定的类型（如 unsigned），我们需要手动创建表
	err = db.Exec(`
		CREATE TABLE ttpos_desk_region (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			sort INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create desk_region table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE ttpos_desk (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			desk_no TEXT DEFAULT '',
			region_uuid INTEGER DEFAULT 0,
			type_uuid INTEGER DEFAULT 0,
			sort INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			is_disable INTEGER DEFAULT 0,
			qrcode_token TEXT DEFAULT '',
			sale_bill_uuid INTEGER DEFAULT 0,
			device_uuid INTEGER DEFAULT 0,
			default_people_num INTEGER DEFAULT 0,
			is_open_default_people_num INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create desk table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE ttpos_desk_map_layout (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			area_uuid INTEGER DEFAULT 0,
			layout_json TEXT DEFAULT ''
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create desk_map_layout table: %v", err)
	}

	// 创建唯一索引
	err = db.Exec(`CREATE UNIQUE INDEX uk_area_uuid ON ttpos_desk_map_layout(area_uuid)`).Error
	if err != nil {
		t.Fatalf("Failed to create unique index: %v", err)
	}

	// 创建 DBManager 并设置 Mock DB
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 使用反射初始化私有字段（仅用于测试）
	// 需要初始化 lock, lastCheck, checkInterval 字段
	dbmValue := reflect.ValueOf(dbm).Elem()

	// 设置 lock 字段
	lockField := dbmValue.FieldByName("lock")
	if lockField.IsValid() && lockField.CanSet() {
		lockField.Set(reflect.ValueOf(&sync.Mutex{}))
	}

	// 设置 lastCheck 字段
	lastCheckField := dbmValue.FieldByName("lastCheck")
	if lastCheckField.IsValid() && lastCheckField.CanSet() {
		lastCheckField.Set(reflect.ValueOf(make(map[uint64]time.Time)))
	}

	// 设置 checkInterval 字段
	checkIntervalField := dbmValue.FieldByName("checkInterval")
	if checkIntervalField.IsValid() && checkIntervalField.CanSet() {
		checkIntervalField.Set(reflect.ValueOf(10 * time.Second))
	}

	return dbm
}

// cleanupTestData 清理测试数据
func cleanupTestData(t *testing.T, dbm *database.DBManager) {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	if db == nil {
		return
	}

	// 清理所有测试数据
	db.Exec("DELETE FROM ttpos_desk_map_layout")
	db.Exec("DELETE FROM ttpos_desk")
	db.Exec("DELETE FROM ttpos_desk_region")
}

// createTestArea 创建测试区域
func createTestArea(t *testing.T, dbm *database.DBManager, name string, sort uint) *model.DeskRegion {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	area := &model.DeskRegion{
		Name: name,
		Sort: sort,
	}
	if err := db.Create(area).Error; err != nil {
		t.Fatalf("Failed to create test area: %v", err)
	}
	return area
}

// createTestDesks 批量创建测试桌台
func createTestDesks(t *testing.T, dbm *database.DBManager, areaUuid uint64, count int) {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	for i := 1; i <= count; i++ {
		desk := &model.Desk{
			DeskNo:     string(rune('A' + i - 1)),
			RegionUuid: areaUuid,
			Sort:       uint(i),
		}
		if err := db.Create(desk).Error; err != nil {
			t.Fatalf("Failed to create test desk: %v", err)
		}
	}
}

// createTestDeskWithDetails 创建带详细信息的测试桌台
func createTestDeskWithDetails(t *testing.T, dbm *database.DBManager, areaUuid uint64, deskNo string, sort uint) *model.Desk {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	desk := &model.Desk{
		DeskNo:     deskNo,
		RegionUuid: areaUuid,
		Sort:       sort,
		Status:     constant.DeskStatusClose,
		IsDisable:  0,
	}
	if err := db.Create(desk).Error; err != nil {
		t.Fatalf("Failed to create test desk: %v", err)
	}
	return desk
}

// createTestLayout 创建简单的测试布局
func createTestLayout(t *testing.T, dbm *database.DBManager, areaUuid uint64) *model.DeskMapLayout {
	t.Helper()

	layoutData := map[string]interface{}{
		"desks": []map[string]interface{}{
			{
				"desk_uuid": 123,
				"shape":     "circle",
				"x":         100,
				"y":         200,
			},
		},
	}
	return createTestLayoutWithData(t, dbm, areaUuid, layoutData)
}

// createTestLayoutWithData 创建带指定数据的测试布局
func createTestLayoutWithData(t *testing.T, dbm *database.DBManager, areaUuid uint64, layoutData map[string]interface{}) *model.DeskMapLayout {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	layoutJSON, err := json.Marshal(layoutData)
	if err != nil {
		t.Fatalf("Failed to marshal layout data: %v", err)
	}

	layout := &model.DeskMapLayout{
		AreaUuid:   areaUuid,
		LayoutJson: string(layoutJSON),
	}
	if err := db.Create(layout).Error; err != nil {
		t.Fatalf("Failed to create test layout: %v", err)
	}
	return layout
}
