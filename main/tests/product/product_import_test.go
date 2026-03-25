//go:build integration

package product_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	pathProductImportList = "/api/v1/shop/product/import/list"
	pathProductImport     = "/api/v1/shop/product/import"

	codeTokenInvalidImport = -102 // constant.CodeTokenInvalid
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Shop_ProductImport_Import_DeliveryScanOrderEnabled verifies that
// is_show_delivery is preserved when scan-order (IsOpenMemberInstant) is enabled
// but rider delivery (DeliveryStatus) is off.
// This is the core fix: previously delivery was silently reset to 0 in this scenario.
// Route: POST /api/v1/shop/product/import
func Test_P0_Shop_ProductImport_Import_DeliveryScanOrderEnabled(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingDeliveryStatus(0),          // rider OFF
		fixture.WithCompanySettingIsOpenMemberInstant(1),     // scan order ON
		fixture.WithCompanySettingEnableKiosk(0),             // kiosk OFF
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	// Seed prerequisite data
	catUUID := seedProductCategory(t, db)
	unitUUID := seedProductUnit(t, db)
	flavorUUID := seedProductFlavor(t, db)

	body := map[string]any{
		"list": []map[string]any{
			{
				"locale_name":    map[string]string{"en": "Import Delivery Test"},
				"category_uuid":  catUUID,
				"unit_uuid":      unitUUID,
				"sku_uuid":       flavorUUID,
				"product_price":  10.0,
				"product_status": 1,
				"num_type":       1,
				"deduct_stock_type": 1,
				"is_show_cashier":   1,
				"is_show_delivery":  1, // user selected delivery
				"is_show_kiosk":     0,
				"row":              1,
			},
		},
	}

	resp := client.Post(t, pathProductImport, body)
	resp.AssertOK(t).AssertSuccess(t)

	// Wait for async import and verify DB
	pkg := pollProductPackage(t, db, "Import Delivery Test", 5*time.Second)
	if pkg.isShowDelivery != 1 {
		t.Errorf("is_show_delivery = %d, want 1 (scan order enabled should preserve delivery)", pkg.isShowDelivery)
	}
}

// Test_P0_Shop_ProductImport_Import_DeliveryRiderEnabled verifies that
// is_show_delivery is preserved when rider delivery (DeliveryStatus) is on.
// Route: POST /api/v1/shop/product/import
func Test_P0_Shop_ProductImport_Import_DeliveryRiderEnabled(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingDeliveryStatus(1),          // rider ON
		fixture.WithCompanySettingIsOpenMemberInstant(0),     // scan order OFF
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	catUUID := seedProductCategory(t, db)
	unitUUID := seedProductUnit(t, db)
	flavorUUID := seedProductFlavor(t, db)

	body := map[string]any{
		"list": []map[string]any{
			{
				"locale_name":    map[string]string{"en": "Import Rider Test"},
				"category_uuid":  catUUID,
				"unit_uuid":      unitUUID,
				"sku_uuid":       flavorUUID,
				"product_price":  10.0,
				"product_status": 1,
				"num_type":       1,
				"deduct_stock_type": 1,
				"is_show_cashier":   1,
				"is_show_delivery":  1,
				"row":              1,
			},
		},
	}

	resp := client.Post(t, pathProductImport, body)
	resp.AssertOK(t).AssertSuccess(t)

	pkg := pollProductPackage(t, db, "Import Rider Test", 5*time.Second)
	if pkg.isShowDelivery != 1 {
		t.Errorf("is_show_delivery = %d, want 1 (rider enabled should preserve delivery)", pkg.isShowDelivery)
	}
}

