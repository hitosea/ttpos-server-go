//go:build integration

package member_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// Test_OrderFirstPayLater_FullFlow verifies the complete "order-first-pay-later" business flow:
//
//  1. Member creates dine-in order (no payment required)
//  2. H5 order is generated → cashier sees it in accept-list
//  3. Cashier accepts → order becomes instant-order hold
//  4. Member sees "preparing" status in order detail
//
// This test requires SERVICE_URL to be set (default: http://localhost:8080).
func Test_OrderFirstPayLater_FullFlow(t *testing.T) {
	const testDeviceId = "test-order-first-device-001"

	// ── 1. Setup tenant ──────────────────────────────────────────────
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// Member
	member := fixture.SeedMember(t, db)

	// Staff + device for cashier auth
	fixture.SeedDevice(t, db,
		fixture.WithDeviceType("cashier"),
		fixture.WithDeviceId(testDeviceId),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffName("Cashier Test"),
		fixture.WithStaffUsername("cashier_ofpl"),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-ofpl"),
	)

	// Business settings: all-day open
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// Key config: enable order-first-pay-later mode
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_order_first_pay_later":1}`)

	// Product
	bomUUID := fixture.SeedProductWithFlavor(t, db, "测试鸡排", 25.00)

	// External gRPC stubs
	fixture.SetupWireMock(t)

	// ── 2. Generate tokens ───────────────────────────────────────────
	memberToken := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))
	cashierToken := fixture.GenerateStaffToken(t,
		companyUUID,
		mustParseString(staff.UUID),
		staff.Username,
		1,
		testDeviceId,
	)

	memberClient := fixture.NewHTTPClient().WithToken(memberToken)
	cashierClient := fixture.NewHTTPClient().WithToken(cashierToken)

	// ── 3. Member creates dine-in order ──────────────────────────────
	createReq := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  bomUUID,
				"num":          1,
				"price":        25.00,
				"product_type": 0,
			},
		},
	}
	createResp := memberClient.Post(t, "/api/v1/member/order/dine_in/create", createReq)
	createResp.AssertOK(t)
	createAPI := createResp.ParseAPIResponse(t)
	if createAPI.Code != 0 {
		t.Fatalf("create dine-in order failed: code=%d msg=%s body=%s", createAPI.Code, createAPI.Message, createResp.String())
	}

	var orderResult struct {
		SaleBillUuid  uint64 `json:"sale_bill_uuid"`
		SaleOrderUuid uint64 `json:"sale_order_uuid"`
	}
	if err := json.Unmarshal(createAPI.Data, &orderResult); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	if orderResult.SaleBillUuid == 0 || orderResult.SaleOrderUuid == 0 {
		t.Fatal("expected non-zero sale_bill_uuid and sale_order_uuid")
	}
	t.Logf("Created order: sale_bill_uuid=%d, sale_order_uuid=%d", orderResult.SaleBillUuid, orderResult.SaleOrderUuid)

	// ── 4. Member order detail: should show "pending" (待接单) ───────
	detailResp := memberClient.Get(t, fmt.Sprintf("/api/v1/member/order/dine_in/detail?sale_bill_uuid=%d", orderResult.SaleBillUuid))
	detailResp.AssertOK(t)
	detailAPI := detailResp.ParseAPIResponse(t)
	if detailAPI.Code != 0 {
		t.Fatalf("get order detail failed: code=%d msg=%s", detailAPI.Code, detailAPI.Message)
	}

	var detail struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(detailAPI.Data, &detail); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if detail.Status != "pending" {
		t.Errorf("expected status 'pending' (待接单) after create, got %q", detail.Status)
	}

	// ── 5. Cashier: see H5 order in accept list ──────────────────────
	h5ListResp := cashierClient.Get(t, "/api/v1/cashier/h5_order/list?page_no=1&page_size=20")
	h5ListResp.AssertOK(t)
	h5ListAPI := h5ListResp.ParseAPIResponse(t)
	if h5ListAPI.Code != 0 {
		t.Fatalf("h5_order list failed: code=%d msg=%s", h5ListAPI.Code, h5ListAPI.Message)
	}

	// Find the H5 order UUID for our sale_bill
	var h5List struct {
		List []struct {
			Uuid         uint64 `json:"uuid"`
			SaleBillUuid uint64 `json:"sale_bill_uuid"`
		} `json:"list"`
	}
	if err := json.Unmarshal(h5ListAPI.Data, &h5List); err != nil {
		t.Fatalf("unmarshal h5 list: %v", err)
	}

	var h5OrderUuid uint64
	for _, item := range h5List.List {
		if item.SaleBillUuid == orderResult.SaleBillUuid {
			h5OrderUuid = item.Uuid
			break
		}
	}
	if h5OrderUuid == 0 {
		t.Fatal("H5 order not found in cashier accept list for our sale_bill_uuid")
	}
	t.Logf("Found H5 order: uuid=%d", h5OrderUuid)

	// ── 6. Cashier accepts the H5 order ──────────────────────────────
	acceptReq := map[string]any{
		"h5_order_uuid": h5OrderUuid,
	}
	acceptResp := cashierClient.Post(t, "/api/v1/cashier/h5_order/accept", acceptReq)
	acceptResp.AssertOK(t)
	acceptAPI := acceptResp.ParseAPIResponse(t)
	if acceptAPI.Code != 0 {
		t.Fatalf("accept H5 order failed: code=%d msg=%s body=%s", acceptAPI.Code, acceptAPI.Message, acceptResp.String())
	}
	t.Log("H5 order accepted successfully")

	// ── 7. Cashier: order should appear in instant-order hold list ────
	holdListResp := cashierClient.Get(t, "/api/v1/cashier/instant/order/list?page_no=1&page_size=20")
	holdListResp.AssertOK(t)
	holdListAPI := holdListResp.ParseAPIResponse(t)
	if holdListAPI.Code != 0 {
		t.Fatalf("instant order list failed: code=%d msg=%s", holdListAPI.Code, holdListAPI.Message)
	}

	var holdList struct {
		List []struct {
			SaleBillUuid uint64 `json:"sale_bill_uuid"`
		} `json:"list"`
	}
	if err := json.Unmarshal(holdListAPI.Data, &holdList); err != nil {
		t.Fatalf("unmarshal hold list: %v", err)
	}

	found := false
	for _, item := range holdList.List {
		if item.SaleBillUuid == orderResult.SaleBillUuid {
			found = true
			break
		}
	}
	if !found {
		t.Error("order not found in instant-order hold list after acceptance")
	}

	// ── 8. Member order detail: should show "preparing" (备餐中) ─────
	detailResp2 := memberClient.Get(t, fmt.Sprintf("/api/v1/member/order/dine_in/detail?sale_bill_uuid=%d", orderResult.SaleBillUuid))
	detailResp2.AssertOK(t)
	detailAPI2 := detailResp2.ParseAPIResponse(t)
	if detailAPI2.Code != 0 {
		t.Fatalf("get order detail after accept failed: code=%d msg=%s", detailAPI2.Code, detailAPI2.Message)
	}

	var detail2 struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(detailAPI2.Data, &detail2); err != nil {
		t.Fatalf("unmarshal detail2: %v", err)
	}
	if detail2.Status != "preparing" {
		t.Errorf("expected status 'preparing' (备餐中) after accept, got %q", detail2.Status)
	}

	t.Log("Order-first-pay-later full flow completed successfully")
}

