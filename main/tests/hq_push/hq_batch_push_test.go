//go:build integration

package hq_push_test

import (
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// --- Constants ---

const (
	pathHqBatchPush = "/api/v1/shop/product/hq_batch_push"

	codeTokenInvalid = -102
)

// --- P0: Critical Path ---

// Test_P0_Shop_HqBatchPush_HappyPath_SingleField verifies a single field_type push to all stores.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P0_Shop_HqBatchPush_HappyPath_SingleField(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"dine_shelf"},
		"is_all_stores": true,
	})
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	if string(apiResp.Data) == "" {
		t.Fatal("expected non-empty data")
	}
}

// Test_P0_Shop_HqBatchPush_HappyPath_MultipleFields verifies multiple field_types push to specified stores.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P0_Shop_HqBatchPush_HappyPath_MultipleFields(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf", "takeout_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_HqBatchPush_HappyPath_AllFieldTypes verifies all 4 field_types in one request.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P0_Shop_HqBatchPush_HappyPath_AllFieldTypes(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"dine_shelf", "takeout_shelf", "takeout_price", "negative_stock"},
		"is_all_stores": true,
	})
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_HqBatchPush_Unauthorized_NotHeadquarter verifies non-HQ company is rejected.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P0_Shop_HqBatchPush_Unauthorized_NotHeadquarter(t *testing.T) {
	// Use a regular (non-HQ) company
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		// No HeadquarterConfig → IsHeadquarter()=false
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	resp := client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"dine_shelf"},
		"is_all_stores": true,
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected error for non-headquarter company")
	}
}

// --- P1: Input Validation ---

// Test_P1_Shop_HqBatchPush_InvalidInput_EmptyFieldTypes verifies empty field_types is rejected.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P1_Shop_HqBatchPush_InvalidInput_EmptyFieldTypes(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{},
		"is_all_stores": true,
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected validation error for empty field_types")
	}
}

// Test_P1_Shop_HqBatchPush_InvalidInput_MissingFieldTypes verifies missing field_types is rejected.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P1_Shop_HqBatchPush_InvalidInput_MissingFieldTypes(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"is_all_stores": true,
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected validation error for missing field_types")
	}
}

// Test_P1_Shop_HqBatchPush_InvalidInput_InvalidFieldType verifies unknown field_type is rejected.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P1_Shop_HqBatchPush_InvalidInput_InvalidFieldType(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"invalid_type"},
		"is_all_stores": true,
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected error for invalid field_type")
	}
}

// Test_P1_Shop_HqBatchPush_InvalidInput_SafetyStockNotAllowed verifies safety_stock is not a valid batch push type.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P1_Shop_HqBatchPush_InvalidInput_SafetyStockNotAllowed(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"safety_stock"},
		"is_all_stores": true,
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected error for safety_stock field_type")
	}
}

// Test_P1_Shop_HqBatchPush_InvalidInput_NoStoreSelected verifies error when no stores are selected.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P1_Shop_HqBatchPush_InvalidInput_NoStoreSelected(t *testing.T) {
	env := setupHqEnv(t)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{},
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected error for empty store_uuids")
	}
}

// Test_P1_Shop_HqBatchPush_Unauthorized_NoToken verifies request without token is rejected.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P1_Shop_HqBatchPush_Unauthorized_NoToken(t *testing.T) {
	fixture.SetupWireMock(t)
	client := fixture.NewHTTPClient()

	resp := client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"dine_shelf"},
		"is_all_stores": true,
	})
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- P2: Data Validation (dine_shelf) ---

// Test_P2_Shop_HqBatchPush_DineShelf_ProductPackageStatusSync verifies product_package.status is synced to store.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_DineShelf_ProductPackageStatusSync(t *testing.T) {
	env := setupHqEnv(t)

	// Seed: HQ product (status=1), store product (status=0, same UUID)
	productUuid := fixture.GenerateSnowflakeID()
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productUuid) == 1
	})
}

