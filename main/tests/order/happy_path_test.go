//go:build integration

package order_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hitosea/wiremock/client"
	"ttpos-server-go/tests/fixture"
)

// codeTokenInvalid matches constant.CodeTokenInvalid in the server (-102).
// The server always responds with HTTP 200; authentication errors are expressed
// as a JSON code in the response body.
const codeTokenInvalid = -102

// codeAccessDenied matches constant.CodeAccessDenied in the server (-103).
// This is returned when JWT is valid but company / staff auth fails.
const codeAccessDenied = -103

// TestCreateOrder_HappyPath tests the complete instant order creation workflow.
// Route: POST /api/v1/cashier/instant/order/create
//
// Auth requirements for this test:
//   - Company and CompanySetting must exist in the tenant DB (preloaded with staff)
//   - Device must be registered with source="cashier" and matching device_id in JWT
//   - Staff must have duty_no set (on-duty) and is_super=1 (bypasses role/permission tables)
//   - Full tenant schema loaded via shop_01.sql (NewTestTenantFull)
func TestCreateOrder_HappyPath(t *testing.T) {
	const testDeviceId = "test-cashier-device-001"

	// 1. Create tenant database with full production schema (shop_01.sql)
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)

	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. Seed company and company_setting (required for WithCompany() preload in auth)
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. Seed device (required for IsDeviceBind check in auth).
	// IsDeviceBind queries only by source + device_id, so company_uuid doesn't matter.
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)

	// 4. Seed staff: is_super=1 bypasses role/permission tables; duty_no set means on-duty
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffName("Cashier 1"),
		fixture.WithStaffUsername("cashier1"),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-001"),
	)

	// 5. Setup WireMock: register default catch-all stubs for all gRPC services
	wiremockURL := getEnv("WIREMOCK_URL", "http://wiremock:8080")
	wmClient := client.NewClient(wiremockURL)
	wmClient.Reset(t)
	client.RegisterTTPOSDefaults(wmClient, t)

	// 6. Generate JWT with device_id matching the seeded device
	token := fixture.GenerateStaffToken(t,
		companyUUID,
		mustParseString(staff.UUID),
		staff.Username,
		1,
		testDeviceId,
	)

	// 7. POST to the instant-order-create endpoint
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/cashier/instant/order/create", nil)

	// 8. Assert a successful response
	resp.AssertOK(t).AssertSuccess(t)

	// 9. Parse the response and verify the created order identifiers
	var result struct {
		SaleBillUUID  uint64 `json:"sale_bill_uuid"`
		SaleOrderUUID uint64 `json:"sale_order_uuid"`
	}
	apiResp := resp.ParseAPIResponse(t)
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}

	if result.SaleBillUUID == 0 {
		t.Error("expected sale_bill_uuid to be non-zero")
	}
	if result.SaleOrderUUID == 0 {
		t.Error("expected sale_order_uuid to be non-zero")
	}

	// Cleanup is automatic via t.Cleanup() in fixture functions
}

// TestGetOrder_NotFound tests that a valid JWT for a company without seeded company/setting
// data is rejected by the auth layer with codeAccessDenied.
// Route: GET /api/v1/cashier/instant/order/cart/info
//
// Auth fails because WithCompany() preload returns nil (no ttpos_company row for this UUID),
// so authSrv.Auth returns an error which the middleware wraps as codeAccessDenied (-103).
func TestGetOrder_NotFound(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenant(t, companyUUID)

	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(mustParseInt64(companyUUID)),
	)

	wmClient := client.NewClient(getEnv("WIREMOCK_URL", "http://wiremock:8080"))
	wmClient.Reset(t)
	client.RegisterTTPOSDefaults(wmClient, t)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), staff.Username, 1)
	httpClient := fixture.NewHTTPClient().WithToken(token)

	// Request cart info for a device that does not exist
	resp := httpClient.Get(t, "/api/v1/cashier/instant/order/cart/info?device_sn=non-existent-device")

	// The server always responds HTTP 200; auth failure returns codeAccessDenied
	resp.AssertOK(t).AssertErrorCode(t, codeAccessDenied)
}

// TestCreateOrder_Unauthorized tests that requests without a token are rejected.
// Route: POST /api/v1/cashier/instant/order/create
//
// The server responds HTTP 200 with code=-102 (token invalid) rather than HTTP 401.
func TestCreateOrder_Unauthorized(t *testing.T) {
	httpClient := fixture.NewHTTPClient() // No token

	resp := httpClient.Post(t, "/api/v1/cashier/instant/order/create", nil)

	// Server always returns HTTP 200; missing token produces codeTokenInvalid
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Helper functions

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
