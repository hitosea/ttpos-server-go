//go:build integration

package hq_push_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// --- Constants ---

const (
	pathMaterialUpdateNegativeStock = "/api/v1/shop/material/update_negative_stock"
	pathMaterialStatusBatch         = "/api/v1/shop/material/status/batch"
)

// ========================================================================================
// Test 1: Sub-store update negative stock via new API
// ========================================================================================

// Test_P0_Shop_MaterialNegativeStock_SubStoreUpdate verifies that a sub-store can update
// the negative stock setting on an HQ-synced material when "separate control" is enabled.
// This tests the new POST /shop/material/update_negative_stock endpoint.
func Test_P0_Shop_MaterialNegativeStock_SubStoreUpdate(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()

	// Seed material in HQ DB (allow_negative_stock=0)
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)

	// Seed material in sub-store DB (HQ-synced, allow_negative_stock=0)
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)

	// Set negative_stock to "separate control" in HQ DB
	seedHqControlSetting(t, env.hqDB, "negative_stock", 0) // 0 = separate control

	// Sub-store enables negative stock
	resp := env.storeClient.Post(t, pathMaterialUpdateNegativeStock, map[string]any{
		"uuid":                 materialUuid,
		"allow_negative_stock": true,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: material in store DB has allow_negative_stock=1
	negStock := queryIntField(t, env.storeDB, "ttpos_material", "allow_negative_stock", materialUuid)
	if negStock != 1 {
		t.Fatalf("expected allow_negative_stock=1 in store, got %d", negStock)
	}

	// Verify: override record exists for negative_stock
	if !queryOverrideExists(t, env.storeDB, materialUuid, "negative_stock") {
		t.Fatal("expected override record for negative_stock to exist")
	}
}

// Test_P0_Shop_MaterialNegativeStock_SubStoreUpdate_UnifiedControlRejected verifies that
// a sub-store cannot update negative stock when "unified control" is active (default).
func Test_P0_Shop_MaterialNegativeStock_SubStoreUpdate_UnifiedControlRejected(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()

	// Seed material in store DB (HQ-synced)
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)

	// Default: negative_stock is "unified control" — no need to seed hq_control_setting

	// Sub-store tries to enable negative stock → should fail
	resp := env.storeClient.Post(t, pathMaterialUpdateNegativeStock, map[string]any{
		"uuid":                 materialUuid,
		"allow_negative_stock": true,
	})
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected error when unified control is active")
	}

	// Verify: value unchanged
	negStock := queryIntField(t, env.storeDB, "ttpos_material", "allow_negative_stock", materialUuid)
	if negStock != 0 {
		t.Fatalf("expected allow_negative_stock=0 (unchanged), got %d", negStock)
	}
}

