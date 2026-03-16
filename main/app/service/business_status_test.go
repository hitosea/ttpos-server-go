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

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// mockBusinessStatusSettingSrv 满足 ISrv 接口
type mockBusinessStatusSettingSrv struct {
	settingSrv.ISrv
}

func setupBusinessStatusTestDB(t *testing.T) *database.DBManager {
	t.Helper()

	utils.InitIdGenerator()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	for _, ddl := range []string{
		`DROP TABLE IF EXISTS ttpos_business_status_period`,
		`CREATE TABLE ttpos_business_status_period (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			start_time INTEGER DEFAULT 0,
			end_time INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0
		)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("Failed to execute DDL: %v", err)
		}
	}

	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	dbmValue := reflect.ValueOf(dbm).Elem()
	lockField := dbmValue.FieldByName("lock")
	if lockField.IsValid() {
		reflect.NewAt(lockField.Type(), unsafe.Pointer(lockField.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(&sync.Mutex{}))
	}
	lastCheckField := dbmValue.FieldByName("lastCheck")
	if lastCheckField.IsValid() {
		reflect.NewAt(lastCheckField.Type(), unsafe.Pointer(lastCheckField.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(make(map[uint64]time.Time)))
	}
	checkIntervalField := dbmValue.FieldByName("checkInterval")
	if checkIntervalField.IsValid() {
		reflect.NewAt(checkIntervalField.Type(), unsafe.Pointer(checkIntervalField.UnsafeAddr())).
			Elem().Set(reflect.ValueOf(10 * time.Second))
	}

	return dbm
}

func createBusinessStatusTestContext(t *testing.T, dbm *database.DBManager) context.Context {
	t.Helper()

	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db := dbm.GetDB(constant.MockDB)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test", nil)
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = r

	ctx := context.NewContext(
		context.WithContext(stdcontext.Background()),
		context.WithLogger(logger.Logger),
		context.WithGinContext(ginCtx),
		context.WithStaff(model.Staff{
			BaseModel: model.BaseModel{Uuid: 10001},
			IsSuper:   1,
		}),
	)
	ctx.SetDB(db)
	ctx.SetCompanyUuid(constant.MockDB)
	ctx.SetCompany(model.Company{
		BaseModel: model.BaseModel{Uuid: constant.MockDB},
	})
	ctx.SetCompanySetting(model.CompanySetting{
		BaseModel:   model.BaseModel{Uuid: constant.MockDB},
		CompanyUuid: constant.MockDB,
	})

	return ctx
}

func getCurrentBusinessStatus(db *gorm.DB) int {
	repo := repository.NewBusinessStatusPeriodRepo(db)
	period, _ := repo.GetOpenPeriod()
	if period != nil {
		return constant.BusinessStatusTest
	}
	return constant.BusinessStatusNormal
}

// TestUpdateBusinessStatus_SwitchToTest 正常营业 → 测试营业
func TestUpdateBusinessStatus_SwitchToTest(t *testing.T) {
	dbm := setupBusinessStatusTestDB(t)
	db := dbm.GetDB(constant.MockDB)
	srv := NewCompanySrv(dbm, &mockBusinessStatusSettingSrv{})

	ctx := createBusinessStatusTestContext(t, dbm)

	// 初始状态：正常营业（无未关闭时段）
	status := getCurrentBusinessStatus(db)
	if status != constant.BusinessStatusNormal {
		t.Fatalf("expected initial status Normal(2), got %d", status)
	}

	// 切换到测试营业
	err := srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
		Uuid:           constant.MockDB,
		BusinessStatus: constant.BusinessStatusTest,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证状态已变为测试营业
	status = getCurrentBusinessStatus(db)
	if status != constant.BusinessStatusTest {
		t.Errorf("expected status Test(1), got %d", status)
	}
}

// TestUpdateBusinessStatus_SwitchToNormal 测试营业 → 正常营业
func TestUpdateBusinessStatus_SwitchToNormal(t *testing.T) {
	dbm := setupBusinessStatusTestDB(t)
	db := dbm.GetDB(constant.MockDB)
	srv := NewCompanySrv(dbm, &mockBusinessStatusSettingSrv{})

	ctx := createBusinessStatusTestContext(t, dbm)

	// 先插入一条未关闭的测试时段
	db.Exec(`INSERT INTO ttpos_business_status_period (uuid, start_time, end_time, create_time, update_time, delete_time)
		VALUES (1001, 1000, 0, 1000, 1000, 0)`)

	status := getCurrentBusinessStatus(db)
	if status != constant.BusinessStatusTest {
		t.Fatalf("expected initial status Test(1), got %d", status)
	}

	// 切换到正常营业
	err := srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
		Uuid:           constant.MockDB,
		BusinessStatus: constant.BusinessStatusNormal,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证状态已变为正常营业
	status = getCurrentBusinessStatus(db)
	if status != constant.BusinessStatusNormal {
		t.Errorf("expected status Normal(2), got %d", status)
	}

	// 验证时段已关闭（end_time > 0）
	var period model.BusinessStatusPeriod
	db.Where("uuid = 1001").First(&period)
	if period.EndTime == 0 {
		t.Error("expected end_time to be set after switching to normal")
	}
}

// TestUpdateBusinessStatus_Idempotent 重复切换同一状态不报错
func TestUpdateBusinessStatus_Idempotent(t *testing.T) {
	dbm := setupBusinessStatusTestDB(t)
	srv := NewCompanySrv(dbm, &mockBusinessStatusSettingSrv{})

	ctx := createBusinessStatusTestContext(t, dbm)

	// 初始为正常营业，再次设置为正常营业
	err := srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
		Uuid:           constant.MockDB,
		BusinessStatus: constant.BusinessStatusNormal,
	})
	if err != nil {
		t.Fatalf("idempotent switch to Normal should not error: %v", err)
	}
}

// TestUpdateBusinessStatus_RoundTrip 完整来回切换
func TestUpdateBusinessStatus_RoundTrip(t *testing.T) {
	dbm := setupBusinessStatusTestDB(t)
	db := dbm.GetDB(constant.MockDB)
	srv := NewCompanySrv(dbm, &mockBusinessStatusSettingSrv{})

	ctx := createBusinessStatusTestContext(t, dbm)

	// 正常 → 测试
	err := srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
		Uuid:           constant.MockDB,
		BusinessStatus: constant.BusinessStatusTest,
	})
	if err != nil {
		t.Fatalf("switch to Test failed: %v", err)
	}
	if getCurrentBusinessStatus(db) != constant.BusinessStatusTest {
		t.Error("expected Test after first switch")
	}

	// 测试 → 正常
	err = srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
		Uuid:           constant.MockDB,
		BusinessStatus: constant.BusinessStatusNormal,
	})
	if err != nil {
		t.Fatalf("switch to Normal failed: %v", err)
	}
	if getCurrentBusinessStatus(db) != constant.BusinessStatusNormal {
		t.Error("expected Normal after second switch")
	}

	// 正常 → 测试（再次）
	err = srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
		Uuid:           constant.MockDB,
		BusinessStatus: constant.BusinessStatusTest,
	})
	if err != nil {
		t.Fatalf("switch to Test again failed: %v", err)
	}
	if getCurrentBusinessStatus(db) != constant.BusinessStatusTest {
		t.Error("expected Test after third switch")
	}

	// 验证有 2 条已关闭时段 + 1 条未关闭时段
	var closedCount int64
	db.Model(&model.BusinessStatusPeriod{}).Where("end_time > 0 AND delete_time = 0").Count(&closedCount)
	if closedCount != 1 {
		t.Errorf("expected 1 closed period, got %d", closedCount)
	}

	var openCount int64
	db.Model(&model.BusinessStatusPeriod{}).Where("end_time = 0 AND delete_time = 0").Count(&openCount)
	if openCount != 1 {
		t.Errorf("expected 1 open period, got %d", openCount)
	}
}

// TestUpdateBusinessStatus_InvalidValues 无效值应返回错误
func TestUpdateBusinessStatus_InvalidValues(t *testing.T) {
	dbm := setupBusinessStatusTestDB(t)
	srv := NewCompanySrv(dbm, &mockBusinessStatusSettingSrv{})
	ctx := createBusinessStatusTestContext(t, dbm)

	for _, val := range []int{0, 3, -1, 99} {
		err := srv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
			Uuid:           constant.MockDB,
			BusinessStatus: val,
		})
		if err == nil {
			t.Errorf("business_status=%d should return error, got nil", val)
		}
	}
}
