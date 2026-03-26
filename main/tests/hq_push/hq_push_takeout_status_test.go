//go:build integration

package hq_push_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// --- Constants ---

const (
	pathTakeoutBatchOnline     = "/api/v1/shop/takeout/products/batch_online"
	pathTakeoutBatchOffline    = "/api/v1/shop/takeout/products/batch_offline"
	pathHqControlSetting       = "/api/v1/shop/product/hq_control_setting"
)

// ========================================================================================
// Fix 1: Sub-store batch online/offline blocked under unified control
// ========================================================================================

// Test_P0_TakeoutStatus_SubStoreBatchOnline_UnifiedControl_Ignored verifies that when
// HQ sets takeout_shelf to unified control, sub-store batch online requests are silently
// ignored — the takeout product status remains unchanged.
// Route: POST /api/v1/shop/takeout/products/batch_online
func Test_P0_TakeoutStatus_SubStoreBatchOnline_UnifiedControl_Ignored(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()
	storeTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: takeout status=0 (下架)
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 0, 10.0)

	// Sub-store: takeout status=0 (下架), linked to HQ
	fixture.SeedProductPackage(t, env.storeDB,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(1),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)
	seedProductPackageTakeout(t, env.storeDB, storeTakeoutUuid, productUuid, env.hqUuid, 1, 0, 10.0)

	// Set takeout_shelf to unified control in HQ DB
	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 1) // 1 = unified

	// Sub-store tries batch online → should be silently ignored
	resp := env.storeClient.Post(t, pathTakeoutBatchOnline, map[string]any{
		"platform":      "grab",
		"product_uuids": []int64{productUuid},
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Wait a moment for any async processing
	time.Sleep(1 * time.Second)

	// Verify: sub-store takeout status should remain 0 (下架)
	storeStatus := queryIntField(t, env.storeDB, "ttpos_product_package_takeout", "status", storeTakeoutUuid)
	if storeStatus != 0 {
		t.Fatalf("expected sub-store takeout status=0 (unified control should block), got %d", storeStatus)
	}
}

// Test_P0_TakeoutStatus_SubStoreBatchOffline_UnifiedControl_Ignored verifies that when
// HQ sets takeout_shelf to unified control, sub-store batch offline requests are silently ignored.
// Route: POST /api/v1/shop/takeout/products/batch_offline
func Test_P0_TakeoutStatus_SubStoreBatchOffline_UnifiedControl_Ignored(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()
	storeTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: takeout status=1 (上架)
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 1, 10.0)

	// Sub-store: takeout status=1 (上架), linked to HQ
	fixture.SeedProductPackage(t, env.storeDB,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(1),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)
	seedProductPackageTakeout(t, env.storeDB, storeTakeoutUuid, productUuid, env.hqUuid, 1, 1, 10.0)

	// Set takeout_shelf to unified control
	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 1)

	// Sub-store tries batch offline → should be silently ignored
	resp := env.storeClient.Post(t, pathTakeoutBatchOffline, map[string]any{
		"platform":      "grab",
		"product_uuids": []int64{productUuid},
	})
	resp.AssertOK(t).AssertSuccess(t)

	time.Sleep(1 * time.Second)

	// Verify: status should remain 1 (上架)
	storeStatus := queryIntField(t, env.storeDB, "ttpos_product_package_takeout", "status", storeTakeoutUuid)
	if storeStatus != 1 {
		t.Fatalf("expected sub-store takeout status=1 (unified control should block), got %d", storeStatus)
	}
}

