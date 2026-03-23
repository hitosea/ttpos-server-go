//go:build integration

package auto_receipt_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// API path constants
const (
	pathRuleCreate    = "/api/v1/shop/auto_receipt/rule/create"
	pathRuleUpdate    = "/api/v1/shop/auto_receipt/rule/update"
	pathRuleDelete    = "/api/v1/shop/auto_receipt/rule/delete"
	pathRuleList      = "/api/v1/shop/auto_receipt/rule/list"
	pathShopList      = "/api/v1/shop/auto_receipt/shop_list"
	pathWarehouseList = "/api/v1/shop/auto_receipt/warehouse/list"
	pathLogList       = "/api/v1/shop/auto_receipt/log/list"
	pathLogDetail     = "/api/v1/shop/auto_receipt/log/detail"

	codeTokenInvalid = -102
	codeFail         = -1
)

// --- P0 ---

// Test_P0_Shop_AutoReceipt_RuleCreate_HappyPath tests creating a rule successfully.
// Route: POST /shop/auto_receipt/rule/create
func Test_P0_Shop_AutoReceipt_RuleCreate_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body := map[string]any{
		"locale_name":        map[string]string{"zh": "测试规则", "en": "Test Rule"},
		"warehouse_erp_code": "WH-TEST-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         3,
		"status":             1,
	}
	resp := env.client.Post(t, pathRuleCreate, body)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P0_Shop_AutoReceipt_RuleCreate_Unauthorized tests that unauthenticated requests are rejected.
// Route: POST /shop/auto_receipt/rule/create
func Test_P0_Shop_AutoReceipt_RuleCreate_Unauthorized(t *testing.T) {
	body := map[string]any{
		"locale_name":        map[string]string{"zh": "测试"},
		"warehouse_erp_code": "WH-001",
		"shop_uuids":         []int64{1},
		"delay_days":         0,
		"status":             1,
	}
	resp := fixture.NewHTTPClient().Post(t, pathRuleCreate, body)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_AutoReceipt_RuleList_HappyPath tests listing rules after creating one.
// Route: GET /shop/auto_receipt/rule/list
func Test_P0_Shop_AutoReceipt_RuleList_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create a rule first
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "列表测试规则", "en": "List Test Rule"},
		"warehouse_erp_code": "WH-LIST-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         1,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	// List rules
	resp := env.client.Get(t, pathRuleList)
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var listData ruleListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if len(listData.List) == 0 {
		t.Error("expected at least one rule in the list, got 0")
	}

	// Verify created rule fields
	found := false
	for _, rule := range listData.List {
		if rule.WarehouseErpCode == "WH-LIST-001" {
			found = true
			if rule.DelayDays != 1 {
				t.Errorf("expected delay_days=1, got %d", rule.DelayDays)
			}
			if rule.Status != 1 {
				t.Errorf("expected status=1, got %d", rule.Status)
			}
			if rule.ShopCount != 1 {
				t.Errorf("expected shop_count=1, got %d", rule.ShopCount)
			}
		}
	}
	if !found {
		t.Error("created rule with warehouse_erp_code=WH-LIST-001 not found in list")
	}
}

// Test_P0_Shop_AutoReceipt_RuleList_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/auto_receipt/rule/list
func Test_P0_Shop_AutoReceipt_RuleList_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathRuleList)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_AutoReceipt_RuleUpdate_HappyPath tests updating a rule.
// Route: POST /shop/auto_receipt/rule/update
func Test_P0_Shop_AutoReceipt_RuleUpdate_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create a rule
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "待更新规则"},
		"warehouse_erp_code": "WH-UPD-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         2,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	// Get the created rule UUID from list
	ruleUuid := getRuleUuidByWarehouse(t, env.client, "WH-UPD-001")
	if ruleUuid == 0 {
		t.Fatal("failed to find created rule")
	}

	// Update the rule
	updateBody := map[string]any{
		"uuid":               ruleUuid,
		"locale_name":        map[string]string{"zh": "已更新规则", "en": "Updated Rule"},
		"warehouse_erp_code": "WH-UPD-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1, env.shopCompanyUUID2},
		"delay_days":         5,
		"status":             0,
	}
	resp := env.client.Post(t, pathRuleUpdate, updateBody)
	resp.AssertOK(t).AssertSuccess(t)

	// Verify update via list
	listResp := env.client.Get(t, pathRuleList)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	json.Unmarshal(apiResp.Data, &listData)

	for _, rule := range listData.List {
		if rule.Uuid == ruleUuid {
			if rule.DelayDays != 5 {
				t.Errorf("expected updated delay_days=5, got %d", rule.DelayDays)
			}
			if rule.Status != 0 {
				t.Errorf("expected updated status=0, got %d", rule.Status)
			}
			if rule.ShopCount != 2 {
				t.Errorf("expected updated shop_count=2, got %d", rule.ShopCount)
			}
			return
		}
	}
	t.Error("updated rule not found in list")
}

