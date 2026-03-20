//go:build integration

package service

import (
	stdcontext "context"
	"fmt"
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
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// setupDiffDineInTestDB 创建测试数据库连接（SQLite 内存数据库 + ttpos_ 前缀）
func setupDiffDineInTestDB(t *testing.T) (*database.DBManager, *gorm.DB) {
	t.Helper()

	if logger.Logger == nil {
		logger.Logger, _ = zap.NewDevelopment()
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ttpos_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // SQLite 并发限制

	// 创建 product_package_attribute 表（diffDineInProducts 查询此表解析属性UUID）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_product_package_attribute (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			product_package_attribute_group_uuid INTEGER DEFAULT 0,
			attribute_uuid INTEGER DEFAULT 0,
			is_default_selected INTEGER DEFAULT 0,
			price REAL DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create product_package_attribute table: %v", err)
	}

	// 创建 product_attribute 表（Preload Attribute 需要此表存在）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_product_attribute (
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
		t.Fatalf("failed to create product_attribute table: %v", err)
	}

	// 创建 multi_language_name 表（Preload Attribute.MultiLanguageName 需要此表存在）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_multi_language_name (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT ''
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create multi_language_name table: %v", err)
	}

	// 创建 DBManager 并设置 Mock DB
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

	return dbm, db
}

