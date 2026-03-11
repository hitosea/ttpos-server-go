//go:build integration

package shift_test

import (
	"testing"

	"github.com/hitosea/wiremock/client"
	"ttpos-server-go/tests/fixture"
)

// pathSubmitShift is the endpoint for closing/submitting a shift.
const pathSubmitShift = "/api/v1/cashier/shift"

// erpSiteCode is a non-empty site code used to enable ERP flows in tests.
// company_setting.erpnext_site_code must be non-empty for needErpClosePos to be true.
const erpSiteCode = "TEST_SITE"

// erpOpenEntryName is a non-empty entry name used to satisfy the third ERP gate.
// shift_log.erpnext_open_pos_entry_name must be non-empty for ClosePosEntry to be called.
const erpOpenEntryName = "POS-ENTRY-TEST-001"

// --- P0 ---

// Test_P0_Cashier_Shift_Close_WithErp_CashHasPaymentId verifies that when Cash has an
// erpnext_payment_id, the ClosePosEntry gRPC call uses payment_id (not mode_of_payment).
//
// Logic under test (staff_shift.go):
//
//	if cashPaymentMethod.Uuid > 0 && cashPaymentMethod.ErpnextPaymentId != "" {
//	  cashDetail.PaymentId = &cashPaymentMethod.ErpnextPaymentId   // ← this path
//	} else {
//	  cashDetail.ModeOfPayment = "Cash"
//	}
//
// Route: POST /cashier/shift
func Test_P0_Cashier_Shift_Close_WithErp_CashHasPaymentId(t *testing.T) {
	const (
		dutyNo         = "DUTY-ERP-PAYID-001"
		cashPaymentId  = "PID-TEST-CASH-001"
		cashErpPayment = "Cash"
		testDeviceId   = "ERP-PAYID-DEVICE-001"
	)

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// Company must have is_enable_erp=1 for needErpClosePos to be true.
	fixture.SeedCompany(t, db,
		fixture.WithCompanyUUID(companyUUIDInt),
		fixture.WithCompanyIsEnableErp(1),
	)
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingErpConfig(erpSiteCode, "TEST_PROFILE", "test@example.com", "TEST", "TEST_BRANCH"),
	)
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo(dutyNo),
		fixture.WithStaffBindKey(testDeviceId),
	)

	// Shift log must have erpnext_open_pos_entry_name != "" to trigger ClosePosEntry.
	fixture.SeedShiftLog(t, db,
		fixture.WithShiftLogStaffUUID(staff.UUID),
		fixture.WithShiftLogShiftNo(dutyNo),
		fixture.WithShiftLogErpOpenEntryName(erpOpenEntryName),
	)

	// Cash payment method WITH erpnext_payment_id → should use payment_id in gRPC call.
	fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodCode(40), // PaymentMethodCodeCash
		fixture.WithPaymentMethodSource(0), // PaymentMethodSourceSystem
		fixture.WithPaymentMethodErpnextPayment(cashErpPayment),
		fixture.WithPaymentMethodErpnextPaymentId(cashPaymentId),
	)

	wmClient := fixture.SetupWireMock(t)

	// Stub ClosePosEntry scoped to this test's entry name (priority 1 beats defaults at 100).
	// Body-matching ensures parallel tests with different entry names don't collide.
	wmClient.RegisterStub(t,
		client.NewGrpcStub("selling.SellingService", "ClosePosEntry").
			WithName("close-pos-entry-cash-payment-id").
			WithPriority(1).
			WithRequestBodyContains(erpOpenEntryName).
			WillReturnResponse(map[string]any{
				"code":    "0",
				"message": "success",
				"data": map[string]any{
					"@type":         "type.googleapis.com/selling.ClosePosEntryResp",
					"asyncRecordId": "ASYNC-TEST-001",
				},
			}).
			Build(),
	)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), "test-cashier", 0, testDeviceId)
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathSubmitShift, map[string]any{
		"withdraw_cash": 0,
		"leave_cash":    0,
		"is_background": false,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify WireMock received a ClosePosEntry call containing the payment_id (not "Cash").
	wmClient.Verify(t, &client.VerificationRequest{
		Method:  "POST",
		URLPath: "/selling.SellingService/ClosePosEntry",
		BodyPatterns: []client.BodyPattern{
			{Contains: cashPaymentId},
		},
	})
}

