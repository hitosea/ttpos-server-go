//go:build integration

package desk_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// API path constants
const (
	pathDeskOpen     = "/api/v1/cashier/desk/open"
	pathDeskCartInfo = "/api/v1/cashier/desk/order/cart/info"
)

// codeTokenInvalid matches constant.CodeTokenInvalid in the server (-102).
const codeTokenInvalid = -102

// --- P0 ---

// Test_P0_Cashier_DeskMustPlan_AutoAdd_ConcurrentNoDuplicate tests that concurrent requests
// to the cart/info endpoint do NOT duplicate must-plan auto-add products.
//
// Route: GET /api/v1/cashier/desk/order/cart/info
//
// Bug scenario:
//   - Cashier opens a desk (creates sale_bill with auto_add_must_product=1)
//   - Two concurrent GET /cart/info requests arrive (e.g., cashier + H5 ping)
//   - Both call DeskOrderMustPlan → both check IsAutoAddMustProduct() == true
//   - Without distributed lock: both execute autoAddSaleOrderProductToDesk → duplicated products
//   - With fix: only one request adds products, the other skips via double-check
func Test_P0_Cashier_DeskMustPlan_AutoAdd_ConcurrentNoDuplicate(t *testing.T) {
	const testDeviceId = "test-must-plan-race-device"

	// 1. Create tenant with full schema
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. Seed company + company_setting (required for auth)
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	fixture.SeedCashierPermissions(t, db)

	// 3. Seed device (required for device binding check)
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)

	// 4. Seed staff (super admin, on-duty)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffName("Race Test Cashier"),
		fixture.WithStaffUsername("race-test-cashier"),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-race-001"),
	)

	// 5. Setup must-plan: region → desk → must_plan → must_plan_item → must_plan_region
	region := fixture.SeedDeskRegion(t, db)
	desk := fixture.SeedDesk(t, db,
		fixture.WithDeskName("Race Test Table"),
		fixture.WithDeskRegionUUID(region.UUID),
	)

	// Create a product to be auto-added by must-plan
	productBomUUID := fixture.SeedProductWithFlavor(t, db, "Must Plan Product", 10.0)
	productPackageUUID := getPackageUUIDFromBom(t, db, productBomUUID)

	mustPlan := fixture.SeedMustPlan(t, db,
		fixture.WithMustPlanUseChannel("20"), // desk mode
		fixture.WithMustPlanAutoCart(1),       // auto add to cart
	)
	fixture.SeedMustPlanItem(t, db,
		fixture.WithMustPlanItemPlanUUID(mustPlan.UUID),
		fixture.WithMustPlanItemPackageUUID(productPackageUUID),
	)
	fixture.SeedMustPlanRegion(t, db,
		fixture.WithMustPlanRegionPlanUUID(mustPlan.UUID),
		fixture.WithMustPlanRegionDeskRegionUUID(region.UUID),
	)

	// 6. Setup WireMock for gRPC service stubs
	fixture.SetupWireMock(t)

	// 7. Generate JWT token
	token := fixture.GenerateStaffToken(t,
		companyUUID,
		mustParseString(staff.UUID),
		staff.Username,
		1,
		testDeviceId,
	)
	httpClient := fixture.NewHTTPClient().WithToken(token)

	// 8. Open desk to create sale_bill (auto_add_must_product defaults to 1)
	openResp := httpClient.Post(t, pathDeskOpen, map[string]any{
		"desk_uuid": desk.UUID,
		"meal_num":  2,
	})
	openResp.AssertOK(t).AssertSuccess(t)

	var openResult struct {
		SaleBillUUID  uint64 `json:"sale_bill_uuid"`
		SaleOrderUUID uint64 `json:"sale_order_uuid"`
	}
	apiResp := openResp.ParseAPIResponse(t)
	if err := json.Unmarshal(apiResp.Data, &openResult); err != nil {
		t.Fatalf("failed to unmarshal desk open response: %v", err)
	}
	if openResult.SaleBillUUID == 0 {
		t.Fatal("expected sale_bill_uuid to be non-zero after desk open")
	}

	// 9. Send concurrent cart/info requests (simulates cashier + H5 ping race condition)
	const concurrency = 5
	cartInfoURL := fmt.Sprintf("%s?sale_bill_uuid=%d", pathDeskCartInfo, openResult.SaleBillUUID)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			resp := httpClient.Get(t, cartInfoURL)
			resp.AssertOK(t).AssertSuccess(t)
		}()
	}
	wg.Wait()

	// 10. Verify: must-plan products should exist exactly once (no duplicates)
	mustPlanProductCount := countMustPlanProducts(t, db, openResult.SaleBillUUID)
	if mustPlanProductCount != 1 {
		t.Errorf("expected exactly 1 must-plan product, got %d (race condition caused duplicates)", mustPlanProductCount)
	}

	// Also verify the auto_add_must_product flag is now 0
	autoAddFlag := getAutoAddMustProductFlag(t, db, openResult.SaleBillUUID)
	if autoAddFlag != 0 {
		t.Errorf("expected auto_add_must_product=0 after auto-add, got %d", autoAddFlag)
	}
}

