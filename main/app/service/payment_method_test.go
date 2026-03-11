//go:build integration

package service

import (
	stdcontext "context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPaymentMethodTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupPaymentMethodTestDB(t *testing.T) *database.DBManager {
	t.Helper()

	// 确保全局 logger 已初始化
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 创建 payment_methods 表
	err = db.Exec(`
		CREATE TABLE payment_methods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			code INTEGER DEFAULT 0,
			payment_name TEXT DEFAULT '',
			source INTEGER DEFAULT 0,
			logo_file_uuid INTEGER DEFAULT 0,
			qrcode_file_uuid INTEGER DEFAULT 0,
			fee_percent REAL DEFAULT 0,
			is_show_cashier INTEGER DEFAULT 0,
			is_show_assistant INTEGER DEFAULT 0,
			is_show_kiosk INTEGER DEFAULT 0,
			is_show_member_recharge INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			sort INTEGER DEFAULT 0,
			default_img TEXT DEFAULT '',
			erpnext_payment TEXT DEFAULT '',
			erpnext_payment_id TEXT DEFAULT '',
			headquarter_uuid INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create payment_methods table: %v", err)
	}

	// 创建唯一索引
	err = db.Exec(`CREATE UNIQUE INDEX uk_uuid ON payment_methods(uuid)`).Error
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

// createPaymentMethodTestContext 创建测试上下文
func createPaymentMethodTestContext(t *testing.T, dbm *database.DBManager) context.Context {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	// 创建测试用的 logger
	testLogger := logger.Logger
	if testLogger == nil {
		testLogger, _ = zap.NewDevelopment()
	}

	// 创建 Gin 上下文
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = req

	ctx := context.NewContext(
		context.WithContext(stdcontext.Background()),
		context.WithLogger(testLogger),
		context.WithGinContext(ginCtx),
	)
	ctx.SetDB(db)
	ctx.SetCompanyUuid(constant.MockDB)

	// 创建测试用的 Company 和 CompanySetting
	company := &model.Company{
		BaseModel: model.BaseModel{
			Uuid: constant.MockDB,
		},
		IsEnableErp: 0, // 默认不开启 ERP
	}
	ctx.SetCompany(*company)

	companySetting := model.CompanySetting{
		BaseModel: model.BaseModel{
			Uuid: constant.MockDB,
		},
		CompanyUuid: constant.MockDB,
	}
	ctx.SetCompanySetting(companySetting)

	return ctx
}

// TestPaymentMethodSrv_SaveGrabPaymentMethod 测试 SaveGrabPaymentMethod 方法
func TestPaymentMethodSrv_SaveGrabPaymentMethod(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 测试1: 首次创建 Grab 支付方式
	tx := db.Begin()
	err := paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to save Grab payment method: %v", err)
	}
	tx.Commit()

	// 验证支付方式已创建
	var paymentMethod model.PaymentMethod
	err = db.Where("code = ?", constant.PaymentMethodCodeGrab).First(&paymentMethod).Error
	if err != nil {
		t.Fatalf("Failed to find Grab payment method: %v", err)
	}

	if paymentMethod.Code != constant.PaymentMethodCodeGrab {
		t.Errorf("Expected code %d, got %d", constant.PaymentMethodCodeGrab, paymentMethod.Code)
	}
	if paymentMethod.PaymentName != constant.PaymentMethodNameGrab {
		t.Errorf("Expected payment_name %s, got %s", constant.PaymentMethodNameGrab, paymentMethod.PaymentName)
	}
	if paymentMethod.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Expected source %d, got %d", constant.PaymentMethodSourceSystem, paymentMethod.Source)
	}
	if paymentMethod.DefaultImg != "" {
		t.Errorf("Expected default_img to be empty, got %s", paymentMethod.DefaultImg)
	}
	if paymentMethod.Status != constant.PaymentMethodStatusEnable {
		t.Errorf("Expected status %d, got %d", constant.PaymentMethodStatusEnable, paymentMethod.Status)
	}

	// 测试2: 幂等性 - 再次调用应该跳过
	tx2 := db.Begin()
	err = paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx2)
	if err != nil {
		tx2.Rollback()
		t.Fatalf("Failed to save Grab payment method (idempotent call): %v", err)
	}
	tx2.Commit()

	// 验证只有一条记录
	var count int64
	db.Model(&model.PaymentMethod{}).Where("code = ?", constant.PaymentMethodCodeGrab).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d (idempotency test failed)", count)
	}
}

// TestPaymentMethodSrv_SaveLineManPaymentMethod 测试 SaveLineManPaymentMethod 方法
func TestPaymentMethodSrv_SaveLineManPaymentMethod(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 测试1: 首次创建 LINE MAN 支付方式
	tx := db.Begin()
	err := paymentMethodSrv.SaveLineManPaymentMethod(ctx, tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to save LINE MAN payment method: %v", err)
	}
	tx.Commit()

	// 验证支付方式已创建
	var paymentMethod model.PaymentMethod
	err = db.Where("code = ?", constant.PaymentMethodCodeLineMan).First(&paymentMethod).Error
	if err != nil {
		t.Fatalf("Failed to find LINE MAN payment method: %v", err)
	}

	if paymentMethod.Code != constant.PaymentMethodCodeLineMan {
		t.Errorf("Expected code %d, got %d", constant.PaymentMethodCodeLineMan, paymentMethod.Code)
	}
	if paymentMethod.PaymentName != constant.PaymentMethodNameLineMan {
		t.Errorf("Expected payment_name %s, got %s", constant.PaymentMethodNameLineMan, paymentMethod.PaymentName)
	}
	if paymentMethod.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Expected source %d, got %d", constant.PaymentMethodSourceSystem, paymentMethod.Source)
	}
	if paymentMethod.DefaultImg != "" {
		t.Errorf("Expected default_img to be empty, got %s", paymentMethod.DefaultImg)
	}
	if paymentMethod.Status != constant.PaymentMethodStatusEnable {
		t.Errorf("Expected status %d, got %d", constant.PaymentMethodStatusEnable, paymentMethod.Status)
	}

	// 测试2: 幂等性 - 再次调用应该跳过
	tx2 := db.Begin()
	err = paymentMethodSrv.SaveLineManPaymentMethod(ctx, tx2)
	if err != nil {
		tx2.Rollback()
		t.Fatalf("Failed to save LINE MAN payment method (idempotent call): %v", err)
	}
	tx2.Commit()

	// 验证只有一条记录
	var count int64
	db.Model(&model.PaymentMethod{}).Where("code = ?", constant.PaymentMethodCodeLineMan).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d (idempotency test failed)", count)
	}
}

// TestPaymentMethodSrv_SaveGrabPaymentMethod_ExistingByPaymentName 测试通过 payment_name 检查已存在的支付方式
// 注意：当前实现只通过 code 检查，不检查 payment_name
// 此测试验证：当存在相同 payment_name 但不同 code 的记录时，会创建新的 Grab 支付方式
func TestPaymentMethodSrv_SaveGrabPaymentMethod_ExistingByPaymentName(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 预先创建一个支付方式（通过 payment_name，但使用不同的 code）
	existingPayment := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:        constant.PaymentMethodNameGrab,
		Code:        99999, // 不同的 code
		PaymentName: constant.PaymentMethodNameGrab,
		Source:      constant.PaymentMethodSourceSystem,
		Status:      constant.PaymentMethodStatusEnable,
		Sort:        1,
	}
	db.Create(&existingPayment)

	// 测试：因为 code 不同，会创建新的 Grab 支付方式
	tx := db.Begin()
	err := paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to save Grab payment method: %v", err)
	}
	tx.Commit()

	// 验证有两条记录（原有的记录 + 新创建的）
	// 因为幂等性检查只基于 code，不基于 payment_name
	var count int64
	db.Model(&model.PaymentMethod{}).Where("payment_name = ?", constant.PaymentMethodNameGrab).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 records (existing + new), got %d", count)
	}

	// 验证新创建的记录使用正确的 code
	var newPayment model.PaymentMethod
	err = db.Where("code = ?", constant.PaymentMethodCodeGrab).First(&newPayment).Error
	if err != nil {
		t.Fatalf("Failed to find new Grab payment: %v", err)
	}
	if newPayment.Code != constant.PaymentMethodCodeGrab {
		t.Errorf("Expected code %d, got %d", constant.PaymentMethodCodeGrab, newPayment.Code)
	}
}

// TestPaymentMethodSrv_SaveGrabPaymentMethod_ExistingByCode 测试通过 code 检查已存在的支付方式
func TestPaymentMethodSrv_SaveGrabPaymentMethod_ExistingByCode(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 预先创建一个支付方式（通过 code）
	existingPayment := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:        "Existing Grab",
		Code:        constant.PaymentMethodCodeGrab,
		PaymentName: "Existing Grab", // 不同的 payment_name
		Source:      constant.PaymentMethodSourceSystem,
		Status:      constant.PaymentMethodStatusEnable,
		Sort:        1,
	}
	db.Create(&existingPayment)

	// 测试：应该跳过创建（因为 code 已存在）
	tx := db.Begin()
	err := paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to save Grab payment method: %v", err)
	}
	tx.Commit()

	// 验证只有一条记录（原有的记录）
	var count int64
	db.Model(&model.PaymentMethod{}).Where("code = ?", constant.PaymentMethodCodeGrab).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

// TestPaymentMethodSrv_GetList_FilterGrabAndLineMan 测试 GetList 方法过滤 Grab 和 LINE MAN
func TestPaymentMethodSrv_GetList_FilterGrabAndLineMan(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建测试数据
	paymentMethods := []model.PaymentMethod{
		{
			BaseModel: model.BaseModel{
				Uuid:       1001,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Cash",
			Code:        constant.PaymentMethodCodeCash,
			PaymentName: "Cash",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        1,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1002,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        constant.PaymentMethodNameGrab,
			Code:        constant.PaymentMethodCodeGrab,
			PaymentName: constant.PaymentMethodNameGrab,
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        2,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1003,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        constant.PaymentMethodNameLineMan,
			Code:        constant.PaymentMethodCodeLineMan,
			PaymentName: constant.PaymentMethodNameLineMan,
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        3,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1004,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "WeChat Pay",
			Code:        constant.PaymentMethodCodeWechat,
			PaymentName: "WeChat Pay",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        4,
		},
	}

	for _, pm := range paymentMethods {
		db.Create(&pm)
	}

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv)

	// 测试 GetList
	result := paymentMethodSrv.GetList(ctx, constant.PaymentMethodShowAll)

	// 验证结果：应该只有 Cash 和 WeChat Pay，不包含 Grab 和 LINE MAN
	if len(result.List) != 2 {
		t.Errorf("Expected 2 payment methods, got %d", len(result.List))
	}

	// 验证不包含 Grab 和 LINE MAN
	for _, item := range result.List {
		if item.Code == constant.PaymentMethodCodeGrab {
			t.Error("Grab payment method should be filtered out")
		}
		if item.Code == constant.PaymentMethodCodeLineMan {
			t.Error("LINE MAN payment method should be filtered out")
		}
	}
}

// TestPaymentMethodSrv_GetManagementList_FilterGrabAndLineMan 测试 GetManagementList 方法过滤 Grab 和 LINE MAN
func TestPaymentMethodSrv_GetManagementList_FilterGrabAndLineMan(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建测试数据
	paymentMethods := []model.PaymentMethod{
		{
			BaseModel: model.BaseModel{
				Uuid:       1001,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        "Cash",
			Code:        constant.PaymentMethodCodeCash,
			PaymentName: "Cash",
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        1,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1002,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        constant.PaymentMethodNameGrab,
			Code:        constant.PaymentMethodCodeGrab,
			PaymentName: constant.PaymentMethodNameGrab,
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        2,
		},
		{
			BaseModel: model.BaseModel{
				Uuid:       1003,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:        constant.PaymentMethodNameLineMan,
			Code:        constant.PaymentMethodCodeLineMan,
			PaymentName: constant.PaymentMethodNameLineMan,
			Source:      constant.PaymentMethodSourceSystem,
			Status:      constant.PaymentMethodStatusEnable,
			Sort:        3,
		},
	}

	for _, pm := range paymentMethods {
		db.Create(&pm)
	}

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv)

	// 测试 GetManagementList
	req := &req.PaymentMethodManagementListReq{
		PageNo:   1,
		PageSize: 10,
	}
	result, err := paymentMethodSrv.GetManagementList(ctx, req)
	if err != nil {
		t.Fatalf("Failed to get management list: %v", err)
	}

	// 验证结果：应该只有 Cash，不包含 Grab 和 LINE MAN
	if len(result.List) != 1 {
		t.Errorf("Expected 1 payment method, got %d", len(result.List))
	}

	// 验证不包含 Grab 和 LINE MAN
	for _, item := range result.List {
		if item.Uuid == 1002 || item.Uuid == 1003 {
			t.Error("Grab and LINE MAN payment methods should be filtered out")
		}
	}
}

// setupCreatePaymentFromERPTestDB 创建 createPaymentFromERP 测试专用数据库
// 包含 erpnext_payment_id 字段
func setupCreatePaymentFromERPTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 创建 payment_methods 表（GORM 默认表名，包含 erpnext_payment_id）
	err = db.Exec(`
		CREATE TABLE payment_methods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			code INTEGER DEFAULT 0,
			payment_name TEXT DEFAULT '',
			source INTEGER DEFAULT 0,
			logo_file_uuid INTEGER DEFAULT 0,
			qrcode_file_uuid INTEGER DEFAULT 0,
			fee_percent REAL DEFAULT 0,
			is_show_cashier INTEGER DEFAULT 0,
			is_show_assistant INTEGER DEFAULT 0,
			is_show_kiosk INTEGER DEFAULT 0,
			is_show_member_recharge INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			sort INTEGER DEFAULT 0,
			default_img TEXT DEFAULT '',
			erpnext_payment TEXT DEFAULT '',
			erpnext_payment_id TEXT DEFAULT '',
			headquarter_uuid INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create payment_methods table: %v", err)
	}

	return db
}