// Test_P0_Shop_MaterialNegativeStock_OverridePreservedOnPush verifies that after sub-store
// modifies negative stock and marks override, subsequent HQ auto-push (triggered by HQ editing
// material status) preserves the sub-store value under "separate control" mode.
// Note: BatchPush (isForce=true) intentionally overwrites all overrides by design.
func Test_P0_Shop_MaterialNegativeStock_OverridePreservedOnPush(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()

	// HQ material: allow_negative_stock=0, status=0 (下架)
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)

	// Sub-store material: allow_negative_stock=0 (HQ-synced)
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)

	// Set negative_stock to "separate control"
	seedHqControlSetting(t, env.hqDB, "negative_stock", 0)

	// Sub-store enables negative stock via API
	resp := env.storeClient.Post(t, pathMaterialUpdateNegativeStock, map[string]any{
		"uuid":                 materialUuid,
		"allow_negative_stock": true,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// HQ triggers auto-push by changing material status → OnHqMaterialChanged → pushSingleMaterialToStore
	// Auto-push respects override marks (unlike batch push which is force)
	pushResp := env.hqClient.Post(t, pathMaterialStatusBatch, map[string]any{
		"uuids":  []int64{materialUuid},
		"status": 1,
	})
	pushResp.AssertOK(t).AssertSuccess(t)

	// Wait for async push
	time.Sleep(3 * time.Second)

	// Verify: sub-store's value should be preserved (allow_negative_stock=1)
	negStock := queryIntField(t, env.storeDB, "ttpos_material", "allow_negative_stock", materialUuid)
	if negStock != 1 {
		t.Fatalf("expected sub-store allow_negative_stock=1 (override preserved), got %d", negStock)
	}
}

// ========================================================================================
// Test 2: Safety stock preserved during force push
// ========================================================================================

// Test_P0_Shop_MaterialSafetyStock_PreservedDuringForcePush verifies that sub-store's
// overridden safety stock is NOT affected when negative_stock batch push runs with force.
func Test_P0_Shop_MaterialSafetyStock_PreservedDuringForcePush(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	hqSafetyStock := 100.0
	storeSafetyStock := 50.0

	// HQ material: safety_stock=100
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, &hqSafetyStock)

	// Sub-store material: safety_stock=50 (independently modified by sub-store)
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, &storeSafetyStock)

	// Mark safety_stock as overridden in sub-store
	seedHqFieldOverride(t, env.storeDB, materialUuid, "material", "safety_stock")

	// HQ triggers batch push with negative_stock (force — unified control is default)
	// This should NOT affect safety_stock
	resp := env.hqClient.Post(t, pathHqBatchPush, map[string]any{
		"field_types":   []string{"negative_stock"},
		"is_all_stores": true,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Wait for async push
	time.Sleep(2 * time.Second)

	// Verify: sub-store's safety_stock should remain 50 (NOT overwritten to 100)
	safetyStock := queryFloatField(t, env.storeDB, "ttpos_material", "safety_stock", materialUuid)
	if safetyStock != 50.0 {
		t.Fatalf("expected sub-store safety_stock=50 (preserved), got %.1f", safetyStock)
	}

	// Verify: safety_stock override record should still exist
	if !queryOverrideExists(t, env.storeDB, materialUuid, "safety_stock") {
		t.Fatal("expected safety_stock override record to still exist")
	}
}

// ========================================================================================
// Test 3: Material name sync during HQ auto-push
// ========================================================================================

// Test_P0_Shop_MaterialNameSync_OnHqPush verifies that when HQ edits a material and triggers
// auto-push, the sub-store's ttpos_multi_language_name record is updated (not just the name JSON field).
// Route: POST /shop/material/status/batch (triggers OnHqMaterialChanged → pushSingleMaterialToStore)
func Test_P0_Shop_MaterialNameSync_OnHqPush(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()

	// Seed multi_language_name in HQ DB (new name)
	hqNameUuid := fixture.GenerateSnowflakeID()
	seedMultiLanguageName(t, env.hqDB, hqNameUuid, "新物品名称", "New Material Name")

	// Seed HQ material with the name
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)
	updateMaterialField(t, env.hqDB, materialUuid, "multi_language_name_uuid", hqNameUuid)
	updateMaterialField(t, env.hqDB, materialUuid, "name", `{"zh_name":"新物品名称","en_name":"New Material Name"}`)

	// Seed multi_language_name in sub-store DB (old name)
	storeNameUuid := fixture.GenerateSnowflakeID()
	seedMultiLanguageName(t, env.storeDB, storeNameUuid, "旧物品名称", "Old Material Name")

	// Seed sub-store material linked to store's multi_language_name
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)
	updateMaterialField(t, env.storeDB, materialUuid, "multi_language_name_uuid", storeNameUuid)

	// HQ changes material status → triggers OnHqMaterialChanged → pushSingleMaterialToStore
	resp := env.hqClient.Post(t, pathMaterialStatusBatch, map[string]any{
		"uuids":  []int64{materialUuid},
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: sub-store's multi_language_name record should be updated with HQ's name
	waitForAsync(t, 5*time.Second, func() bool {
		zhName := queryStringField(t, env.storeDB, "ttpos_multi_language_name", "zh_name", storeNameUuid)
		return zhName == "新物品名称"
	})

	// Also verify en_name
	enName := queryStringField(t, env.storeDB, "ttpos_multi_language_name", "en_name", storeNameUuid)
	if enName != "New Material Name" {
		t.Fatalf("expected en_name 'New Material Name', got '%s'", enName)
	}
}

// ========================================================================================
// Test 4: Category not in sub-store → defaults to 0
// ========================================================================================

// Test_P0_Shop_MaterialPush_CategoryNotInStore verifies that when HQ pushes a material
// whose category_uuid doesn't exist in the sub-store, the sub-store value defaults to 0.
func Test_P0_Shop_MaterialPush_CategoryNotInStore(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	categoryUuid := fixture.GenerateSnowflakeID()

	// Create category in HQ only (NOT in sub-store)
	seedMaterialCategory(t, env.hqDB, categoryUuid)

	// Create material in HQ with this category
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)
	updateMaterialField(t, env.hqDB, materialUuid, "category_uuid", categoryUuid)

	// Create material in sub-store (HQ-synced) with a different category
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)
	dummyCategoryUuid := fixture.GenerateSnowflakeID()
	updateMaterialField(t, env.storeDB, materialUuid, "category_uuid", dummyCategoryUuid)

	// Trigger push via status change
	resp := env.hqClient.Post(t, pathMaterialStatusBatch, map[string]any{
		"uuids":  []int64{materialUuid},
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: category_uuid should be 0 (HQ category doesn't exist in sub-store)
	waitForAsync(t, 5*time.Second, func() bool {
		catUuid := queryInt64Field(t, env.storeDB, "ttpos_material", "category_uuid", materialUuid)
		return catUuid == 0
	})
}

// Test_P0_Shop_MaterialPush_CategoryExistsInStore verifies that when HQ pushes a material
// whose category_uuid DOES exist in the sub-store, the value is preserved correctly.
func Test_P0_Shop_MaterialPush_CategoryExistsInStore(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	categoryUuid := fixture.GenerateSnowflakeID()

	// Create category in BOTH HQ and sub-store
	seedMaterialCategory(t, env.hqDB, categoryUuid)
	seedMaterialCategory(t, env.storeDB, categoryUuid)

	// Create material in HQ with this category
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)
	updateMaterialField(t, env.hqDB, materialUuid, "category_uuid", categoryUuid)

	// Create material in sub-store (HQ-synced, category=0 initially)
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)

	// Trigger push
	resp := env.hqClient.Post(t, pathMaterialStatusBatch, map[string]any{
		"uuids":  []int64{materialUuid},
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: category_uuid should match HQ value
	waitForAsync(t, 5*time.Second, func() bool {
		catUuid := queryInt64Field(t, env.storeDB, "ttpos_material", "category_uuid", materialUuid)
		return catUuid == categoryUuid
	})
}

// ========================================================================================
// Test 5: Unit fields when ProductUnit not in sub-store
// ========================================================================================

// Test_P0_Shop_MaterialPush_UnitNotInStore verifies that when HQ pushes a material whose
// units reference ProductUnits not in the sub-store, all unit fields default to 0.
func Test_P0_Shop_MaterialPush_UnitNotInStore(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	productUnitUuid := fixture.GenerateSnowflakeID()
	baseUnitUuid := fixture.GenerateSnowflakeID()    // MaterialUnit UUID (base)
	purchaseUnitUuid := fixture.GenerateSnowflakeID() // MaterialUnit UUID (purchase)

	// Create ProductUnit in HQ only (NOT in sub-store)
	seedProductUnit(t, env.hqDB, productUnitUuid, "kg")

	// Create MaterialUnits in HQ
	seedMaterialUnit(t, env.hqDB, baseUnitUuid, materialUuid, productUnitUuid, 1)
	seedMaterialUnit(t, env.hqDB, purchaseUnitUuid, materialUuid, productUnitUuid, 0)

	// Create material in HQ with unit references
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)
	updateMaterialField(t, env.hqDB, materialUuid, "unit_uuid", baseUnitUuid)
	updateMaterialField(t, env.hqDB, materialUuid, "purchase_unit_uuid", purchaseUnitUuid)
	updateMaterialField(t, env.hqDB, materialUuid, "cost_unit_uuid", purchaseUnitUuid)
	updateMaterialField(t, env.hqDB, materialUuid, "default_sales_unit_uuid", purchaseUnitUuid)

	// Create material in sub-store with dummy unit values
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)
	dummyUnitUuid := fixture.GenerateSnowflakeID()
	updateMaterialField(t, env.storeDB, materialUuid, "unit_uuid", dummyUnitUuid)

	// Trigger push
	resp := env.hqClient.Post(t, pathMaterialStatusBatch, map[string]any{
		"uuids":  []int64{materialUuid},
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: all unit fields should be 0 (ProductUnit doesn't exist in sub-store)
	waitForAsync(t, 5*time.Second, func() bool {
		return queryInt64Field(t, env.storeDB, "ttpos_material", "unit_uuid", materialUuid) == 0
	})

	purchaseUnit := queryInt64Field(t, env.storeDB, "ttpos_material", "purchase_unit_uuid", materialUuid)
	if purchaseUnit != 0 {
		t.Fatalf("expected purchase_unit_uuid=0, got %d", purchaseUnit)
	}
	costUnit := queryInt64Field(t, env.storeDB, "ttpos_material", "cost_unit_uuid", materialUuid)
	if costUnit != 0 {
		t.Fatalf("expected cost_unit_uuid=0 (base unit also invalid), got %d", costUnit)
	}
	salesUnit := queryInt64Field(t, env.storeDB, "ttpos_material", "default_sales_unit_uuid", materialUuid)
	if salesUnit != 0 {
		t.Fatalf("expected default_sales_unit_uuid=0, got %d", salesUnit)
	}
}

// Test_P0_Shop_MaterialPush_CostUnitFallbackToBaseUnit verifies that when HQ's cost unit
// references a ProductUnit not in sub-store but the base unit's ProductUnit DOES exist,
// cost_unit_uuid falls back to the base unit (following doUpdateMaterialByErpItem pattern).
func Test_P0_Shop_MaterialPush_CostUnitFallbackToBaseUnit(t *testing.T) {
	env := setupMaterialTestEnv(t)

	materialUuid := fixture.GenerateSnowflakeID()
	productUnitUuid1 := fixture.GenerateSnowflakeID() // base unit — exists in sub-store
	productUnitUuid2 := fixture.GenerateSnowflakeID() // cost unit — NOT in sub-store
	baseUnitUuid := fixture.GenerateSnowflakeID()     // MaterialUnit UUID (base)
	costMaterialUnitUuid := fixture.GenerateSnowflakeID() // MaterialUnit UUID (cost)

	// Create ProductUnit1 in both HQ and sub-store
	seedProductUnit(t, env.hqDB, productUnitUuid1, "kg")
	seedProductUnit(t, env.storeDB, productUnitUuid1, "kg")

	// Create ProductUnit2 in HQ only (NOT in sub-store)
	seedProductUnit(t, env.hqDB, productUnitUuid2, "piece")

	// Create MaterialUnits in HQ
	seedMaterialUnit(t, env.hqDB, baseUnitUuid, materialUuid, productUnitUuid1, 1)
	seedMaterialUnit(t, env.hqDB, costMaterialUnitUuid, materialUuid, productUnitUuid2, 0)

	// Create material in HQ
	seedMaterial(t, env.hqDB, materialUuid, 0, 0, nil)
	updateMaterialField(t, env.hqDB, materialUuid, "unit_uuid", baseUnitUuid)
	updateMaterialField(t, env.hqDB, materialUuid, "cost_unit_uuid", costMaterialUnitUuid)

	// Create material in sub-store
	seedMaterial(t, env.storeDB, materialUuid, env.hqUuid, 0, nil)

	// Trigger push
	resp := env.hqClient.Post(t, pathMaterialStatusBatch, map[string]any{
		"uuids":  []int64{materialUuid},
		"status": 1,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: cost_unit_uuid should fallback to base unit UUID
	waitForAsync(t, 5*time.Second, func() bool {
		costUuid := queryInt64Field(t, env.storeDB, "ttpos_material", "cost_unit_uuid", materialUuid)
		return costUuid == baseUnitUuid
	})

	// Verify: base unit itself should be preserved
	unitUuid := queryInt64Field(t, env.storeDB, "ttpos_material", "unit_uuid", materialUuid)
	if unitUuid != baseUnitUuid {
		t.Fatalf("expected unit_uuid=%d, got %d", baseUnitUuid, unitUuid)
	}
}

// ========================================================================================
// Helpers
// ========================================================================================

// materialTestEnv holds test environment for material push tests.
// Includes both HQ client and sub-store client for testing both perspectives.
type materialTestEnv struct {
	hqClient    *fixture.HTTPClient
	storeClient *fixture.HTTPClient
	hqDB        *sql.DB
	storeDB     *sql.DB
	hqUuid      int64
	storeUuid   int64
}

// setupMaterialTestEnv creates an HQ + 1 sub-store test environment with both HQ and store tokens.
func setupMaterialTestEnv(t *testing.T) materialTestEnv {
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
	hqStaff := fixture.SeedStaff(t, hqDB,
		fixture.WithStaffCompanyUUID(hqUuid),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	// Create sub-store tenant
	storeUUIDStr := fixture.GenerateCompanyUUID(t)
	storeDB := fixture.NewTestTenantFull(t, storeUUIDStr)
	storeUuid := mustParseInt64(storeUUIDStr)

	fixture.SeedCompany(t, storeDB, fixture.WithCompanyUUID(storeUuid))
	// Sub-store needs: ErpnextSiteCode != "" && != "1", CompanyAbbr != HeadquarterAbbr
	// so that IsSubShop() returns true and IsTtposSite() returns false.
	fixture.SeedCompanySetting(t, storeDB,
		fixture.WithCompanySettingCompanyUUID(storeUuid),
		fixture.WithCompanySettingHeadquarterUuid(hqUuid),
		func(o *fixture.CompanySettingOptions) {
			o.ErpnextSiteCode = "test-site"
			o.ErpnextCompanyAbbr = "STORE"
			o.ErpnextHeadquarterAbbr = "HQ"
		},
	)
	storeStaff := fixture.SeedStaff(t, storeDB,
		fixture.WithStaffCompanyUUID(storeUuid),
		fixture.WithStaffIsSuper(1),
	)

	// Seed in saas DB for getSubStoreUuids to discover stores
	saasDB := fixture.NewSaasDB(t)
	fixture.SeedCompany(t, saasDB, fixture.WithCompanyUUID(hqUuid))
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(hqUuid),
		fixture.WithCompanySettingHeadquarterConfig("test-site", "HQ"),
	)
	fixture.SeedCompany(t, saasDB, fixture.WithCompanyUUID(storeUuid))
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(storeUuid),
		fixture.WithCompanySettingHeadquarterUuid(hqUuid),
		func(o *fixture.CompanySettingOptions) {
			o.ErpnextSiteCode = "test-site"
			o.ErpnextCompanyAbbr = "STORE"
			o.ErpnextHeadquarterAbbr = "HQ"
		},
	)

	hqToken := fixture.GenerateShopToken(t, hqUUIDStr, mustParseString(hqStaff.UUID))
	hqClient := fixture.NewHTTPClient().WithToken(hqToken)

	storeToken := fixture.GenerateShopToken(t, storeUUIDStr, mustParseString(storeStaff.UUID))
	storeClient := fixture.NewHTTPClient().WithToken(storeToken)

	return materialTestEnv{
		hqClient:    hqClient,
		storeClient: storeClient,
		hqDB:        hqDB,
		storeDB:     storeDB,
		hqUuid:      hqUuid,
		storeUuid:   storeUuid,
	}
}

// seedHqControlSetting inserts a control setting record in HQ DB.
// controlMode: 0=separate, 1=unified
func seedHqControlSetting(t *testing.T, db *sql.DB, fieldType string, controlMode int) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_hq_control_setting (uuid, field_type, control_mode, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, fixture.GenerateSnowflakeID(), fieldType, controlMode, now, now)
	if err != nil {
		t.Fatalf("failed to seed hq_control_setting: %v", err)
	}
}

// updateMaterialField updates a single field on ttpos_material by uuid.
func updateMaterialField(t *testing.T, db *sql.DB, uuid int64, field string, value any) {
	t.Helper()
	_, err := db.Exec("UPDATE ttpos_material SET `"+field+"` = ? WHERE uuid = ?", value, uuid)
	if err != nil {
		t.Fatalf("failed to update material field %s: %v", field, err)
	}
}

// seedMaterialCategory inserts a material category record.
func seedMaterialCategory(t *testing.T, db *sql.DB, uuid int64) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_material_category (uuid, name, code, create_time, update_time, delete_time)
		VALUES (?, 'Test Category', '', ?, ?, 0)
	`, uuid, now, now)
	if err != nil {
		t.Fatalf("failed to seed material_category: %v", err)
	}
}

// seedProductUnit inserts a product unit record.
func seedProductUnit(t *testing.T, db *sql.DB, uuid int64, name string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_unit (uuid, name, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, 0)
	`, uuid, name, now, now)
	if err != nil {
		t.Fatalf("failed to seed product_unit: %v", err)
	}
}

// seedMaterialUnit inserts a material unit record.
func seedMaterialUnit(t *testing.T, db *sql.DB, uuid, materialUuid, unitUuid int64, isDefault int) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_material_unit (uuid, material_uuid, unit_uuid, name, conversion_rate, is_default, create_time, update_time, delete_time)
		VALUES (?, ?, ?, 'unit', 1, ?, ?, ?, 0)
	`, uuid, materialUuid, unitUuid, isDefault, now, now)
	if err != nil {
		t.Fatalf("failed to seed material_unit: %v", err)
	}
}

// queryInt64Field queries a single int64 field (suitable for UUID fields) from a table by uuid.
func queryInt64Field(t *testing.T, db *sql.DB, table, field string, uuid int64) int64 {
	t.Helper()
	var val int64
	err := db.QueryRow(fmt.Sprintf("SELECT `%s` FROM `%s` WHERE uuid = ? AND delete_time = 0", field, table), uuid).Scan(&val)
	if err != nil {
		return -1
	}
	return val
}
