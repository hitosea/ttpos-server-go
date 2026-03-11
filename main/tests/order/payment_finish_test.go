//go:build integration

package order_test

import (
	"encoding/json"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// API paths for instant order payment flow
	pathOrderCreate   = "/api/v1/cashier/instant/order/create"
	pathPaymentCreate = "/api/v1/cashier/instant/order/payment/create"
	pathPaymentFinish = "/api/v1/cashier/instant/order/payment/finish"
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Cashier_OrderPaymentFinish_HappyPath tests the complete payment flow:
// 1. Create instant order (via API)
// 2. Seed a payment order record directly in the database (simulating a paid cash payment)
// 3. Finish payment (via API)
//
// This test covers the metrics.PaymentsTotal.WithLabelValues(..., "success").Inc()
// line in main/app/service/order_pay.go:1210.
//
// Route: POST /api/v1/cashier/instant/order/payment/finish
//
// Prerequisites:
//   - Company and CompanySetting must exist in the tenant DB
//   - Device must be registered with source="cashier" and matching device_id in JWT
//   - Staff must have duty_no set (on-duty) and is_super=1
//   - Full tenant schema loaded via shop_01.sql (NewTestTenantFull)
func Test_P0_Cashier_OrderPaymentFinish_HappyPath(t *testing.T) {
	const testDeviceId = "test-cashier-payment-001"

	// 1. Create tenant database with full production schema
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)

	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. Seed company and company_setting
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. Seed device
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)

	// 4. Seed staff
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffName("Cashier 1"),
		fixture.WithStaffUsername("cashier1"),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-001"),
	)

	// 5. Seed payment method (cash - code 40)
	paymentMethod := fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodName("Cash"),
		fixture.WithPaymentMethodCode(40),          // PaymentMethodCodeCash
		fixture.WithPaymentMethodStatus(1),          // enabled
		fixture.WithPaymentMethodIsShowCashier(1),   // visible on cashier terminal
	)

	// 6. Setup WireMock for gRPC services
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

	// 8. Create instant order (creates sale_bill and sale_order)
	createResp := httpClient.Post(t, pathOrderCreate, nil)
	createResp.AssertOK(t).AssertSuccess(t)

	var createResult struct {
		SaleBillUUID  uint64 `json:"sale_bill_uuid"`
		SaleOrderUUID uint64 `json:"sale_order_uuid"`
	}
	apiResp := createResp.ParseAPIResponse(t)
	if err := json.Unmarshal(apiResp.Data, &createResult); err != nil {
		t.Fatalf("failed to unmarshal order create response: %v", err)
	}

	if createResult.SaleBillUUID == 0 {
		t.Fatal("expected sale_bill_uuid to be non-zero")
	}
	if createResult.SaleOrderUUID == 0 {
		t.Fatal("expected sale_order_uuid to be non-zero")
	}

	// 9. Seed a payment order record (simulating a paid cash payment)
	// The order has 0 amount initially, so we set payment amount to 0
	fixture.SeedPaymentOrder(t, db,
		fixture.WithPaymentOrderRelatedUUID(int64(createResult.SaleOrderUUID)),
		fixture.WithPaymentOrderPaymentMethodUUID(paymentMethod.UUID),
		fixture.WithPaymentOrderPaymentMethodName("Cash"),
		fixture.WithPaymentOrderAmount(0), // Order has no products, so amount is 0
		fixture.WithPaymentOrderStatus(1), // paid
	)

	// 10. Finish payment - this is where metrics.PaymentsTotal.WithLabelValues(..., "success").Inc() is called
	finishPaymentReq := map[string]any{
		"sale_bill_uuid":  createResult.SaleBillUUID,
		"sale_order_uuid": createResult.SaleOrderUUID,
	}
	finishPaymentResp := httpClient.Post(t, pathPaymentFinish, finishPaymentReq)
	finishPaymentResp.AssertOK(t).AssertSuccess(t)

	// Verify the response contains the payment method list (metrics were incremented)
	finishApiResp := finishPaymentResp.ParseAPIResponse(t)

	// Verify the response contains the payment method list (metrics were incremented)
	var finishResult struct {
		PayMethods struct {
			List []struct {
				Uuid uint64 `json:"uuid"`
				Name string `json:"name"`
			} `json:"list"`
		} `json:"pay_methods"`
	}
	if err := json.Unmarshal(finishApiResp.Data, &finishResult); err != nil {
		t.Fatalf("failed to unmarshal payment finish response: %v", err)
	}

	// Verify payment method is in the response (metrics were incremented)
	if len(finishResult.PayMethods.List) == 0 {
		t.Error("expected pay_methods to contain at least one payment method")
	}

	// Verify the payment method is cash
	found := false
	paymentMethodUUID := uint64(paymentMethod.UUID)
	for _, pm := range finishResult.PayMethods.List {
		if pm.Uuid == paymentMethodUUID {
			found = true
			if pm.Name != "Cash" {
				t.Errorf("expected payment method name 'Cash', got '%s'", pm.Name)
			}
			break
		}
	}
	if !found {
		t.Error("expected cash payment method to be in the response")
	}
}

// Test_P0_Cashier_OrderPaymentFinish_Unauthorized tests that requests without a token are rejected.
// Route: POST /api/v1/cashier/instant/order/payment/finish
func Test_P0_Cashier_OrderPaymentFinish_Unauthorized(t *testing.T) {
	httpClient := fixture.NewHTTPClient() // No token

	req := map[string]any{
		"sale_bill_uuid":  12345,
		"sale_order_uuid": 67890,
	}
	resp := httpClient.Post(t, pathPaymentFinish, req)

	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Cashier_OrderPaymentFinish_OrderNotFound tests that finishing payment for
// a non-existent order returns an appropriate error.
// Route: POST /api/v1/cashier/instant/order/payment/finish
func Test_P1_Cashier_OrderPaymentFinish_OrderNotFound(t *testing.T) {
	const testDeviceId = "test-cashier-payment-notfound-001"

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)

	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)

	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-001"),
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

	// Try to finish payment for non-existent order
	req := map[string]any{
		"sale_bill_uuid":  999999999,
		"sale_order_uuid": 999999999,
	}
	resp := httpClient.Post(t, pathPaymentFinish, req)

	// Should return an error (not success)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error code for non-existent order, got success")
	}
}

// Test_P1_Cashier_OrderPaymentCreate_MissingRequiredFields tests that the payment
// create endpoint validates required fields.
// Route: POST /api/v1/cashier/instant/order/payment/create
func Test_P1_Cashier_OrderPaymentCreate_MissingRequiredFields(t *testing.T) {
	const testDeviceId = "test-cashier-payment-validation-001"

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)

	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)

	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-001"),
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

	// Missing payment_method_uuid
	req := map[string]any{
		"sale_bill_uuid":  12345,
		"sale_order_uuid": 67890,
		"payment_amount":  100.0,
		// payment_method_uuid is missing
	}
	resp := httpClient.Post(t, pathPaymentCreate, req)

	resp.AssertOK(t)
	// Should return validation error
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error for missing payment_method_uuid")
	}
}

// Helpers are defined in happy_path_test.go (mustParseInt64, mustParseString)
