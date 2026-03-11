//go:build integration

package product_test

import (
	"fmt"
	"os"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// codeTokenInvalid matches constant.CodeTokenInvalid in the server (-102).
const codeTokenInvalid = -102

// API path constants
const (
	pathProductFlavorList = "/api/v1/shop/product/flavor/list"
	pathProductFlavorAdd  = "/api/v1/shop/product/flavor/add"
	pathProductFlavorEdit = "/api/v1/shop/product/flavor/edit"
)

// --- P0 ---

// Test_P0_Shop_ProductFlavor_List_HappyPath tests listing product flavors for an authenticated shop user.
// Route: GET /shop/product/flavor/list
func Test_P0_Shop_ProductFlavor_List_HappyPath(t *testing.T) {
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
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathProductFlavorList)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_ProductFlavor_Add_HappyPath tests creating a new product flavor.
// Route: POST /shop/product/flavor/add
func Test_P0_Shop_ProductFlavor_Add_HappyPath(t *testing.T) {
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
	body := map[string]any{
		"locale_name": map[string]string{
			"zh_name": "大杯",
			"en_name": "Large Cup",
		},
	}
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathProductFlavorAdd, body)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_ProductFlavor_List_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/product/flavor/list
func Test_P0_Shop_ProductFlavor_List_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathProductFlavorList)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- P1 ---

// Helpers

func mustParseInt64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

func mustParseString(i int64) string {
	return fmt.Sprintf("%d", i)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