// Test_P2_Shop_HqBatchPush_DineShelf_ProductBomStatusSync verifies product_bom.status is synced along with product_package.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_DineShelf_ProductBomStatusSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	bom1Uuid := fixture.GenerateSnowflakeID()
	bom2Uuid := fixture.GenerateSnowflakeID()

	// HQ: product status=1
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))

	// Store: product status=0, 2 boms status=0
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))
	fixture.SeedProductBom(t, env.storeDB1, fixture.WithProductBomPackageUUID(productUuid), func(o *fixture.ProductBomOptions) { o.UUID = bom1Uuid; o.Status = 0 })
	fixture.SeedProductBom(t, env.storeDB1, fixture.WithProductBomPackageUUID(productUuid), func(o *fixture.ProductBomOptions) { o.UUID = bom2Uuid; o.Status = 0 })

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_bom", "status", bom1Uuid) == 1 &&
			queryIntField(t, env.storeDB1, "ttpos_product_bom", "status", bom2Uuid) == 1
	})
}

// Test_P2_Shop_HqBatchPush_DineShelf_MissingProductSkipped verifies products not in store are skipped without error.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_DineShelf_MissingProductSkipped(t *testing.T) {
	env := setupHqEnv(t)

	productA := fixture.GenerateSnowflakeID()
	productB := fixture.GenerateSnowflakeID()

	// HQ has both products
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productA), fixture.WithProductPackageStatus(1))
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productB), fixture.WithProductPackageStatus(1))

	// Store only has productA
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productA), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productA) == 1
	})
}

// Test_P2_Shop_HqBatchPush_DineShelf_OverrideCleared verifies hq_field_override is soft-deleted after force push.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_DineShelf_OverrideCleared(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	// Seed override record in store
	seedHqFieldOverride(t, env.storeDB1, productUuid, "product", "dine_shelf")

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		// Override should be soft-deleted (delete_time > 0)
		return !queryOverrideExists(t, env.storeDB1, productUuid, "dine_shelf")
	})
}

// Test_P2_Shop_HqBatchPush_DineShelf_MultiProductIndependent verifies per-product transaction independence.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_DineShelf_MultiProductIndependent(t *testing.T) {
	env := setupHqEnv(t)

	productA := fixture.GenerateSnowflakeID()
	productB := fixture.GenerateSnowflakeID()

	// HQ: both products status=1
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productA), fixture.WithProductPackageStatus(1))
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productB), fixture.WithProductPackageStatus(1))

	// Store: both products status=0
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productA), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productB), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productA) == 1 &&
			queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productB) == 1
	})
}

// --- P2: Data Validation (takeout_shelf) ---

// Test_P2_Shop_HqBatchPush_TakeoutShelf_StatusSync verifies takeout product status is synced.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_TakeoutShelf_StatusSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()
	storeTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: takeout status=1
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 1, 10.0)

	// Store: takeout status=0 (different UUID, same product_package_uuid+takeout_type)
	seedProductPackageTakeout(t, env.storeDB1, storeTakeoutUuid, productUuid, env.hqUuid, 1, 0, 10.0)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"takeout_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_package_takeout", "status", storeTakeoutUuid) == 1
	})
}

// Test_P2_Shop_HqBatchPush_TakeoutShelf_MatchByPackageUuidAndType verifies matching by product_package_uuid + takeout_type.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_TakeoutShelf_MatchByPackageUuidAndType(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// HQ: Grab takeout(type=1) status=1, FoodPanda takeout(type=2) status=0
	seedProductPackageTakeout(t, env.hqDB, fixture.GenerateSnowflakeID(), productUuid, 0, 1, 1, 10.0)
	seedProductPackageTakeout(t, env.hqDB, fixture.GenerateSnowflakeID(), productUuid, 0, 2, 0, 10.0)

	// Store: Grab takeout(type=1) status=0
	storeGrabUuid := fixture.GenerateSnowflakeID()
	seedProductPackageTakeout(t, env.storeDB1, storeGrabUuid, productUuid, env.hqUuid, 1, 0, 10.0)

	// Store: FoodPanda takeout(type=2) status=1
	storeFpUuid := fixture.GenerateSnowflakeID()
	seedProductPackageTakeout(t, env.storeDB1, storeFpUuid, productUuid, env.hqUuid, 2, 1, 10.0)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"takeout_shelf"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		// Grab should be updated to 1, FoodPanda should be updated to 0
		return queryIntField(t, env.storeDB1, "ttpos_product_package_takeout", "status", storeGrabUuid) == 1 &&
			queryIntField(t, env.storeDB1, "ttpos_product_package_takeout", "status", storeFpUuid) == 0
	})
}

// --- P2: Data Validation (takeout_price) ---

