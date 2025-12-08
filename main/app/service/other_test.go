package service

import (
	stdcontext "context"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	// 初始化 ID 生成器用于测试
	utils.InitIdGenerator()
}

// setupOrderItemRemarkTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupOrderItemRemarkTestDB(t *testing.T) *database.DBManager {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 创建表结构（GORM 使用复数形式）
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

	// Setting 表没有 TableName 方法，GORM 使用复数形式
	err = db.Exec(`
		CREATE TABLE settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			key TEXT DEFAULT '',
			"values" TEXT DEFAULT '{}',
			describe TEXT DEFAULT ''
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create setting table: %v", err)
	}

	// 创建唯一索引
	err = db.Exec(`CREATE UNIQUE INDEX uk_uuid ON ttpos_order_item_remark(uuid)`).Error
	if err != nil {
		t.Fatalf("Failed to create unique index: %v", err)
	}

	// 创建 DBManager 并设置 Mock DB
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 使用反射初始化私有字段（仅用于测试）
	dbmValue := reflect.ValueOf(dbm).Elem()

	// 设置 lock 字段
	lockField := dbmValue.FieldByName("lock")
	if lockField.IsValid() {
		reflect.NewAt(lockField.Type(), unsafe.Pointer(lockField.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(&sync.Mutex{}))
	}

	// 设置 lastCheck 字段
	lastCheckField := dbmValue.FieldByName("lastCheck")
	if lastCheckField.IsValid() {
		reflect.NewAt(lastCheckField.Type(), unsafe.Pointer(lastCheckField.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(make(map[uint64]time.Time)))
	}

	// 设置 checkInterval 字段
	checkIntervalField := dbmValue.FieldByName("checkInterval")
	if checkIntervalField.IsValid() {
		reflect.NewAt(checkIntervalField.Type(), unsafe.Pointer(checkIntervalField.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(10 * time.Second))
	}

	return dbm
}

// createOrderItemRemarkTestContext 创建测试上下文
func createOrderItemRemarkTestContext(t *testing.T, dbm *database.DBManager) context.Context {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	// 创建测试用的 logger（如果 logger.Logger 为 nil）
	testLogger := logger.Logger
	if testLogger == nil {
		testLogger, _ = zap.NewDevelopment()
	}
	ctx := context.NewContext(
		context.WithContext(stdcontext.Background()),
		context.WithLogger(testLogger),
	)
	ctx.SetDB(db)
	// 使用 MockDB 作为 CompanyUuid，确保 GetDB 返回 MockDB
	ctx.SetCompanyUuid(constant.MockDB)
	return ctx
}

// TestOtherSrv_GetOrderItemRemarkList 测试获取单品备注原因列表
func TestOtherSrv_GetOrderItemRemarkList(t *testing.T) {
	dbm := setupOrderItemRemarkTestDB(t)
	ctx := createOrderItemRemarkTestContext(t, dbm)

	// 创建测试数据
	db := dbm.GetDB(constant.MockDB)
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

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	otherSrv := NewOtherSrv(dbm, mockCache, settingSrv)

	// 测试获取列表
	result, err := otherSrv.GetOrderItemRemarkList(ctx)
	if err != nil {
		t.Fatalf("Failed to get order item remark list: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.List) != 2 {
		t.Errorf("Expected 2 records, got %d", len(result.List))
	}

	// 验证数据
	if result.List[0].Uuid != remark1.Uuid && result.List[0].Uuid != remark2.Uuid {
		t.Errorf("Unexpected uuid: %d", result.List[0].Uuid)
	}
}

// TestOtherSrv_AddOrderItemRemark 测试新增单品备注原因
func TestOtherSrv_AddOrderItemRemark(t *testing.T) {
	dbm := setupOrderItemRemarkTestDB(t)
	ctx := createOrderItemRemarkTestContext(t, dbm)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	otherSrv := NewOtherSrv(dbm, mockCache, settingSrv)

	// 设置门店语言（zh, en）
	db := dbm.GetDB(constant.MockDB)
	setting := model.Setting{
		Key:        constant.SettingStore,
		Values:     `{"language":[{"key":1,"name":"zh"},{"key":2,"name":"en"}]}`,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}
	db.Create(&setting)

	// 测试新增
	addReq := req.AddOrderItemRemarkReq{
		LocaleName: dto.LocaleResponse{
			ZH: "不要香菜",
			EN: "No Cilantro",
		},
	}

	err := otherSrv.AddOrderItemRemark(ctx, addReq)
	if err != nil {
		t.Fatalf("Failed to add order item remark: %v", err)
	}

	// 验证数据已创建
	var count int64
	db.Model(&model.OrderItemRemark{}).Where("delete_time = 0").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestOtherSrv_AddOrderItemRemark_Validation 测试新增时的验证逻辑
func TestOtherSrv_AddOrderItemRemark_Validation(t *testing.T) {
	dbm := setupOrderItemRemarkTestDB(t)
	ctx := createOrderItemRemarkTestContext(t, dbm)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	otherSrv := NewOtherSrv(dbm, mockCache, settingSrv)

	// 设置门店语言（zh, en）
	db := dbm.GetDB(constant.MockDB)
	setting := model.Setting{
		Key:        constant.SettingStore,
		Values:     `{"language":[{"key":1,"name":"zh"},{"key":2,"name":"en"}]}`,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}
	db.Create(&setting)

	// 测试1: 多语言名称不完整（缺少 en）
	req1 := req.AddOrderItemRemarkReq{
		LocaleName: dto.LocaleResponse{
			ZH: "不要香菜",
			// EN 缺失
		},
	}

	err := otherSrv.AddOrderItemRemark(ctx, req1)
	if err == nil {
		t.Error("Expected error for incomplete locale name, got nil")
	}

	// 测试2: 字数超过100字
	longText := ""
	for i := 0; i < 101; i++ {
		longText += "字"
	}
	req2 := req.AddOrderItemRemarkReq{
		LocaleName: dto.LocaleResponse{
			ZH: longText,
			EN: "Long Text",
		},
	}

	err = otherSrv.AddOrderItemRemark(ctx, req2)
	if err == nil {
		t.Error("Expected error for text length exceeding 100 characters, got nil")
	}

	// 测试3: 数量限制（超过100个）
	// 创建100个备注
	for i := 0; i < 100; i++ {
		multiLangName := model.MultiLanguageName{
			BaseModel: model.BaseModel{
				Uuid:       uint64(10000 + i),
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			ZhName: "测试",
			EnName: "Test",
		}
		db.Create(&multiLangName)

		remark := model.OrderItemRemark{
			BaseModel: model.BaseModel{
				Uuid:       uint64(20000 + i),
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:                  "测试",
			MultiLanguageNameUuid: multiLangName.Uuid,
		}
		db.Create(&remark)
	}

	req3 := req.AddOrderItemRemarkReq{
		LocaleName: dto.LocaleResponse{
			ZH: "新备注",
			EN: "New Remark",
		},
	}

	err = otherSrv.AddOrderItemRemark(ctx, req3)
	if err == nil {
		t.Error("Expected error for exceeding 100 items limit, got nil")
	}
}

// TestOtherSrv_EditOrderItemRemark 测试编辑单品备注原因
func TestOtherSrv_EditOrderItemRemark(t *testing.T) {
	dbm := setupOrderItemRemarkTestDB(t)
	ctx := createOrderItemRemarkTestContext(t, dbm)

	// 创建测试数据
	db := dbm.GetDB(constant.MockDB)
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

	// 设置门店语言
	settingModel := model.Setting{
		Key:        constant.SettingStore,
		Values:     `{"language":[{"key":1,"name":"zh"},{"key":2,"name":"en"}]}`,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}
	db.Create(&settingModel)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	otherSrv := NewOtherSrv(dbm, mockCache, settingSrv)

	// 测试编辑
	editReq := req.EditOrderItemRemarkReq{
		Uuid: remark.Uuid,
		LocaleName: dto.LocaleResponse{
			ZH: "不要葱",
			EN: "No Scallion",
		},
	}

	err := otherSrv.EditOrderItemRemark(ctx, editReq)
	if err != nil {
		t.Fatalf("Failed to edit order item remark: %v", err)
	}

	// 验证更新
	var updated model.OrderItemRemark
	db.Where("uuid = ?", remark.Uuid).First(&updated)
	if updated.Name != "不要葱" {
		t.Errorf("Expected name '不要葱', got '%s'", updated.Name)
	}
}

// TestOtherSrv_EditOrderItemRemark_NotFound 测试编辑不存在的备注
func TestOtherSrv_EditOrderItemRemark_NotFound(t *testing.T) {
	dbm := setupOrderItemRemarkTestDB(t)
	ctx := createOrderItemRemarkTestContext(t, dbm)

	// 设置门店语言
	db := dbm.GetDB(constant.MockDB)
	settingModel := model.Setting{
		Key:        constant.SettingStore,
		Values:     `{"language":[{"key":1,"name":"zh"},{"key":2,"name":"en"}]}`,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}
	db.Create(&settingModel)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	otherSrv := NewOtherSrv(dbm, mockCache, settingSrv)

	// 测试编辑不存在的备注
	editReq := req.EditOrderItemRemarkReq{
		Uuid: 99999, // 不存在的UUID
		LocaleName: dto.LocaleResponse{
			ZH: "不要葱",
			EN: "No Scallion",
		},
	}

	err := otherSrv.EditOrderItemRemark(ctx, editReq)
	if err == nil {
		t.Error("Expected error for non-existent remark, got nil")
	}
}

// TestOtherSrv_DeleteOrderItemRemark 测试删除单品备注原因
func TestOtherSrv_DeleteOrderItemRemark(t *testing.T) {
	dbm := setupOrderItemRemarkTestDB(t)
	ctx := createOrderItemRemarkTestContext(t, dbm)

	// 创建测试数据
	db := dbm.GetDB(constant.MockDB)
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

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	otherSrv := NewOtherSrv(dbm, mockCache, settingSrv)

	// 测试删除
	deleteReq := req.DeleteOrderItemRemarkReq{
		Uuid: remark.Uuid,
	}

	err := otherSrv.DeleteOrderItemRemark(ctx, deleteReq)
	if err != nil {
		t.Fatalf("Failed to delete order item remark: %v", err)
	}

	// 验证软删除
	var deleted model.OrderItemRemark
	db.Where("uuid = ?", remark.Uuid).First(&deleted)
	if deleted.DeleteTime == 0 {
		t.Error("Expected delete_time to be set, got 0")
	}
}
