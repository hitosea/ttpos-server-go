package service

import (
	"testing"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupShiftDecoupleTestDB 创建测试数据库（复用已有的 setupStaffShiftTestDB 并添加额外的表）
func setupShiftDecoupleTestDB(t *testing.T) (*staffShiftSrv, cache.Cache) {
	t.Helper()

	// 确保全局 logger 已初始化（goroutine 中使用）
	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	dbm := setupStaffShiftTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 创建 staff_shift_logs 表
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS staff_shift_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			staff_uuid INTEGER DEFAULT 0,
			shift_no TEXT DEFAULT '',
			status INTEGER DEFAULT 0,
			previous_shift_cash REAL DEFAULT 0,
			current_cash_total REAL DEFAULT 0,
			incomes TEXT DEFAULT '',
			total_income REAL DEFAULT 0,
			cash_taken_out REAL DEFAULT 0,
			cash_left REAL DEFAULT 0,
			cash_income REAL DEFAULT 0,
			total_business REAL DEFAULT 0,
			is_printed INTEGER DEFAULT 0,
			remark TEXT DEFAULT '',
			withdraw_cash REAL DEFAULT 0,
			deposit_cash REAL DEFAULT 0,
			exception_remark TEXT DEFAULT '',
			abnormal TEXT DEFAULT '',
			shift_start_time INTEGER DEFAULT 0,
			shift_end_time INTEGER DEFAULT 0,
			erpnext_open_pos_entry_name TEXT DEFAULT '',
			erpnext_close_pos_entry_name TEXT DEFAULT '',
			erpnext_async_record_id TEXT DEFAULT '',
			opening_payment_methods TEXT DEFAULT '',
			shift_version INTEGER DEFAULT 2,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create staff_shift_logs table: %v", err)
	}

	// 创建 cash_boxes 表
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cash_boxes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			name TEXT DEFAULT '',
			balance REAL DEFAULT 0,
			frozen_balance REAL DEFAULT 0,
			previous_balance REAL DEFAULT 0,
			cash_withdrawal REAL DEFAULT 0,
			cash_deposit REAL DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create cash_boxes table: %v", err)
	}

	// 创建 ttpos_takeout_order 表（CreateWorkingLog 的异步 goroutine 需要）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_takeout_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			staff_shift_log_uuid INTEGER DEFAULT 0,
			staff_uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create ttpos_takeout_order table: %v", err)
	}

	// 插入默认钱箱记录
	err = db.Exec(`INSERT INTO cash_boxes (uuid, name, balance, frozen_balance, create_time, update_time) VALUES (1, 'default', 100.00, 0, ?, ?)`,
		time.Now().Unix(), time.Now().Unix()).Error
	if err != nil {
		t.Fatalf("Failed to insert cash_box record: %v", err)
	}

	memCache := cache.NewCache(cache.GoCache, cache.Config{})
	srv := &staffShiftSrv{
		dbm:            dbm,
		cache:          memCache,
		cacheKeyPrefix: "__TEST_SHIFTLOG_GENERATENUMBER__",
		MaxAmount:      "1000000000000",
	}

	return srv, memCache
}

// TestValidatePaymentMethod_ErpDisabled 未开启 ERP 时，直接允许
func TestValidatePaymentMethod_ErpDisabled(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, false) // ERP disabled

	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT001", 12345)
	assert.NoError(t, err)
	assert.True(t, ok, "ERP disabled should allow all payment methods")
}

// TestValidatePaymentMethod_NewShiftVersion 新版班次（ShiftVersion=2）直接允许
func TestValidatePaymentMethod_NewShiftVersion(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true) // ERP enabled
	db := srv.dbm.GetDB(constant.MockDB)

	// 创建新版班次记录
	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLogRepo.Create(model.StaffShiftLog{
		BaseModel:             model.BaseModel{Uuid: 10001, CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()},
		StaffUuid:             1001,
		ShiftNo:               "SHIFT_NEW_001",
		ShiftVersion:          constant.ShiftVersionNew,
		OpeningPaymentMethods: "100,200,300", // 即使有列表也应该直接通过
	})

	// 使用列表中不存在的支付方式 UUID，新版班次应该直接允许
	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT_NEW_001", 999)
	assert.NoError(t, err)
	assert.True(t, ok, "New shift version should allow all payment methods regardless of list")
}