// TestPaymentMethodSrv_createPaymentFromERP_GrabSystemDefault 测试 Grab 系统默认支付方式创建
func TestPaymentMethodSrv_createPaymentFromERP_GrabSystemDefault(t *testing.T) {
	// 初始化 logger（如果未初始化）
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 模拟 ERP 返回的 Grab 系统默认支付方式
	// 名称格式：{PayType}-0000 - {company_abbr}
	companyAbbr := "TEST"
	erpPaymentName := paymentMethodSrv.buildERPPaymentName(constant.PaymentMethodNameGrab, companyAbbr)

	erpPayment := &selling.ModeOfPayment{
		Name:      erpPaymentName, // "Grab-0000 - TEST"
		Enabled:   true,
		PaymentId: "PM-00001",
	}

	// 调用 createPaymentFromERP
	err := paymentMethodSrv.createPaymentFromERP(db, erpPayment, companyAbbr)
	if err != nil {
		t.Fatalf("createPaymentFromERP failed: %v", err)
	}

	// 验证创建的支付方式
	var pm model.PaymentMethod
	err = db.Model(&model.PaymentMethod{}).First(&pm).Error
	if err != nil {
		t.Fatalf("Failed to query payment method: %v", err)
	}

	// 验证字段值
	if pm.Name != constant.PaymentMethodNameGrab {
		t.Errorf("Expected name '%s', got '%s'", constant.PaymentMethodNameGrab, pm.Name)
	}
	if pm.PaymentName != constant.PaymentMethodNameGrab {
		t.Errorf("Expected payment_name '%s', got '%s'", constant.PaymentMethodNameGrab, pm.PaymentName)
	}
	if pm.Code != constant.PaymentMethodCodeGrab {
		t.Errorf("Expected code %d, got %d", constant.PaymentMethodCodeGrab, pm.Code)
	}
	if pm.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Expected source %d (system), got %d", constant.PaymentMethodSourceSystem, pm.Source)
	}
	if pm.IsShowCashier != 0 {
		t.Errorf("Expected is_show_cashier 0, got %d", pm.IsShowCashier)
	}
	if pm.IsShowAssistant != 0 {
		t.Errorf("Expected is_show_assistant 0, got %d", pm.IsShowAssistant)
	}
	if pm.IsShowMemberRecharge != 0 {
		t.Errorf("Expected is_show_member_recharge 0, got %d", pm.IsShowMemberRecharge)
	}
	if pm.DefaultImg != "" {
		t.Errorf("Expected default_img '', got '%s'", pm.DefaultImg)
	}
	if pm.ErpnextPayment != erpPaymentName {
		t.Errorf("Expected erpnext_payment '%s', got '%s'", erpPaymentName, pm.ErpnextPayment)
	}
	if pm.ErpnextPaymentId != "PM-00001" {
		t.Errorf("Expected erpnext_payment_id 'PM-00001', got '%s'", pm.ErpnextPaymentId)
	}
	if pm.Status != 1 {
		t.Errorf("Expected status 1 (enabled), got %d", pm.Status)
	}
}