// createDiffDineInTestContext 创建测试上下文
func createDiffDineInTestContext(t *testing.T, dbm *database.DBManager) context.Context {
	t.Helper()

	db := dbm.GetDB(constant.MockDB)
	testLogger := logger.Logger
	if testLogger == nil {
		testLogger, _ = zap.NewDevelopment()
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test", nil)
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = r

	ctx := context.NewContext(
		context.WithContext(stdcontext.Background()),
		context.WithLogger(testLogger),
		context.WithGinContext(ginCtx),
	)
	ctx.SetDB(db)
	ctx.SetCompanyUuid(constant.MockDB)

	company := &model.Company{
		BaseModel: model.BaseModel{Uuid: constant.MockDB},
	}
	ctx.SetCompany(*company)

	companySetting := model.CompanySetting{
		BaseModel:   model.BaseModel{Uuid: constant.MockDB},
		CompanyUuid: constant.MockDB,
	}
	ctx.SetCompanySetting(companySetting)

	return ctx
}

// newTestOrderSrv 创建一个只用于测试 diffDineInProducts 的 orderSrv 实例
func newTestOrderSrv(dbm *database.DBManager) *orderSrv {
	return &orderSrv{
		dbm: dbm,
	}
}

// makeSaleOrderProduct 构建一个用于测试的 SaleOrderProduct
func makeSaleOrderProduct(uuid uint64, flavorBomUuid uint64, productType uint8, deleteTime int64, sauceUuids []uint64, attributeUuids []uint64) *model.SaleOrderProduct {
	boms := make([]*model.SaleOrderProductBom, 0)
	// 添加 flavor bom
	boms = append(boms, &model.SaleOrderProductBom{
		BaseModel:      model.BaseModel{Uuid: flavorBomUuid},
		IsFlavorBom:    constant.ProductBomTypeFlavor, // 1 = 规格
		ProductBomUuid: flavorBomUuid,
	})
	// 添加 sauce boms
	for _, sauceUuid := range sauceUuids {
		boms = append(boms, &model.SaleOrderProductBom{
			BaseModel:      model.BaseModel{Uuid: sauceUuid},
			IsFlavorBom:    constant.ProductBomTypeSauce, // 0 = 小料
			ProductBomUuid: sauceUuid,
		})
	}

	attributes := make([]*model.SaleOrderProductAttribute, 0)
	for _, attrUuid := range attributeUuids {
		attributes = append(attributes, &model.SaleOrderProductAttribute{
			BaseModel:            model.BaseModel{Uuid: attrUuid},
			ProductAttributeUuid: attrUuid,
		})
	}

	return &model.SaleOrderProduct{
		BaseModel:                  model.BaseModel{Uuid: uuid, DeleteTime: deleteTime},
		ProductType:                productType,
		SaleOrderProductBoms:       boms,
		SaleOrderProductAttributes: attributes,
	}
}

// makePackageSaleOrderProduct 构建一个用于测试的套餐 SaleOrderProduct
// packageUuid: 套餐UUID, subProductParamsJSON: PackageSubProductParams 的 JSON 字符串
func makePackageSaleOrderProduct(uuid uint64, flavorBomUuid uint64, packageUuid uint64, subProductParamsJSON string, deleteTime int64) *model.SaleOrderProduct {
	boms := make([]*model.SaleOrderProductBom, 0)
	boms = append(boms, &model.SaleOrderProductBom{
		BaseModel:      model.BaseModel{Uuid: flavorBomUuid},
		IsFlavorBom:    constant.ProductBomTypeFlavor,
		ProductBomUuid: flavorBomUuid,
	})

	return &model.SaleOrderProduct{
		BaseModel:              model.BaseModel{Uuid: uuid, DeleteTime: deleteTime},
		ProductType:            constant.ProductTypePackage,
		ProductPackageUuid:     packageUuid,
		PackageSubProductParams: subProductParamsJSON,
		SaleOrderProductBoms:   boms,
	}
}

// makeSaleBill 构建一个用于测试的 SaleBill，包含一个 SaleOrder 和指定的 SaleOrderProducts
func makeSaleBill(products []*model.SaleOrderProduct) *model.SaleBill {
	return &model.SaleBill{
		SaleOrders: []*model.SaleOrder{
			{
				SaleOrderProducts: products,
			},
		},
	}
}

// Test_diffDineInProducts_AllNew 全新增：订单中无已有商品，提交的商品全部为新增
func Test_diffDineInProducts_AllNew(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交 2 个新商品
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0},
		{FlavorProductBomUuid: 1002, Num: 2, ProductType: 0},
	}
	// 空订单（无已有商品）
	saleBill := makeSaleBill(nil)

	addProducts, deleteProducts, updateProducts, updateOlderProducts, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addProducts) != 2 {
		t.Errorf("expected 2 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
	if len(updateOlderProducts) != 0 {
		t.Errorf("expected 0 updateOlderProducts, got %d", len(updateOlderProducts))
	}
}

// Test_diffDineInProducts_AllDelete 全删除：提交空商品列表，订单中有已有商品
func Test_diffDineInProducts_AllDelete(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 空提交
	products := []req.ProductParams{}
	// 已有 2 个商品
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, nil),
		makeSaleOrderProduct(2, 1002, constant.ProductTypeProduct, 0, nil, nil),
	})

	addProducts, deleteProducts, updateProducts, updateOlderProducts, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 2 {
		t.Errorf("expected 2 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
	if len(updateOlderProducts) != 0 {
		t.Errorf("expected 0 updateOlderProducts, got %d", len(updateOlderProducts))
	}
}

// Test_diffDineInProducts_AllUpdate 全修改：提交的商品和订单中一致
func Test_diffDineInProducts_AllUpdate(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交和已有商品 key 一致（同 FlavorProductBomUuid）
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 3, ProductType: 0},
		{FlavorProductBomUuid: 1002, Num: 5, ProductType: 0},
	}
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, nil),
		makeSaleOrderProduct(2, 1002, constant.ProductTypeProduct, 0, nil, nil),
	})

	addProducts, deleteProducts, updateProducts, updateOlderProducts, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 2 {
		t.Errorf("expected 2 updateProducts, got %d", len(updateProducts))
	}
	if len(updateOlderProducts) != 2 {
		t.Errorf("expected 2 updateOlderProducts, got %d", len(updateOlderProducts))
	}

	// 验证 updateProducts 的数量是新值
	for _, p := range updateProducts {
		if p.FlavorProductBomUuid == 1001 && p.Num != 3 {
			t.Errorf("expected updated num=3 for flavor 1001, got %.0f", p.Num)
		}
		if p.FlavorProductBomUuid == 1002 && p.Num != 5 {
			t.Errorf("expected updated num=5 for flavor 1002, got %.0f", p.Num)
		}
	}
}