// Test_P0_Cashier_DeskMustPlan_AutoAdd_HappyPath tests the basic happy path:
// opening a desk with must-plan auto-adds the required products exactly once.
//
// Route: GET /api/v1/cashier/desk/order/cart/info
func Test_P0_Cashier_DeskMustPlan_AutoAdd_HappyPath(t *testing.T) {
	const testDeviceId = "test-must-plan-happy-device"

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	fixture.SeedCashierPermissions(t, db)
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffName("Happy Path Cashier"),
		fixture.WithStaffUsername("happy-cashier"),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-happy-001"),
	)

	region := fixture.SeedDeskRegion(t, db)
	desk := fixture.SeedDesk(t, db,
		fixture.WithDeskName("Happy Path Table"),
		fixture.WithDeskRegionUUID(region.UUID),
	)

	productBomUUID := fixture.SeedProductWithFlavor(t, db, "Happy Must Plan Product", 8.0)
	productPackageUUID := getPackageUUIDFromBom(t, db, productBomUUID)

	mustPlan := fixture.SeedMustPlan(t, db,
		fixture.WithMustPlanUseChannel("20"),
		fixture.WithMustPlanAutoCart(1),
	)
	fixture.SeedMustPlanItem(t, db,
		fixture.WithMustPlanItemPlanUUID(mustPlan.UUID),
		fixture.WithMustPlanItemPackageUUID(productPackageUUID),
	)
	fixture.SeedMustPlanRegion(t, db,
		fixture.WithMustPlanRegionPlanUUID(mustPlan.UUID),
		fixture.WithMustPlanRegionDeskRegionUUID(region.UUID),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateStaffToken(t,
		companyUUID,
		mustParseString(staff.UUID),
		staff.Username,
		1,
		testDeviceId,
	)
	httpClient := fixture.NewHTTPClient().WithToken(token)

	// Open desk
	openResp := httpClient.Post(t, pathDeskOpen, map[string]any{
		"desk_uuid": desk.UUID,
		"meal_num":  2,
	})
	openResp.AssertOK(t).AssertSuccess(t)

	var openResult struct {
		SaleBillUUID uint64 `json:"sale_bill_uuid"`
	}
	apiResp := openResp.ParseAPIResponse(t)
	if err := json.Unmarshal(apiResp.Data, &openResult); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Query cart info (triggers auto-add)
	cartInfoURL := fmt.Sprintf("%s?sale_bill_uuid=%d", pathDeskCartInfo, openResult.SaleBillUUID)
	cartResp := httpClient.Get(t, cartInfoURL)
	cartResp.AssertOK(t).AssertSuccess(t)

	// Verify exactly 1 must-plan product was added
	mustPlanProductCount := countMustPlanProducts(t, db, openResult.SaleBillUUID)
	if mustPlanProductCount != 1 {
		t.Errorf("expected 1 must-plan product, got %d", mustPlanProductCount)
	}

	// Second request should NOT add again
	cartResp2 := httpClient.Get(t, cartInfoURL)
	cartResp2.AssertOK(t).AssertSuccess(t)

	mustPlanProductCount2 := countMustPlanProducts(t, db, openResult.SaleBillUUID)
	if mustPlanProductCount2 != 1 {
		t.Errorf("expected still 1 must-plan product after second request, got %d", mustPlanProductCount2)
	}
}

// Test_P0_Cashier_DeskCartInfo_Unauthorized tests that requests without a token are rejected.
// Route: GET /api/v1/cashier/desk/order/cart/info
func Test_P0_Cashier_DeskCartInfo_Unauthorized(t *testing.T) {
	httpClient := fixture.NewHTTPClient() // No token
	resp := httpClient.Get(t, pathDeskCartInfo+"?sale_bill_uuid=123")
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- Helpers ---

func mustParseInt64(s string) int64 {
	var i int64
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0
	}
	return i
}

func mustParseString(i int64) string {
	return fmt.Sprintf("%d", i)
}

// getPackageUUIDFromBom queries the product_package_uuid for a given product_bom UUID.
func getPackageUUIDFromBom(t *testing.T, db *sql.DB, bomUUID int64) int64 {
	t.Helper()
	var packageUUID int64
	err := db.QueryRow(
		"SELECT product_package_uuid FROM ttpos_product_bom WHERE uuid = ?", bomUUID,
	).Scan(&packageUUID)
	if err != nil {
		t.Fatalf("failed to get product_package_uuid from bom %d: %v", bomUUID, err)
	}
	return packageUUID
}

// countMustPlanProducts counts must-plan products (must_plan_uuid > 0) for a given sale_bill.
func countMustPlanProducts(t *testing.T, db *sql.DB, saleBillUUID uint64) int {
	t.Helper()
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM ttpos_sale_order_product sop
		JOIN ttpos_sale_order so ON sop.sale_order_uuid = so.uuid
		WHERE so.sale_bill_uuid = ?
		  AND sop.delete_time = 0
		  AND so.delete_time = 0
		  AND sop.must_plan_uuid > 0
	`, saleBillUUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count must-plan products: %v", err)
	}
	return count
}

// getAutoAddMustProductFlag reads the auto_add_must_product flag from sale_bill.
func getAutoAddMustProductFlag(t *testing.T, db *sql.DB, saleBillUUID uint64) int {
	t.Helper()
	var flag int
	err := db.QueryRow(
		"SELECT auto_add_must_product FROM ttpos_sale_bill WHERE uuid = ?", saleBillUUID,
	).Scan(&flag)
	if err != nil {
		t.Fatalf("failed to get auto_add_must_product flag: %v", err)
	}
	return flag
}
