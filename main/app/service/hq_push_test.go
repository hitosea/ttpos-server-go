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
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	testHqUuid    uint64 = 100
	testStoreUuid uint64 = 200
)

var hqPushTestOnce sync.Once

// ========== Setup Helpers ==========

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func createTestDBManager(t *testing.T, dbs map[uint64]*gorm.DB) *database.DBManager {
	t.Helper()
	dbm := &database.DBManager{}
	dbmVal := reflect.ValueOf(dbm).Elem()

	// inject dbs map
	dbsField := dbmVal.FieldByName("dbs")
	reflect.NewAt(dbsField.Type(), unsafe.Pointer(dbsField.UnsafeAddr())).Elem().Set(reflect.ValueOf(dbs))

	// inject lock
	lockField := dbmVal.FieldByName("lock")
	reflect.NewAt(lockField.Type(), unsafe.Pointer(lockField.UnsafeAddr())).Elem().Set(reflect.ValueOf(&sync.Mutex{}))

	// inject lastCheck
	lcField := dbmVal.FieldByName("lastCheck")
	reflect.NewAt(lcField.Type(), unsafe.Pointer(lcField.UnsafeAddr())).Elem().Set(reflect.ValueOf(make(map[uint64]time.Time)))

	// inject checkInterval
	ciField := dbmVal.FieldByName("checkInterval")
	reflect.NewAt(ciField.Type(), unsafe.Pointer(ciField.UnsafeAddr())).Elem().Set(reflect.ValueOf(10 * time.Second))

	return dbm
}

func createSaasDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := openTestDB(t)
	db.Exec(`CREATE TABLE ttpos_company (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		name TEXT DEFAULT '',
		logo TEXT DEFAULT '',
		expire_time INTEGER DEFAULT 0,
		auth_day INTEGER DEFAULT 0,
		status INTEGER DEFAULT 0,
		auth_start_time INTEGER DEFAULT 0,
		old_company_id INTEGER DEFAULT 0,
		is_enable_erp INTEGER DEFAULT 0,
		last_sync_time INTEGER DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_company_setting (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		company_uuid INTEGER DEFAULT 0,
		real_name TEXT DEFAULT '',
		link_name TEXT DEFAULT '',
		link_phone TEXT DEFAULT '',
		sale_stock INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		erpnext_site_code TEXT DEFAULT '',
		erpnext_company_abbr TEXT DEFAULT '',
		erpnext_headquarter_abbr TEXT DEFAULT ''
	)`)
	return db
}

func seedHqAndStore(t *testing.T, saasDB *gorm.DB) {
	t.Helper()
	// HQ company
	saasDB.Exec("INSERT INTO ttpos_company (uuid, name, status, delete_time) VALUES (?, 'HQ', 1, 0)", testHqUuid)
	saasDB.Exec("INSERT INTO ttpos_company_setting (uuid, company_uuid, headquarter_uuid, erpnext_site_code, erpnext_company_abbr, erpnext_headquarter_abbr, delete_time) VALUES (1001, ?, 0, 'hq-site', 'HQ', 'HQ', 0)", testHqUuid)
	// Store company
	saasDB.Exec("INSERT INTO ttpos_company (uuid, name, status, delete_time) VALUES (?, 'Store', 1, 0)", testStoreUuid)
	saasDB.Exec("INSERT INTO ttpos_company_setting (uuid, company_uuid, headquarter_uuid, erpnext_site_code, erpnext_company_abbr, erpnext_headquarter_abbr, delete_time) VALUES (1002, ?, ?, 'store-site', 'STORE', 'HQ', 0)", testStoreUuid, testHqUuid)
}

func createShopDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := openTestDB(t)
	db.Exec(`CREATE TABLE ttpos_hq_control_setting (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		field_type TEXT DEFAULT '' UNIQUE,
		control_mode INTEGER DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_hq_field_override (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		entity_type TEXT DEFAULT '',
		entity_uuid INTEGER DEFAULT 0,
		field_type TEXT DEFAULT ''
	)`)
	db.Exec(`CREATE TABLE ttpos_product_package (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		name TEXT DEFAULT '',
		status INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		price REAL DEFAULT 0,
		category_uuid INTEGER DEFAULT 0,
		unit_uuid INTEGER DEFAULT 0,
		dine_tax_uuid INTEGER DEFAULT 0,
		takeout_tax_uuid INTEGER DEFAULT 0,
		image_file_uuid INTEGER DEFAULT 0,
		image_name TEXT DEFAULT '',
		image_url TEXT DEFAULT '',
		product_type INTEGER DEFAULT 0,
		sauce_required INTEGER DEFAULT 0,
		sauce_min_selection INTEGER DEFAULT 0,
		sauce_max_selection INTEGER DEFAULT 0,
		sort INTEGER DEFAULT 0,
		limit_num INTEGER DEFAULT 0,
		describe TEXT DEFAULT '',
		describe_multi_language_name_uuid INTEGER DEFAULT 0,
		multi_language_name_uuid INTEGER DEFAULT 0,
		is_batch INTEGER DEFAULT 0,
		product_label_uuid INTEGER DEFAULT 0,
		open_discount INTEGER DEFAULT 0,
		open_overall_discount INTEGER DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_product_package_takeout (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		product_package_uuid INTEGER DEFAULT 0,
		multi_language_name_uuid INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		name TEXT DEFAULT '',
		product_type INTEGER DEFAULT 0,
		price REAL DEFAULT 0,
		takeout_type INTEGER DEFAULT 0,
		status INTEGER DEFAULT 0,
		category_uuid INTEGER DEFAULT 0,
		special_category_uuid INTEGER DEFAULT 0,
		image_file_uuid INTEGER DEFAULT 0,
		describe TEXT DEFAULT '',
		describe_multi_language_name_uuid INTEGER DEFAULT 0,
		source TEXT DEFAULT '',
		source_product_id TEXT DEFAULT ''
	)`)
	db.Exec(`CREATE TABLE ttpos_product_bom_takeout (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		product_package_takeout_uuid INTEGER DEFAULT 0,
		product_bom_uuid INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		price REAL DEFAULT 0,
		grab_modifier_id TEXT DEFAULT ''
	)`)
	db.Exec(`CREATE TABLE ttpos_product_package_attribute_takeout (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		product_package_takeout_uuid INTEGER DEFAULT 0,
		product_package_attribute_uuid INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		price REAL DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_product_package_group_item_takeout (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		product_package_takeout_uuid INTEGER DEFAULT 0,
		product_package_group_item_uuid INTEGER DEFAULT 0,
		product_package_group_uuid INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		add_price REAL DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_material (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		name TEXT DEFAULT '',
		code TEXT DEFAULT '',
		multi_language_name_uuid INTEGER DEFAULT 0,
		category_uuid INTEGER DEFAULT 0,
		supplier_uuid INTEGER DEFAULT 0,
		status INTEGER DEFAULT 0,
		barcode_value TEXT DEFAULT '',
		internal_code TEXT DEFAULT '',
		origin_country_code TEXT DEFAULT '',
		unit_uuid INTEGER DEFAULT 0,
		purchase_unit_uuid INTEGER DEFAULT 0,
		cost_unit_uuid INTEGER DEFAULT 0,
		default_sales_unit_uuid INTEGER DEFAULT 0,
		safety_stock REAL,
		allow_negative_stock INTEGER DEFAULT 0,
		headquarter_uuid INTEGER DEFAULT 0,
		warehouse_uuid INTEGER DEFAULT 0,
		allow_substore_visible INTEGER DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_material_unit (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		update_time INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		name TEXT DEFAULT '',
		unit_uuid INTEGER DEFAULT 0,
		conversion_rate REAL DEFAULT 1,
		from_unit_uuid INTEGER DEFAULT 0,
		is_default INTEGER DEFAULT 0,
		material_uuid INTEGER DEFAULT 0
	)`)
	db.Exec(`CREATE TABLE ttpos_product_unit (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid INTEGER DEFAULT 0,
		name TEXT DEFAULT ''
	)`)
	return db
}