// TestValidatePaymentMethod_NewShiftVersion_ZeroDefault 默认值（ShiftVersion=0）视为新版
func TestValidatePaymentMethod_NewShiftVersion_ZeroDefault(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true)
	db := srv.dbm.GetDB(constant.MockDB)

	// ShiftVersion=0 的记录（数据库默认值或历史数据迁移后）
	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLogRepo.Create(model.StaffShiftLog{
		BaseModel:    model.BaseModel{Uuid: 10002, CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()},
		StaffUuid:    1001,
		ShiftNo:      "SHIFT_ZERO_001",
		ShiftVersion: 0, // 未设置版本，应视为新版
	})

	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT_ZERO_001", 999)
	assert.NoError(t, err)
	assert.True(t, ok, "ShiftVersion=0 should be treated as new version")
}

// TestValidatePaymentMethod_LegacyShift_InList 旧版班次，支付方式在列表中
func TestValidatePaymentMethod_LegacyShift_InList(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true)
	db := srv.dbm.GetDB(constant.MockDB)

	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLogRepo.Create(model.StaffShiftLog{
		BaseModel:             model.BaseModel{Uuid: 10003, CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()},
		StaffUuid:             1001,
		ShiftNo:               "SHIFT_LEGACY_001",
		ShiftVersion:          constant.ShiftVersionLegacy,
		OpeningPaymentMethods: "100,200,300",
	})

	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT_LEGACY_001", 200)
	assert.NoError(t, err)
	assert.True(t, ok, "Legacy shift should allow payment method in the opening list")
}

// TestValidatePaymentMethod_LegacyShift_NotInList 旧版班次，支付方式不在列表中
func TestValidatePaymentMethod_LegacyShift_NotInList(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true)
	db := srv.dbm.GetDB(constant.MockDB)

	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLogRepo.Create(model.StaffShiftLog{
		BaseModel:             model.BaseModel{Uuid: 10004, CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()},
		StaffUuid:             1001,
		ShiftNo:               "SHIFT_LEGACY_002",
		ShiftVersion:          constant.ShiftVersionLegacy,
		OpeningPaymentMethods: "100,200,300",
	})

	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT_LEGACY_002", 999)
	assert.NoError(t, err)
	assert.False(t, ok, "Legacy shift should reject payment method not in the opening list")
}

// TestValidatePaymentMethod_LegacyShift_EmptyList 旧版班次，空列表
func TestValidatePaymentMethod_LegacyShift_EmptyList(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true)
	db := srv.dbm.GetDB(constant.MockDB)

	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLogRepo.Create(model.StaffShiftLog{
		BaseModel:             model.BaseModel{Uuid: 10005, CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()},
		StaffUuid:             1001,
		ShiftNo:               "SHIFT_LEGACY_003",
		ShiftVersion:          constant.ShiftVersionLegacy,
		OpeningPaymentMethods: "", // 空列表
	})

	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT_LEGACY_003", 100)
	assert.NoError(t, err)
	assert.False(t, ok, "Legacy shift with empty payment methods list should reject")
}

// TestValidatePaymentMethod_ShiftNotFound 班次记录不存在
func TestValidatePaymentMethod_ShiftNotFound(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true)

	ok, err := srv.ValidatePaymentMethod(ctx, "NONEXISTENT_SHIFT", 100)
	assert.Error(t, err, "Should return error for non-existent shift")
	assert.False(t, ok)
}