// Test_P2_Shop_HqBatchPush_TakeoutPrice_MainPriceSync verifies takeout main table price is synced.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_TakeoutPrice_MainPriceSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()
	storeTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: takeout price=15.50
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 1, 15.50)

	// Store: takeout price=10.00
	seedProductPackageTakeout(t, env.storeDB1, storeTakeoutUuid, productUuid, env.hqUuid, 1, 1, 10.00)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"takeout_price"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryFloatField(t, env.storeDB1, "ttpos_product_package_takeout", "price", storeTakeoutUuid) == 15.50
	})
}

// Test_P2_Shop_HqBatchPush_TakeoutPrice_BomPriceSync verifies takeout bom price is synced.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_TakeoutPrice_BomPriceSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	bomUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()
	storeTakeoutUuid := fixture.GenerateSnowflakeID()
	hqBomTakeoutUuid := fixture.GenerateSnowflakeID()
	storeBomTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: takeout + bom_takeout price=8.00
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 1, 15.00)
	seedProductBomTakeout(t, env.hqDB, hqBomTakeoutUuid, hqTakeoutUuid, bomUuid, 0, 8.00)

	// Store: takeout + bom_takeout price=5.00
	seedProductPackageTakeout(t, env.storeDB1, storeTakeoutUuid, productUuid, env.hqUuid, 1, 1, 10.00)
	seedProductBomTakeout(t, env.storeDB1, storeBomTakeoutUuid, storeTakeoutUuid, bomUuid, env.hqUuid, 5.00)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"takeout_price"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryFloatField(t, env.storeDB1, "ttpos_product_bom_takeout", "price", storeBomTakeoutUuid) == 8.00
	})
}

// --- P2: Data Validation (negative_stock) ---

// Test_P2_Shop_HqBatchPush_NegativeStock_AllowNegativeStockSync verifies allow_negative_stock is synced.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_NegativeStock_AllowNegativeStockSync(t *testing.T) {
	env := setupHqEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()

	// HQ: allow_negative_stock=1
	seedMaterial(t, env.hqDB, materialUuid, 0, 1, nil)

	// Store: allow_negative_stock=0
	seedMaterial(t, env.storeDB1, materialUuid, env.hqUuid, 0, nil)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"negative_stock"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_material", "allow_negative_stock", materialUuid) == 1
	})
}

// Test_P2_Shop_HqBatchPush_NegativeStock_SafetyStockSync verifies safety_stock is synced with negative_stock push.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_NegativeStock_SafetyStockSync(t *testing.T) {
	env := setupHqEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	safetyStock := 10.0

	// HQ: safety_stock=10
	seedMaterial(t, env.hqDB, materialUuid, 0, 1, &safetyStock)

	// Store: safety_stock=NULL
	seedMaterial(t, env.storeDB1, materialUuid, env.hqUuid, 0, nil)

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"negative_stock"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryFloatField(t, env.storeDB1, "ttpos_material", "safety_stock", materialUuid) == 10.0
	})
}

// Test_P2_Shop_HqBatchPush_NegativeStock_OverrideCleared verifies both negative_stock and safety_stock overrides are cleared.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_NegativeStock_OverrideCleared(t *testing.T) {
	env := setupHqEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	safetyStock := 10.0

	seedMaterial(t, env.hqDB, materialUuid, 0, 1, &safetyStock)
	seedMaterial(t, env.storeDB1, materialUuid, env.hqUuid, 0, nil)

	// Seed two override records
	seedHqFieldOverride(t, env.storeDB1, materialUuid, "material", "negative_stock")
	seedHqFieldOverride(t, env.storeDB1, materialUuid, "material", "safety_stock")

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"negative_stock"},
		"store_uuids": []int64{env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return !queryOverrideExists(t, env.storeDB1, materialUuid, "negative_stock") &&
			!queryOverrideExists(t, env.storeDB1, materialUuid, "safety_stock")
	})
}

// --- P2: Data Validation (cross-store) ---

// Test_P2_Shop_HqBatchPush_CrossStore_MultipleStoresSync verifies data is pushed to multiple stores.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_CrossStore_MultipleStoresSync(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// HQ: status=1
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))

	// Store1 and Store2: status=0
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))
	fixture.SeedProductPackage(t, env.storeDB2, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.storeUuid1, env.storeUuid2},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productUuid) == 1 &&
			queryIntField(t, env.storeDB2, "ttpos_product_package", "status", productUuid) == 1
	})
}