func setupHqPushTest(t *testing.T) (IHqPushSrv, *gorm.DB, *gorm.DB, *gorm.DB) {
	t.Helper()

	hqPushTestOnce.Do(func() {
		utils.InitIdGenerator()
		if logger.Logger == nil {
			logger.Logger, _ = zap.NewDevelopment()
		}
	})

	config.Database.TablePrefix = "ttpos_"
	cache.Global = cache.NewCache(cache.GoCache, cache.Config{})
	t.Cleanup(func() { cache.Global = nil })

	saasDB := createSaasDB(t)
	seedHqAndStore(t, saasDB)
	hqDB := createShopDB(t)
	storeDB := createShopDB(t)

	dbm := createTestDBManager(t, map[uint64]*gorm.DB{
		constant.DefaultDB: saasDB,
		testHqUuid:         hqDB,
		testStoreUuid:      storeDB,
	})

	srv := NewHqPushSrv(dbm)
	return srv, hqDB, storeDB, saasDB
}

func createHqTestContext(companyUuid uint64, isHq bool, db *gorm.DB) context.Context {
	cs := model.CompanySetting{
		ErpnextSiteCode: "hq-site",
	}
	if isHq {
		cs.ErpnextCompanyAbbr = "HQ"
		cs.ErpnextHeadquarterAbbr = "HQ"
	} else {
		cs.ErpnextCompanyAbbr = "STORE"
		cs.ErpnextHeadquarterAbbr = "HQ"
	}

	testLogger, _ := zap.NewDevelopment()
	return context.NewContext(
		context.WithContext(stdcontext.Background()),
		context.WithDB(db),
		context.WithCompanyUuid(companyUuid),
		context.WithCompanySetting(cs),
		context.WithLogger(testLogger),
	)
}

// ========== Data Seeders ==========

func seedHqProduct(t *testing.T, hqDB *gorm.DB, uuid uint64, status uint) {
	t.Helper()
	hqDB.Exec("INSERT INTO ttpos_product_package (uuid, status, headquarter_uuid, delete_time) VALUES (?, ?, 0, 0)", uuid, status)
}

func seedStoreProduct(t *testing.T, storeDB *gorm.DB, uuid uint64, status uint) {
	t.Helper()
	storeDB.Exec("INSERT INTO ttpos_product_package (uuid, status, headquarter_uuid, delete_time) VALUES (?, ?, ?, 0)", uuid, status, testHqUuid)
}

func seedOverride(t *testing.T, storeDB *gorm.DB, entityUuid uint64, entityType, fieldType string) {
	t.Helper()
	repo := repository.NewHqFieldOverrideRepo(storeDB)
	if err := repo.MarkOverridden(entityUuid, entityType, fieldType); err != nil {
		t.Fatalf("seedOverride: %v", err)
	}
}

func seedHqTakeout(t *testing.T, hqDB *gorm.DB, uuid uint64, status uint, price float64) {
	t.Helper()
	hqDB.Exec("INSERT INTO ttpos_product_package_takeout (uuid, status, price, headquarter_uuid, delete_time) VALUES (?, ?, ?, 0, 0)", uuid, status, price)
}

func seedStoreTakeout(t *testing.T, storeDB *gorm.DB, uuid uint64, status uint, price float64) {
	t.Helper()
	storeDB.Exec("INSERT INTO ttpos_product_package_takeout (uuid, status, price, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, 0)", uuid, status, price, testHqUuid)
}

func seedHqBomTakeout(t *testing.T, hqDB *gorm.DB, bomUuid, takeoutUuid uint64, price float64) {
	t.Helper()
	hqDB.Exec("INSERT INTO ttpos_product_bom_takeout (uuid, product_package_takeout_uuid, price, headquarter_uuid, delete_time) VALUES (?, ?, ?, 0, 0)", bomUuid, takeoutUuid, price)
}

func seedStoreBomTakeout(t *testing.T, storeDB *gorm.DB, bomUuid, takeoutUuid uint64, price float64) {
	t.Helper()
	storeDB.Exec("INSERT INTO ttpos_product_bom_takeout (uuid, product_package_takeout_uuid, price, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, 0)", bomUuid, takeoutUuid, price, testHqUuid)
}

func seedHqAttrTakeout(t *testing.T, hqDB *gorm.DB, uuid, takeoutUuid, attrUuid uint64, price float64) {
	t.Helper()
	hqDB.Exec("INSERT INTO ttpos_product_package_attribute_takeout (uuid, product_package_takeout_uuid, product_package_attribute_uuid, price, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, 0, 0)", uuid, takeoutUuid, attrUuid, price)
}

func seedStoreAttrTakeout(t *testing.T, storeDB *gorm.DB, uuid, takeoutUuid, attrUuid uint64, price float64) {
	t.Helper()
	storeDB.Exec("INSERT INTO ttpos_product_package_attribute_takeout (uuid, product_package_takeout_uuid, product_package_attribute_uuid, price, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, ?, 0)", uuid, takeoutUuid, attrUuid, price, testHqUuid)
}