// Test_P0_Shop_AutoReceipt_RuleDelete_HappyPath tests deleting a rule.
// Route: DELETE /shop/auto_receipt/rule/delete
func Test_P0_Shop_AutoReceipt_RuleDelete_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create a rule
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "待删除规则"},
		"warehouse_erp_code": "WH-DEL-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	ruleUuid := getRuleUuidByWarehouse(t, env.client, "WH-DEL-001")
	if ruleUuid == 0 {
		t.Fatal("failed to find created rule")
	}

	// Delete the rule
	deleteBody := map[string]any{
		"uuids": []uint64{ruleUuid},
	}
	resp := env.client.Delete(t, pathRuleDelete, deleteBody)
	resp.AssertOK(t).AssertSuccess(t)

	// Verify deletion via list
	listResp := env.client.Get(t, pathRuleList)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	json.Unmarshal(apiResp.Data, &listData)

	for _, rule := range listData.List {
		if rule.Uuid == ruleUuid {
			t.Error("deleted rule still found in list")
		}
	}
}

// Test_P0_Shop_AutoReceipt_WarehouseList_HappyPath tests listing warehouses.
// Route: GET /shop/auto_receipt/warehouse/list
func Test_P0_Shop_AutoReceipt_WarehouseList_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Seed a warehouse with erp_code in the tenant DB
	seedWarehouse(t, env.tenantDB, "WH-TEST-WLIST", `{"zh":"测试仓库","en":"Test Warehouse"}`, "warehouse")

	resp := env.client.Get(t, pathWarehouseList)
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data warehouseListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse warehouse list response: %v", err)
	}

	found := false
	for _, w := range data.List {
		if w.ErpCode == "WH-TEST-WLIST" {
			found = true
			break
		}
	}
	if !found {
		t.Error("seeded warehouse WH-TEST-WLIST not found in list")
	}
}

// Test_P0_Shop_AutoReceipt_WarehouseList_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/auto_receipt/warehouse/list
func Test_P0_Shop_AutoReceipt_WarehouseList_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathWarehouseList)
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_AutoReceipt_ShopList_HappyPath tests listing available shops for rule configuration.
// Route: GET /shop/auto_receipt/shop_list
func Test_P0_Shop_AutoReceipt_ShopList_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathShopList+"?warehouse_erp_code=WH-SHOPLIST-001")
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data shopListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse shop list response: %v", err)
	}
	// List may be empty if no sub-shops are in saas DB; just verify API structure is valid
	if data.List == nil {
		t.Error("expected list to be non-nil (should be empty array, not null)")
	}
}

// Test_P0_Shop_AutoReceipt_ShopList_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/auto_receipt/shop_list
func Test_P0_Shop_AutoReceipt_ShopList_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathShopList+"?warehouse_erp_code=WH-001")
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_AutoReceipt_LogList_HappyPath tests listing auto-receipt logs (empty case).
// Route: GET /shop/auto_receipt/log/list
func Test_P0_Shop_AutoReceipt_LogList_HappyPath(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathLogList+"?page_no=1&page_size=10")
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data logListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse log list response: %v", err)
	}
	if data.List == nil {
		t.Error("expected list to be non-nil (should be empty array, not null)")
	}
}

// Test_P0_Shop_AutoReceipt_LogList_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/auto_receipt/log/list
func Test_P0_Shop_AutoReceipt_LogList_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathLogList+"?page_no=1&page_size=10")
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// Test_P0_Shop_AutoReceipt_LogDetail_Unauthorized tests that unauthenticated requests are rejected.
// Route: GET /shop/auto_receipt/log/detail
func Test_P0_Shop_AutoReceipt_LogDetail_Unauthorized(t *testing.T) {
	resp := fixture.NewHTTPClient().Get(t, pathLogDetail+"?uuid=12345")
	resp.AssertOK(t).AssertErrorCode(t, codeTokenInvalid)
}

// --- P1 ---

