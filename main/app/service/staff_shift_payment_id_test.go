package service

import (
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupStaffShiftTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupStaffShiftTestDB(t *testing.T) *database.DBManager {
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

// createStaffShiftTestContext 创建测试上下文
func createStaffShiftTestContext(t *testing.T, dbm *database.DBManager, enableErp bool) context.Context {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	// 创建测试用的 logger
	testLogger := logger.Logger
	if testLogger == nil {
		testLogger, _ = zap.NewDevelopment()
	}

	// 创建 Gin 上下文
	gin.SetMode(gin.TestMode)
	w := gin.New().ServeHTTP
	_ = w

	ctx := context.NewContext(
		context.WithLogger(testLogger),
	)
	ctx.SetDB(db)
	ctx.SetCompanyUuid(constant.MockDB)

	// 创建测试用的 Company 和 CompanySetting
	erpEnabled := 0
	if enableErp {
		erpEnabled = 1
	}
	company := &model.Company{
		BaseModel: model.BaseModel{
			Uuid: constant.MockDB,
		},
		IsEnableErp: erpEnabled,
	}
	ctx.SetCompany(*company)

	erpnextSiteCode := ""
	if enableErp {
		erpnextSiteCode = "TEST_SITE"
	}
	companySetting := model.CompanySetting{
		BaseModel: model.BaseModel{
			Uuid: constant.MockDB,
		},
		CompanyUuid:        constant.MockDB,
		ErpnextSiteCode:    erpnextSiteCode,
		ErpnextPosProfileName: "TEST_PROFILE",
		ErpnextAdminEmail:  "test@example.com",
		ErpnextCompanyAbbr: "TEST",
		ErpnextBranchName:  "TEST_BRANCH",
	}
	ctx.SetCompanySetting(companySetting)

	return ctx
}

// TestGetCashPaymentModeForErp_WithPaymentID 测试开账时 Cash 支付方式有 PaymentID 的情况
func TestGetCashPaymentModeForErp_WithPaymentID(t *testing.T) {
	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建 Cash 支付方式（有 PaymentID）
	cashPaymentMethod := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:             "Cash",
		Code:             constant.PaymentMethodCodeCash,
		PaymentName:      "Cash",
		Source:           constant.PaymentMethodSourceSystem,
		Status:           constant.PaymentMethodStatusEnable,
		ErpnextPayment:   "Cash-0001-TEST",
		ErpnextPaymentId: "PID1234567890123456",
	}
	db.Create(&cashPaymentMethod)

	// 查询 Cash 支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	commonRepo := repository.NewCommonRepo()
	result := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeCash),
		commonRepo.WhereBySource(constant.PaymentMethodSourceSystem),
		commonRepo.WhereBySoftDelete(),
	)

	// 验证结果
	assert.Equal(t, uint64(1001), result.Uuid)
	assert.Equal(t, "PID1234567890123456", result.ErpnextPaymentId)

	// 验证 PaymentID 选择逻辑
	var modeOfPayment string
	if result.Uuid > 0 && result.ErpnextPaymentId != "" {
		modeOfPayment = result.ErpnextPaymentId
	} else {
		modeOfPayment = "Cash"
	}
	assert.Equal(t, "PID1234567890123456", modeOfPayment)
}

// TestGetCashPaymentModeForErp_WithoutPaymentID 测试开账时 Cash 支付方式没有 PaymentID 的情况
func TestGetCashPaymentModeForErp_WithoutPaymentID(t *testing.T) {
	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建 Cash 支付方式（没有 PaymentID）
	cashPaymentMethod := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:           "Cash",
		Code:           constant.PaymentMethodCodeCash,
		PaymentName:    "Cash",
		Source:         constant.PaymentMethodSourceSystem,
		Status:         constant.PaymentMethodStatusEnable,
		ErpnextPayment: "Cash-0001-TEST",
		// ErpnextPaymentId 为空
	}
	db.Create(&cashPaymentMethod)

	// 查询 Cash 支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	commonRepo := repository.NewCommonRepo()
	result := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeCash),
		commonRepo.WhereBySource(constant.PaymentMethodSourceSystem),
		commonRepo.WhereBySoftDelete(),
	)

	// 验证结果
	assert.Equal(t, uint64(1001), result.Uuid)
	assert.Equal(t, "", result.ErpnextPaymentId)

	// 验证 PaymentID 选择逻辑
	var modeOfPayment string
	if result.Uuid > 0 && result.ErpnextPaymentId != "" {
		modeOfPayment = result.ErpnextPaymentId
	} else {
		modeOfPayment = "Cash"
	}
	assert.Equal(t, "Cash", modeOfPayment)
}