func seedHqGroupItemTakeout(t *testing.T, hqDB *gorm.DB, uuid, takeoutUuid, groupUuid, groupItemUuid uint64, addPrice float64) {
	t.Helper()
	hqDB.Exec("INSERT INTO ttpos_product_package_group_item_takeout (uuid, product_package_takeout_uuid, product_package_group_uuid, product_package_group_item_uuid, add_price, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, ?, 0, 0)", uuid, takeoutUuid, groupUuid, groupItemUuid, addPrice)
}

func seedStoreGroupItemTakeout(t *testing.T, storeDB *gorm.DB, uuid, takeoutUuid, groupUuid, groupItemUuid uint64, addPrice float64) {
	t.Helper()
	storeDB.Exec("INSERT INTO ttpos_product_package_group_item_takeout (uuid, product_package_takeout_uuid, product_package_group_uuid, product_package_group_item_uuid, add_price, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, ?, ?, 0)", uuid, takeoutUuid, groupUuid, groupItemUuid, addPrice, testHqUuid)
}

func seedHqMaterial(t *testing.T, hqDB *gorm.DB, uuid uint64, allowNegStock int, safetyStock *float64) {
	t.Helper()
	if safetyStock != nil {
		hqDB.Exec("INSERT INTO ttpos_material (uuid, allow_negative_stock, safety_stock, headquarter_uuid, delete_time) VALUES (?, ?, ?, 0, 0)", uuid, allowNegStock, *safetyStock)
	} else {
		hqDB.Exec("INSERT INTO ttpos_material (uuid, allow_negative_stock, safety_stock, headquarter_uuid, delete_time) VALUES (?, ?, NULL, 0, 0)", uuid, allowNegStock)
	}
}

func seedStoreMaterial(t *testing.T, storeDB *gorm.DB, uuid uint64, allowNegStock int, safetyStock *float64) {
	t.Helper()
	if safetyStock != nil {
		storeDB.Exec("INSERT INTO ttpos_material (uuid, allow_negative_stock, safety_stock, headquarter_uuid, delete_time) VALUES (?, ?, ?, ?, 0)", uuid, allowNegStock, *safetyStock, testHqUuid)
	} else {
		storeDB.Exec("INSERT INTO ttpos_material (uuid, allow_negative_stock, safety_stock, headquarter_uuid, delete_time) VALUES (?, ?, NULL, ?, 0)", uuid, allowNegStock, testHqUuid)
	}
}

func seedHqMaterialUnit(t *testing.T, hqDB *gorm.DB, uuid, materialUuid, unitUuid uint64, conversionRate float64, name string) {
	t.Helper()
	hqDB.Exec("INSERT INTO ttpos_material_unit (uuid, material_uuid, unit_uuid, conversion_rate, name, is_default, delete_time) VALUES (?, ?, ?, ?, ?, 0, 0)", uuid, materialUuid, unitUuid, conversionRate, name)
}

func seedStoreMaterialUnit(t *testing.T, storeDB *gorm.DB, uuid, materialUuid, unitUuid uint64, conversionRate float64, name string) {
	t.Helper()
	storeDB.Exec("INSERT INTO ttpos_material_unit (uuid, material_uuid, unit_uuid, conversion_rate, name, is_default, delete_time) VALUES (?, ?, ?, ?, ?, 0, 0)", uuid, materialUuid, unitUuid, conversionRate, name)
}

// ========== Query Helpers ==========

func getStoreProductStatus(storeDB *gorm.DB, uuid uint64) uint {
	var status uint
	storeDB.Raw("SELECT status FROM ttpos_product_package WHERE uuid = ? AND delete_time = 0", uuid).Scan(&status)
	return status
}

func hasOverride(storeDB *gorm.DB, entityUuid uint64, fieldType string) bool {
	return repository.NewHqFieldOverrideRepo(storeDB).IsOverridden(entityUuid, fieldType)
}

func getStoreTakeoutStatus(storeDB *gorm.DB, uuid uint64) uint {
	var status uint
	storeDB.Raw("SELECT status FROM ttpos_product_package_takeout WHERE uuid = ? AND delete_time = 0", uuid).Scan(&status)
	return status
}

func getStoreTakeoutPrice(storeDB *gorm.DB, uuid uint64) float64 {
	var price float64
	storeDB.Raw("SELECT price FROM ttpos_product_package_takeout WHERE uuid = ? AND delete_time = 0", uuid).Scan(&price)
	return price
}

func getStoreMaterialNegStock(storeDB *gorm.DB, uuid uint64) int {
	var val int
	storeDB.Raw("SELECT allow_negative_stock FROM ttpos_material WHERE uuid = ? AND delete_time = 0", uuid).Scan(&val)
	return val
}

func floatPtr(v float64) *float64 { return &v }

// ========== Section A: Pure Function Tests ==========

func TestHqPush_GetControlModeWithDefault(t *testing.T) {
	m := map[string]int{"dine_shelf": 1}

	if got := getControlModeWithDefault(m, "dine_shelf"); got != 1 {
		t.Errorf("expected 1 for existing key, got %d", got)
	}
	if got := getControlModeWithDefault(m, "takeout_shelf"); got != constant.HqControlSeparate {
		t.Errorf("expected default 0 for missing key, got %d", got)
	}
	if got := getControlModeWithDefault(m, constant.HqFieldNegativeStock); got != constant.HqControlUnified {
		t.Errorf("expected default 1 for negative_stock, got %d", got)
	}
}

// ========== Section B: IsFieldEditable ==========

func TestHqPush_IsFieldEditable_SafetyStock_AlwaysTrue(t *testing.T) {
	srv, hqDB, _, _ := setupHqPushTest(t)

	// Set unified control for safety_stock (should be ignored)
	repository.NewHqControlSettingRepo(hqDB).Upsert(constant.HqFieldSafetyStock, constant.HqControlUnified)

	if !srv.IsFieldEditable(testHqUuid, constant.HqFieldSafetyStock) {
		t.Error("safety_stock should always be editable")
	}
}

func TestHqPush_IsFieldEditable_Unified_NotEditable(t *testing.T) {
	srv, hqDB, _, _ := setupHqPushTest(t)

	repo := repository.NewHqControlSettingRepo(hqDB)
	repo.Upsert(constant.HqFieldDineShelf, constant.HqControlUnified)
	repo.InvalidateCache(testHqUuid)

	if srv.IsFieldEditable(testHqUuid, constant.HqFieldDineShelf) {
		t.Error("dine_shelf unified should NOT be editable")
	}
}