// Test_diffDineInProducts_Mixed 混合场景：新增+删除+修改同时存在
func Test_diffDineInProducts_Mixed(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交: flavor 1001（修改）, 1003（新增）
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 5, ProductType: 0},
		{FlavorProductBomUuid: 1003, Num: 1, ProductType: 0},
	}
	// 已有: flavor 1001（修改）, 1002（删除）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, nil),
		makeSaleOrderProduct(2, 1002, constant.ProductTypeProduct, 0, nil, nil),
	})

	addProducts, deleteProducts, updateProducts, updateOlderProducts, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 1 {
		t.Errorf("expected 1 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
	if len(updateOlderProducts) != 1 {
		t.Errorf("expected 1 updateOlderProducts, got %d", len(updateOlderProducts))
	}

	// 验证新增的是 flavor 1003
	for _, p := range addProducts {
		if p.FlavorProductBomUuid != 1003 {
			t.Errorf("expected add product flavor 1003, got %d", p.FlavorProductBomUuid)
		}
	}

	// 验证删除的是 flavor 1002
	for _, p := range deleteProducts {
		found := false
		for _, bom := range p.SaleOrderProductBoms {
			if bom.IsFlavor() && bom.ProductBomUuid == 1002 {
				found = true
			}
		}
		if !found {
			t.Error("expected delete product with flavor 1002")
		}
	}
}

// Test_diffDineInProducts_SkipDeleted 已删除的已有商品应被跳过
func Test_diffDineInProducts_SkipDeleted(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0},
	}
	// 已有 1001 已软删除，1002 正常
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 1000, nil, nil), // 已删除
		makeSaleOrderProduct(2, 1002, constant.ProductTypeProduct, 0, nil, nil),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1001 在已有中被跳过（已删除），所以提交的 1001 应为新增
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts (1001 is new since old was deleted), got %d", len(addProducts))
	}
	// 1002 不在提交中，应为删除
	if len(deleteProducts) != 1 {
		t.Errorf("expected 1 deleteProducts (1002), got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_SkipPackageSubProduct 套餐子商品应被跳过
func Test_diffDineInProducts_SkipPackageSubProduct(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0},
	}
	// 已有：1001 普通商品 + 1002 套餐子商品（product_type=2）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, nil),
		makeSaleOrderProduct(2, 1002, constant.ProductTypePackageSubProduct, 0, nil, nil),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 套餐子商品 1002 被跳过，不参与比较
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts (flavor 1001), got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_PackageProduct 套餐商品匹配：ProductPackageUuid + 子商品签名一致时应为更新
func Test_diffDineInProducts_PackageProduct(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交一个套餐商品
	products := []req.ProductParams{
		{
			ProductPackageUuid: 2001,
			Num:                1,
			ProductType:        constant.ProductTypePackage,
			Products: []req.ProductRequest{
				{
					ProductPackageGroupUuid: 3001,
					EditProductReq:          req.EditProductReq{FlavorUuid: 4001},
					Num:                     1,
					UnitNum:                 1,
				},
			},
		},
	}
	// 已有一个相同配置的套餐商品（ProductPackageUuid 一致，子商品签名一致）
	// 子商品签名 = "3001:4001:1.00:1.00"
	subProductParams := `[{"flavor_uuid":4001,"product_package_group_uuid":3001,"num":1,"unit_num":1}]`
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makePackageSaleOrderProduct(1, 2001, 2001, subProductParams, 0),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// key 一致，应为修改
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_EmptyBoth 两边都为空：无商品提交，无已有商品
func Test_diffDineInProducts_EmptyBoth(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	products := []req.ProductParams{}
	saleBill := makeSaleBill(nil)

	addProducts, deleteProducts, updateProducts, updateOlderProducts, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
	if len(updateOlderProducts) != 0 {
		t.Errorf("expected 0 updateOlderProducts, got %d", len(updateOlderProducts))
	}
}

// Test_diffDineInProducts_EmptySaleOrders SaleBill 无 SaleOrders
func Test_diffDineInProducts_EmptySaleOrders(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0},
	}
	saleBill := &model.SaleBill{} // 无 SaleOrders

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 已有为空，提交全部为新增
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_WithSauce 带小料的商品：小料参与 key 计算
func Test_diffDineInProducts_WithSauce(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交：flavor 1001 + sauce [5001, 5002]
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, SauceProductBomUuidList: []uint64{5001, 5002}},
	}
	// 已有：flavor 1001 + sauce [5001, 5002]（相同小料）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, []uint64{5001, 5002}, nil),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// key 相同，应为修改
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_DifferentSauce 不同小料的相同商品：key 不同，应为新增+删除
func Test_diffDineInProducts_DifferentSauce(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交：flavor 1001 + sauce [5001]
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, SauceProductBomUuidList: []uint64{5001}},
	}
	// 已有：flavor 1001 + sauce [5002]（不同小料）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, []uint64{5002}, nil),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// key 不同（小料不同），应为 1 新增 + 1 删除
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 1 {
		t.Errorf("expected 1 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_WithAttribute 带属性的商品：属性通过数据库查询解析
func Test_diffDineInProducts_WithAttribute(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 在数据库中插入 product_package_attribute 记录
	err := db.Exec(`INSERT INTO ttpos_product_package_attribute (uuid, attribute_uuid, delete_time) VALUES (6001, 7001, 0)`).Error
	if err != nil {
		t.Fatalf("failed to insert product_package_attribute: %v", err)
	}
	err = db.Exec(`INSERT INTO ttpos_product_package_attribute (uuid, attribute_uuid, delete_time) VALUES (6002, 7002, 0)`).Error
	if err != nil {
		t.Fatalf("failed to insert product_package_attribute: %v", err)
	}

	// 提交：flavor 1001 + attribute [6001, 6002]（product_package_attribute UUID）
	// 查询后解析为 attribute_uuid [7001, 7002]
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, ProductPackageAttributeUuidList: []uint64{6001, 6002}},
	}
	// 已有：flavor 1001 + product_attribute_uuid [7001, 7002]（与解析后一致）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, []uint64{7001, 7002}),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// key 相同，应为修改
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_DifferentAttribute 不同属性的相同商品
func Test_diffDineInProducts_DifferentAttribute(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 插入属性记录
	err := db.Exec(`INSERT INTO ttpos_product_package_attribute (uuid, attribute_uuid, delete_time) VALUES (6001, 7001, 0)`).Error
	if err != nil {
		t.Fatalf("failed to insert product_package_attribute: %v", err)
	}

	// 提交：flavor 1001 + attribute [6001] → 解析为 [7001]
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, ProductPackageAttributeUuidList: []uint64{6001}},
	}
	// 已有：flavor 1001 + attribute [7002]（不同属性）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, []uint64{7002}),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// key 不同（属性不同），应为 1 新增 + 1 删除
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 1 {
		t.Errorf("expected 1 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_DuplicateSubmit 提交重复商品：后者覆盖前者（map 特性）
func Test_diffDineInProducts_DuplicateSubmit(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交 2 个相同 flavor 的商品，num 不同
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0},
		{FlavorProductBomUuid: 1001, Num: 3, ProductType: 0}, // key 相同，会覆盖上面
	}
	saleBill := makeSaleBill(nil)

	addProducts, _, _, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 两个提交 key 相同，commitProductMap 只有 1 条记录
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts (duplicate key merged), got %d", len(addProducts))
	}

	// 应该保留后者的 Num=3
	for _, p := range addProducts {
		if p.Num != 3 {
			t.Errorf("expected Num=3 (last one wins), got %.0f", p.Num)
		}
	}
}