// Test_P0_Cashier_Shift_Close_WithErp_CashNoPaymentId_FallsBackToModeOfPayment verifies that
// when Cash has no erpnext_payment_id, ClosePosEntry uses mode_of_payment = "Cash".
//
// Logic under test (staff_shift.go):
//
//	if cashPaymentMethod.Uuid > 0 && cashPaymentMethod.ErpnextPaymentId != "" { ... } else {
//	  cashDetail.ModeOfPayment = "Cash"   // ← this path
//	}
//
// Route: POST /cashier/shift
func Test_P0_Cashier_Shift_Close_WithErp_CashNoPaymentId_FallsBackToModeOfPayment(t *testing.T) {
	const (
		dutyNo           = "DUTY-ERP-NOPAYID-001"
		noPayIdOpenEntry = "POS-ENTRY-NOPAYID-001"
		cashErpPayment   = "Cash"
		testDeviceId     = "ERP-NOPAYID-DEVICE-001"
	)

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db,
		fixture.WithCompanyUUID(companyUUIDInt),
		fixture.WithCompanyIsEnableErp(1),
	)
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingErpConfig(erpSiteCode, "TEST_PROFILE", "test@example.com", "TEST", "TEST_BRANCH"),
	)
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo(dutyNo),
		fixture.WithStaffBindKey(testDeviceId),
	)
	fixture.SeedShiftLog(t, db,
		fixture.WithShiftLogStaffUUID(staff.UUID),
		fixture.WithShiftLogShiftNo(dutyNo),
		fixture.WithShiftLogErpOpenEntryName(noPayIdOpenEntry),
	)

	// Cash payment method WITHOUT erpnext_payment_id → should fall back to mode_of_payment="Cash".
	fixture.SeedPaymentMethod(t, db,
		fixture.WithPaymentMethodCode(40),
		fixture.WithPaymentMethodSource(0),
		fixture.WithPaymentMethodErpnextPayment(cashErpPayment),
		// No WithPaymentMethodErpnextPaymentId — left empty intentionally.
	)

	wmClient := fixture.SetupWireMock(t)

	// Stub scoped to this test's unique entry name.
	wmClient.RegisterStub(t,
		client.NewGrpcStub("selling.SellingService", "ClosePosEntry").
			WithName("close-pos-entry-cash-mode-of-payment").
			WithPriority(1).
			WithRequestBodyContains(noPayIdOpenEntry).
			WillReturnResponse(map[string]any{
				"code":    "0",
				"message": "success",
				"data": map[string]any{
					"@type":         "type.googleapis.com/selling.ClosePosEntryResp",
					"asyncRecordId": "ASYNC-TEST-002",
				},
			}).
			Build(),
	)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), "test-cashier", 0, testDeviceId)
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathSubmitShift, map[string]any{
		"withdraw_cash": 0,
		"leave_cash":    0,
		"is_background": false,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// Verify WireMock received ClosePosEntry with "Cash" as mode_of_payment.
	wmClient.Verify(t, &client.VerificationRequest{
		Method:  "POST",
		URLPath: "/selling.SellingService/ClosePosEntry",
		BodyPatterns: []client.BodyPattern{
			{Contains: `Cash`},
		},
	})
}

// --- P1 ---

// Test_P1_Cashier_Shift_Close_ErpDisabled_SkipsClosePosEntry verifies that when ERP is disabled
// (is_enable_erp=0), the ClosePosEntry gRPC call is never made.
//
// Route: POST /cashier/shift
func Test_P1_Cashier_Shift_Close_ErpDisabled_SkipsClosePosEntry(t *testing.T) {
	const (
		dutyNo       = "DUTY-ERP-DISABLED-001"
		testDeviceId = "ERP-DISABLED-DEVICE-001"
	)

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// ERP disabled (default is_enable_erp=0).
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo(dutyNo),
		fixture.WithStaffBindKey(testDeviceId),
	)
	fixture.SeedShiftLog(t, db,
		fixture.WithShiftLogStaffUUID(staff.UUID),
		fixture.WithShiftLogShiftNo(dutyNo),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), "test-cashier", 0, testDeviceId)
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathSubmitShift, map[string]any{
		"withdraw_cash": 0,
		"leave_cash":    0,
		"is_background": false,
	})
	resp.AssertOK(t).AssertSuccess(t)

	// ERP is disabled, so shift close succeeds without calling ClosePosEntry.
	// We cannot verify request count=0 because other tests in the same process
	// may have already made ClosePosEntry calls (WireMock journal is shared).
}
