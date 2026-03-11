//go:build integration

package remark_test

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
	pathGetOrderItemRemark    = "/api/v1/shop/setting/order_item_remark"
	pathAddOrderItemRemark    = "/api/v1/shop/setting/order_item_remark/add"
	pathEditOrderItemRemark   = "/api/v1/shop/setting/order_item_remark/edit"
	pathDeleteOrderItemRemark = "/api/v1/shop/setting/order_item_remark" // DELETE method
)

// remarkListData is used to parse the list response and extract UUIDs.
type remarkListData struct {
	List []struct {
		Uuid uint64 `json:"uuid"`
	} `json:"list"`
}

// --- P0 ---

// Test_P0_Shop_OrderItemRemark_List_HappyPath tests listing order item remarks for an authenticated shop user.
// Route: GET /shop/setting/order_item_remark
func Test_P0_Shop_OrderItemRemark_List_HappyPath(t *testing.T) {
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
	resp := fixture.NewHTTPClient().WithToken(token).Get(t, pathGetOrderItemRemark)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_OrderItemRemark_Add_HappyPath tests creating a new order item remark.
// Route: POST /shop/setting/order_item_remark/add
func Test_P0_Shop_OrderItemRemark_Add_HappyPath(t *testing.T) {
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
			"zh_name": "不要香菜",
			"en_name": "No Cilantro",
		},
	}
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathAddOrderItemRemark, body)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_OrderItemRemark_List_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/setting/order_item_remark
func Test_P0_Shop_OrderItemRemark_List_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathGetOrderItemRemark)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- P1 ---

// Test_P0_Shop_OrderItemRemark_Delete_HappyPath tests the full Add → List → Delete → verify-empty lifecycle.
// Add returns no UUID, so we retrieve it from GET /list before deleting.
// Route: DELETE /shop/setting/order_item_remark
func Test_P0_Shop_OrderItemRemark_Delete_HappyPath(t *testing.T) {
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

	// 1. Add a remark
	httpClient.Post(t, pathAddOrderItemRemark, map[string]any{
		"locale_name": map[string]string{"zh_name": "不要葱", "en_name": "No Scallion"},
	}).AssertOK(t).AssertSuccess(t)

	// 2. GET list to obtain the generated UUID (Add returns no UUID in body)
	listResp := httpClient.Get(t, pathGetOrderItemRemark)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)

	var listData remarkListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if len(listData.List) == 0 {
		t.Fatal("expected at least one remark after add")
	}
	uuid := listData.List[0].Uuid

	// 3. Delete by UUID
	httpClient.Delete(t, pathDeleteOrderItemRemark, map[string]any{"uuid": uuid}).
		AssertOK(t).AssertSuccess(t)

	// 4. Verify the list is now empty
	verifyResp := httpClient.Get(t, pathGetOrderItemRemark)
	verifyResp.AssertOK(t).AssertSuccess(t)
	verifyApiResp := verifyResp.ParseAPIResponse(t)

	var verifyData remarkListData
	if err := json.Unmarshal(verifyApiResp.Data, &verifyData); err != nil {
		t.Fatalf("failed to parse verify list response: %v", err)
	}
	if len(verifyData.List) != 0 {
		t.Errorf("expected empty list after delete, got %d items", len(verifyData.List))
	}
}

// Test_P0_Shop_OrderItemRemark_Edit_HappyPath tests that an existing remark's name can be updated.
// Route: POST /shop/setting/order_item_remark/edit
func Test_P0_Shop_OrderItemRemark_Edit_HappyPath(t *testing.T) {
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

	// 1. Add a remark
	httpClient.Post(t, pathAddOrderItemRemark, map[string]any{
		"locale_name": map[string]string{"zh_name": "不要香菜", "en_name": "No Cilantro"},
	}).AssertOK(t).AssertSuccess(t)

	// 2. GET list to retrieve UUID
	listResp := httpClient.Get(t, pathGetOrderItemRemark)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)

	var listData remarkListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if len(listData.List) == 0 {
		t.Fatal("expected at least one remark after add")
	}
	uuid := listData.List[0].Uuid

	// 3. Edit with new name
	httpClient.Post(t, pathEditOrderItemRemark, map[string]any{
		"uuid":        uuid,
		"locale_name": map[string]string{"zh_name": "不要葱", "en_name": "No Scallion"},
	}).AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_OrderItemRemark_Edit_NotFound tests editing a remark with a non-existent UUID.
// Route: POST /shop/setting/order_item_remark/edit
func Test_P1_Shop_OrderItemRemark_Edit_NotFound(t *testing.T) {
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
		"uuid": 999999999,
		"locale_name": map[string]string{
			"en_name": "No Onion",
		},
	}
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathEditOrderItemRemark, body)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error when editing non-existent remark, got success")
	}
}

// Test_P1_Shop_OrderItemRemark_Add_MaxLimit verifies that adding more than 100 remarks per company
// returns an error. Seeds 100 records directly into the DB, then tries to add a 101st via API.
// Route: POST /shop/setting/order_item_remark/add
func Test_P1_Shop_OrderItemRemark_Add_MaxLimit(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	// Pre-seed 100 remark records to reach the limit
	for i := 0; i < 100; i++ {
		fixture.SeedOrderItemRemark(t, db,
			fixture.WithOrderItemRemarkName(fmt.Sprintf(`{"en_name":"Remark%d"}`, i)),
		)
	}

	fixture.SetupWireMock(t)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	body := map[string]any{
		"locale_name": map[string]string{
			"en_name": "Overflow Remark",
		},
	}
	resp := fixture.NewHTTPClient().WithToken(token).Post(t, pathAddOrderItemRemark, body)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error when exceeding 100-item remark limit, got success")
	}
}

// Test_P1_Shop_OrderItemRemark_Delete_Idempotent verifies that deleting a non-existent UUID
// succeeds silently (idempotent soft-delete: UPDATE affects 0 rows but returns no error).
// Route: DELETE /shop/setting/order_item_remark
func Test_P1_Shop_OrderItemRemark_Delete_Idempotent(t *testing.T) {
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
	// Deleting a non-existent UUID is idempotent: the service performs a soft-delete (UPDATE
	// delete_time) which affects 0 rows but returns no error.
	resp := fixture.NewHTTPClient().WithToken(token).Delete(t, pathDeleteOrderItemRemark, map[string]any{
		"uuid": uint64(999999999999),
	})
	resp.AssertOK(t).AssertSuccess(t)
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
