//go:build integration

package statistics_test

import (
	"fmt"
	"os"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// codeTokenInvalid matches constant.CodeTokenInvalid in the server (-102).
const codeTokenInvalid = -102

// codeAccessDenied matches constant.CodeAccessDenied in the server (-103).
// Returned by MinVersionCheck middleware when client version is too low.
const codeAccessDenied = -103

// API path constants
const (
	pathStatsBusiness      = "/api/v1/shop/statistics/business"
	pathStatsPaymentMethod = "/api/v1/shop/statistics/payment_method"
	pathStatsProductRank   = "/api/v1/shop/statistics/product_rank"
	pathStatsTimePeriod    = "/api/v1/shop/statistics/business/time_period"
	pathSetDataManage      = "/api/v1/shop/setting/data_manage/set"
)

// versionHeaders provides a Client-Version header high enough to pass any
// MinVersionCheck middleware. Tests that need to bypass version gating should
// attach these headers to requests.
var versionHeaders = map[string]string{"Client-Version": "99.0.0"}

// --- P0 ---

// Test_P0_Shop_Statistics_Business_HappyPath tests fetching business statistics.
// Route: GET /shop/statistics/business
func Test_P0_Shop_Statistics_Business_HappyPath(t *testing.T) {
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
	// Query with a date range for today
	resp := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders).Get(t, pathStatsBusiness+"?start_time=1700000000&end_time=1900000000")
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_Statistics_PaymentMethod_HappyPath tests fetching payment method statistics.
// Route: GET /shop/statistics/payment_method
func Test_P0_Shop_Statistics_PaymentMethod_HappyPath(t *testing.T) {
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
	resp := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders).Get(t, pathStatsPaymentMethod+"?start_time=1700000000&end_time=1900000000")
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_Statistics_Business_Unauthorized tests that unauthenticated requests are rejected.
// The version check middleware runs before auth, so depending on the server version
// configuration, we may get -103 (version too low) or -102 (token invalid).
// Route: GET /shop/statistics/business
func Test_P0_Shop_Statistics_Business_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathStatsBusiness)
	apiResp := resp.AssertOK(t).ParseAPIResponse(t)
	if apiResp.Code != codeTokenInvalid && apiResp.Code != codeAccessDenied {
		t.Fatalf("expected error code %d or %d but got %d: %s",
			codeTokenInvalid, codeAccessDenied, apiResp.Code, apiResp.Message)
	}
}

// --- P1 ---

// Test_P1_Shop_Statistics_ProductRank_HappyPath tests fetching product rank statistics.
// Route: GET /shop/statistics/product_rank
func Test_P1_Shop_Statistics_ProductRank_HappyPath(t *testing.T) {
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
	resp := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders).Get(t, pathStatsProductRank+"?start_time=1700000000&end_time=1900000000")
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_Statistics_Business_WithDataManagement tests that enabling the data management
// exclusion feature (is_enable_data_manage) does not break the statistics endpoint.
//
// The ExcludeDataManage flag is derived from: companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage
// This test verifies the full feature toggle path works end-to-end without order data.
// Route: POST /shop/setting/data_manage/set  →  GET /shop/statistics/business
func Test_P1_Shop_Statistics_Business_WithDataManagement(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		// enable_data_management=1 is required for SetDataManage permission check to pass.
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders)

	// Enable data management exclusion feature
	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{},
	}).AssertOK(t).AssertSuccess(t)

	// Statistics endpoint must still respond successfully with the feature enabled
	httpClient.Get(t, pathStatsBusiness+"?start_time=1700000000&end_time=1900000000").
		AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_Statistics_Business_ExcludesDataManaged verifies that when a completed sale bill
// is added to the data_manage exclusion list, the business statistics endpoint still responds
// successfully (the NOT IN subquery over ttpos_data_manage executes without error on real data).
// Route: GET /shop/statistics/business
func Test_P1_Shop_Statistics_Business_ExcludesDataManaged(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	// Seed a completed bill within the query time range so it participates in exclusion filtering.
	bill := fixture.SeedSaleBill(t, db)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders)

	// Mark the seeded bill as excluded via the data_manage API.
	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{uint64(bill.UUID)},
	}).AssertOK(t).AssertSuccess(t)

	// Business statistics must succeed with real data_manage exclusion records in the DB.
	httpClient.Get(t, pathStatsBusiness+"?start_time=1700000000&end_time=1900000000").
		AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_Statistics_PaymentMethod_ExcludesDataManaged verifies that when a completed sale
// bill is added to the data_manage exclusion list, the payment method statistics endpoint still
// responds successfully.
// Route: GET /shop/statistics/payment_method
func Test_P1_Shop_Statistics_PaymentMethod_ExcludesDataManaged(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	bill := fixture.SeedSaleBill(t, db)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders)

	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{uint64(bill.UUID)},
	}).AssertOK(t).AssertSuccess(t)

	httpClient.Get(t, pathStatsPaymentMethod+"?start_time=1700000000&end_time=1900000000").
		AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_Statistics_TimePeriod_ExcludesDataManaged verifies that when a completed sale
// bill is added to the data_manage exclusion list, the business time-period statistics endpoint
// still responds successfully.
// Route: GET /shop/statistics/business/time_period
func Test_P1_Shop_Statistics_TimePeriod_ExcludesDataManaged(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	bill := fixture.SeedSaleBill(t, db)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders)

	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{uint64(bill.UUID)},
	}).AssertOK(t).AssertSuccess(t)

	// time_period=3 (hourly), statistics_type=1 (checkout time)
	httpClient.Get(t, pathStatsTimePeriod+"?start_time=1700000000&end_time=1900000000&time_period=3&statistics_type=1").
		AssertOK(t).AssertSuccess(t)
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