func TestHqPush_IsFieldEditable_Separate_Editable(t *testing.T) {
	srv, hqDB, _, _ := setupHqPushTest(t)

	repo := repository.NewHqControlSettingRepo(hqDB)
	repo.Upsert(constant.HqFieldDineShelf, constant.HqControlSeparate)
	repo.InvalidateCache(testHqUuid)

	if !srv.IsFieldEditable(testHqUuid, constant.HqFieldDineShelf) {
		t.Error("dine_shelf separate should be editable")
	}
}

func TestHqPush_IsFieldEditable_DefaultNegativeStock(t *testing.T) {
	srv, _, _, _ := setupHqPushTest(t)

	// No records — negative_stock defaults to unified
	if srv.IsFieldEditable(testHqUuid, constant.HqFieldNegativeStock) {
		t.Error("negative_stock default (unified) should NOT be editable")
	}
}

// ========== Section C: Control Setting API ==========

func TestHqPush_GetControlSetting_NonHQ_Error(t *testing.T) {
	srv, _, storeDB, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testStoreUuid, false, storeDB)

	_, err := srv.GetControlSetting(ctx)
	if err == nil {
		t.Error("expected error for non-HQ")
	}
}

func TestHqPush_GetControlSetting_Defaults(t *testing.T) {
	srv, _, _, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testHqUuid, true, nil)

	result, err := srv.GetControlSetting(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HqControlDineShelf != constant.HqControlSeparate {
		t.Errorf("dine_shelf: want %d, got %d", constant.HqControlSeparate, result.HqControlDineShelf)
	}
	if result.HqControlTakeoutShelf != constant.HqControlSeparate {
		t.Errorf("takeout_shelf: want %d, got %d", constant.HqControlSeparate, result.HqControlTakeoutShelf)
	}
	if result.HqControlTakeoutPrice != constant.HqControlSeparate {
		t.Errorf("takeout_price: want %d, got %d", constant.HqControlSeparate, result.HqControlTakeoutPrice)
	}
	if result.HqControlNegativeStock != constant.HqControlUnified {
		t.Errorf("negative_stock: want %d, got %d", constant.HqControlUnified, result.HqControlNegativeStock)
	}
}

func TestHqPush_UpdateControlSetting_CacheInvalidated(t *testing.T) {
	srv, hqDB, _, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testHqUuid, true, hqDB)

	unified := 1
	err := srv.UpdateControlSetting(ctx, req.HqControlSettingUpdateReq{HqControlDineShelf: &unified})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// After update, GetControlSetting should reflect new value
	result, _ := srv.GetControlSetting(ctx)
	if result.HqControlDineShelf != constant.HqControlUnified {
		t.Errorf("after update: want %d, got %d", constant.HqControlUnified, result.HqControlDineShelf)
	}
}

// ========== Section D: Push Decision Matrix (P0) ==========

// Branch 1: ForceMode → overwrite + clear override
func TestHqPush_DineShelf_Force_OverwriteAndClear(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	productUuid := uint64(3001)
	seedHqProduct(t, hqDB, productUuid, 1)       // HQ: status=1 (上架)
	seedStoreProduct(t, storeDB, productUuid, 0) // Store: status=0 (下架)
	seedOverride(t, storeDB, productUuid, constant.HqEntityProduct, constant.HqFieldDineShelf)

	// Force push
	srv.(*hqPushSrv).pushDineShelfToStore(testHqUuid, testStoreUuid, true)

	if got := getStoreProductStatus(storeDB, productUuid); got != 1 {
		t.Errorf("store status: want 1, got %d", got)
	}
	if hasOverride(storeDB, productUuid, constant.HqFieldDineShelf) {
		t.Error("override should be cleared after force push")
	}
}

// Branch 2: Unified → overwrite + clear override
func TestHqPush_DineShelf_Unified_OverwriteAndClear(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	productUuid := uint64(3002)
	seedHqProduct(t, hqDB, productUuid, 1)
	seedStoreProduct(t, storeDB, productUuid, 0)
	seedOverride(t, storeDB, productUuid, constant.HqEntityProduct, constant.HqFieldDineShelf)

	// Set unified control
	repo := repository.NewHqControlSettingRepo(hqDB)
	repo.Upsert(constant.HqFieldDineShelf, constant.HqControlUnified)
	repo.InvalidateCache(testHqUuid)

	// isUnified=true passed as forceOverwrite
	srv.(*hqPushSrv).pushDineShelfToStore(testHqUuid, testStoreUuid, true)

	if got := getStoreProductStatus(storeDB, productUuid); got != 1 {
		t.Errorf("store status: want 1, got %d", got)
	}
	if hasOverride(storeDB, productUuid, constant.HqFieldDineShelf) {
		t.Error("override should be cleared under unified control")
	}
}

// Branch 3: Separate + no override + same value → update (sync)
func TestHqPush_DineShelf_Separate_NoOverride_Same(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	productUuid := uint64(3003)
	seedHqProduct(t, hqDB, productUuid, 1)
	seedStoreProduct(t, storeDB, productUuid, 1) // same value

	srv.(*hqPushSrv).pushDineShelfToStore(testHqUuid, testStoreUuid, false)

	if got := getStoreProductStatus(storeDB, productUuid); got != 1 {
		t.Errorf("store status: want 1, got %d", got)
	}
	if hasOverride(storeDB, productUuid, constant.HqFieldDineShelf) {
		t.Error("no override should be created for same value")
	}
}

// Branch 4: Separate + no override + different value → 直接同步（子店没改过=跟总店走）
func TestHqPush_DineShelf_Separate_NoOverride_Differ(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	productUuid := uint64(3004)
	seedHqProduct(t, hqDB, productUuid, 1)
	seedStoreProduct(t, storeDB, productUuid, 0) // different

	srv.(*hqPushSrv).pushDineShelfToStore(testHqUuid, testStoreUuid, false)

	// Store value should be synced from HQ
	if got := getStoreProductStatus(storeDB, productUuid); got != 1 {
		t.Errorf("store status: want 1 (synced), got %d", got)
	}
	// No override should be created
	if hasOverride(storeDB, productUuid, constant.HqFieldDineShelf) {
		t.Error("no override should be created — sub-store hasn't modified")
	}
}