// TestPaymentMethodSrv_createPaymentFromERP_LineManSystemDefault 测试 LINE MAN 系统默认支付方式创建
func TestPaymentMethodSrv_createPaymentFromERP_LineManSystemDefault(t *testing.T) {
	// 初始化 logger（如果未初始化）
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 模拟 ERP 返回的 LINE MAN 系统默认支付方式
	companyAbbr := "ABC"
	erpPaymentName := paymentMethodSrv.buildERPPaymentName(constant.PaymentMethodNameLineMan, companyAbbr)

	erpPayment := &selling.ModeOfPayment{
		Name:      erpPaymentName, // "LINE MAN-0000 - ABC"
		Enabled:   true,
		PaymentId: "PM-00002",
	}

	// 调用 createPaymentFromERP
	err := paymentMethodSrv.createPaymentFromERP(db, erpPayment, companyAbbr)
	if err != nil {
		t.Fatalf("createPaymentFromERP failed: %v", err)
	}

	// 验证创建的支付方式
	var pm model.PaymentMethod
	err = db.Model(&model.PaymentMethod{}).First(&pm).Error
	if err != nil {
		t.Fatalf("Failed to query payment method: %v", err)
	}

	// 验证字段值
	if pm.Name != constant.PaymentMethodNameLineMan {
		t.Errorf("Expected name '%s', got '%s'", constant.PaymentMethodNameLineMan, pm.Name)
	}
	if pm.PaymentName != constant.PaymentMethodNameLineMan {
		t.Errorf("Expected payment_name '%s', got '%s'", constant.PaymentMethodNameLineMan, pm.PaymentName)
	}
	if pm.Code != constant.PaymentMethodCodeLineMan {
		t.Errorf("Expected code %d, got %d", constant.PaymentMethodCodeLineMan, pm.Code)
	}
	if pm.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Expected source %d (system), got %d", constant.PaymentMethodSourceSystem, pm.Source)
	}
	if pm.IsShowCashier != 0 {
		t.Errorf("Expected is_show_cashier 0, got %d", pm.IsShowCashier)
	}
	if pm.IsShowAssistant != 0 {
		t.Errorf("Expected is_show_assistant 0, got %d", pm.IsShowAssistant)
	}
	if pm.IsShowMemberRecharge != 0 {
		t.Errorf("Expected is_show_member_recharge 0, got %d", pm.IsShowMemberRecharge)
	}
	if pm.DefaultImg != "" {
		t.Errorf("Expected default_img '', got '%s'", pm.DefaultImg)
	}
	if pm.ErpnextPayment != erpPaymentName {
		t.Errorf("Expected erpnext_payment '%s', got '%s'", erpPaymentName, pm.ErpnextPayment)
	}
}

