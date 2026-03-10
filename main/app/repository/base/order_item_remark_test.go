//go:build integration

package base

import (
	"testing"
	"time"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupOrderItemRemarkTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 创建表结构
	err = db.Exec(`
		CREATE TABLE multi_language_names (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			zh_name TEXT DEFAULT '',
			th_name TEXT DEFAULT '',
			en_name TEXT DEFAULT '',
			zh_tw_name TEXT DEFAULT '',
			ja_name TEXT DEFAULT '',
			ko_name TEXT DEFAULT '',
			my_name TEXT DEFAULT '',
			tr_name TEXT DEFAULT '',
			sv_name TEXT DEFAULT ''
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create multi_language_name table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE ttpos_order_item_remark (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			multi_language_name_uuid INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create order_item_remark table: %v", err)
	}

	// 创建唯一索引
	err = db.Exec(`CREATE UNIQUE INDEX uk_uuid ON ttpos_order_item_remark(uuid)`).Error
	if err != nil {
		t.Fatalf("Failed to create unique index: %v", err)
	}

	return db
}

// TestOrderItemRemarkRepo_CreateOrderItemRemark 测试创建单品备注原因
func TestOrderItemRemarkRepo_CreateOrderItemRemark(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建单品备注原因（CreateOrderItemRemark 会自动创建 MultiLanguageName）
	remark := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name: "不要香菜",
		MultiLanguageName: model.MultiLanguageName{
			BaseModel: model.BaseModel{
				Uuid:       1001,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			ZhName:   "不要香菜",
			ThName:   "ไม่ใส่ผักชี",
			EnName:   "No Cilantro",
			ZhTwName: "不要香菜",
		},
	}

	uuid, err := repo.CreateOrderItemRemark(remark)
	if err != nil {
		t.Fatalf("Failed to create order item remark: %v", err)
	}

	if uuid != remark.Uuid {
		t.Errorf("Expected uuid %d, got %d", remark.Uuid, uuid)
	}

	// 验证数据已创建
	var count int64
	db.Model(&model.OrderItemRemark{}).Where("uuid = ?", uuid).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestOrderItemRemarkRepo_GetOrderItemRemarkList 测试获取单品备注原因列表
func TestOrderItemRemarkRepo_GetOrderItemRemarkList(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName1 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
		EnName: "No Cilantro",
	}
	db.Create(&multiLangName1)

	multiLangName2 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要葱",
		EnName: "No Scallion",
	}
	db.Create(&multiLangName2)

	remark1 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName1.Uuid,
	}
	db.Create(&remark1)

	remark2 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要葱",
		MultiLanguageNameUuid: multiLangName2.Uuid,
	}
	db.Create(&remark2)

	// 测试获取列表
	list, err := repo.GetOrderItemRemarkList()
	if err != nil {
		t.Fatalf("Failed to get order item remark list: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 records, got %d", len(list))
	}

	// 验证多语言名称已预加载
	if list[0].MultiLanguageName.Uuid == 0 {
		t.Error("MultiLanguageName should be preloaded")
	}
}

// TestOrderItemRemarkRepo_GetOrderItemRemarkByUuid 测试根据UUID获取单品备注原因
func TestOrderItemRemarkRepo_GetOrderItemRemarkByUuid(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
		EnName: "No Cilantro",
	}
	db.Create(&multiLangName)

	remark := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName.Uuid,
	}
	db.Create(&remark)

	// 测试获取
	result, err := repo.GetOrderItemRemarkByUuid(remark.Uuid)
	if err != nil {
		t.Fatalf("Failed to get order item remark: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Uuid != remark.Uuid {
		t.Errorf("Expected uuid %d, got %d", remark.Uuid, result.Uuid)
	}

	if result.MultiLanguageName.Uuid == 0 {
		t.Error("MultiLanguageName should be preloaded")
	}
}

// TestOrderItemRemarkRepo_UpdateOrderItemRemark 测试更新单品备注原因
func TestOrderItemRemarkRepo_UpdateOrderItemRemark(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
		EnName: "No Cilantro",
	}
	db.Create(&multiLangName)

	remark := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName.Uuid,
	}
	db.Create(&remark)

	// 更新数据
	remark.Name = "不要葱"
	remark.MultiLanguageName.ZhName = "不要葱"
	remark.MultiLanguageName.EnName = "No Scallion"
	remark.MultiLanguageName.Uuid = multiLangName.Uuid

	err := repo.UpdateOrderItemRemark(remark.Uuid, remark)
	if err != nil {
		t.Fatalf("Failed to update order item remark: %v", err)
	}

	// 验证更新
	var updated model.OrderItemRemark
	db.Where("uuid = ?", remark.Uuid).First(&updated)
	if updated.Name != "不要葱" {
		t.Errorf("Expected name '不要葱', got '%s'", updated.Name)
	}
}

// TestOrderItemRemarkRepo_DeleteOrderItemRemark 测试软删除单品备注原因
func TestOrderItemRemarkRepo_DeleteOrderItemRemark(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
		EnName: "No Cilantro",
	}
	db.Create(&multiLangName)

	remark := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName.Uuid,
	}
	db.Create(&remark)

	// 删除
	err := repo.DeleteOrderItemRemark(remark.Uuid)
	if err != nil {
		t.Fatalf("Failed to delete order item remark: %v", err)
	}

	// 验证软删除（delete_time 不为 0）
	var deleted model.OrderItemRemark
	db.Where("uuid = ?", remark.Uuid).First(&deleted)
	if deleted.DeleteTime == 0 {
		t.Error("Expected delete_time to be set, got 0")
	}

	// 验证 GetOrderItemRemarkList 不返回已删除的记录
	list, err := repo.GetOrderItemRemarkList()
	if err != nil {
		t.Fatalf("Failed to get order item remark list: %v", err)
	}

	for _, item := range list {
		if item.Uuid == remark.Uuid {
			t.Error("Deleted remark should not appear in list")
		}
	}
}

// TestOrderItemRemarkRepo_CountOrderItemRemark 测试统计单品备注原因数量
func TestOrderItemRemarkRepo_CountOrderItemRemark(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName1 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
	}
	db.Create(&multiLangName1)

	multiLangName2 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要葱",
	}
	db.Create(&multiLangName2)

	remark1 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName1.Uuid,
	}
	db.Create(&remark1)

	remark2 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
			DeleteTime: time.Now().Unix(), // 已删除
		},
		Name:                  "不要葱",
		MultiLanguageNameUuid: multiLangName2.Uuid,
	}
	db.Create(&remark2)

	// 测试统计（应该只统计未删除的）
	count, err := repo.CountOrderItemRemark()
	if err != nil {
		t.Fatalf("Failed to count order item remark: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

// TestOrderItemRemarkRepo_GetOrderItemRemarks 测试使用选项模式查询
func TestOrderItemRemarkRepo_GetOrderItemRemarks(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName1 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
	}
	db.Create(&multiLangName1)

	multiLangName2 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要葱",
	}
	db.Create(&multiLangName2)

	remark1 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName1.Uuid,
	}
	db.Create(&remark1)

	remark2 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要葱",
		MultiLanguageNameUuid: multiLangName2.Uuid,
	}
	db.Create(&remark2)

	// 测试使用选项模式查询
	remarks, err := repo.GetOrderItemRemarks(
		repository.CommonRepo.WhereInUuids([]uint64{remark1.Uuid, remark2.Uuid}),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		t.Fatalf("Failed to get order item remarks: %v", err)
	}

	if len(remarks) != 2 {
		t.Errorf("Expected 2 records, got %d", len(remarks))
	}
}

// TestOrderItemRemarkRepo_GetOrderItemRemarkListByUuids 测试根据UUID数组获取列表
func TestOrderItemRemarkRepo_GetOrderItemRemarkListByUuids(t *testing.T) {
	db := setupOrderItemRemarkTestDB(t)
	repo := NewOrderItemRemarkRepo(db)

	// 创建测试数据
	multiLangName1 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要香菜",
	}
	db.Create(&multiLangName1)

	multiLangName2 := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ZhName: "不要葱",
	}
	db.Create(&multiLangName2)

	remark1 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要香菜",
		MultiLanguageNameUuid: multiLangName1.Uuid,
	}
	db.Create(&remark1)

	remark2 := model.OrderItemRemark{
		BaseModel: model.BaseModel{
			Uuid:       2002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  "不要葱",
		MultiLanguageNameUuid: multiLangName2.Uuid,
	}
	db.Create(&remark2)

	// 测试根据UUID数组获取
	remarks, err := repo.GetOrderItemRemarkListByUuids([]uint64{remark1.Uuid, remark2.Uuid})
	if err != nil {
		t.Fatalf("Failed to get order item remark list by uuids: %v", err)
	}

	if len(remarks) != 2 {
		t.Errorf("Expected 2 records, got %d", len(remarks))
	}

	// 验证多语言名称已预加载
	for _, remark := range remarks {
		if remark.MultiLanguageName.Uuid == 0 {
			t.Error("MultiLanguageName should be preloaded")
		}
	}
}
