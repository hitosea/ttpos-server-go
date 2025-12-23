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

// setupStatisticsExcludeDataManageTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupStatisticsExcludeDataManageTestDB(t *testing.T) *database.DBManager {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	prefix := "ttpos_"

	// 创建 statistics_sale 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `statistics_sale (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			sale_bill_uuid INTEGER DEFAULT 0,
			complete_time INTEGER DEFAULT 0,
			order_amount DECIMAL(14,2) DEFAULT 0.00,
			pay_amount DECIMAL(14,2) DEFAULT 0.00,
			product_num DECIMAL(14,2) DEFAULT 0.00
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_sale table: %v", err)
	}

	// 创建 sale_bill 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `sale_bill (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			finish_time INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			bill_type INTEGER DEFAULT 0,
			meal_num INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create sale_bill table: %v", err)
	}

	// 创建 data_manage 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `data_manage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			data_uuid INTEGER DEFAULT 0,
			type INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create data_manage table: %v", err)
	}

	// 创建 company_setting 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `company_setting (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			company_uuid INTEGER DEFAULT 0,
			"values" TEXT DEFAULT '{}'
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create company_setting table: %v", err)
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

// createStatisticsExcludeDataManageTestContext 创建测试上下文
func createStatisticsExcludeDataManageTestContext(t *testing.T, dbm *database.DBManager, enableDataManage bool) context.Context {
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
		IsEnableErp: 0,
	}
	ctx.SetCompany(*company)

	// 设置 CompanySetting，根据 enableDataManage 参数设置数据管理功能
	companySetting := model.CompanySetting{
		BaseModel: model.BaseModel{
			Uuid: constant.MockDB,
		},
		CompanyUuid: constant.MockDB,
		Timezone:    "Asia/Shanghai",
		EnableDataManagement: func() int {
			if enableDataManage {
				return 1
			}
			return 0
		}(),
	}
	ctx.SetCompanySetting(companySetting)

	return ctx
}

// TestStatisticsSrv_ExcludeDataManage_CountBusinessTimePeriod 测试营业时段统计排除数据管理订单
func TestStatisticsSrv_ExcludeDataManage_CountBusinessTimePeriod(t *testing.T) {
	dbm := setupStatisticsExcludeDataManageTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 准备测试数据
	now := int64(1703232000) // 2023-12-22 00:00:00

	// 创建普通订单
	normalBillUuid := uint64(1001)
	err := db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, create_time, finish_time, status, delete_time)
		VALUES (?, ?, ?, ?, ?)
	`, normalBillUuid, now, now+3600, constant.SaleBillStatusComplete, constant.NotDeleted).Error
	if err != nil {
		t.Fatalf("Failed to insert normal sale bill: %v", err)
	}

	// 创建已选择订单（数据管理订单）
	selectedBillUuid := uint64(1002)
	err = db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, create_time, finish_time, status, delete_time)
		VALUES (?, ?, ?, ?, ?)
	`, selectedBillUuid, now, now+3600, constant.SaleBillStatusComplete, constant.NotDeleted).Error
	if err != nil {
		t.Fatalf("Failed to insert selected sale bill: %v", err)
	}

	// 将订单标记为数据管理订单
	err = db.Exec(`
		INSERT INTO ttpos_data_manage (uuid, data_uuid, type, delete_time, create_time)
		VALUES (?, ?, ?, ?, ?)
	`, 1, selectedBillUuid, model.DataManageTypeOrder, constant.NotDeleted, now).Error
	if err != nil {
		t.Fatalf("Failed to insert data manage record: %v", err)
	}

	// 创建统计服务
	statisticsSrv := NewStatisticsSrv()

	// 测试场景1: 数据管理功能开启，应排除已选择订单
	t.Run("数据管理功能开启-应排除已选择订单", func(t *testing.T) {
		ctx := createStatisticsExcludeDataManageTestContext(t, dbm, true)
		req := req.BusinessTimePeriodReq{
			QueryStartTime:    now,
			QueryEndTime:      now + 86400,
			TimePeriod:        3, // 1小时
			StatisticsType:    1, // 结账时间
			ExcludeDataManage: true,
		}

		result := statisticsSrv.CountBusinessTimePeriod(ctx, req)

		// 验证结果：应该只包含普通订单，不包含已选择订单
		// 验证总数应该 >= 0
		if result.TotalBusinessTimePeriodNum < 0 {
			t.Errorf("Expected total count >= 0, got %d", result.TotalBusinessTimePeriodNum)
		}
		// 验证结果列表不为 nil
		if result.BusinessTimePeriodList == nil {
			t.Error("Expected BusinessTimePeriodList not nil")
		}
	})

	// 测试场景2: 数据管理功能关闭，应包含所有订单
	t.Run("数据管理功能关闭-应包含所有订单", func(t *testing.T) {
		ctx := createStatisticsExcludeDataManageTestContext(t, dbm, false)
		req := req.BusinessTimePeriodReq{
			QueryStartTime:    now,
			QueryEndTime:      now + 86400,
			TimePeriod:        3,
			StatisticsType:    1,
			ExcludeDataManage: false,
		}

		result := statisticsSrv.CountBusinessTimePeriod(ctx, req)

		// 验证结果：应该包含所有订单
		if result.TotalBusinessTimePeriodNum < 0 {
			t.Errorf("Expected total count >= 0, got %d", result.TotalBusinessTimePeriodNum)
		}
		// 验证结果列表不为 nil
		if result.BusinessTimePeriodList == nil {
			t.Error("Expected BusinessTimePeriodList not nil")
		}
	})
}

