//go:build integration

package shift_test

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
	pathGetShiftInfo = "/api/v1/cashier/shift"
)

// --- P0 ---

// Test_P0_Cashier_Shift_GetInfo_HappyPath tests that an on-duty cashier can retrieve their shift info.
// Prerequisites: staff must have a non-empty duty_no, a matching active shift log, and a bound device.
// Route: GET /cashier/shift
func Test_P0_Cashier_Shift_GetInfo_HappyPath(t *testing.T) {
	const (
		dutyNo       = "DUTY-TEST-001"
		testDeviceId = "SHIFT-TEST-DEVICE-001"
	)

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
		fixture.WithStaffDutyNo(dutyNo),
	)

	// Shift log must exist and be active (status=0) for GetShiftInfo to succeed.
	fixture.SeedShiftLog(t, db,
		fixture.WithShiftLogStaffUUID(staff.UUID),
		fixture.WithShiftLogShiftNo(dutyNo),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), "test-cashier", 0, testDeviceId)
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathGetShiftInfo)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Cashier_Shift_GetInfo_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /cashier/shift
func Test_P0_Cashier_Shift_GetInfo_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathGetShiftInfo)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- P1 ---

// Test_P1_Cashier_Shift_GetInfo_AlreadyHandedOver tests that requesting shift info after handover returns an error.
// A shift log with status=1 (handed over) should be rejected by the service.
// Route: GET /cashier/shift
func Test_P1_Cashier_Shift_GetInfo_AlreadyHandedOver(t *testing.T) {
	const (
		dutyNo       = "DUTY-HANDEDOVER-001"
		testDeviceId = "SHIFT-HO-DEVICE-001"
	)

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
		fixture.WithStaffDutyNo(dutyNo),
	)

	// Seed a handed-over shift log (status=1) — service should reject this.
	fixture.SeedShiftLog(t, db,
		fixture.WithShiftLogStaffUUID(staff.UUID),
		fixture.WithShiftLogShiftNo(dutyNo),
		fixture.WithShiftLogStatus(1),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), "test-cashier", 0, testDeviceId)
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathGetShiftInfo)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for handed-over shift, got success")
	}
}

// Test_P1_Cashier_Shift_GetInfo_NoActiveShift tests that an on-duty cashier with no matching shift log gets an error.
// Staff has a duty_no so auth passes, but no shift log exists for that duty_no — service returns "当前班次不存在".
// Route: GET /cashier/shift
func Test_P1_Cashier_Shift_GetInfo_NoActiveShift(t *testing.T) {
	const (
		dutyNo       = "DUTY-NO-LOG-001"
		testDeviceId = "SHIFT-NL-DEVICE-001"
	)

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)
	// Staff has duty_no so auth passes, but no shift log is seeded for this duty_no.
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo(dutyNo),
	)

	fixture.SetupWireMock(t)

	token := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), "test-cashier", 0, testDeviceId)
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathGetShiftInfo)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for staff with no active shift log, got success")
	}
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