// Test_OrderFirstPayLater_Cancel verifies member can cancel an order-first-pay-later order
// before it is accepted by cashier (H5 status = waiting).
func Test_OrderFirstPayLater_Cancel(t *testing.T) {
	// ── 1. Setup tenant ──────────────────────────────────────────────
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	member := fixture.SeedMember(t, db)
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_order_first_pay_later":1}`)
	bomUUID := fixture.SeedProductWithFlavor(t, db, "取消测试商品", 18.00)
	fixture.SetupWireMock(t)

	memberToken := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))
	memberClient := fixture.NewHTTPClient().WithToken(memberToken)

	// ── 2. Create order ──────────────────────────────────────────────
	createReq := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{"flavor_uuid": bomUUID, "num": 1, "price": 18.00, "product_type": 0},
		},
	}
	createResp := memberClient.Post(t, "/api/v1/member/order/dine_in/create", createReq)
	createResp.AssertOK(t)
	createAPI := createResp.ParseAPIResponse(t)
	if createAPI.Code != 0 {
		t.Fatalf("create failed: code=%d msg=%s", createAPI.Code, createAPI.Message)
	}

	var order struct {
		SaleBillUuid uint64 `json:"sale_bill_uuid"`
	}
	json.Unmarshal(createAPI.Data, &order)

	// ── 3. Cancel order ──────────────────────────────────────────────
	cancelReq := map[string]any{
		"sale_bill_uuid": order.SaleBillUuid,
	}
	cancelResp := memberClient.Post(t, "/api/v1/member/order/dine_in/cancel", cancelReq)
	cancelResp.AssertOK(t)
	cancelAPI := cancelResp.ParseAPIResponse(t)
	if cancelAPI.Code != 0 {
		t.Fatalf("cancel failed: code=%d msg=%s body=%s", cancelAPI.Code, cancelAPI.Message, cancelResp.String())
	}

	// ── 4. Verify order status is cancelled ──────────────────────────
	detailResp := memberClient.Get(t, fmt.Sprintf("/api/v1/member/order/dine_in/detail?sale_bill_uuid=%d", order.SaleBillUuid))
	detailResp.AssertOK(t)
	detailAPI := detailResp.ParseAPIResponse(t)
	if detailAPI.Code != 0 {
		t.Fatalf("detail failed: code=%d msg=%s", detailAPI.Code, detailAPI.Message)
	}

	var detail struct {
		Status string `json:"status"`
	}
	json.Unmarshal(detailAPI.Data, &detail)
	if detail.Status != "cancelled" {
		t.Errorf("expected status 'cancelled' after cancel, got %q", detail.Status)
	}
}

