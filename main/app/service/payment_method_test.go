package service

import (
	stdcontext "context"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPaymentMethodTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupPaymentMethodTestDB(t *testing.T) *database.DBManager {
	t.Helper()

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
func TestPaymentMethodSrv_SaveGrabPaymentMethod_ExistingByPaymentName(t *testing.T) {
	dbm := setupPaymentMethodTestDB(t)
	ctx := createPaymentMethodTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)

	// 创建服务
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})
	settingSrv := setting.NewSrv(dbm, mockCache)
	paymentMethodSrv := NewPaymentMethodSrv(dbm, settingSrv).(*paymentMethodSrv)

	// 预先创建一个支付方式（通过 payment_name）
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

	// 测试：应该跳过创建（因为 payment_name 已存在）
	tx := db.Begin()
	err := paymentMethodSrv.SaveGrabPaymentMethod(ctx, tx)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to save Grab payment method: %v", err)
	}
	tx.Commit()

	// 验证只有一条记录（原有的记录）
	var count int64
	db.Model(&model.PaymentMethod{}).Where("payment_name = ?", constant.PaymentMethodNameGrab).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
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