// TestPaymentMethodSrv_createPaymentFromERP_RegularPayment 测试普通支付方式创建（非 Grab/LINE MAN）
func TestPaymentMethodSrv_createPaymentFromERP_RegularPayment(t *testing.T) {
	// 初始化 logger（如果未初始化）
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 模拟 ERP 返回的普通支付方式
	companyAbbr := "TEST"
	erpPayment := &selling.ModeOfPayment{
		Name:      "Cash-0001 - TEST", // 普通支付方式，非 Grab/LINE MAN
		Enabled:   true,
		PaymentId: "PM-00003",
	}

	// 调用 createPaymentFromERP
	err := paymentMethodSrv.createPaymentFromERP(db, erpPayment, companyAbbr)
	if err != nil {
		t.Fatalf("createPaymentFromERP failed: %v", err)
	}

	// 验证创建的支付方式
	var pm model.PaymentMethod
	err = db.Model(&model.PaymentMethod{}).First(&pm).Error
	if err != nil {
		t.Fatalf("Failed to query payment method: %v", err)
	}

	// 验证字段值 - 普通支付方式应使用 ERP 原始名称
	if pm.Name != "Cash-0001 - TEST" {
		t.Errorf("Expected name 'Cash-0001 - TEST', got '%s'", pm.Name)
	}
	if pm.PaymentName != "Cash-0001 - TEST" {
		t.Errorf("Expected payment_name 'Cash-0001 - TEST', got '%s'", pm.PaymentName)
	}
	// 普通支付方式 code 应该是自动生成的（>= 20000）
	if pm.Code < 20000 {
		t.Errorf("Expected code >= 20000, got %d", pm.Code)
	}
	// 普通支付方式 source 应该是 1（手动添加）
	if pm.Source != model.PaymentSourceDefault {
		t.Errorf("Expected source %d (default/manual), got %d", model.PaymentSourceDefault, pm.Source)
	}
	// 普通支付方式的显示开关应该是 1
	if pm.IsShowCashier != 1 {
		t.Errorf("Expected is_show_cashier 1, got %d", pm.IsShowCashier)
	}
	if pm.IsShowAssistant != 1 {
		t.Errorf("Expected is_show_assistant 1, got %d", pm.IsShowAssistant)
	}
	if pm.IsShowMemberRecharge != 1 {
		t.Errorf("Expected is_show_member_recharge 1, got %d", pm.IsShowMemberRecharge)
	}
	// 普通支付方式应该有默认图标
	if pm.DefaultImg != "/image/pay/ja_pay.png" {
		t.Errorf("Expected default_img '/image/pay/ja_pay.png', got '%s'", pm.DefaultImg)
	}
}