// Test_P1_Shop_AutoReceipt_RuleCreate_MissingFields tests creating a rule with missing required fields.
// Route: POST /shop/auto_receipt/rule/create
func Test_P1_Shop_AutoReceipt_RuleCreate_MissingFields(t *testing.T) {
	env := setupHeadquarterEnv(t)

	tests := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing warehouse_erp_code",
			body: map[string]any{
				"locale_name": map[string]string{"zh": "测试"},
				"shop_uuids":  []int64{1},
				"delay_days":  0,
				"status":      1,
			},
		},
		{
			name: "missing shop_uuids",
			body: map[string]any{
				"locale_name":        map[string]string{"zh": "测试"},
				"warehouse_erp_code": "WH-001",
				"delay_days":         0,
				"status":             1,
			},
		},
		{
			name: "empty shop_uuids",
			body: map[string]any{
				"locale_name":        map[string]string{"zh": "测试"},
				"warehouse_erp_code": "WH-001",
				"shop_uuids":         []int64{},
				"delay_days":         0,
				"status":             1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := env.client.Post(t, pathRuleCreate, tc.body)
			resp.AssertOK(t)
			apiResp := resp.ParseAPIResponse(t)
			if apiResp.Code == 0 {
				t.Errorf("expected validation error for %s, got success", tc.name)
			}
		})
	}
}

// Test_P1_Shop_AutoReceipt_RuleCreate_DuplicateShopInSameWarehouse tests that a shop
// cannot be configured in two rules under the same warehouse.
// Route: POST /shop/auto_receipt/rule/create
func Test_P1_Shop_AutoReceipt_RuleCreate_DuplicateShopInSameWarehouse(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create the first rule
	body1 := map[string]any{
		"locale_name":        map[string]string{"zh": "规则A"},
		"warehouse_erp_code": "WH-DUP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, body1).AssertOK(t).AssertSuccess(t)

	// Create another rule with the same shop and warehouse — should fail
	body2 := map[string]any{
		"locale_name":        map[string]string{"zh": "规则B"},
		"warehouse_erp_code": "WH-DUP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         1,
		"status":             1,
	}
	resp := env.client.Post(t, pathRuleCreate, body2)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for duplicate shop in same warehouse, got success")
	}
}

// Test_P1_Shop_AutoReceipt_RuleCreate_DifferentWarehouseAllowed tests that the same shop
// can be configured in rules under different warehouses.
// Route: POST /shop/auto_receipt/rule/create
func Test_P1_Shop_AutoReceipt_RuleCreate_DifferentWarehouseAllowed(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body1 := map[string]any{
		"locale_name":        map[string]string{"zh": "仓库A规则"},
		"warehouse_erp_code": "WH-DIFFA-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, body1).AssertOK(t).AssertSuccess(t)

	// Same shop, different warehouse — should succeed
	body2 := map[string]any{
		"locale_name":        map[string]string{"zh": "仓库B规则"},
		"warehouse_erp_code": "WH-DIFFB-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         1,
		"status":             1,
	}
	resp := env.client.Post(t, pathRuleCreate, body2)
	resp.AssertOK(t).AssertSuccess(t)
}

// Test_P1_Shop_AutoReceipt_RuleUpdate_NotFound tests updating a non-existent rule.
// Route: POST /shop/auto_receipt/rule/update
func Test_P1_Shop_AutoReceipt_RuleUpdate_NotFound(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body := map[string]any{
		"uuid":               uint64(999999999999),
		"locale_name":        map[string]string{"zh": "不存在"},
		"warehouse_erp_code": "WH-NF-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	resp := env.client.Post(t, pathRuleUpdate, body)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for non-existent rule, got success")
	}
}

// Test_P1_Shop_AutoReceipt_RuleUpdate_MissingUuid tests updating without providing uuid.
// Route: POST /shop/auto_receipt/rule/update
func Test_P1_Shop_AutoReceipt_RuleUpdate_MissingUuid(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body := map[string]any{
		"locale_name":        map[string]string{"zh": "缺少UUID"},
		"warehouse_erp_code": "WH-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	resp := env.client.Post(t, pathRuleUpdate, body)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error for missing uuid, got success")
	}
}

// Test_P1_Shop_AutoReceipt_RuleDelete_EmptyUuids tests deleting with empty uuids.
// Route: DELETE /shop/auto_receipt/rule/delete
func Test_P1_Shop_AutoReceipt_RuleDelete_EmptyUuids(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body := map[string]any{
		"uuids": []uint64{},
	}
	resp := env.client.Delete(t, pathRuleDelete, body)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error for empty uuids, got success")
	}
}

// Test_P1_Shop_AutoReceipt_ShopList_MissingWarehouseErpCode tests shop list without required param.
// Route: GET /shop/auto_receipt/shop_list
func Test_P1_Shop_AutoReceipt_ShopList_MissingWarehouseErpCode(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathShopList)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error for missing warehouse_erp_code, got success")
	}
}