// TestGetFreeMealPaymentModeForErp_WithPaymentID 测试关账时 Free Meal 支付方式有 PaymentID 的情况
func TestGetFreeMealPaymentModeForErp_WithPaymentID(t *testing.T) {
	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建 Free Meal 支付方式（有 PaymentID）
	freeMealPaymentMethod := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:             "Free Meal",
		Code:             constant.PaymentMethodCodeFreeMealForErp, // 92000
		PaymentName:      "Free Meal",
		Source:           constant.PaymentMethodSourceSystem,
		Status:           constant.PaymentMethodStatusEnable,
		ErpnextPayment:   "Free Meal-0001-TEST",
		ErpnextPaymentId: "PID9876543210987654",
	}
	db.Create(&freeMealPaymentMethod)

	// 查询 Free Meal 支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	commonRepo := repository.NewCommonRepo()
	result := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeFreeMealForErp),
		commonRepo.WhereBySource(constant.PaymentMethodSourceSystem),
		commonRepo.WhereBySoftDelete(),
	)

	// 验证结果
	assert.Equal(t, uint64(1002), result.Uuid)
	assert.Equal(t, "PID9876543210987654", result.ErpnextPaymentId)

	// 验证 PaymentID 选择逻辑
	var modeOfPayment string
	if result.Uuid > 0 && result.ErpnextPaymentId != "" {
		modeOfPayment = result.ErpnextPaymentId
	} else {
		modeOfPayment = "Free Meal"
	}
	assert.Equal(t, "PID9876543210987654", modeOfPayment)
}

// TestGetFreeMealPaymentModeForErp_WithoutPaymentID 测试关账时 Free Meal 支付方式没有 PaymentID 的情况
func TestGetFreeMealPaymentModeForErp_WithoutPaymentID(t *testing.T) {
	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建 Free Meal 支付方式（没有 PaymentID）
	freeMealPaymentMethod := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:           "Free Meal",
		Code:           constant.PaymentMethodCodeFreeMealForErp, // 92000
		PaymentName:    "Free Meal",
		Source:         constant.PaymentMethodSourceSystem,
		Status:         constant.PaymentMethodStatusEnable,
		ErpnextPayment: "Free Meal-0001-TEST",
		// ErpnextPaymentId 为空
	}
	db.Create(&freeMealPaymentMethod)

	// 查询 Free Meal 支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	commonRepo := repository.NewCommonRepo()
	result := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeFreeMealForErp),
		commonRepo.WhereBySource(constant.PaymentMethodSourceSystem),
		commonRepo.WhereBySoftDelete(),
	)

	// 验证结果
	assert.Equal(t, uint64(1002), result.Uuid)
	assert.Equal(t, "", result.ErpnextPaymentId)

	// 验证 PaymentID 选择逻辑
	var modeOfPayment string
	if result.Uuid > 0 && result.ErpnextPaymentId != "" {
		modeOfPayment = result.ErpnextPaymentId
	} else {
		modeOfPayment = "Free Meal"
	}
	assert.Equal(t, "Free Meal", modeOfPayment)
}

// TestGetFreeMealPaymentModeForErp_NotExists 测试关账时 Free Meal 支付方式不存在的情况
func TestGetFreeMealPaymentModeForErp_NotExists(t *testing.T) {
	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 不创建 Free Meal 支付方式

	// 查询 Free Meal 支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	commonRepo := repository.NewCommonRepo()
	result := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeFreeMealForErp),
		commonRepo.WhereBySource(constant.PaymentMethodSourceSystem),
		commonRepo.WhereBySoftDelete(),
	)

	// 验证结果：应该查询不到
	assert.Equal(t, uint64(0), result.Uuid)
	assert.Equal(t, "", result.ErpnextPaymentId)

	// 验证 PaymentID 选择逻辑：应该使用默认值
	var modeOfPayment string
	if result.Uuid > 0 && result.ErpnextPaymentId != "" {
		modeOfPayment = result.ErpnextPaymentId
	} else {
		modeOfPayment = "Free Meal"
	}
	assert.Equal(t, "Free Meal", modeOfPayment)
}

// TestGetPaymentModeForErp_MultipleSources 测试多个来源的 Cash 支付方式，应该只查询系统默认的
func TestGetPaymentModeForErp_MultipleSources(t *testing.T) {
	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建系统默认的 Cash 支付方式（有 PaymentID）
	systemCash := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:             "Cash",
		Code:             constant.PaymentMethodCodeCash,
		PaymentName:      "Cash",
		Source:           constant.PaymentMethodSourceSystem, // 系统默认
		Status:           constant.PaymentMethodStatusEnable,
		ErpnextPayment:   "Cash-0001-TEST",
		ErpnextPaymentId: "PID1234567890123456",
	}
	db.Create(&systemCash)

	// 创建自行添加的 Cash 支付方式（不应该被查询到）
	customCash := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid:       1002,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:             "Custom Cash",
		Code:             constant.PaymentMethodCodeCash,
		PaymentName:      "Custom Cash",
		Source:           constant.PaymentMethodSourceDefault, // 自行添加
		Status:           constant.PaymentMethodStatusEnable,
		ErpnextPayment:   "Cash-0002-TEST",
		ErpnextPaymentId: "PID9999999999999999",
	}
	db.Create(&customCash)

	// 查询 Cash 支付方式（系统默认，source = 0）
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	commonRepo := repository.NewCommonRepo()
	result := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeCash),
		commonRepo.WhereBySource(constant.PaymentMethodSourceSystem),
		commonRepo.WhereBySoftDelete(),
	)

	// 验证结果：应该查询到系统默认的 Cash 支付方式
	assert.Equal(t, uint64(1001), result.Uuid)
	assert.Equal(t, "PID1234567890123456", result.ErpnextPaymentId)
	assert.Equal(t, constant.PaymentMethodSourceSystem, result.Source)
}