// TestPaymentMethodSrv_createPaymentFromERP_GrabDisabled 测试 Grab 禁用状态创建
func TestPaymentMethodSrv_createPaymentFromERP_GrabDisabled(t *testing.T) {
	// 初始化 logger（如果未初始化）
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 模拟 ERP 返回的禁用状态 Grab 支付方式
	companyAbbr := "TEST"
	erpPaymentName := paymentMethodSrv.buildERPPaymentName(constant.PaymentMethodNameGrab, companyAbbr)

	erpPayment := &selling.ModeOfPayment{
		Name:      erpPaymentName,
		Enabled:   false, // 禁用状态
		PaymentId: "PM-00004",
	}

	// 调用 createPaymentFromERP
	err := paymentMethodSrv.createPaymentFromERP(db, erpPayment, companyAbbr)
	if err != nil {
		t.Fatalf("createPaymentFromERP failed: %v", err)
	}

	// 验证创建的支付方式
	var pm model.PaymentMethod
	err = db.Model(&model.PaymentMethod{}).First(&pm).Error
	if err != nil {
		t.Fatalf("Failed to query payment method: %v", err)
	}

	// 验证状态为禁用
	if pm.Status != 0 {
		t.Errorf("Expected status 0 (disabled), got %d", pm.Status)
	}
	// 其他字段仍应符合 Grab 系统默认配置
	if pm.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Expected source %d (system), got %d", constant.PaymentMethodSourceSystem, pm.Source)
	}
}