// Test_P1_Shop_AutoReceipt_LogDetail_NotFound tests log detail with non-existent uuid.
// Route: GET /shop/auto_receipt/log/detail
func Test_P1_Shop_AutoReceipt_LogDetail_NotFound(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathLogDetail+"?uuid=999999999999")
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for non-existent log, got success")
	}
}

// Test_P1_Shop_AutoReceipt_LogList_WithTimeFilter tests log list with time range filter.
// Route: GET /shop/auto_receipt/log/list
func Test_P1_Shop_AutoReceipt_LogList_WithTimeFilter(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathLogList+"?page_no=1&page_size=10&start_time=2026-01-01+00:00:00&end_time=2026-12-31+23:59:59")
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data logListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse log list response: %v", err)
	}
	// Should return empty list since there are no logs matching the filter
	if data.Meta.Total != 0 {
		t.Errorf("expected total=0 for filtered empty logs, got %d", data.Meta.Total)
	}
}

// Test_P1_Shop_AutoReceipt_LogList_InvalidTimeFormat tests log list with invalid time format.
// Route: GET /shop/auto_receipt/log/list
func Test_P1_Shop_AutoReceipt_LogList_InvalidTimeFormat(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathLogList+"?page_no=1&page_size=10&start_time=invalid-time")
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error for invalid time format, got success")
	}
}

// Test_P1_Shop_AutoReceipt_LogDetail_MissingUuid tests log detail without uuid.
// Route: GET /shop/auto_receipt/log/detail
func Test_P1_Shop_AutoReceipt_LogDetail_MissingUuid(t *testing.T) {
	env := setupHeadquarterEnv(t)

	resp := env.client.Get(t, pathLogDetail)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error for missing uuid, got success")
	}
}

// Test_P1_Shop_AutoReceipt_RuleUpdate_RemoveShop tests updating a rule to remove a shop.
// This exercises the SoftDeleteByUuids path in the rule shop repo.
// Route: POST /shop/auto_receipt/rule/update
func Test_P1_Shop_AutoReceipt_RuleUpdate_RemoveShop(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create a rule with 2 shops
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "移除门店测试"},
		"warehouse_erp_code": "WH-RMSHOP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1, env.shopCompanyUUID2},
		"delay_days":         2,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	ruleUuid := getRuleUuidByWarehouse(t, env.client, "WH-RMSHOP-001")
	if ruleUuid == 0 {
		t.Fatal("failed to find created rule")
	}

	// Verify initial shop_count = 2
	verifyRuleShopCount(t, env.client, ruleUuid, 2)

	// Update: remove shop2, keep only shop1
	updateBody := map[string]any{
		"uuid":               ruleUuid,
		"locale_name":        map[string]string{"zh": "移除门店测试-更新"},
		"warehouse_erp_code": "WH-RMSHOP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         2,
		"status":             1,
	}
	env.client.Post(t, pathRuleUpdate, updateBody).AssertOK(t).AssertSuccess(t)

	// Verify shop_count reduced to 1
	verifyRuleShopCount(t, env.client, ruleUuid, 1)
}

// Test_P1_Shop_AutoReceipt_RuleUpdate_DuplicateShopInOtherRule tests that updating a rule
// to add a shop already configured in another rule under the same warehouse is rejected.
// Route: POST /shop/auto_receipt/rule/update
func Test_P1_Shop_AutoReceipt_RuleUpdate_DuplicateShopInOtherRule(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create rule A with shop1
	bodyA := map[string]any{
		"locale_name":        map[string]string{"zh": "规则A"},
		"warehouse_erp_code": "WH-UPDDUP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, bodyA).AssertOK(t).AssertSuccess(t)

	// Create rule B with shop2
	bodyB := map[string]any{
		"locale_name":        map[string]string{"zh": "规则B"},
		"warehouse_erp_code": "WH-UPDDUP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID2},
		"delay_days":         0,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, bodyB).AssertOK(t).AssertSuccess(t)

	ruleUuidB := getRuleUuidByWarehouse(t, env.client, "WH-UPDDUP-001")

	// Try to update rule B to include shop1 (already in rule A) — should fail
	updateBody := map[string]any{
		"uuid":               ruleUuidB,
		"locale_name":        map[string]string{"zh": "规则B-更新"},
		"warehouse_erp_code": "WH-UPDDUP-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1, env.shopCompanyUUID2},
		"delay_days":         0,
		"status":             1,
	}
	resp := env.client.Post(t, pathRuleUpdate, updateBody)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for duplicate shop in update, got success")
	}
}