// Test_P0_Shop_ProductImport_Import_Unauthorized verifies unauthenticated requests are rejected.
// Route: POST /api/v1/shop/product/import
func Test_P0_Shop_ProductImport_Import_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Post(t, pathProductImport, nil)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalidImport)
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Shop_ProductImport_Import_DeliveryBothOff verifies that is_show_delivery
// is reset to 0 when both rider and scan-order are disabled.
// Route: POST /api/v1/shop/product/import
func Test_P1_Shop_ProductImport_Import_DeliveryBothOff(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingDeliveryStatus(0),          // rider OFF
		fixture.WithCompanySettingIsOpenMemberInstant(0),     // scan order OFF
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	catUUID := seedProductCategory(t, db)
	unitUUID := seedProductUnit(t, db)
	flavorUUID := seedProductFlavor(t, db)

	body := map[string]any{
		"list": []map[string]any{
			{
				"locale_name":    map[string]string{"en": "Import BothOff Test"},
				"category_uuid":  catUUID,
				"unit_uuid":      unitUUID,
				"sku_uuid":       flavorUUID,
				"product_price":  10.0,
				"product_status": 1,
				"num_type":       1,
				"deduct_stock_type": 1,
				"is_show_cashier":   1,
				"is_show_delivery":  1, // user selected, but both features off
				"row":              1,
			},
		},
	}

	resp := client.Post(t, pathProductImport, body)
	resp.AssertOK(t).AssertSuccess(t)

	pkg := pollProductPackage(t, db, "Import BothOff Test", 5*time.Second)
	if pkg.isShowDelivery != 0 {
		t.Errorf("is_show_delivery = %d, want 0 (both rider and scan order off)", pkg.isShowDelivery)
	}
}

// Test_P1_Shop_ProductImport_Import_KioskEnabled verifies that is_show_kiosk is
// preserved when kiosk feature is enabled.
// Route: POST /api/v1/shop/product/import
func Test_P1_Shop_ProductImport_Import_KioskEnabled(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableKiosk(1), // kiosk ON
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	catUUID := seedProductCategory(t, db)
	unitUUID := seedProductUnit(t, db)
	flavorUUID := seedProductFlavor(t, db)

	body := map[string]any{
		"list": []map[string]any{
			{
				"locale_name":    map[string]string{"en": "Import Kiosk On Test"},
				"category_uuid":  catUUID,
				"unit_uuid":      unitUUID,
				"sku_uuid":       flavorUUID,
				"product_price":  10.0,
				"product_status": 1,
				"num_type":       1,
				"deduct_stock_type": 1,
				"is_show_cashier":   1,
				"is_show_kiosk":     1, // user selected kiosk
				"row":              1,
			},
		},
	}

	resp := client.Post(t, pathProductImport, body)
	resp.AssertOK(t).AssertSuccess(t)

	pkg := pollProductPackage(t, db, "Import Kiosk On Test", 5*time.Second)
	if pkg.isShowKiosk != 1 {
		t.Errorf("is_show_kiosk = %d, want 1 (kiosk enabled)", pkg.isShowKiosk)
	}
}

// Test_P1_Shop_ProductImport_Import_KioskDisabled verifies that is_show_kiosk is
// reset to 0 when kiosk feature is disabled.
// Route: POST /api/v1/shop/product/import
func Test_P1_Shop_ProductImport_Import_KioskDisabled(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableKiosk(0), // kiosk OFF
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	catUUID := seedProductCategory(t, db)
	unitUUID := seedProductUnit(t, db)
	flavorUUID := seedProductFlavor(t, db)

	body := map[string]any{
		"list": []map[string]any{
			{
				"locale_name":    map[string]string{"en": "Import Kiosk Off Test"},
				"category_uuid":  catUUID,
				"unit_uuid":      unitUUID,
				"sku_uuid":       flavorUUID,
				"product_price":  10.0,
				"product_status": 1,
				"num_type":       1,
				"deduct_stock_type": 1,
				"is_show_cashier":   1,
				"is_show_kiosk":     1, // user selected, but kiosk off
				"row":              1,
			},
		},
	}

	resp := client.Post(t, pathProductImport, body)
	resp.AssertOK(t).AssertSuccess(t)

	pkg := pollProductPackage(t, db, "Import Kiosk Off Test", 5*time.Second)
	if pkg.isShowKiosk != 0 {
		t.Errorf("is_show_kiosk = %d, want 0 (kiosk disabled)", pkg.isShowKiosk)
	}
}

// ---------------------------------------------------------------------------
// P2 — Edge Cases
// ---------------------------------------------------------------------------