// TestPaymentMethodSrv_createPaymentFromERP_PrefixNotMatch 测试前缀不完全匹配的情况
func TestPaymentMethodSrv_createPaymentFromERP_PrefixNotMatch(t *testing.T) {
	// 初始化 logger（如果未初始化）
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	companyAbbr := "TEST"

	// 测试用例：名称前缀类似但不是精确匹配
	testCases := []struct {
		name        string
		erpName     string
		expectCode  int
		expectName  string
		shouldMatch bool
	}{
		{
			name:        "Grab-0001 (not 0000)",
			erpName:     "Grab-0001 - TEST",
			expectCode:  20000, // 应该使用自动生成的 code
			expectName:  "Grab-0001 - TEST",
			shouldMatch: false,
		},
		{
			name:        "Grab-0000 different company",
			erpName:     "Grab-0000 - OTHER",
			expectCode:  20000, // 应该使用自动生成的 code
			expectName:  "Grab-0000 - OTHER",
			shouldMatch: false,
		},
		{
			name:        "GrabFood-0000 (different prefix)",
			erpName:     "GrabFood-0000 - TEST",
			expectCode:  20000,
			expectName:  "GrabFood-0000 - TEST",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 清理表
			db.Where("1=1").Delete(&model.PaymentMethod{})

			erpPayment := &selling.ModeOfPayment{
				Name:      tc.erpName,
				Enabled:   true,
				PaymentId: "PM-TEST",
			}

			err := paymentMethodSrv.createPaymentFromERP(db, erpPayment, companyAbbr)
			if err != nil {
				t.Fatalf("createPaymentFromERP failed: %v", err)
			}

			var pm model.PaymentMethod
			err = db.Model(&model.PaymentMethod{}).First(&pm).Error
			if err != nil {
				t.Fatalf("Failed to query payment method: %v", err)
			}

			// 验证不匹配时使用原始 ERP 名称
			if pm.Name != tc.expectName {
				t.Errorf("Expected name '%s', got '%s'", tc.expectName, pm.Name)
			}

			// 验证不匹配时 source 应为手动添加
			if pm.Source != model.PaymentSourceDefault {
				t.Errorf("Expected source %d (manual), got %d", model.PaymentSourceDefault, pm.Source)
			}

			// 验证不匹配时 code 应该是自动生成的
			if pm.Code < 20000 {
				t.Errorf("Expected code >= 20000 (auto-generated), got %d", pm.Code)
			}
		})
	}
}