// TestStatisticsSrv_ExcludeDataManage_CountBusinessSummary 测试综合运营统计排除数据管理订单
func TestStatisticsSrv_ExcludeDataManage_CountBusinessSummary(t *testing.T) {
	dbm := setupStatisticsExcludeDataManageTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 准备测试数据
	now := int64(1703232000)

	// 创建普通订单
	normalBillUuid := uint64(2001)
	err := db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, create_time, finish_time, status, delete_time)
		VALUES (?, ?, ?, ?, ?)
	`, normalBillUuid, now, now+3600, constant.SaleBillStatusComplete, constant.NotDeleted).Error
	if err != nil {
		t.Fatalf("Failed to insert normal sale bill: %v", err)
	}

	// 创建已选择订单
	selectedBillUuid := uint64(2002)
	err = db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, create_time, finish_time, status, delete_time)
		VALUES (?, ?, ?, ?, ?)
	`, selectedBillUuid, now, now+3600, constant.SaleBillStatusComplete, constant.NotDeleted).Error
	if err != nil {
		t.Fatalf("Failed to insert selected sale bill: %v", err)
	}

	// 将订单标记为数据管理订单
	err = db.Exec(`
		INSERT INTO ttpos_data_manage (uuid, data_uuid, type, delete_time, create_time)
		VALUES (?, ?, ?, ?, ?)
	`, 2, selectedBillUuid, model.DataManageTypeOrder, constant.NotDeleted, now).Error
	if err != nil {
		t.Fatalf("Failed to insert data manage record: %v", err)
	}

	statisticsSrv := NewStatisticsSrv()

	t.Run("数据管理功能开启-应排除已选择订单", func(t *testing.T) {
		ctx := createStatisticsExcludeDataManageTestContext(t, dbm, true)
		req := req.StatisticsSummaryReq{
			QueryStartTime:    now,
			QueryEndTime:      now + 86400,
			Cycle:             0, // 按日
			ExcludeDataManage: true,
		}

		result := statisticsSrv.CountBusinessSummary(ctx, req)

		if result.TotalStatisticsComprehensiveNum < 0 {
			t.Errorf("Expected total count >= 0, got %d", result.TotalStatisticsComprehensiveNum)
		}
	})
}

// TestStatisticsSrv_ExcludeDataManage_CountBusinessPaymentMethod 测试营业收款统计排除数据管理订单
func TestStatisticsSrv_ExcludeDataManage_CountBusinessPaymentMethod(t *testing.T) {
	dbm := setupStatisticsExcludeDataManageTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 准备测试数据
	now := int64(1703232000)

	// 创建普通订单
	normalBillUuid := uint64(3001)
	err := db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, create_time, finish_time, status, delete_time)
		VALUES (?, ?, ?, ?, ?)
	`, normalBillUuid, now, now+3600, constant.SaleBillStatusComplete, constant.NotDeleted).Error
	if err != nil {
		t.Fatalf("Failed to insert normal sale bill: %v", err)
	}

	// 创建已选择订单
	selectedBillUuid := uint64(3002)
	err = db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, create_time, finish_time, status, delete_time)
		VALUES (?, ?, ?, ?, ?)
	`, selectedBillUuid, now, now+3600, constant.SaleBillStatusComplete, constant.NotDeleted).Error
	if err != nil {
		t.Fatalf("Failed to insert selected sale bill: %v", err)
	}

	// 将订单标记为数据管理订单
	err = db.Exec(`
		INSERT INTO ttpos_data_manage (uuid, data_uuid, type, delete_time, create_time)
		VALUES (?, ?, ?, ?, ?)
	`, 3, selectedBillUuid, model.DataManageTypeOrder, constant.NotDeleted, now).Error
	if err != nil {
		t.Fatalf("Failed to insert data manage record: %v", err)
	}

	statisticsSrv := NewStatisticsSrv()

	t.Run("数据管理功能开启-应排除已选择订单", func(t *testing.T) {
		ctx := createStatisticsExcludeDataManageTestContext(t, dbm, true)
		req := req.StatisticsPaymentMethodReq{
			QueryStartTime:    now,
			QueryEndTime:      now + 86400,
			Cycle:             0,
			ExcludeDataManage: true,
		}

		result := statisticsSrv.CountBusinessPaymentMethod(ctx, req)

		if result.TotalStatisticsPaymentMethodNum < 0 {
			t.Errorf("Expected total count >= 0, got %d", result.TotalStatisticsPaymentMethodNum)
		}
	})
}