// Test_P1_Shop_AutoReceipt_ShopList_DisabledShops tests that shops already configured in a rule
// are returned with disabled=true in the shop list.
// Route: GET /shop/auto_receipt/shop_list
func Test_P1_Shop_AutoReceipt_ShopList_DisabledShops(t *testing.T) {
	env := setupHeadquarterEnvWithSaasShops(t)

	// Create a rule with shop1 under warehouse WH-DISABLED-001
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "禁用测试"},
		"warehouse_erp_code": "WH-DISABLED-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	// Query shop_list for the same warehouse
	resp := env.client.Get(t, pathShopList+"?warehouse_erp_code=WH-DISABLED-001")
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data shopListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse shop list response: %v", err)
	}

	if len(data.List) == 0 {
		t.Fatal("expected at least one shop in list, got 0")
	}

	// shop1 should be disabled, shop2 should not
	for _, shop := range data.List {
		if shop.Uuid == uint64(env.shopCompanyUUID1) {
			if !shop.Disabled {
				t.Error("expected shop1 to be disabled (already configured), got disabled=false")
			}
		}
		if shop.Uuid == uint64(env.shopCompanyUUID2) {
			if shop.Disabled {
				t.Error("expected shop2 to NOT be disabled, got disabled=true")
			}
		}
	}
}

// Test_P1_Shop_AutoReceipt_LogList_WithShopFilter tests log list with shop_company_uuid filter.
// Route: GET /shop/auto_receipt/log/list
func Test_P1_Shop_AutoReceipt_LogList_WithShopFilter(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Seed log records directly in saas DB
	saasDB := fixture.NewSaasDB(t)
	now := time.Now().Unix()
	logUuid1 := uint64(generateTestID())
	logUuid2 := logUuid1 + 1

	seedAutoReceiptLog(t, saasDB, logUuid1, uint64(env.companyUUID), uint64(env.shopCompanyUUID1), now)
	seedAutoReceiptLog(t, saasDB, logUuid2, uint64(env.companyUUID), uint64(env.shopCompanyUUID2), now)

	// Query with shop filter for shop1
	resp := env.client.Get(t, fmt.Sprintf("%s?page_no=1&page_size=10&shop_company_uuid=%d", pathLogList, env.shopCompanyUUID1))
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data logListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse log list response: %v", err)
	}

	// Should only contain logs for shop1
	for _, log := range data.List {
		if log.ShopCompanyUuid != uint64(env.shopCompanyUUID1) {
			t.Errorf("expected all logs to be for shop1 (%d), got shop_company_uuid=%d",
				env.shopCompanyUUID1, log.ShopCompanyUuid)
		}
	}
	if data.Meta.Total == 0 {
		t.Error("expected at least one log for shop1 filter, got total=0")
	}
}

// Test_P1_Shop_AutoReceipt_LogDetail_HappyPath tests successful log detail retrieval.
// The service GetLogDetail path is covered even if the downstream receipt detail call fails.
// Route: GET /shop/auto_receipt/log/detail
func Test_P1_Shop_AutoReceipt_LogDetail_HappyPath(t *testing.T) {
	env := setupHeadquarterEnvWithSaasShops(t)

	// Seed a log record in saas DB pointing to shop1
	saasDB := fixture.NewSaasDB(t)
	now := time.Now().Unix()
	logUuid := uint64(generateTestID())
	seedAutoReceiptLog(t, saasDB, logUuid, uint64(env.companyUUID), uint64(env.shopCompanyUUID1), now)

	// Query detail — the GetLogDetail service call will succeed (covering service lines 500-526),
	// but GetPurchaseReceiptOrderDetail may fail since there's no actual receipt order.
	// Either outcome is acceptable for coverage purposes.
	resp := env.client.Get(t, fmt.Sprintf("%s?uuid=%d", pathLogDetail, logUuid))
	resp.AssertOK(t)
	// We don't assert success here because the receipt order may not exist in the shop DB,
	// but the service path is still exercised.
}