// ============================================================================
// TC-008: SaveGrabPaymentMethod - 未开启 ERP 时仅创建 TTPOS
// ============================================================================

// TestPaymentMethodSrv_SaveGrabPaymentMethod_NoERP 测试未开启 ERP 时仅创建 TTPOS
func TestPaymentMethodSrv_SaveGrabPaymentMethod_NoERP(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建测试上下文，ERP 未开启
	ctx := createPaymentMethodTestContext(t, dbm)
	// 确保 Company.IsEnableErp = 0（默认值）

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 调用 SaveGrabPaymentMethod
	tx := db.Begin()
	err := paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("SaveGrabPaymentMethod failed: %v", err)
	}
	tx.Commit()

	// 验证 TTPOS 创建成功
	var pm model.PaymentMethod
	err = db.Where("code = ?", constant.PaymentMethodCodeGrab).First(&pm).Error
	if err != nil {
		t.Fatalf("Failed to find Grab payment method: %v", err)
	}

	// 验证字段
	if pm.Code != constant.PaymentMethodCodeGrab {
		t.Errorf("Expected code %d, got %d", constant.PaymentMethodCodeGrab, pm.Code)
	}
	if pm.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Expected source %d, got %d", constant.PaymentMethodSourceSystem, pm.Source)
	}
	// 未开启 ERP 时，erpnext_payment 和 erpnext_payment_id 应为空
	if pm.ErpnextPayment != "" {
		t.Errorf("Expected erpnext_payment to be empty, got '%s'", pm.ErpnextPayment)
	}
	if pm.ErpnextPaymentId != "" {
		t.Errorf("Expected erpnext_payment_id to be empty, got '%s'", pm.ErpnextPaymentId)
	}
}

// ============================================================================
// TC-009: 并发调用 SaveGrabPaymentMethod
// ============================================================================

// TestPaymentMethodSrv_SaveGrabPaymentMethod_Concurrent 测试并发调用
// 注意：SQLite 内存数据库不支持真正的并发测试，此测试改为串行调用验证幂等性
func TestPaymentMethodSrv_SaveGrabPaymentMethod_Concurrent(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 连续调用多次，验证幂等性
	callCount := 5
	for i := 0; i < callCount; i++ {
		tx := db.Begin()
		err := paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx)
		if err != nil {
			tx.Rollback()
			t.Fatalf("Call %d failed: %v", i+1, err)
		}
		tx.Commit()
	}

	// 验证只有一条记录（幂等性）
	var count int64
	db.Model(&model.PaymentMethod{}).Where("code = ?", constant.PaymentMethodCodeGrab).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 Grab payment method after %d calls, got %d", callCount, count)
	}
}

// ============================================================================
// TC-019: 同时同步 Grab 和 LINE MAN（createPaymentFromERP）
// ============================================================================