// Test_P0_TakeoutStatus_SubStoreBatchOnline_SeparateControl_Allowed verifies that when
// HQ sets takeout_shelf to separate control, sub-store batch online works normally.
// Route: POST /api/v1/shop/takeout/products/batch_online
func Test_P0_TakeoutStatus_SubStoreBatchOnline_SeparateControl_Allowed(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()
	storeTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: takeout status=0
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 0, 10.0)

	// Sub-store: takeout status=0
	fixture.SeedProductPackage(t, env.storeDB,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(1),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)
	seedProductPackageTakeout(t, env.storeDB, storeTakeoutUuid, productUuid, env.hqUuid, 1, 0, 10.0)

	// Set takeout_shelf to separate control
	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 0) // 0 = separate

	// Sub-store batch online → should succeed
	resp := env.storeClient.Post(t, pathTakeoutBatchOnline, map[string]any{
		"platform":      "grab",
		"product_uuids": []int64{productUuid},
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: status should be updated to 1
	waitForAsync(t, 5*time.Second, func() bool {
		return queryIntField(t, env.storeDB, "ttpos_product_package_takeout", "status", storeTakeoutUuid) == 1
	})
}

// ========================================================================================
// Fix 2: First sync follows HQ status under unified control
// ========================================================================================

// Test_P0_TakeoutStatus_FirstSync_UnifiedControl_FollowsHqStatus verifies that when
// HQ modifies a takeout product status (triggering auto-push to sub-stores),
// and the sub-store doesn't have the takeout product yet, under unified control,
// the newly created sub-store takeout inherits HQ's status (not default 0).
// Route: POST /api/v1/shop/takeout/products/batch_online (HQ triggers OnHqProductTakeoutChanged → pushSingleTakeoutToStore → createTakeoutInStore)
func Test_P0_TakeoutStatus_FirstSync_UnifiedControl_FollowsHqStatus(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: has product and takeout (status=0, will be set to 1 via batch_online)
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 0, 15.0)

	// Sub-store: has the dine-in product but NO takeout product yet
	fixture.SeedProductPackage(t, env.storeDB,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(1),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)

	// Set takeout_shelf to unified control
	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 1)

	// HQ sets takeout to online → triggers auto-push → createTakeoutInStore with unified control
	resp := env.hqClient.Post(t, pathTakeoutBatchOnline, map[string]any{
		"platform":      "grab",
		"product_uuids": []int64{productUuid},
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify: sub-store should have a new takeout product with status=1 (following HQ)
	waitForAsync(t, 10*time.Second, func() bool {
		status := queryTakeoutStatusByProductAndType(t, env.storeDB, productUuid, 1)
		return status == 1
	})
}

// Test_P0_TakeoutStatus_FirstSync_SeparateControl_DefaultOffline verifies that when
// HQ modifies a takeout product (triggering auto-push), under separate control,
// the newly created sub-store takeout defaults to status=0 (下架).
// Route: POST /api/v1/shop/takeout/products/batch_online (HQ triggers auto-push)
func Test_P0_TakeoutStatus_FirstSync_SeparateControl_DefaultOffline(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	productUuid := fixture.GenerateSnowflakeID()
	hqTakeoutUuid := fixture.GenerateSnowflakeID()

	// HQ: has product and takeout (status=0, will be set to 1 via batch_online)
	fixture.SeedProductPackage(t, env.hqDB, fixture.WithProductPackageUUID(productUuid), fixture.WithProductPackageStatus(1))
	seedProductPackageTakeout(t, env.hqDB, hqTakeoutUuid, productUuid, 0, 1, 0, 15.0)

	// Sub-store: has dine-in product, no takeout
	fixture.SeedProductPackage(t, env.storeDB,
		fixture.WithProductPackageUUID(productUuid),
		fixture.WithProductPackageStatus(1),
		fixture.WithProductPackageHeadquarterUuid(env.hqUuid),
	)

	// Set takeout_shelf to separate control
	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 0)

	// HQ sets takeout online → triggers auto-push → createTakeoutInStore with separate control
	resp := env.hqClient.Post(t, pathTakeoutBatchOnline, map[string]any{
		"platform":      "grab",
		"product_uuids": []int64{productUuid},
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Wait for async push and verify: sub-store takeout should be created with status=0
	// We need to wait for the takeout to appear first, then check its status
	waitForAsync(t, 10*time.Second, func() bool {
		status := queryTakeoutStatusByProductAndType(t, env.storeDB, productUuid, 1)
		return status >= 0 // takeout exists (status is 0 or 1, not -1)
	})

	status := queryTakeoutStatusByProductAndType(t, env.storeDB, productUuid, 1)
	if status != 0 {
		t.Fatalf("expected sub-store takeout status=0 (separate control default), got %d", status)
	}
}

// ========================================================================================
// Fix 3: GetControlSetting accessible to sub-stores
// ========================================================================================

// Test_P0_TakeoutStatus_GetControlSetting_SubStore_ReturnsHqSettings verifies that
// sub-store can query HQ's control settings via GET /shop/product/hq_control_setting.
// Route: GET /api/v1/shop/product/hq_control_setting
func Test_P0_TakeoutStatus_GetControlSetting_SubStore_ReturnsHqSettings(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	// Set control settings in HQ DB
	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 1) // unified
	seedHqControlSetting(t, env.hqDB, "dine_shelf", 0)    // separate

	// Sub-store queries control settings
	resp := env.storeClient.Get(t, pathHqControlSetting)
	resp.AssertOK(t).AssertSuccess(t)

	// Parse response and verify
	apiResp := resp.ParseAPIResponse(t)
	var data hqControlSettingData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse control setting data: %v", err)
	}

	if data.HqControlTakeoutShelf != 1 {
		t.Fatalf("expected hq_control_takeout_shelf=1 (unified), got %d", data.HqControlTakeoutShelf)
	}
	if data.HqControlDineShelf != 0 {
		t.Fatalf("expected hq_control_dine_shelf=0 (separate), got %d", data.HqControlDineShelf)
	}
}

