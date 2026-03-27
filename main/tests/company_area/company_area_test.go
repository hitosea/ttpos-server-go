//go:build integration

package company_area_test

import (
	"fmt"
	"strings"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// Error codes matching constant package values.
const (
	codeTokenInvalid = -102 // constant.CodeTokenInvalid
	codeParamError   = -3   // constant.CodeParamError
)

// API path constants
const (
	pathCompanyAreaList       = "/api/v1/shop/company_area/list"
	pathCompanyAreaInfo       = "/api/v1/shop/company_area/info"
	pathCompanyAreaCreate     = "/api/v1/shop/company_area/create"
	pathCompanyAreaUpdate     = "/api/v1/shop/company_area/update"
	pathCompanyAreaDelete     = "/api/v1/shop/company_area/delete"
	pathCompanyAreaUnassigned = "/api/v1/shop/company_area/unassigned"
	pathCompanyAreaShopList   = "/api/v1/shop/company_area/shop_list"
)

// --- P0 ---

// Test_P0_Shop_CompanyArea_List_HappyPath tests listing company areas for an authenticated HQ user.
// Route: GET /shop/company_area/list
func Test_P0_Shop_CompanyArea_List_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathCompanyAreaList)
	resp.AssertOK(t).AssertSuccess(t)

	// Verify response contains summary fields and list
	data := resp.JSONData(t)
	if _, ok := data["total_shop_count"]; !ok {
		t.Error("response missing total_shop_count field")
	}
	if _, ok := data["unassigned_shop_count"]; !ok {
		t.Error("response missing unassigned_shop_count field")
	}
	if _, ok := data["list"]; !ok {
		t.Error("response missing list field")
	}
}

// Test_P0_Shop_CompanyArea_Create_HappyPath tests creating a new company area.
// Route: POST /shop/company_area/create
func Test_P0_Shop_CompanyArea_Create_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	body := map[string]any{
		"locale_name": map[string]string{
			"zh": "华南区",
			"en": "South China",
		},
		"sort": 1,
	}
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathCompanyAreaCreate, body)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_CompanyArea_Info_HappyPath tests getting area detail after creation.
// Route: GET /shop/company_area/info
func Test_P0_Shop_CompanyArea_Info_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	client := fixture.NewHTTPClient().WithToken(
		fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID)),
	)

	// Create an area first
	createBody := map[string]any{
		"locale_name": map[string]string{"en": "North"},
	}
	client.Post(t, pathCompanyAreaCreate, createBody).AssertOK(t).AssertSuccess(t)

	// Get area list to obtain the UUID
	listResp := client.Get(t, pathCompanyAreaList)
	listResp.AssertOK(t).AssertSuccess(t)
	listData := listResp.JSONData(t)
	list := listData["list"].([]any)
	if len(list) == 0 {
		t.Fatal("expected at least 1 area after creation")
	}
	areaUuid := list[0].(map[string]any)["uuid"]

	// Get area info
	infoResp := client.Get(t, fmt.Sprintf("%s?uuid=%v", pathCompanyAreaInfo, areaUuid))
	infoResp.AssertOK(t).AssertSuccess(t)

	infoData := infoResp.JSONData(t)
	if _, ok := infoData["shop_list"]; !ok {
		t.Error("response missing shop_list field")
	}
}

// Test_P0_Shop_CompanyArea_Update_HappyPath tests updating an existing area.
// Route: POST /shop/company_area/update
func Test_P0_Shop_CompanyArea_Update_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	client := fixture.NewHTTPClient().WithToken(
		fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID)),
	)

	// Create area
	client.Post(t, pathCompanyAreaCreate, map[string]any{
		"locale_name": map[string]string{"en": "East"},
	}).AssertOK(t).AssertSuccess(t)

	// Get UUID
	listData := client.Get(t, pathCompanyAreaList).AssertOK(t).AssertSuccess(t).JSONData(t)
	areaUuid := listData["list"].([]any)[0].(map[string]any)["uuid"]

	// Update area
	updateBody := map[string]any{
		"uuid":        areaUuid,
		"locale_name": map[string]string{"en": "East Updated"},
		"sort":        2,
	}
	resp := client.Post(t, pathCompanyAreaUpdate, updateBody)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_CompanyArea_Delete_HappyPath tests deleting an area.