// Test_diffDineInProducts_SauceOrderIndependent 小料顺序不影响 key：[5002,5001] 与 [5001,5002] 应匹配
func Test_diffDineInProducts_SauceOrderIndependent(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交：sauce 顺序 [5002, 5001]（逆序）
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, SauceProductBomUuidList: []uint64{5002, 5001}},
	}
	// 已有：sauce 顺序 [5001, 5002]（正序）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, []uint64{5001, 5002}, nil),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 排序后 key 相同，应为修改
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_MixedNormalAndPackage 普通商品和套餐商品混合
func Test_diffDineInProducts_MixedNormalAndPackage(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交：1 个普通商品 + 1 个套餐商品
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 2, ProductType: 0},
		{
			ProductPackageUuid: 2001,
			Num:                1,
			ProductType:        constant.ProductTypePackage,
			Products: []req.ProductRequest{
				{
					ProductPackageGroupUuid: 3001,
					EditProductReq:          req.EditProductReq{FlavorUuid: 4001},
					Num:                     1,
					UnitNum:                 1,
				},
			},
		},
	}
	// 已有：1 个普通商品（1001）+ 1 个普通商品（1003，将被删除）
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, nil),
		makeSaleOrderProduct(2, 1003, constant.ProductTypeProduct, 0, nil, nil),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1001 匹配 → 修改, 2001 套餐不匹配 → 新增, 1003 不在提交 → 删除
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts (package 2001), got %d", len(addProducts))
	}
	if len(deleteProducts) != 1 {
		t.Errorf("expected 1 deleteProducts (normal 1003), got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts (normal 1001), got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_DeletedBomSkipped 已有商品中已删除的 bom 应被 ProductKey 跳过
func Test_diffDineInProducts_DeletedBomSkipped(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交：flavor 1001 + sauce [5001]
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, SauceProductBomUuidList: []uint64{5001}},
	}

	// 已有：flavor 1001 + sauce [5001]（正常）+ sauce [5002]（已删除）
	existingProduct := &model.SaleOrderProduct{
		BaseModel:   model.BaseModel{Uuid: 1, DeleteTime: 0},
		ProductType: constant.ProductTypeProduct,
		SaleOrderProductBoms: []*model.SaleOrderProductBom{
			{
				BaseModel:      model.BaseModel{Uuid: 101},
				IsFlavorBom:    constant.ProductBomTypeFlavor,
				ProductBomUuid: 1001,
			},
			{
				BaseModel:      model.BaseModel{Uuid: 102},
				IsFlavorBom:    constant.ProductBomTypeSauce,
				ProductBomUuid: 5001,
			},
			{
				BaseModel:      model.BaseModel{Uuid: 103, DeleteTime: 1000}, // 已删除的 sauce
				IsFlavorBom:    constant.ProductBomTypeSauce,
				ProductBomUuid: 5002,
			},
		},
	}
	saleBill := makeSaleBill([]*model.SaleOrderProduct{existingProduct})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 已删除的 sauce 5002 被 ProductKey 跳过
	// 已有 key = "1001--5001"，提交 key = "1001--5001"，应匹配
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_UpdateOlderProductsMatch 验证 updateOlderProducts 返回正确的旧商品引用
func Test_diffDineInProducts_UpdateOlderProductsMatch(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 10, ProductType: 0},
	}
	olderProduct := makeSaleOrderProduct(99, 1001, constant.ProductTypeProduct, 0, nil, nil)
	saleBill := makeSaleBill([]*model.SaleOrderProduct{olderProduct})

	_, _, updateProducts, updateOlderProducts, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updateProducts) != 1 {
		t.Fatalf("expected 1 updateProducts, got %d", len(updateProducts))
	}
	if len(updateOlderProducts) != 1 {
		t.Fatalf("expected 1 updateOlderProducts, got %d", len(updateOlderProducts))
	}

	// 验证 updateProducts 和 updateOlderProducts 使用相同的 key
	for key, newProduct := range updateProducts {
		oldProduct, ok := updateOlderProducts[key]
		if !ok {
			t.Errorf("updateOlderProducts missing key %q", key)
			continue
		}
		if newProduct.Num != 10 {
			t.Errorf("expected new product Num=10, got %.0f", newProduct.Num)
		}
		if oldProduct.Uuid != 99 {
			t.Errorf("expected old product Uuid=99, got %d", oldProduct.Uuid)
		}
	}
}