// Test_OrderFirstPayLater_CancelAfterAccept_Rejected verifies that member cannot cancel
// an order after it has been accepted by cashier.
func Test_OrderFirstPayLater_CancelAfterAccept_Rejected(t *testing.T) {
	const testDeviceId = "test-cancel-reject-device"

	// ── 1. Setup tenant ──────────────────────────────────────────────
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	member := fixture.SeedMember(t, db)
	fixture.SeedDevice(t, db, fixture.WithDeviceType("cashier"), fixture.WithDeviceId(testDeviceId))
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffUsername("cashier_cancel_test"),
		fixture.WithStaffIsSuper(1),
		fixture.WithStaffDutyNo("shift-ct"),
	)
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_order_first_pay_later":1}`)
	bomUUID := fixture.SeedProductWithFlavor(t, db, "拒绝取消商品", 20.00)
	fixture.SetupWireMock(t)

	memberToken := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))
	cashierToken := fixture.GenerateStaffToken(t, companyUUID, mustParseString(staff.UUID), staff.Username, 1, testDeviceId)

	memberClient := fixture.NewHTTPClient().WithToken(memberToken)
	cashierClient := fixture.NewHTTPClient().WithToken(cashierToken)

	// ── 2. Create order ──────────────────────────────────────────────
	createReq := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{"flavor_uuid": bomUUID, "num": 1, "price": 20.00, "product_type": 0},
		},
	}
	createResp := memberClient.Post(t, "/api/v1/member/order/dine_in/create", createReq)
	createResp.AssertOK(t)
	createAPI := createResp.ParseAPIResponse(t)
	if createAPI.Code != 0 {
		t.Fatalf("create failed: code=%d msg=%s", createAPI.Code, createAPI.Message)
	}
	var order struct {
		SaleBillUuid uint64 `json:"sale_bill_uuid"`
	}
	json.Unmarshal(createAPI.Data, &order)

	// ── 3. Find and accept H5 order ──────────────────────────────────
	h5ListResp := cashierClient.Get(t, "/api/v1/cashier/h5_order/list?page_no=1&page_size=20")
	h5ListResp.AssertOK(t)
	h5ListAPI := h5ListResp.ParseAPIResponse(t)
	var h5List struct {
		List []struct {
			Uuid         uint64 `json:"uuid"`
			SaleBillUuid uint64 `json:"sale_bill_uuid"`
		} `json:"list"`
	}
	json.Unmarshal(h5ListAPI.Data, &h5List)

	var h5UUID uint64
	for _, item := range h5List.List {
		if item.SaleBillUuid == order.SaleBillUuid {
			h5UUID = item.Uuid
		}
	}
	if h5UUID == 0 {
		t.Fatal("H5 order not found")
	}

	acceptResp := cashierClient.Post(t, "/api/v1/cashier/h5_order/accept", map[string]any{"h5_order_uuid": h5UUID})
	acceptResp.AssertOK(t)

	// ── 4. Try to cancel → should be rejected ────────────────────────
	cancelResp := memberClient.Post(t, "/api/v1/member/order/dine_in/cancel", map[string]any{"sale_bill_uuid": order.SaleBillUuid})
	cancelResp.AssertOK(t) // HTTP 200 but business error expected
	cancelAPI := cancelResp.ParseAPIResponse(t)
	if cancelAPI.Code == 0 {
		t.Error("expected error when cancelling after acceptance, but got success")
	}
	t.Logf("Cancel after accept correctly rejected: code=%d msg=%s", cancelAPI.Code, cancelAPI.Message)
}

func mustParseString(v int64) string {
	return strconv.FormatInt(v, 10)
}