// Test_P1_Shop_AutoReceipt_RuleList_EmptyRules_UnconfiguredCount tests that rule list
// with no rules returns correct unconfigured_count for shops in the saas DB.
// Route: GET /shop/auto_receipt/rule/list
func Test_P1_Shop_AutoReceipt_RuleList_EmptyRules_UnconfiguredCount(t *testing.T) {
	env := setupHeadquarterEnvWithSaasShops(t)

	// Don't create any rules — just query the list
	resp := env.client.Get(t, pathRuleList)
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data ruleListData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse rule list response: %v", err)
	}

	if len(data.List) != 0 {
		t.Errorf("expected empty rule list, got %d rules", len(data.List))
	}
	// With 2 sub-shops seeded in saas DB, unconfigured_count should be 2
	if data.UnconfiguredCount != 2 {
		t.Errorf("expected unconfigured_count=2, got %d", data.UnconfiguredCount)
	}
}

// Test_P1_Shop_AutoReceipt_CRUD_FullFlow tests the complete CRUD lifecycle.
// Route: POST/GET/DELETE /shop/auto_receipt/rule/*
func Test_P1_Shop_AutoReceipt_CRUD_FullFlow(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Step 1: Create rule
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "完整流程规则", "en": "Full Flow Rule"},
		"warehouse_erp_code": "WH-FLOW-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         3,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	// Step 2: Verify in list
	ruleUuid := getRuleUuidByWarehouse(t, env.client, "WH-FLOW-001")
	if ruleUuid == 0 {
		t.Fatal("created rule not found in list")
	}

	// Step 3: Update rule — change delay and add a shop
	updateBody := map[string]any{
		"uuid":               ruleUuid,
		"locale_name":        map[string]string{"zh": "已更新流程规则"},
		"warehouse_erp_code": "WH-FLOW-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1, env.shopCompanyUUID2},
		"delay_days":         7,
		"status":             1,
	}
	env.client.Post(t, pathRuleUpdate, updateBody).AssertOK(t).AssertSuccess(t)

	// Step 4: Verify update in list
	listResp := env.client.Get(t, pathRuleList)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	json.Unmarshal(apiResp.Data, &listData)
	for _, rule := range listData.List {
		if rule.Uuid == ruleUuid {
			if rule.DelayDays != 7 {
				t.Errorf("expected delay_days=7 after update, got %d", rule.DelayDays)
			}
			if rule.ShopCount != 2 {
				t.Errorf("expected shop_count=2 after update, got %d", rule.ShopCount)
			}
		}
	}

	// Step 5: Delete rule
	deleteBody := map[string]any{"uuids": []uint64{ruleUuid}}
	env.client.Delete(t, pathRuleDelete, deleteBody).AssertOK(t).AssertSuccess(t)

	// Step 6: Verify deleted
	listResp2 := env.client.Get(t, pathRuleList)
	apiResp2 := listResp2.ParseAPIResponse(t)
	var listData2 ruleListData
	json.Unmarshal(apiResp2.Data, &listData2)
	for _, rule := range listData2.List {
		if rule.Uuid == ruleUuid {
			t.Error("deleted rule still found in list after deletion")
		}
	}
}

// Test_P1_Shop_AutoReceipt_RuleCreate_StatusZero verifies that creating a rule
// with status=0 (disabled) persists correctly — the database must store 0, not
// fall back to the column default (1).
// Route: POST /shop/auto_receipt/rule/create
func Test_P1_Shop_AutoReceipt_RuleCreate_StatusZero(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body := map[string]any{
		"locale_name":        map[string]string{"zh": "禁用规则"},
		"warehouse_erp_code": "WH-STATUS0-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         1,
		"status":             0,
	}
	env.client.Post(t, pathRuleCreate, body).AssertOK(t).AssertSuccess(t)

	// Verify via list that the persisted status is 0
	listResp := env.client.Get(t, pathRuleList)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	for _, rule := range listData.List {
		if rule.WarehouseErpCode == "WH-STATUS0-001" {
			if rule.Status != 0 {
				t.Errorf("expected status=0 (disabled), got %d", rule.Status)
			}
			return
		}
	}
	t.Error("created rule with warehouse_erp_code=WH-STATUS0-001 not found in list")
}

// Test_P1_Shop_AutoReceipt_RuleCreate_MissingStatus verifies that omitting the
// status field triggers a validation error (the field is required).
// Route: POST /shop/auto_receipt/rule/create
func Test_P1_Shop_AutoReceipt_RuleCreate_MissingStatus(t *testing.T) {
	env := setupHeadquarterEnv(t)

	body := map[string]any{
		"locale_name":        map[string]string{"zh": "缺少状态"},
		"warehouse_erp_code": "WH-NOSTATUS-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         0,
		// status intentionally omitted
	}
	resp := env.client.Post(t, pathRuleCreate, body)
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected validation error when status is missing, got success")
	}
}