// Route: DELETE /shop/company_area/delete
func Test_P0_Shop_CompanyArea_Delete_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	client := fixture.NewHTTPClient().WithToken(
		fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID)),
	)

	// Create area
	client.Post(t, pathCompanyAreaCreate, map[string]any{
		"locale_name": map[string]string{"en": "ToDelete"},
	}).AssertOK(t).AssertSuccess(t)

	// Get UUID
	listData := client.Get(t, pathCompanyAreaList).AssertOK(t).AssertSuccess(t).JSONData(t)
	areaUuid := listData["list"].([]any)[0].(map[string]any)["uuid"]

	// Delete area
	deleteBody := map[string]any{"uuid": areaUuid}
	resp := client.Delete(t, pathCompanyAreaDelete, deleteBody)
	resp.AssertOK(t).AssertSuccess(t)

	// Verify area is gone
	listData2 := client.Get(t, pathCompanyAreaList).AssertOK(t).AssertSuccess(t).JSONData(t)
	list := listData2["list"].([]any)
	if len(list) != 0 {
		t.Errorf("expected 0 areas after delete, got %d", len(list))
	}
}

// Test_P0_Shop_CompanyArea_Unassigned_HappyPath tests getting unassigned shops.
// Route: GET /shop/company_area/unassigned
func Test_P0_Shop_CompanyArea_Unassigned_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathCompanyAreaUnassigned)
	resp.AssertOK(t).AssertSuccess(t)

	data := resp.JSONData(t)
	if _, ok := data["list"]; !ok {
		t.Error("response missing list field")
	}
}

// Test_P0_Shop_CompanyArea_List_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/company_area/list
func Test_P0_Shop_CompanyArea_List_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathCompanyAreaList)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_CompanyArea_Create_Unauthorized tests that unauthenticated create is rejected.
// Route: POST /shop/company_area/create
func Test_P0_Shop_CompanyArea_Create_Unauthorized(t *testing.T) {
	body := map[string]any{
		"locale_name": map[string]string{"en": "Test"},
	}
	resp := fixture.NewHTTPClient().Post(t, pathCompanyAreaCreate, body)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_CompanyArea_Delete_Unauthorized tests that unauthenticated delete is rejected.
// Route: DELETE /shop/company_area/delete
func Test_P0_Shop_CompanyArea_Delete_Unauthorized(t *testing.T) {
	body := map[string]any{"uuid": 12345}
	resp := fixture.NewHTTPClient().Delete(t, pathCompanyAreaDelete, body)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_CompanyArea_ShopList_HappyPath tests getting shop list for area management.
// Route: GET /shop/company_area/shop_list
func Test_P0_Shop_CompanyArea_ShopList_HappyPath(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("site1", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathCompanyAreaShopList)
	resp.AssertOK(t).AssertSuccess(t)

	data := resp.JSONData(t)
	if _, ok := data["list"]; !ok {
		t.Error("response missing list field")
	}
}

// Test_P0_Shop_CompanyArea_ShopList_Unauthorized tests that unauthenticated shop_list is rejected.
// Route: GET /shop/company_area/shop_list
func Test_P0_Shop_CompanyArea_ShopList_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathCompanyAreaShopList)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_CompanyArea_Info_Unauthorized tests that unauthenticated info is rejected.
// Route: GET /shop/company_area/info
func Test_P0_Shop_CompanyArea_Info_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathCompanyAreaInfo+"?uuid=12345")
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_CompanyArea_Update_Unauthorized tests that unauthenticated update is rejected.
// Route: POST /shop/company_area/update
func Test_P0_Shop_CompanyArea_Update_Unauthorized(t *testing.T) {
	body := map[string]any{
		"uuid":        12345,
		"locale_name": map[string]string{"en": "Test"},
	}
	resp := fixture.NewHTTPClient().Post(t, pathCompanyAreaUpdate, body)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_CompanyArea_Unassigned_Unauthorized tests that unauthenticated unassigned is rejected.
// Route: GET /shop/company_area/unassigned
func Test_P0_Shop_CompanyArea_Unassigned_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathCompanyAreaUnassigned)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Helpers

func mustParseInt64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

func mustParseString(i int64) string {
	return fmt.Sprintf("%d", i)
}
