//go:build integration

package payment_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// codeTokenInvalid matches constant.CodeTokenInvalid in the server (-102).
const codeTokenInvalid = -102

// API path constants
const (
	pathPaymentMethodList   = "/api/v1/shop/payment_method/list"
	pathPaymentMethodCreate = "/api/v1/shop/payment_method/create"
)

// --- P0 ---

// Test_P0_Shop_PaymentMethod_List_HappyPath tests listing payment methods for an authenticated shop user.
// Route: GET /shop/payment_method/list
func Test_P0_Shop_PaymentMethod_List_HappyPath(t *testing.T) {
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
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathPaymentMethodList+"?page_no=1&page_size=10")
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_PaymentMethod_List_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/payment_method/list
func Test_P0_Shop_PaymentMethod_List_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathPaymentMethodList+"?page_no=1&page_size=10")
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- P1 ---

// Test_P1_Shop_PaymentMethod_List_MissingPagination tests that missing pagination params return an error.
// Route: GET /shop/payment_method/list
func Test_P1_Shop_PaymentMethod_List_MissingPagination(t *testing.T) {
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
	// Missing required page_no and page_size query params
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathPaymentMethodList)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error when pagination params are missing, got success")
	}
}

// paymentMethodListData parses the management list response for filtering assertions.
type paymentMethodListData struct {
	List []struct {
		Uuid uint64 `json:"uuid"`
		Name string `json:"name"`
	} `json:"list"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

// containsUUID returns true if any item in the list has the given UUID.
func containsUUID(list []struct {
	Uuid uint64 `json:"uuid"`
	Name string `json:"name"`
}, uuid uint64) bool {
	for _, item := range list {
		if item.Uuid == uuid {
			return true
		}
	}
	return false
}

// Test_P1_Shop_PaymentMethod_ManagementList_FiltersFreeMeal verifies that payment methods
// with code=-1 (Free Meal / 免单) are excluded from the management list.
// Route: GET /shop/payment_method/list
func Test_P1_Shop_PaymentMethod_ManagementList_FiltersFreeMeal(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	// Seed a Free Meal payment method (code=-1) — should be excluded
	freeMeal := fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodUUID(100001),
		fixture.WithPaymentMethodName("Test FreeMeal"),
		fixture.WithPaymentMethodCode(-1),
	)
	// Seed a Cash payment method (code=40) — should be included
	cash := fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodUUID(100002),
		fixture.WithPaymentMethodName("Test Cash"),
		fixture.WithPaymentMethodCode(40),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathPaymentMethodList+"?page_no=1&page_size=100")
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var listData paymentMethodListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if containsUUID(listData.List, uint64(freeMeal.UUID)) {
		t.Error("expected Free Meal (code=-1) to be excluded from management list, but it was returned")
	}
	if !containsUUID(listData.List, uint64(cash.UUID)) {
		t.Error("expected Cash (code=40) to be included in management list, but it was not returned")
	}
}

// Test_P1_Shop_PaymentMethod_ManagementList_FiltersGrabAndLineMan verifies that Grab (code=91100)
// and LINE MAN (code=91200) payment methods are excluded from the management list.
// Route: GET /shop/payment_method/list
func Test_P1_Shop_PaymentMethod_ManagementList_FiltersGrabAndLineMan(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	// Grab (code=91100) — should be excluded
	grab := fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodUUID(200001),
		fixture.WithPaymentMethodName("Test Grab"),
		fixture.WithPaymentMethodCode(91100),
	)
	// LINE MAN (code=91200) — should be excluded
	lineMan := fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodUUID(200002),
		fixture.WithPaymentMethodName("Test LineMan"),
		fixture.WithPaymentMethodCode(91200),
	)
	// Cash (code=40) — should be included
	cash := fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodUUID(200003),
		fixture.WithPaymentMethodName("Test Cash"),
		fixture.WithPaymentMethodCode(40),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathPaymentMethodList+"?page_no=1&page_size=100")
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var listData paymentMethodListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if containsUUID(listData.List, uint64(grab.UUID)) {
		t.Error("expected Grab (code=91100) to be excluded from management list, but it was returned")
	}
	if containsUUID(listData.List, uint64(lineMan.UUID)) {
		t.Error("expected LINE MAN (code=91200) to be excluded from management list, but it was returned")
	}
	if !containsUUID(listData.List, uint64(cash.UUID)) {
		t.Error("expected Cash (code=40) to be included in management list, but it was not returned")
	}
}

// Test_P1_Shop_PaymentMethod_Create_NewMethod tests creating a new custom payment method.
// Route: POST /shop/payment_method/create
func Test_P1_Shop_PaymentMethod_Create_NewMethod(t *testing.T) {
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
		"items": []map[string]any{
			{
				"name":            "Bank Transfer",
				"payment_name":    "bank_transfer",
				"source":          1, // PaymentMethodSourceDefault (custom)
				"default_img":     "cash.png", // required: logo_file_uuid or default_img must be non-empty
				"fee_percent":     0,
				"is_show_cashier": 1,
				"status":          1,
			},
		},
	}
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathPaymentMethodCreate, body)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_PaymentMethod_Create_Idempotent tests that creating the same payment method
// twice does not fail (idempotent create).
// Route: POST /shop/payment_method/create
func Test_P1_Shop_PaymentMethod_Create_Idempotent(t *testing.T) {
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
	httpClient := fixture.NewHTTPClient().WithToken(token)

	body := map[string]any{
		"items": []map[string]any{
			{"name": "E-Wallet", "payment_name": "ewallet", "source": 1, "default_img": "ewallet.png", "status": 1},
		},
	}

	// First create
	httpClient.Post(t, pathPaymentMethodCreate, body).AssertOK(t).AssertSuccess(t)
	// Second create with same data — must not fail
	httpClient.Post(t, pathPaymentMethodCreate, body).AssertOK(t).AssertSuccess(t)
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