// Branch 5: Separate + already overridden → skip
func TestHqPush_DineShelf_Separate_Overridden_Skip(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	productUuid := uint64(3005)
	seedHqProduct(t, hqDB, productUuid, 1)
	seedStoreProduct(t, storeDB, productUuid, 0)
	seedOverride(t, storeDB, productUuid, constant.HqEntityProduct, constant.HqFieldDineShelf)

	srv.(*hqPushSrv).pushDineShelfToStore(testHqUuid, testStoreUuid, false)

	// Store value unchanged
	if got := getStoreProductStatus(storeDB, productUuid); got != 0 {
		t.Errorf("store status: want 0 (skipped), got %d", got)
	}
	// Override still present
	if !hasOverride(storeDB, productUuid, constant.HqFieldDineShelf) {
		t.Error("override should remain after skip")
	}
}

// Edge: product not in store → no error
func TestHqPush_DineShelf_ProductNotInStore_Skip(t *testing.T) {
	srv, hqDB, _, _ := setupHqPushTest(t)

	productUuid := uint64(3006)
	seedHqProduct(t, hqDB, productUuid, 1)
	// No store product seeded

	// Should not panic
	srv.(*hqPushSrv).pushDineShelfToStore(testHqUuid, testStoreUuid, true)
}

// ========== Section E: Takeout Push ==========

func TestHqPush_TakeoutShelf_Force(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(4001)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 10.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 0, 10.0)
	seedOverride(t, storeDB, takeoutUuid, constant.HqEntityProductTakeout, constant.HqFieldTakeoutShelf)

	srv.(*hqPushSrv).pushTakeoutShelfToStore(testHqUuid, testStoreUuid, true)

	if got := getStoreTakeoutStatus(storeDB, takeoutUuid); got != 1 {
		t.Errorf("takeout status: want 1, got %d", got)
	}
	if hasOverride(storeDB, takeoutUuid, constant.HqFieldTakeoutShelf) {
		t.Error("override should be cleared after force push")
	}
}

func TestHqPush_TakeoutPrice_Force(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(4002)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 25.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 1, 15.0)

	// Also seed bom takeout
	bomUuid := uint64(5001)
	seedHqBomTakeout(t, hqDB, bomUuid, takeoutUuid, 30.0)
	seedStoreBomTakeout(t, storeDB, bomUuid, takeoutUuid, 20.0)

	srv.(*hqPushSrv).pushTakeoutPriceToStore(testHqUuid, testStoreUuid, true)

	if got := getStoreTakeoutPrice(storeDB, takeoutUuid); got != 25.0 {
		t.Errorf("takeout price: want 25.0, got %f", got)
	}
	// Check bom price
	var bomPrice float64
	storeDB.Raw("SELECT price FROM ttpos_product_bom_takeout WHERE uuid = ?", bomUuid).Scan(&bomPrice)
	if bomPrice != 30.0 {
		t.Errorf("bom price: want 30.0, got %f", bomPrice)
	}
}

// ========== Section E2: Takeout Realtime Push (Associated Tables) ==========

// BOM price synced when main price is synced (no override)
func TestHqPush_TakeoutRealtime_SyncsBomPrice_WhenPriceSynced(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(7001)
	bomUuid := uint64(7101)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 50.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 1, 30.0)
	seedHqBomTakeout(t, hqDB, bomUuid, takeoutUuid, 60.0)
	seedStoreBomTakeout(t, storeDB, bomUuid, takeoutUuid, 40.0)

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	hqTakeoutRepo := repository.NewProductPackageTakeoutRepo(hqDB2)
	hqTakeout, _ := hqTakeoutRepo.GetProductPackageTakeout(
		repository.NewCommonRepo().WhereByUuid(takeoutUuid),
		repository.NewCommonRepo().WhereByHeadquarterUuid(0),
		hqTakeoutRepo.WithProductBomTakeouts(),
		hqTakeoutRepo.WithProductPackageAttributeTakeouts(),
		hqTakeoutRepo.WithProductPackageGroupItemTakeouts(),
	)
	srv.(*hqPushSrv).pushSingleTakeoutToStore(testHqUuid, testStoreUuid, hqTakeout, controlRepo)

	// Main price should be synced (no override)
	if got := getStoreTakeoutPrice(storeDB, takeoutUuid); got != 50.0 {
		t.Errorf("takeout price: want 50.0, got %f", got)
	}
	// BOM price should also be synced
	var bomPrice float64
	storeDB.Raw("SELECT price FROM ttpos_product_bom_takeout WHERE uuid = ?", bomUuid).Scan(&bomPrice)
	if bomPrice != 60.0 {
		t.Errorf("bom price: want 60.0, got %f", bomPrice)
	}
}

// BOM price NOT synced when main price is overridden
func TestHqPush_TakeoutRealtime_SkipsBomPrice_WhenPriceOverridden(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(7002)
	bomUuid := uint64(7102)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 50.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 1, 30.0)
	seedHqBomTakeout(t, hqDB, bomUuid, takeoutUuid, 60.0)
	seedStoreBomTakeout(t, storeDB, bomUuid, takeoutUuid, 40.0)
	// Override takeout price → main price won't sync → BOM price won't sync
	seedOverride(t, storeDB, takeoutUuid, constant.HqEntityProductTakeout, constant.HqFieldTakeoutPrice)

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	hqTakeoutRepo := repository.NewProductPackageTakeoutRepo(hqDB2)
	hqTakeout, _ := hqTakeoutRepo.GetProductPackageTakeout(
		repository.NewCommonRepo().WhereByUuid(takeoutUuid),
		repository.NewCommonRepo().WhereByHeadquarterUuid(0),
		hqTakeoutRepo.WithProductBomTakeouts(),
		hqTakeoutRepo.WithProductPackageAttributeTakeouts(),
		hqTakeoutRepo.WithProductPackageGroupItemTakeouts(),
	)
	srv.(*hqPushSrv).pushSingleTakeoutToStore(testHqUuid, testStoreUuid, hqTakeout, controlRepo)

	// Main price should NOT be synced (overridden)
	if got := getStoreTakeoutPrice(storeDB, takeoutUuid); got != 30.0 {
		t.Errorf("takeout price: want 30.0 (preserved), got %f", got)
	}
	// BOM price should NOT be synced either
	var bomPrice float64
	storeDB.Raw("SELECT price FROM ttpos_product_bom_takeout WHERE uuid = ?", bomUuid).Scan(&bomPrice)
	if bomPrice != 40.0 {
		t.Errorf("bom price: want 40.0 (preserved), got %f", bomPrice)
	}
}

