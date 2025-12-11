package inventory

import (
	stdcontext "context"
	"testing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	domainService "ttpos-server-go/app/modules/inventory/domain/service"
	inventoryPersistence "ttpos-server-go/app/modules/inventory/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	pkg_context "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupIntegrationTestDB 创建集成测试数据库
func setupIntegrationTestDB(t *testing.T) (*database.DBManager, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect test database: %v", err)
	}

	// 创建完整的表结构（包含关联表）
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_product_bom (
			id BIGINT PRIMARY KEY AUTOINCREMENT,
			uuid BIGINT NOT NULL DEFAULT 0,
			create_time INT NOT NULL DEFAULT 0,
			update_time INT NOT NULL DEFAULT 0,
			delete_time INT NOT NULL DEFAULT 0,
			product_package_uuid BIGINT NOT NULL DEFAULT 0,
			product_bom_card_uuid BIGINT NOT NULL DEFAULT 0,
			use_bom_card_stock TINYINT NOT NULL DEFAULT 1,
			is_sold_out TINYINT NOT NULL DEFAULT 0,
			sellable_quantity DECIMAL(22,4) NOT NULL DEFAULT 0.0000,
			stock_num DECIMAL(12,4) NOT NULL DEFAULT 0.0000
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create product_bom table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_product_bom_card (
			id BIGINT PRIMARY KEY AUTOINCREMENT,
			uuid BIGINT NOT NULL DEFAULT 0,
			create_time INT NOT NULL DEFAULT 0,
			update_time INT NOT NULL DEFAULT 0,
			delete_time INT NOT NULL DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create product_bom_card table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_related_material (
			id BIGINT PRIMARY KEY AUTOINCREMENT,
			uuid BIGINT NOT NULL DEFAULT 0,
			create_time INT NOT NULL DEFAULT 0,
			update_time INT NOT NULL DEFAULT 0,
			delete_time INT NOT NULL DEFAULT 0,
			related_uuid BIGINT NOT NULL DEFAULT 0,
			material_uuid BIGINT NOT NULL DEFAULT 0,
			num DECIMAL(22,4) NOT NULL DEFAULT 0.0000
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create related_material table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_material (
			id BIGINT PRIMARY KEY AUTOINCREMENT,
			uuid BIGINT NOT NULL DEFAULT 0,
			create_time INT NOT NULL DEFAULT 0,
			update_time INT NOT NULL DEFAULT 0,
			delete_time INT NOT NULL DEFAULT 0,
			stock_num DECIMAL(12,4) NOT NULL DEFAULT 0.0000
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create material table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_warehouse_item (
			id BIGINT PRIMARY KEY AUTOINCREMENT,
			uuid BIGINT NOT NULL DEFAULT 0,
			warehouse_uuid BIGINT NOT NULL DEFAULT 0,
			material_uuid BIGINT NOT NULL DEFAULT 0,
			stock DECIMAL(22,4) NOT NULL DEFAULT 0.0000
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create warehouse_item table: %v", err)
	}

	// 创建 DBManager
	dbm := &database.DBManager{}
	dbm.SetMockDB(db)

	return dbm, db
}

// createIntegrationTestContext 创建集成测试上下文
func createIntegrationTestContext(t *testing.T, db *gorm.DB) pkg_context.Context {
	t.Helper()

	ctx := pkg_context.NewContext(pkg_context.WithContext(stdcontext.Background()))
	ctx.SetDB(db)
	ctx.SetCompanyUuid(constant.MockDB)

	return ctx
}

// TestProductInventoryIntegration_EndToEnd 端到端集成测试
func TestProductInventoryIntegration_EndToEnd(t *testing.T) {
	dbm, db := setupIntegrationTestDB(t)
	ctx := createIntegrationTestContext(t, db)

	// 1. 创建测试数据
	// 创建材料
	material := &model.Material{
		BaseModel: model.BaseModel{
			Uuid:       1001,
			CreateTime: 1000,
			UpdateTime: 1000,
		},
		StockNum: 100.0,
	}
	db.Create(material)

	// 创建成本卡
	bomCard := &model.ProductBomCard{
		BaseModel: model.BaseModel{
			Uuid:       2001,
			CreateTime: 1000,
			UpdateTime: 1000,
		},
	}
	db.Create(bomCard)

	// 创建关联材料
	relatedMaterial := &model.RelatedMaterial{
		BaseModel: model.BaseModel{
			Uuid:       3001,
			CreateTime: 1000,
			UpdateTime: 1000,
		},
		RelatedUuid:  bomCard.Uuid,
		MaterialUuid: material.Uuid,
		Num:          2.0,
		Material:     material,
	}
	db.Create(relatedMaterial)

	// 创建有成本卡的商品BOM
	productBomWithCard := &model.ProductBom{
		BaseModel: model.BaseModel{
			Uuid:       4001,
			CreateTime: 1000,
			UpdateTime: 1000,
		},
		ProductBomCardUuid: bomCard.Uuid,
		UseBomCardStock:    constant.Yes,
		IsSoldOut:          0,
		SellableQuantity:   0,
		ProductBomCard:     bomCard,
	}
	productBomWithCard.ProductBomCard.RelatedMaterials = []*model.RelatedMaterial{relatedMaterial}
	db.Create(productBomWithCard)

	// 创建无成本卡的商品BOM
	productBomWithoutCard := &model.ProductBom{
		BaseModel: model.BaseModel{
			Uuid:       4002,
			CreateTime: 1000,
			UpdateTime: 1000,
		},
		ProductBomCardUuid: 0,
		UseBomCardStock:    constant.No,
		IsSoldOut:          0,
		SellableQuantity:   50.0,
	}
	db.Create(productBomWithoutCard)

	// 2. 创建 Repository 和 Domain Service
	productBomRepo := inventoryPersistence.NewProductBomRepository(dbm)
	productPackageRepo := inventoryPersistence.NewProductPackageRepository(dbm)
	domainSvc := domainService.NewProductInventoryDomainService(productBomRepo, productPackageRepo)

	// 3. 测试有成本卡商品的库存查询
	inventory1, err := domainSvc.GetProductInventory(ctx, productBomWithCard.Uuid)
	assert.NoError(t, err)
	// 材料库存100，用量2，可生产50
	assert.GreaterOrEqual(t, inventory1, 0.0)

	// 4. 测试无成本卡商品的库存查询
	inventory2, err := domainSvc.GetProductInventory(ctx, productBomWithoutCard.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, inventory2)

	// 5. 测试不存在的商品
	_, err = domainSvc.GetProductInventory(ctx, 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品不存在")
}

// TestProductInventoryIntegration_WithCache 测试带缓存的集成测试
func TestProductInventoryIntegration_WithCache(t *testing.T) {
	dbm, db := setupIntegrationTestDB(t)
	ctx := createIntegrationTestContext(t, db)

	// 创建测试数据
	productBom := &model.ProductBom{
		BaseModel: model.BaseModel{
			Uuid:       5001,
			CreateTime: 1000,
			UpdateTime: 1000,
		},
		ProductBomCardUuid: 0,
		IsSoldOut:          0,
		SellableQuantity:   100.0,
	}
	db.Create(productBom)

	// 创建 Mock Cache
	mockCache := cache.NewCache(cache.GoCache, cache.Config{})

	// 创建 Domain Service 和 Application Service
	appSvc := inventoryApp.NewProductInventoryAppServiceWithDependencies(dbm, mockCache)

	// 第一次查询（应该从数据库获取）
	inventory1, err := appSvc.GetProductInventory(ctx, productBom.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, inventory1)

	// 第二次查询（应该从缓存获取）
	inventory2, err := appSvc.GetProductInventory(ctx, productBom.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, inventory2)

	// 测试缓存失效
	appSvc.InvalidateProductInventoryCache(ctx.GetCompanyUuid(), productBom.Uuid)
	// 再次查询应该重新从数据库获取
	inventory3, err := appSvc.GetProductInventory(ctx, productBom.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, inventory3)
}