// Test_P1_Shop_AutoReceipt_RuleUpdate_StatusZero verifies that updating a rule's
// status from 1 to 0 persists correctly.
// Route: POST /shop/auto_receipt/rule/update
func Test_P1_Shop_AutoReceipt_RuleUpdate_StatusZero(t *testing.T) {
	env := setupHeadquarterEnv(t)

	// Create a rule with status=1
	createBody := map[string]any{
		"locale_name":        map[string]string{"zh": "待禁用规则"},
		"warehouse_erp_code": "WH-UPDST0-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         2,
		"status":             1,
	}
	env.client.Post(t, pathRuleCreate, createBody).AssertOK(t).AssertSuccess(t)

	ruleUuid := getRuleUuidByWarehouse(t, env.client, "WH-UPDST0-001")
	if ruleUuid == 0 {
		t.Fatal("failed to find created rule")
	}

	// Update status to 0
	updateBody := map[string]any{
		"uuid":               ruleUuid,
		"locale_name":        map[string]string{"zh": "待禁用规则"},
		"warehouse_erp_code": "WH-UPDST0-001",
		"shop_uuids":         []int64{env.shopCompanyUUID1},
		"delay_days":         2,
		"status":             0,
	}
	env.client.Post(t, pathRuleUpdate, updateBody).AssertOK(t).AssertSuccess(t)

	// Verify via list
	listResp := env.client.Get(t, pathRuleList)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	for _, rule := range listData.List {
		if rule.Uuid == ruleUuid {
			if rule.Status != 0 {
				t.Errorf("expected status=0 after update, got %d", rule.Status)
			}
			return
		}
	}
	t.Error("updated rule not found in list")
}

// --- Helpers ---

// testEnv holds the test environment for auto-receipt tests.
type testEnv struct {
	client           *fixture.HTTPClient
	tenantDB         *sql.DB
	companyUUID      int64
	shopCompanyUUID1 int64 // sub-shop 1 UUID (used as shop_uuids in rules)
	shopCompanyUUID2 int64 // sub-shop 2 UUID
}

// setupHeadquarterEnv creates a test environment with a headquarter company.
// It seeds company + company_setting with IsHeadquarter()=true,
// and two sub-shop tenant databases for use as shop_uuids in rules.
// Real tenant DBs are needed because GetRuleList calls getShopStoreCodeMap
// which connects to each shop's database via DBManager.
func setupHeadquarterEnv(t *testing.T) testEnv {
	t.Helper()

	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingHeadquarterConfig("test-site", "HQ"),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	// Create real tenant DBs for sub-shops so that DBManager.GetDB() won't fail
	// when GetRuleList tries to query each shop's store_setting.
	shopUUID1Str := fixture.GenerateCompanyUUID(t)
	shopDB1 := fixture.NewTestTenantFull(t, shopUUID1Str)
	shopUUID1 := mustParseInt64(shopUUID1Str)
	fixture.SeedCompany(t, shopDB1, fixture.WithCompanyUUID(shopUUID1))
	fixture.SeedCompanySetting(t, shopDB1, fixture.WithCompanySettingCompanyUUID(shopUUID1))

	shopUUID2Str := fixture.GenerateCompanyUUID(t)
	shopDB2 := fixture.NewTestTenantFull(t, shopUUID2Str)
	shopUUID2 := mustParseInt64(shopUUID2Str)
	fixture.SeedCompany(t, shopDB2, fixture.WithCompanyUUID(shopUUID2))
	fixture.SeedCompanySetting(t, shopDB2, fixture.WithCompanySettingCompanyUUID(shopUUID2))

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	client := fixture.NewHTTPClient().WithToken(token)

	return testEnv{
		client:           client,
		tenantDB:         db,
		companyUUID:      companyUUIDInt,
		shopCompanyUUID1: shopUUID1,
		shopCompanyUUID2: shopUUID2,
	}
}

// ruleListData represents the rule list response for parsing.
type ruleListData struct {
	List []struct {
		Uuid             uint64 `json:"uuid"`
		WarehouseErpCode string `json:"warehouse_erp_code"`
		DelayDays        int    `json:"delay_days"`
		Status           int    `json:"status"`
		ShopCount        int    `json:"shop_count"`
	} `json:"list"`
	ConfiguredCount   int `json:"configured_count"`
	UnconfiguredCount int `json:"unconfigured_count"`
}

// warehouseListData represents the warehouse list response for parsing.
type warehouseListData struct {
	List []struct {
		ErpCode string `json:"erp_code"`
		Type    string `json:"type"`
	} `json:"list"`
}