// Test_P0_TakeoutStatus_GetControlSetting_HQ_StillWorks verifies that
// HQ can still query its own control settings (regression test).
// Route: GET /api/v1/shop/product/hq_control_setting
func Test_P0_TakeoutStatus_GetControlSetting_HQ_StillWorks(t *testing.T) {
	env := setupTakeoutStatusEnv(t)

	seedHqControlSetting(t, env.hqDB, "takeout_shelf", 1)

	resp := env.hqClient.Get(t, pathHqControlSetting)
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data hqControlSettingData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse control setting data: %v", err)
	}

	if data.HqControlTakeoutShelf != 1 {
		t.Fatalf("expected hq_control_takeout_shelf=1, got %d", data.HqControlTakeoutShelf)
	}
}

// Test_P0_TakeoutStatus_GetControlSetting_NonHqNonSubStore_Rejected verifies that
// standalone stores (non-HQ, non-sub-store) are rejected.
// Route: GET /api/v1/shop/product/hq_control_setting
func Test_P0_TakeoutStatus_GetControlSetting_NonHqNonSubStore_Rejected(t *testing.T) {
	// Create a standalone company (not HQ, not sub-store)
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	resp := client.Get(t, pathHqControlSetting)
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Fatal("expected error for standalone (non-HQ, non-sub-store) company")
	}
}

// ========================================================================================
// Helpers
// ========================================================================================

// hqControlSettingData mirrors resp.HqControlSettingResp for JSON parsing in tests.
type hqControlSettingData struct {
	HqControlDineShelf     int `json:"hq_control_dine_shelf"`
	HqControlTakeoutShelf  int `json:"hq_control_takeout_shelf"`
	HqControlTakeoutPrice  int `json:"hq_control_takeout_price"`
	HqControlNegativeStock int `json:"hq_control_negative_stock"`
}

// takeoutStatusTestEnv holds test environment with both HQ and sub-store clients.
type takeoutStatusTestEnv struct {
	hqClient    *fixture.HTTPClient
	storeClient *fixture.HTTPClient
	hqDB        *sql.DB
	storeDB     *sql.DB
	hqUuid      int64
	storeUuid   int64
}

// setupTakeoutStatusEnv creates an HQ + 1 sub-store environment with both HQ and store tokens.
func setupTakeoutStatusEnv(t *testing.T) takeoutStatusTestEnv {
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

	return takeoutStatusTestEnv{
		hqClient:    hqClient,
		storeClient: storeClient,
		hqDB:        hqDB,
		storeDB:     storeDB,
		hqUuid:      hqUuid,
		storeUuid:   storeUuid,
	}
}

// queryTakeoutStatusByProductAndType queries takeout status by product_package_uuid and takeout_type.
func queryTakeoutStatusByProductAndType(t *testing.T, db *sql.DB, productPackageUuid int64, takeoutType int) int {
	t.Helper()
	var status int
	err := db.QueryRow(
		"SELECT `status` FROM `ttpos_product_package_takeout` WHERE product_package_uuid = ? AND takeout_type = ? AND delete_time = 0",
		productPackageUuid, takeoutType,
	).Scan(&status)
	if err != nil {
		return -1
	}
	return status
}