// Test_P2_Shop_HqBatchPush_CrossStore_HqUuidFiltered verifies HQ's own uuid is filtered from store_uuids.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_CrossStore_HqUuidFiltered(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	// HQ: status=1
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))

	// Store: status=0
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	// Include HQ's own uuid in store_uuids — should be filtered
	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types": []string{"dine_shelf"},
		"store_uuids": []int64{env.hqUuid, env.storeUuid1},
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		// Store should be updated
		return queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productUuid) == 1
	})

	// HQ product should remain unchanged (status=1 was already 1, but verify no unexpected changes)
	hqStatus := queryIntField(t, env.hqDB, "ttpos_product_package", "status", productUuid)
	if hqStatus != 1 {
		t.Fatalf("HQ product status should remain unchanged, got %d", hqStatus)
	}
}

// Test_P2_Shop_HqBatchPush_CrossStore_IsAllStoresAutoDiscover verifies is_all_stores discovers stores from saas DB.
// Route: POST /api/v1/shop/product/hq_batch_push
func Test_P2_Shop_HqBatchPush_CrossStore_IsAllStoresAutoDiscover(t *testing.T) {
	env := setupHqEnv(t)

	productUuid := fixture.GenerateSnowflakeID()

	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	fixture.SeedProductPackage(t, env.storeDB1, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))
	fixture.SeedProductPackage(t, env.storeDB2, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(0), fixture.WithProductPackageHeadquarterUuid(env.hqUuid))

	resp := env.client.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"dine_shelf"},
		"is_all_stores": true,
	})
	resp.AssertOK(t).AssertSuccess(t)

	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB1, "ttpos_product_package", "status", productUuid) == 1 &&
			queryIntField(t, env.storeDB2, "ttpos_product_package", "status", productUuid) == 1
	})
}

// --- Helpers ---

type hqTestEnv struct {
	client    *fixture.HTTPClient
	hqDB      *sql.DB
	storeDB1  *sql.DB
	storeDB2  *sql.DB
	hqUuid    int64
	storeUuid1 int64
	storeUuid2 int64
}

// setupHqEnv creates a complete HQ + 2 sub-store test environment.
// Seeds company/company_setting in both tenant DBs and saas DB.
func setupHqEnv(t *testing.T) hqTestEnv {
	t.Helper()

	// Create HQ tenant
	hqUUIDStr := fixture.GenerateCompanyUUID(t)
	hqDB := fixture.NewTestTenantFull(t, hqUUIDStr)
	hqUuid := mustParseInt64(hqUUIDStr)

	fixture.SeedCompany(t, hqDB, fixture.WithCompanyUUID(hqUuid))
	fixture.SeedCompanySetting(t, hqDB,
		fixture.WithCompanySettingCompanyUUID(hqUuid),
		fixture.WithCompanySettingHeadquarterConfig("test-site", "HQ"),
	)
	staff := fixture.SeedStaff(t, hqDB,
		fixture.WithStaffCompanyUUID(hqUuid),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	// Create store 1 tenant
	store1UUIDStr := fixture.GenerateCompanyUUID(t)
	storeDB1 := fixture.NewTestTenantFull(t, store1UUIDStr)
	storeUuid1 := mustParseInt64(store1UUIDStr)
	fixture.SeedCompany(t, storeDB1, fixture.WithCompanyUUID(storeUuid1))
	fixture.SeedCompanySetting(t, storeDB1, fixture.WithCompanySettingCompanyUUID(storeUuid1))

	// Create store 2 tenant
	store2UUIDStr := fixture.GenerateCompanyUUID(t)
	storeDB2 := fixture.NewTestTenantFull(t, store2UUIDStr)
	storeUuid2 := mustParseInt64(store2UUIDStr)
	fixture.SeedCompany(t, storeDB2, fixture.WithCompanyUUID(storeUuid2))
	fixture.SeedCompanySetting(t, storeDB2, fixture.WithCompanySettingCompanyUUID(storeUuid2))

	// Seed in saas DB for getSubStoreUuids to discover stores
	saasDB := fixture.NewSaasDB(t)
	fixture.SeedCompany(t, saasDB, fixture.WithCompanyUUID(hqUuid))
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(hqUuid),
		fixture.WithCompanySettingHeadquarterConfig("test-site", "HQ"),
	)
	fixture.SeedCompany(t, saasDB, fixture.WithCompanyUUID(storeUuid1))
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(storeUuid1),
		fixture.WithCompanySettingHeadquarterUuid(hqUuid),
	)
	fixture.SeedCompany(t, saasDB, fixture.WithCompanyUUID(storeUuid2))
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(storeUuid2),
		fixture.WithCompanySettingHeadquarterUuid(hqUuid),
	)

	token := fixture.GenerateShopToken(t, hqUUIDStr, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	return hqTestEnv{
		client:     client,
		hqDB:       hqDB,
		storeDB1:   storeDB1,
		storeDB2:   storeDB2,
		hqUuid:     hqUuid,
		storeUuid1: storeUuid1,
		storeUuid2: storeUuid2,
	}
}