// TestCreateWorkingLog_NewShiftVersion 创建新版班次记录
func TestCreateWorkingLog_NewShiftVersion(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true) // ERP enabled
	db := srv.dbm.GetDB(constant.MockDB)

	// 创建启用的支付方式
	db.Exec(`INSERT INTO payment_methods (uuid, name, code, payment_name, source, status, erpnext_payment, create_time, update_time) VALUES (100, 'Cash', 10000, 'Cash', 0, 1, 'Cash-0001', ?, ?)`,
		time.Now().Unix(), time.Now().Unix())
	db.Exec(`INSERT INTO payment_methods (uuid, name, code, payment_name, source, status, erpnext_payment, create_time, update_time) VALUES (200, 'QR', 20000, 'QR Payment', 1, 1, 'QR-0001', ?, ?)`,
		time.Now().Unix(), time.Now().Unix())

	staff := model.Staff{
		BaseModel:        model.BaseModel{Uuid: 1001},
		CompanyUuid:      constant.MockDB,
		CashierLoginTime: time.Now().Unix(),
	}

	shiftLog, err := srv.CreateWorkingLog(ctx, staff)
	assert.NoError(t, err)
	assert.NotEmpty(t, shiftLog.ShiftNo, "ShiftNo should be generated")
	assert.Equal(t, constant.ShiftVersionNew, shiftLog.ShiftVersion, "Should be new shift version")
	assert.Empty(t, shiftLog.ErpnextOpenPosEntryName, "New shift should not have ERP open pos entry")
	assert.Equal(t, staff.CashierLoginTime, shiftLog.ShiftStartTime, "Should use cashier login time")
	assert.Equal(t, float64(100), shiftLog.PreviousShiftCash, "Should get previous cash from cash box")

	// 验证支付方式列表被保存（用于内部对账参考）
	assert.Contains(t, shiftLog.OpeningPaymentMethods, "100", "Should contain cash payment UUID")
	assert.Contains(t, shiftLog.OpeningPaymentMethods, "200", "Should contain QR payment UUID")
}

// TestCreateWorkingLog_NoErp 非 ERP 商家创建班次
func TestCreateWorkingLog_NoErp(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, false) // ERP disabled

	staff := model.Staff{
		BaseModel:        model.BaseModel{Uuid: 1002},
		CompanyUuid:      constant.MockDB,
		CashierLoginTime: time.Now().Unix(),
	}

	shiftLog, err := srv.CreateWorkingLog(ctx, staff)
	assert.NoError(t, err)
	assert.NotEmpty(t, shiftLog.ShiftNo)
	assert.Equal(t, constant.ShiftVersionNew, shiftLog.ShiftVersion)
	assert.Empty(t, shiftLog.OpeningPaymentMethods, "Non-ERP should not save payment methods")
	assert.Empty(t, shiftLog.ErpnextOpenPosEntryName)
}

// TestCreateWorkingLog_DefaultStartTime 未设置 CashierLoginTime 时使用当前时间
func TestCreateWorkingLog_DefaultStartTime(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, false)

	beforeTime := time.Now().Unix()
	staff := model.Staff{
		BaseModel:        model.BaseModel{Uuid: 1003},
		CompanyUuid:      constant.MockDB,
		CashierLoginTime: 0, // 未设置
	}

	shiftLog, err := srv.CreateWorkingLog(ctx, staff)
	afterTime := time.Now().Unix()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, shiftLog.ShiftStartTime, beforeTime, "Should use current time when CashierLoginTime is 0")
	assert.LessOrEqual(t, shiftLog.ShiftStartTime, afterTime)
}

// TestIsNewShiftVersion 测试 IsNewShiftVersion 方法
func TestIsNewShiftVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  int
		expected bool
	}{
		{"Version 0 (unset) → new", 0, true},
		{"Version 1 (legacy) → old", constant.ShiftVersionLegacy, false},
		{"Version 2 (new) → new", constant.ShiftVersionNew, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			log := &model.StaffShiftLog{ShiftVersion: tc.version}
			assert.Equal(t, tc.expected, log.IsNewShiftVersion())
		})
	}
}

// TestValidatePaymentMethod_LegacyShift_WhitespaceTrimming 旧版班次，UUID 列表含空格
func TestValidatePaymentMethod_LegacyShift_WhitespaceTrimming(t *testing.T) {
	srv, _ := setupShiftDecoupleTestDB(t)
	ctx := createStaffShiftTestContext(t, srv.dbm, true)
	db := srv.dbm.GetDB(constant.MockDB)

	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLogRepo.Create(model.StaffShiftLog{
		BaseModel:             model.BaseModel{Uuid: 10006, CreateTime: time.Now().Unix(), UpdateTime: time.Now().Unix()},
		StaffUuid:             1001,
		ShiftNo:               "SHIFT_LEGACY_SPACE",
		ShiftVersion:          constant.ShiftVersionLegacy,
		OpeningPaymentMethods: " 100 , 200 , 300 ", // 含空格
	})

	ok, err := srv.ValidatePaymentMethod(ctx, "SHIFT_LEGACY_SPACE", 200)
	assert.NoError(t, err)
	assert.True(t, ok, "Should trim whitespace and find payment method")
}