// TestPaymentMethodSrv_createPaymentFromERP_MultiplePayments 测试同时同步多个支付方式
func TestPaymentMethodSrv_createPaymentFromERP_MultiplePayments(t *testing.T) {
	// 初始化 logger
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	companyAbbr := "TEST"

	// 模拟 ERP 返回的多个支付方式
	erpPayments := []*selling.ModeOfPayment{
		{
			Name:      paymentMethodSrv.buildERPPaymentName(constant.PaymentMethodNameGrab, companyAbbr),
			Enabled:   true,
			PaymentId: "PM-GRAB-001",
		},
		{
			Name:      paymentMethodSrv.buildERPPaymentName(constant.PaymentMethodNameLineMan, companyAbbr),
			Enabled:   true,
			PaymentId: "PM-LINEMAN-001",
		},
		{
			Name:      "Cash-0001 - TEST",
			Enabled:   true,
			PaymentId: "PM-CASH-001",
		},
	}

	// 依次创建
	for _, erpPayment := range erpPayments {
		err := paymentMethodSrv.createPaymentFromERP(db, erpPayment, companyAbbr)
		if err != nil {
			t.Fatalf("createPaymentFromERP failed for %s: %v", erpPayment.Name, err)
		}
	}

	// 验证创建的支付方式
	var payments []model.PaymentMethod
	err := db.Model(&model.PaymentMethod{}).Order("code").Find(&payments).Error
	if err != nil {
		t.Fatalf("Failed to query payment methods: %v", err)
	}

	if len(payments) != 3 {
		t.Fatalf("Expected 3 payment methods, got %d", len(payments))
	}

	// 验证 Cash（普通支付方式，code >= 20000）
	cashPm := payments[0]
	if cashPm.Code < 20000 {
		t.Errorf("Cash: Expected code >= 20000, got %d", cashPm.Code)
	}
	if cashPm.Source != model.PaymentSourceDefault {
		t.Errorf("Cash: Expected source %d, got %d", model.PaymentSourceDefault, cashPm.Source)
	}
	if cashPm.IsShowCashier != 1 {
		t.Errorf("Cash: Expected is_show_cashier 1, got %d", cashPm.IsShowCashier)
	}

	// 验证 Grab（code = 91100）
	var grabPm model.PaymentMethod
	for _, pm := range payments {
		if pm.Code == constant.PaymentMethodCodeGrab {
			grabPm = pm
			break
		}
	}
	if grabPm.Code != constant.PaymentMethodCodeGrab {
		t.Errorf("Grab: Expected code %d, not found", constant.PaymentMethodCodeGrab)
	}
	if grabPm.Name != constant.PaymentMethodNameGrab {
		t.Errorf("Grab: Expected name '%s', got '%s'", constant.PaymentMethodNameGrab, grabPm.Name)
	}
	if grabPm.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("Grab: Expected source %d, got %d", constant.PaymentMethodSourceSystem, grabPm.Source)
	}
	if grabPm.IsShowCashier != 0 {
		t.Errorf("Grab: Expected is_show_cashier 0, got %d", grabPm.IsShowCashier)
	}

	// 验证 LINE MAN（code = 91200）
	var lineManPm model.PaymentMethod
	for _, pm := range payments {
		if pm.Code == constant.PaymentMethodCodeLineMan {
			lineManPm = pm
			break
		}
	}
	if lineManPm.Code != constant.PaymentMethodCodeLineMan {
		t.Errorf("LINE MAN: Expected code %d, not found", constant.PaymentMethodCodeLineMan)
	}
	if lineManPm.Name != constant.PaymentMethodNameLineMan {
		t.Errorf("LINE MAN: Expected name '%s', got '%s'", constant.PaymentMethodNameLineMan, lineManPm.Name)
	}
	if lineManPm.Source != constant.PaymentMethodSourceSystem {
		t.Errorf("LINE MAN: Expected source %d, got %d", constant.PaymentMethodSourceSystem, lineManPm.Source)
	}
	if lineManPm.IsShowCashier != 0 {
		t.Errorf("LINE MAN: Expected is_show_cashier 0, got %d", lineManPm.IsShowCashier)
	}
}

// ============================================================================
// TC-017: syncFromERP 已存在时不重复创建
// ============================================================================

// TestPaymentMethodSrv_createPaymentFromERP_ExistingGrab 测试 Grab 已存在时的行为
func TestPaymentMethodSrv_createPaymentFromERP_ExistingGrab(t *testing.T) {
	// 初始化 logger
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := setupCreatePaymentFromERPTestDB(t)

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	companyAbbr := "TEST"
	erpPaymentName := paymentMethodSrv.buildERPPaymentName(constant.PaymentMethodNameGrab, companyAbbr)

	// 第一次创建
	erpPayment1 := &selling.ModeOfPayment{
		Name:      erpPaymentName,
		Enabled:   true,
		PaymentId: "PM-GRAB-001",
	}
	err := paymentMethodSrv.createPaymentFromERP(db, erpPayment1, companyAbbr)
	if err != nil {
		t.Fatalf("First createPaymentFromERP failed: %v", err)
	}

	// 验证第一次创建成功
	var grabPayment model.PaymentMethod
	err = db.Where("code = ?", constant.PaymentMethodCodeGrab).First(&grabPayment).Error
	if err != nil {
		t.Fatalf("Failed to find Grab payment: %v", err)
	}
	if grabPayment.Name != constant.PaymentMethodNameGrab {
		t.Errorf("Expected name=%s, got %s", constant.PaymentMethodNameGrab, grabPayment.Name)
	}
	if grabPayment.ErpnextPaymentId != "PM-GRAB-001" {
		t.Errorf("Expected erpnext_payment_id=PM-GRAB-001, got %s", grabPayment.ErpnextPaymentId)
	}

	// 注意：createPaymentFromERP 是低级方法，不检查是否已存在
	// 幂等性检查由上层 syncPaymentMethodFromERP 负责
	// 在实际业务中不会重复调用 createPaymentFromERP 创建相同的支付方式
}