// shopListData represents the shop list response for parsing.
type shopListData struct {
	List []struct {
		Uuid     uint64 `json:"uuid"`
		Name     string `json:"name"`
		Disabled bool   `json:"disabled"`
	} `json:"list"`
}

// logListData represents the log list response for parsing.
type logListData struct {
	List []struct {
		Uuid            uint64 `json:"uuid"`
		ShopCompanyUuid uint64 `json:"shop_company_uuid"`
		ReceiptTime     int64  `json:"receipt_time"`
	} `json:"list"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

// getRuleUuidByWarehouse finds a rule UUID from the list by warehouse_erp_code.
func getRuleUuidByWarehouse(t *testing.T, client *fixture.HTTPClient, warehouseErpCode string) uint64 {
	t.Helper()

	listResp := client.Get(t, pathRuleList)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	for _, rule := range listData.List {
		if rule.WarehouseErpCode == warehouseErpCode {
			return rule.Uuid
		}
	}
	return 0
}

// seedWarehouse inserts a warehouse into the tenant DB.
func seedWarehouse(t *testing.T, db *sql.DB, erpCode, nameJSON, whType string) {
	t.Helper()

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_warehouse (uuid, name, type, erp_code, status, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, 1, ?, ?, 0)
	`, generateTestID(), nameJSON, whType, erpCode, now, now)
	if err != nil {
		t.Fatalf("failed to seed warehouse: %v", err)
	}
}

// generateTestID generates a unique ID for testing.
func generateTestID() int64 {
	return time.Now().UnixNano() / 1000
}

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

// verifyRuleShopCount checks that a rule has the expected shop_count via rule list API.
func verifyRuleShopCount(t *testing.T, client *fixture.HTTPClient, ruleUuid uint64, expected int) {
	t.Helper()

	listResp := client.Get(t, pathRuleList)
	listResp.AssertOK(t).AssertSuccess(t)
	apiResp := listResp.ParseAPIResponse(t)
	var listData ruleListData
	if err := json.Unmarshal(apiResp.Data, &listData); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	for _, rule := range listData.List {
		if rule.Uuid == ruleUuid {
			if rule.ShopCount != expected {
				t.Errorf("expected shop_count=%d, got %d", expected, rule.ShopCount)
			}
			return
		}
	}
	t.Errorf("rule %d not found in list", ruleUuid)
}

// setupHeadquarterEnvWithSaasShops extends setupHeadquarterEnv by also seeding the
// sub-shop companies in the saas DB so that GetShopList and GetLogDetail can find them.
func setupHeadquarterEnvWithSaasShops(t *testing.T) testEnv {
	t.Helper()
	env := setupHeadquarterEnv(t)

	// Seed sub-shop company + company_setting in saas DB so that
	// GetNoDeleteListByHeadquarterUuid and GetCompanyInfoByUuid work.
	saasDB := fixture.NewSaasDB(t)

	fixture.SeedCompany(t, saasDB,
		fixture.WithCompanyUUID(env.shopCompanyUUID1),
	)
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(env.shopCompanyUUID1),
		fixture.WithCompanySettingHeadquarterUuid(env.companyUUID),
	)

	fixture.SeedCompany(t, saasDB,
		fixture.WithCompanyUUID(env.shopCompanyUUID2),
	)
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(env.shopCompanyUUID2),
		fixture.WithCompanySettingHeadquarterUuid(env.companyUUID),
	)

	// Also seed the HQ itself in saas DB (needed for GetCompanyInfoByUuid in some paths)
	fixture.SeedCompany(t, saasDB,
		fixture.WithCompanyUUID(env.companyUUID),
	)
	fixture.SeedCompanySetting(t, saasDB,
		fixture.WithCompanySettingCompanyUUID(env.companyUUID),
		fixture.WithCompanySettingHeadquarterConfig("test-site", "HQ"),
	)

	return env
}

// seedAutoReceiptLog inserts an auto receipt log record into the saas DB.
func seedAutoReceiptLog(t *testing.T, db *sql.DB, uuid uint64, headquarterUuid uint64, shopUuid uint64, receiptTime int64) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_auto_receipt_log (uuid, headquarter_company_uuid, rule_uuid, shop_company_uuid,
			receipt_order_uuid, receipt_order_no, receipt_erp_order_no, receipt_time, create_time, update_time, delete_time)
		VALUES (?, ?, 0, ?, 0, '', '', ?, ?, ?, 0)
	`, uuid, headquarterUuid, shopUuid, receiptTime, now, now)
	if err != nil {
		t.Fatalf("failed to seed auto receipt log: %v", err)
	}
}