// Test_diffDineInProducts_PackageDiffSubProducts 套餐商品子商品不同时，应为新增+删除
func Test_diffDineInProducts_PackageDiffSubProducts(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交一个套餐商品（子商品: group=3001, flavor=4001）
	products := []req.ProductParams{
		{
			ProductPackageUuid: 2001,
			Num:                1,
			ProductType:        constant.ProductTypePackage,
			Products: []req.ProductRequest{
				{
					ProductPackageGroupUuid: 3001,
					EditProductReq:          req.EditProductReq{FlavorUuid: 4001},
					Num:                     1,
					UnitNum:                 1,
				},
			},
		},
	}

	// 已有套餐：相同 PackageUuid 但不同子商品（group=3001, flavor=4002）
	subProductParams := `[{"flavor_uuid":4002,"product_package_group_uuid":3001,"num":1,"unit_num":1}]`
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makePackageSaleOrderProduct(1, 2001, 2001, subProductParams, 0),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 子商品签名不同，key 不匹配 → 1 新增 + 1 删除
	if len(addProducts) != 1 {
		t.Errorf("expected 1 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 1 {
		t.Errorf("expected 1 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 0 {
		t.Errorf("expected 0 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_PackageEmptySubProducts 空子商品的套餐匹配
func Test_diffDineInProducts_PackageEmptySubProducts(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 提交一个空子商品的套餐
	products := []req.ProductParams{
		{
			ProductPackageUuid: 2001,
			Num:                1,
			ProductType:        constant.ProductTypePackage,
			Products:           []req.ProductRequest{},
		},
	}

	// 已有套餐也是空子商品
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makePackageSaleOrderProduct(1, 2001, 2001, "[]", 0),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 都是空子商品，key 一致 "pkg:2001-"，应为修改
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// Test_diffDineInProducts_AttributeOrderIndependent 属性顺序不影响 key
func Test_diffDineInProducts_AttributeOrderIndependent(t *testing.T) {
	dbm, db := setupDiffDineInTestDB(t)
	ctx := createDiffDineInTestContext(t, dbm)
	srv := newTestOrderSrv(dbm)

	// 使用不同的 UUID 避免和其他测试的共享内存数据库冲突
	db.Exec(`INSERT INTO ttpos_product_package_attribute (uuid, attribute_uuid, delete_time) VALUES (16001, 17001, 0)`)
	db.Exec(`INSERT INTO ttpos_product_package_attribute (uuid, attribute_uuid, delete_time) VALUES (16002, 17002, 0)`)

	// 提交：attribute 逆序 [16002, 16001] → 解析为 [17002, 17001]，排序后 [17001, 17002]
	products := []req.ProductParams{
		{FlavorProductBomUuid: 1001, Num: 1, ProductType: 0, ProductPackageAttributeUuidList: []uint64{16002, 16001}},
	}
	// 已有：attribute 正序 [17001, 17002]
	saleBill := makeSaleBill([]*model.SaleOrderProduct{
		makeSaleOrderProduct(1, 1001, constant.ProductTypeProduct, 0, nil, []uint64{17001, 17002}),
	})

	addProducts, deleteProducts, updateProducts, _, err := srv.diffDineInProducts(ctx, db, products, saleBill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 排序后 key 相同，应为修改
	if len(addProducts) != 0 {
		t.Errorf("expected 0 addProducts, got %d", len(addProducts))
	}
	if len(deleteProducts) != 0 {
		t.Errorf("expected 0 deleteProducts, got %d", len(deleteProducts))
	}
	if len(updateProducts) != 1 {
		t.Errorf("expected 1 updateProducts, got %d", len(updateProducts))
	}
}

// suppress unused import warning
var _ = fmt.Sprintf