// Test_P2_Shop_ProductImport_Import_DecimalPricingOverridesDelivery verifies
// that decimal pricing forces is_show_delivery=0 even when delivery features are enabled.
// Note: kiosk is NOT affected by decimal pricing in save path — only gated by IsOpenKiosk().
// Route: POST /api/v1/shop/product/import
func Test_P2_Shop_ProductImport_Import_DecimalPricingOverridesDelivery(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingDeliveryStatus(1),      // rider ON
		fixture.WithCompanySettingEnableKiosk(1),          // kiosk ON
		fixture.WithCompanySettingIsOpenMemberInstant(1),  // scan order ON
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	catUUID := seedProductCategory(t, db)
	unitUUID := seedProductUnit(t, db)
	flavorUUID := seedProductFlavor(t, db)

	body := map[string]any{
		"list": []map[string]any{
			{
				"locale_name":    map[string]string{"en": "Import Decimal Test"},
				"category_uuid":  catUUID,
				"unit_uuid":      unitUUID,
				"sku_uuid":       flavorUUID,
				"product_price":  10.0,
				"product_status": 1,
				"num_type":       2, // decimal pricing
				"deduct_stock_type": 1,
				"is_show_cashier":   1,
				"is_show_delivery":  1, // should be overridden by decimal pricing
				"is_show_kiosk":     1, // NOT overridden — kiosk only gated by IsOpenKiosk()
				"row":              1,
			},
		},
	}

	resp := client.Post(t, pathProductImport, body)
	resp.AssertOK(t).AssertSuccess(t)

	pkg := pollProductPackage(t, db, "Import Decimal Test", 5*time.Second)
	if pkg.isShowDelivery != 0 {
		t.Errorf("is_show_delivery = %d, want 0 (decimal pricing overrides delivery)", pkg.isShowDelivery)
	}
	// kiosk should remain 1 — decimal pricing does not affect kiosk in save path
	if pkg.isShowKiosk != 1 {
		t.Errorf("is_show_kiosk = %d, want 1 (kiosk enabled, not affected by decimal pricing)", pkg.isShowKiosk)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type productPackageRow struct {
	uuid           int64
	isShowDelivery int
	isShowKiosk    int
}

// pollProductPackage polls the DB for a product_package with the given name (EN).
// Retries every 500ms until found or timeout.
func pollProductPackage(t *testing.T, db *sql.DB, enName string, timeout time.Duration) productPackageRow {
	t.Helper()
	deadline := time.Now().Add(timeout)
	namePattern := fmt.Sprintf("%%%s%%", enName)

	for time.Now().Before(deadline) {
		var row productPackageRow
		err := db.QueryRow(`
			SELECT uuid, is_show_delivery, is_show_kiosk
			FROM ttpos_product_package
			WHERE name LIKE ? AND delete_time = 0
			ORDER BY id DESC LIMIT 1
		`, namePattern).Scan(&row.uuid, &row.isShowDelivery, &row.isShowKiosk)
		if err == nil && row.uuid != 0 {
			return row
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("product_package with name containing %q not found within %v", enName, timeout)
	return productPackageRow{}
}

// seedProductCategory inserts a product category and returns its UUID.
func seedProductCategory(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	uuid := fixture.GenerateSnowflakeID()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_category (uuid, name, status, create_time, update_time, delete_time)
		VALUES (?, '{"en":"Test Category"}', 1, ?, ?, 0)
	`, uuid, now, now)
	if err != nil {
		t.Fatalf("seed product_category: %v", err)
	}
	return uuid
}

// seedProductUnit inserts a product unit and returns its UUID.
func seedProductUnit(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	uuid := fixture.GenerateSnowflakeID()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_unit (uuid, name, create_time, update_time, delete_time)
		VALUES (?, '{"en":"pcs"}', ?, ?, 0)
	`, uuid, now, now)
	if err != nil {
		t.Fatalf("seed product_unit: %v", err)
	}
	return uuid
}

// seedProductFlavor inserts a product flavor (spec) and returns its UUID.
func seedProductFlavor(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	uuid := fixture.GenerateSnowflakeID()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_flavor (uuid, name, create_time, update_time, delete_time)
		VALUES (?, '{"en":"Default"}', ?, ?, 0)
	`, uuid, now, now)
	if err != nil {
		t.Fatalf("seed product_flavor: %v", err)
	}
	return uuid
}

// mustParseInt64 and mustParseString are defined in product_flavor_test.go