// Missing BOM takeout in sub-store gets created
func TestHqPush_TakeoutRealtime_CreatesMissingBomTakeout(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(7003)
	bomUuid := uint64(7103)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 50.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 1, 50.0)
	seedHqBomTakeout(t, hqDB, bomUuid, takeoutUuid, 60.0)
	// No store bom takeout seeded — it should be created

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	hqTakeoutRepo := repository.NewProductPackageTakeoutRepo(hqDB2)
	hqTakeout, _ := hqTakeoutRepo.GetProductPackageTakeout(
		repository.NewCommonRepo().WhereByUuid(takeoutUuid),
		repository.NewCommonRepo().WhereByHeadquarterUuid(0),
		hqTakeoutRepo.WithProductBomTakeouts(),
		hqTakeoutRepo.WithProductPackageAttributeTakeouts(),
		hqTakeoutRepo.WithProductPackageGroupItemTakeouts(),
	)
	srv.(*hqPushSrv).pushSingleTakeoutToStore(testHqUuid, testStoreUuid, hqTakeout, controlRepo)

	// New BOM takeout record should be created
	var bomPrice float64
	storeDB.Raw("SELECT price FROM ttpos_product_bom_takeout WHERE uuid = ?", bomUuid).Scan(&bomPrice)
	if bomPrice != 60.0 {
		t.Errorf("created bom price: want 60.0, got %f", bomPrice)
	}
	var hqUuid uint64
	storeDB.Raw("SELECT headquarter_uuid FROM ttpos_product_bom_takeout WHERE uuid = ?", bomUuid).Scan(&hqUuid)
	if hqUuid != testHqUuid {
		t.Errorf("created bom headquarter_uuid: want %d, got %d", testHqUuid, hqUuid)
	}
}

// Attribute takeout sync: existing updated, missing created
func TestHqPush_TakeoutRealtime_SyncsAttrTakeout(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(7004)
	attrUuid1 := uint64(7201)
	attrUuid2 := uint64(7202)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 50.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 1, 50.0)
	// attr1: exists in both → should be updated
	seedHqAttrTakeout(t, hqDB, attrUuid1, takeoutUuid, 8001, 5.0)
	seedStoreAttrTakeout(t, storeDB, attrUuid1, takeoutUuid, 8001, 3.0)
	// attr2: only in HQ → should be created
	seedHqAttrTakeout(t, hqDB, attrUuid2, takeoutUuid, 8002, 7.0)

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	hqTakeoutRepo := repository.NewProductPackageTakeoutRepo(hqDB2)
	hqTakeout, _ := hqTakeoutRepo.GetProductPackageTakeout(
		repository.NewCommonRepo().WhereByUuid(takeoutUuid),
		repository.NewCommonRepo().WhereByHeadquarterUuid(0),
		hqTakeoutRepo.WithProductBomTakeouts(),
		hqTakeoutRepo.WithProductPackageAttributeTakeouts(),
		hqTakeoutRepo.WithProductPackageGroupItemTakeouts(),
	)
	srv.(*hqPushSrv).pushSingleTakeoutToStore(testHqUuid, testStoreUuid, hqTakeout, controlRepo)

	// attr1 should still exist (timestamps synced)
	var count1 int64
	storeDB.Raw("SELECT COUNT(*) FROM ttpos_product_package_attribute_takeout WHERE uuid = ?", attrUuid1).Scan(&count1)
	if count1 != 1 {
		t.Errorf("attr1 count: want 1, got %d", count1)
	}
	// attr2 should be created
	var price2 float64
	storeDB.Raw("SELECT price FROM ttpos_product_package_attribute_takeout WHERE uuid = ?", attrUuid2).Scan(&price2)
	if price2 != 7.0 {
		t.Errorf("created attr2 price: want 7.0, got %f", price2)
	}
}

// GroupItem takeout sync: existing updated, missing created
func TestHqPush_TakeoutRealtime_SyncsGroupItemTakeout(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	takeoutUuid := uint64(7005)
	giUuid1 := uint64(7301)
	giUuid2 := uint64(7302)
	groupUuid := uint64(9001)
	itemUuid1 := uint64(9101)
	itemUuid2 := uint64(9102)
	seedHqTakeout(t, hqDB, takeoutUuid, 1, 50.0)
	seedStoreTakeout(t, storeDB, takeoutUuid, 1, 50.0)
	// gi1: exists in both → should be updated
	seedHqGroupItemTakeout(t, hqDB, giUuid1, takeoutUuid, groupUuid, itemUuid1, 2.5)
	seedStoreGroupItemTakeout(t, storeDB, giUuid1, takeoutUuid, groupUuid, itemUuid1, 1.5)
	// gi2: only in HQ → should be created
	seedHqGroupItemTakeout(t, hqDB, giUuid2, takeoutUuid, groupUuid, itemUuid2, 3.5)

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	hqTakeoutRepo := repository.NewProductPackageTakeoutRepo(hqDB2)
	hqTakeout, _ := hqTakeoutRepo.GetProductPackageTakeout(
		repository.NewCommonRepo().WhereByUuid(takeoutUuid),
		repository.NewCommonRepo().WhereByHeadquarterUuid(0),
		hqTakeoutRepo.WithProductBomTakeouts(),
		hqTakeoutRepo.WithProductPackageAttributeTakeouts(),
		hqTakeoutRepo.WithProductPackageGroupItemTakeouts(),
	)
	srv.(*hqPushSrv).pushSingleTakeoutToStore(testHqUuid, testStoreUuid, hqTakeout, controlRepo)

	// gi1 should still exist
	var count1 int64
	storeDB.Raw("SELECT COUNT(*) FROM ttpos_product_package_group_item_takeout WHERE uuid = ?", giUuid1).Scan(&count1)
	if count1 != 1 {
		t.Errorf("gi1 count: want 1, got %d", count1)
	}
	// gi2 should be created
	var addPrice2 float64
	storeDB.Raw("SELECT add_price FROM ttpos_product_package_group_item_takeout WHERE uuid = ?", giUuid2).Scan(&addPrice2)
	if addPrice2 != 3.5 {
		t.Errorf("created gi2 add_price: want 3.5, got %f", addPrice2)
	}
}

// ========== Section F: Material Push ==========

