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
	shop_req "ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// mockSettingSrv 嵌入 ISrv 接口满足编译，只实现测试所需的 2 个方法
type mockSettingSrv struct {
	settingSrv.ISrv
	dataManageSetting model.DataManageSetting
	updateErr         error
}

func (m *mockSettingSrv) GetDataManageSetting(ctx context.Context) model.DataManageSetting {
	return m.dataManageSetting
}

func (m *mockSettingSrv) UpdateSetting(ctx context.Context, settingKey string, values any) error {
	return m.updateErr
}

func setupDataManageTestDB(t *testing.T) *database.DBManager {
	t.Helper()

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
		`DROP TABLE IF EXISTS ttpos_data_manage`,
		`CREATE TABLE ttpos_data_manage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			type INTEGER DEFAULT 0,
			data_uuid INTEGER DEFAULT 0,
			staff_uuid INTEGER DEFAULT 0
		)`,
		`DROP TABLE IF EXISTS ttpos_staff`,
		`CREATE TABLE ttpos_staff (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			is_super INTEGER DEFAULT 0,
			has_data_permission INTEGER DEFAULT 0
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

func createDataManageTestContext(t *testing.T, dbm *database.DBManager) context.Context {
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
		BaseModel:            model.BaseModel{Uuid: constant.MockDB},
		CompanyUuid:          constant.MockDB,
		EnableDataManagement: 1,
	})

	return ctx
}

func getDataManageUuids(db *gorm.DB) []uint64 {
	var uuids []uint64
	db.Model(&model.DataManage{}).Where("delete_time = 0").Pluck("data_uuid", &uuids)
	return uuids
}

func newTestDataManageSrv(dbm *database.DBManager) IDataManageSrv {
	return NewDataManageSrvImpl(dbm, &mockSettingSrv{
		dataManageSetting: model.DataManageSetting{IsEnableDataManage: true},
	})
}

// TestSetDataManage_IncrementalAdd 从空开始添加订单
func TestSetDataManage_IncrementalAdd(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      []uint64{1001, 1002, 1003},
	})
	if err != nil {
		t.Fatalf("SetDataManage failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 3 {
		t.Errorf("Expected 3 records, got %d", len(uuids))
	}
}

// TestSetDataManage_IncrementalPartial 部分新增、部分删除、部分保留
func TestSetDataManage_IncrementalPartial(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	// 第一次：1001, 1002, 1003
	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      []uint64{1001, 1002, 1003},
	}); err != nil {
		t.Fatalf("First SetDataManage failed: %v", err)
	}

	// 第二次：1002, 1003, 1004（删除1001，新增1004，保留1002/1003）
	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      []uint64{1002, 1003, 1004},
	}); err != nil {
		t.Fatalf("Second SetDataManage failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 3 {
		t.Errorf("Expected 3 records, got %d", len(uuids))
	}

	uuidSet := make(map[uint64]struct{})
	for _, uid := range uuids {
		uuidSet[uid] = struct{}{}
	}
	if _, ok := uuidSet[1001]; ok {
		t.Error("data_uuid 1001 should have been deleted")
	}
	if _, ok := uuidSet[1004]; !ok {
		t.Error("data_uuid 1004 should have been added")
	}
}

// TestSetDataManage_IncrementalNoChange 数据完全相同时记录不被重建
func TestSetDataManage_IncrementalNoChange(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      []uint64{1001, 1002},
	}); err != nil {
		t.Fatalf("First SetDataManage failed: %v", err)
	}

	var originalUuids []uint64
	db.Model(&model.DataManage{}).Where("delete_time = 0").Order("data_uuid").Pluck("uuid", &originalUuids)

	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      []uint64{1001, 1002},
	}); err != nil {
		t.Fatalf("Second SetDataManage failed: %v", err)
	}

	var newUuids []uint64
	db.Model(&model.DataManage{}).Where("delete_time = 0").Order("data_uuid").Pluck("uuid", &newUuids)

	if len(newUuids) != len(originalUuids) {
		t.Fatalf("Record count changed: %d -> %d", len(originalUuids), len(newUuids))
	}
	for i := range originalUuids {
		if originalUuids[i] != newUuids[i] {
			t.Errorf("Record uuid changed at index %d: %d -> %d (unnecessarily recreated)", i, originalUuids[i], newUuids[i])
		}
	}
}

// TestSetDataManage_BatchDelete 验证超过 1000 条的批量删除分批逻辑
func TestSetDataManage_BatchDelete(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	// 第一次：插入 2500 条
	uuids := make([]uint64, 2500)
	for i := range uuids {
		uuids[i] = uint64(i + 1)
	}
	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      uuids,
	}); err != nil {
		t.Fatalf("First SetDataManage failed: %v", err)
	}

	count := getDataManageUuids(db)
	if len(count) != 2500 {
		t.Fatalf("Expected 2500 records, got %d", len(count))
	}

	// 第二次：只保留最后 100 条，删除前 2400 条（触发 3 批删除：1000+1000+400）
	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      uuids[2400:], // 保留 2401~2500
	}); err != nil {
		t.Fatalf("Second SetDataManage failed: %v", err)
	}

	remaining := getDataManageUuids(db)
	if len(remaining) != 100 {
		t.Errorf("Expected 100 remaining records, got %d", len(remaining))
	}

	// 验证保留的都是 2401~2500
	for _, uid := range remaining {
		if uid < 2401 || uid > 2500 {
			t.Errorf("Unexpected data_uuid %d, expected range 2401~2500", uid)
			break
		}
	}
}

// TestSetDataManage_ClearAll 清空所有订单
func TestSetDataManage_ClearAll(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: true,
		SaleBillUuids:      []uint64{1001, 1002, 1003},
	}); err != nil {
		t.Fatalf("First SetDataManage failed: %v", err)
	}

	if err := srv.SetDataManage(ctx, shop_req.SetDataManageReq{
		IsEnableDataManage: false,
		SaleBillUuids:      []uint64{},
	}); err != nil {
		t.Fatalf("Clear SetDataManage failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 0 {
		t.Errorf("Expected 0 records after clear, got %d", len(uuids))
	}
}