// seedProductPackageTakeout inserts a product_package_takeout record.
func seedProductPackageTakeout(t *testing.T, db *sql.DB, uuid, productPackageUuid, headquarterUuid int64, takeoutType, status int, price float64) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_package_takeout (uuid, product_package_uuid, headquarter_uuid, takeout_type, status, price, name, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, '', ?, ?, 0)
	`, uuid, productPackageUuid, headquarterUuid, takeoutType, status, price, now, now)
	if err != nil {
		t.Fatalf("failed to seed product_package_takeout: %v", err)
	}
}

// seedProductBomTakeout inserts a product_bom_takeout record.
func seedProductBomTakeout(t *testing.T, db *sql.DB, uuid, productPackageTakeoutUuid, productBomUuid, headquarterUuid int64, price float64) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_bom_takeout (uuid, product_package_takeout_uuid, product_bom_uuid, headquarter_uuid, price, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	`, uuid, productPackageTakeoutUuid, productBomUuid, headquarterUuid, price, now, now)
	if err != nil {
		t.Fatalf("failed to seed product_bom_takeout: %v", err)
	}
}

// seedMaterial inserts a material record.
func seedMaterial(t *testing.T, db *sql.DB, uuid, headquarterUuid int64, allowNegativeStock int, safetyStock *float64) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_material (uuid, name, headquarter_uuid, allow_negative_stock, safety_stock, status, create_time, update_time, delete_time)
		VALUES (?, '{}', ?, ?, ?, 1, ?, ?, 0)
	`, uuid, headquarterUuid, allowNegativeStock, safetyStock, now, now)
	if err != nil {
		t.Fatalf("failed to seed material: %v", err)
	}
}

// seedHqFieldOverride inserts a hq_field_override record.
func seedHqFieldOverride(t *testing.T, db *sql.DB, entityUuid int64, entityType, fieldType string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_hq_field_override (uuid, entity_type, entity_uuid, field_type, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, fixture.GenerateSnowflakeID(), entityType, entityUuid, fieldType, now, now)
	if err != nil {
		t.Fatalf("failed to seed hq_field_override: %v", err)
	}
}

// queryIntField queries a single int field from a table by uuid.
// table should be the full table name (e.g., "ttpos_product_package").
func queryIntField(t *testing.T, db *sql.DB, table, field string, uuid int64) int {
	t.Helper()
	var val int
	err := db.QueryRow(fmt.Sprintf("SELECT `%s` FROM `%s` WHERE uuid = ? AND delete_time = 0", field, table), uuid).Scan(&val)
	if err != nil {
		return -1 // not found or error
	}
	return val
}

// queryFloatField queries a single float field from a table by uuid.
// table should be the full table name (e.g., "ttpos_product_package_takeout").
func queryFloatField(t *testing.T, db *sql.DB, table, field string, uuid int64) float64 {
	t.Helper()
	var val float64
	err := db.QueryRow(fmt.Sprintf("SELECT `%s` FROM `%s` WHERE uuid = ? AND delete_time = 0", field, table), uuid).Scan(&val)
	if err != nil {
		return -1
	}
	return val
}

// queryOverrideExists checks if an active (non-deleted) override record exists.
func queryOverrideExists(t *testing.T, db *sql.DB, entityUuid int64, fieldType string) bool {
	t.Helper()
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM ttpos_hq_field_override WHERE entity_uuid = ? AND field_type = ? AND delete_time = 0",
		entityUuid, fieldType,
	).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// waitForAsync polls a condition until it's true or timeout.
func waitForAsync(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("async operation did not complete within timeout")
}

func mustParseInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse int64: %s", s))
	}
	return v
}

func mustParseString(i int64) string {
	return strconv.FormatInt(i, 10)
}