func TestHqPush_NegativeStock_DefaultUnified_ActsAsForce(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6001)
	seedHqMaterial(t, hqDB, matUuid, 1, nil)
	seedStoreMaterial(t, storeDB, matUuid, 0, nil)

	// No control setting record → negative_stock defaults to unified → forceOverwrite=true
	srv.(*hqPushSrv).pushNegativeStockToStore(testHqUuid, testStoreUuid, true)

	if got := getStoreMaterialNegStock(storeDB, matUuid); got != 1 {
		t.Errorf("negative_stock: want 1, got %d", got)
	}
}

func TestHqPush_NegativeStock_Separate_Overridden_Skip(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6002)
	seedHqMaterial(t, hqDB, matUuid, 1, nil)
	seedStoreMaterial(t, storeDB, matUuid, 0, nil)
	seedOverride(t, storeDB, matUuid, constant.HqEntityMaterial, constant.HqFieldNegativeStock)

	// Set separate control
	ctrlRepo := repository.NewHqControlSettingRepo(hqDB)
	ctrlRepo.Upsert(constant.HqFieldNegativeStock, constant.HqControlSeparate)
	ctrlRepo.InvalidateCache(testHqUuid)

	srv.(*hqPushSrv).pushNegativeStockToStore(testHqUuid, testStoreUuid, false)

	// Store value unchanged (overridden → skip)
	if got := getStoreMaterialNegStock(storeDB, matUuid); got != 0 {
		t.Errorf("negative_stock: want 0 (skipped), got %d", got)
	}
}

func TestHqPush_NegativeStock_Separate_NoOverride_Differ_Syncs(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6003)
	seedHqMaterial(t, hqDB, matUuid, 1, nil)
	seedStoreMaterial(t, storeDB, matUuid, 0, nil) // different

	ctrlRepo := repository.NewHqControlSettingRepo(hqDB)
	ctrlRepo.Upsert(constant.HqFieldNegativeStock, constant.HqControlSeparate)
	ctrlRepo.InvalidateCache(testHqUuid)

	srv.(*hqPushSrv).pushNegativeStockToStore(testHqUuid, testStoreUuid, false)

	// Should sync HQ value
	if got := getStoreMaterialNegStock(storeDB, matUuid); got != 1 {
		t.Errorf("negative_stock: want 1 (synced), got %d", got)
	}
	// No override created
	if hasOverride(storeDB, matUuid, constant.HqFieldNegativeStock) {
		t.Error("no override should be created — sub-store hasn't modified")
	}
}

func TestHqPush_SafetyStock_NoOverride_Differ_Syncs(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6004)
	seedHqMaterial(t, hqDB, matUuid, 0, floatPtr(10.0))
	seedStoreMaterial(t, storeDB, matUuid, 0, floatPtr(5.0)) // different

	ctrlRepo := repository.NewHqControlSettingRepo(hqDB)
	ctrlRepo.InvalidateCache(testHqUuid)

	hqMaterial := model.Material{SafetyStock: floatPtr(10.0), AllowNegativeStock: 0}
	hqMaterial.Uuid = matUuid

	srv.(*hqPushSrv).pushSingleMaterialToStore(
		testHqUuid, testStoreUuid, &hqMaterial,
		repository.NewHqControlSettingRepo(hqDB),
	)

	// No override — sub-store hasn't modified, should sync
	if hasOverride(storeDB, matUuid, constant.HqFieldSafetyStock) {
		t.Error("no override should be created — sub-store hasn't modified")
	}
}

func TestHqPush_SafetyStock_NoOverride_Same_Updates(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6005)
	seedHqMaterial(t, hqDB, matUuid, 0, floatPtr(10.0))
	seedStoreMaterial(t, storeDB, matUuid, 0, floatPtr(10.0)) // same

	hqMaterial := model.Material{SafetyStock: floatPtr(10.0), AllowNegativeStock: 0}
	hqMaterial.Uuid = matUuid

	srv.(*hqPushSrv).pushSingleMaterialToStore(
		testHqUuid, testStoreUuid, &hqMaterial,
		repository.NewHqControlSettingRepo(hqDB),
	)

	if hasOverride(storeDB, matUuid, constant.HqFieldSafetyStock) {
		t.Error("no override should be created for same safety_stock")
	}
}

func TestHqPush_SafetyStock_BothNil_Equal(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6006)
	seedHqMaterial(t, hqDB, matUuid, 0, nil)
	seedStoreMaterial(t, storeDB, matUuid, 0, nil)

	hqMaterial := model.Material{SafetyStock: nil, AllowNegativeStock: 0}
	hqMaterial.Uuid = matUuid

	srv.(*hqPushSrv).pushSingleMaterialToStore(
		testHqUuid, testStoreUuid, &hqMaterial,
		repository.NewHqControlSettingRepo(hqDB),
	)

	if hasOverride(storeDB, matUuid, constant.HqFieldSafetyStock) {
		t.Error("no override for both-nil safety_stock")
	}
}

func TestHqPush_SafetyStock_OneNil_Differ_Syncs(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(6007)
	seedHqMaterial(t, hqDB, matUuid, 0, floatPtr(10.0))
	seedStoreMaterial(t, storeDB, matUuid, 0, nil) // nil vs 10.0

	hqMaterial := model.Material{SafetyStock: floatPtr(10.0), AllowNegativeStock: 0}
	hqMaterial.Uuid = matUuid

	srv.(*hqPushSrv).pushSingleMaterialToStore(
		testHqUuid, testStoreUuid, &hqMaterial,
		repository.NewHqControlSettingRepo(hqDB),
	)

	// No override — sub-store hasn't modified, should sync
	if hasOverride(storeDB, matUuid, constant.HqFieldSafetyStock) {
		t.Error("no override should be created — sub-store hasn't modified")
	}
}

// ========== Section F2: Material Realtime Push (MaterialUnit) ==========

// HQ has 2 non-base units, sub-store has 1 old one → after push, sub-store has 2 (delete+recreate)
func TestHqPush_MaterialRealtime_SyncsMaterialUnit(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(8001)
	seedHqMaterial(t, hqDB, matUuid, 0, nil)
	seedStoreMaterial(t, storeDB, matUuid, 0, nil)
	// HQ has 2 non-base units
	seedHqMaterialUnit(t, hqDB, 8101, matUuid, 501, 1000, "kg")
	seedHqMaterialUnit(t, hqDB, 8102, matUuid, 502, 500, "half-kg")
	// Sub-store has 1 old unit (will be replaced)
	seedStoreMaterialUnit(t, storeDB, 8199, matUuid, 599, 999, "old-unit")

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	commonRepo := repository.NewCommonRepo()
	hqMaterialRepo := repository.NewMaterialRepo(hqDB2)
	hqMaterial := hqMaterialRepo.GetMaterial(
		commonRepo.WhereByUuid(matUuid),
		commonRepo.WhereByHeadquarterUuid(0),
		hqMaterialRepo.WithNotBaseUnitList(commonRepo.WhereBySoftDelete()),
	)
	srv.(*hqPushSrv).pushSingleMaterialToStore(testHqUuid, testStoreUuid, &hqMaterial, controlRepo)

	// Sub-store should now have exactly 2 units (the HQ ones), old unit deleted
	var count int64
	storeDB.Raw("SELECT COUNT(*) FROM ttpos_material_unit WHERE material_uuid = ?", matUuid).Scan(&count)
	if count != 2 {
		t.Errorf("material unit count: want 2, got %d", count)
	}
	// Verify specific units exist
	var name1, name2 string
	storeDB.Raw("SELECT name FROM ttpos_material_unit WHERE uuid = ?", 8101).Scan(&name1)
	storeDB.Raw("SELECT name FROM ttpos_material_unit WHERE uuid = ?", 8102).Scan(&name2)
	if name1 != "kg" {
		t.Errorf("unit 8101 name: want 'kg', got '%s'", name1)
	}
	if name2 != "half-kg" {
		t.Errorf("unit 8102 name: want 'half-kg', got '%s'", name2)
	}
	// Old unit should be gone
	var oldCount int64
	storeDB.Raw("SELECT COUNT(*) FROM ttpos_material_unit WHERE uuid = ?", 8199).Scan(&oldCount)
	if oldCount != 0 {
		t.Errorf("old unit 8199 should be deleted, got count %d", oldCount)
	}
}

// HQ has no non-base units → sub-store's existing units should be cleared
func TestHqPush_MaterialRealtime_ClearsMaterialUnit_WhenHqHasNone(t *testing.T) {
	srv, hqDB, storeDB, _ := setupHqPushTest(t)

	matUuid := uint64(8002)
	seedHqMaterial(t, hqDB, matUuid, 0, nil)
	seedStoreMaterial(t, storeDB, matUuid, 0, nil)
	// HQ has no non-base units
	// Sub-store has 1 stale unit that should be cleared
	seedStoreMaterialUnit(t, storeDB, 8299, matUuid, 599, 100, "stale-unit")

	controlRepo := srv.(*hqPushSrv).getHqControlRepo(testHqUuid)
	hqDB2 := srv.(*hqPushSrv).dbm.GetDB(testHqUuid)
	commonRepo := repository.NewCommonRepo()
	hqMaterialRepo := repository.NewMaterialRepo(hqDB2)
	hqMaterial := hqMaterialRepo.GetMaterial(
		commonRepo.WhereByUuid(matUuid),
		commonRepo.WhereByHeadquarterUuid(0),
		hqMaterialRepo.WithNotBaseUnitList(commonRepo.WhereBySoftDelete()),
	)
	srv.(*hqPushSrv).pushSingleMaterialToStore(testHqUuid, testStoreUuid, &hqMaterial, controlRepo)

	// Sub-store should have 0 units now
	var count int64
	storeDB.Raw("SELECT COUNT(*) FROM ttpos_material_unit WHERE material_uuid = ?", matUuid).Scan(&count)
	if count != 0 {
		t.Errorf("material unit count: want 0, got %d", count)
	}
}

// ========== Section G: Batch Push Validation ==========

func TestHqPush_BatchPush_NonHQ_Error(t *testing.T) {
	srv, _, storeDB, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testStoreUuid, false, storeDB)

	_, err := srv.BatchPush(ctx, req.HqBatchPushReq{FieldTypes: []string{constant.HqFieldDineShelf}})
	if err == nil {
		t.Error("expected error for non-HQ batch push")
	}
}

func TestHqPush_BatchPush_InvalidFieldType_Error(t *testing.T) {
	srv, _, _, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testHqUuid, true, nil)

	// safety_stock is NOT a valid batch push type
	_, err := srv.BatchPush(ctx, req.HqBatchPushReq{
		FieldTypes:  []string{constant.HqFieldSafetyStock},
		IsAllStores: true,
	})
	if err == nil {
		t.Error("expected error for safety_stock batch push")
	}

	// arbitrary invalid type
	_, err = srv.BatchPush(ctx, req.HqBatchPushReq{
		FieldTypes:  []string{"full_product"},
		IsAllStores: true,
	})
	if err == nil {
		t.Error("expected error for full_product batch push")
	}
}

func TestHqPush_BatchPush_EmptyStores_Error(t *testing.T) {
	srv, _, _, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testHqUuid, true, nil)

	_, err := srv.BatchPush(ctx, req.HqBatchPushReq{
		FieldTypes:  []string{constant.HqFieldDineShelf},
		IsAllStores: false,
		StoreUuids:  []uint64{},
	})
	if err == nil {
		t.Error("expected error for empty store list")
	}
}

func TestHqPush_BatchPush_FiltersOutHqSelf(t *testing.T) {
	srv, _, _, _ := setupHqPushTest(t)

	// resolveTargetStores should filter out HQ UUID
	filtered, err := srv.(*hqPushSrv).resolveTargetStores(testHqUuid, false, []uint64{testHqUuid, testStoreUuid})
	if err != nil {
		t.Fatalf("resolveTargetStores: %v", err)
	}
	if len(filtered) != 1 || filtered[0] != testStoreUuid {
		t.Errorf("expected [%d], got %v", testStoreUuid, filtered)
	}
}

// ========== Section H: Store List ==========

func TestHqPush_GetStoreList_ExcludesHqSelf(t *testing.T) {
	srv, _, _, _ := setupHqPushTest(t)
	ctx := createHqTestContext(testHqUuid, true, nil)

	result, err := srv.GetBatchPushStoreList(ctx)
	if err != nil {
		t.Fatalf("GetBatchPushStoreList: %v", err)
	}

	for _, item := range result.List {
		if item.CompanyUuid == testHqUuid {
			t.Error("store list should not include HQ itself")
		}
	}

	// Should contain the store
	found := false
	for _, item := range result.List {
		if item.CompanyUuid == testStoreUuid {
			found = true
		}
	}
	if !found {
		t.Error("store list should include sub-store")
	}
}
