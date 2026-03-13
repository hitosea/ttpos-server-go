//go:build integration

package member_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"ttpos-server-go/tests/fixture"
)

// Test_Member_DineIn_Create_Success 验证堂食订单创建成功
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_Success(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 创建商品（product + product_package + product_flavor + product_bom 完整链路）
	bomUUID := fixture.SeedProductWithFlavor(t, db, "测试鸡排", 25.00)

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 构造请求
	reqBody := map[string]any{
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

	// 8. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 9. 断言 HTTP 200
	resp.AssertOK(t)

	// 10. 解析响应
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code != 0 {
		t.Fatalf("expected success code 0 but got %d: %s (body: %s)", apiResp.Code, apiResp.Message, resp.String())
	}

	// 11. 验证返回的 sale_bill_uuid 和 sale_order_uuid 不为 0
	var result struct {
		SaleBillUuid  uint64 `json:"sale_bill_uuid"`
		SaleOrderUuid uint64 `json:"sale_order_uuid"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if result.SaleBillUuid == 0 {
		t.Error("expected sale_bill_uuid != 0, got 0")
	}
	if result.SaleOrderUuid == 0 {
		t.Error("expected sale_order_uuid != 0, got 0")
	}
}

// Test_Member_DineIn_Create_OutsideHours 验证营业时间外创建堂食订单被拒绝
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_OutsideHours(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：营业时间不包含当前时间（+2h ~ +3h）
	openingHours := outsideOpeningHours()
	fixture.SeedSetting(t, db, "business", fmt.Sprintf(`{"opening_hours":"%s"}`, openingHours))

	// 5. 创建商品
	bomUUID := fixture.SeedProductWithFlavor(t, db, "测试汉堡", 30.00)

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 构造请求
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  bomUUID,
				"num":          1,
				"price":        30.00,
				"product_type": 0,
			},
		},
	}

	// 8. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 9. 断言 HTTP 200（业务错误也返回200）
	resp.AssertOK(t)

	// 10. 验证业务错误码 != 0（店铺休息中）
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Errorf("expected error when outside opening hours (%s), but got success", openingHours)
	}
}

// Test_Member_DineIn_Create_EmptyProducts 验证空商品列表被拒绝
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_EmptyProducts(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 6. 构造请求：空商品列表
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products":        []map[string]any{},
	}

	// 7. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 8. 断言 HTTP 200
	resp.AssertOK(t)

	// 9. 验证业务错误码 != 0（未选购商品）
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for empty products list, but got success")
	}
}

// Test_Member_DineIn_Create_Package_Success 验证套餐商品创建堂食订单成功
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_Package_Success(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 创建套餐商品（父套餐 + 分组 + 2个子商品）
	pkgResult := fixture.SeedPackageProductWithSubItems(t, db, "测试套餐A", 50.00, []fixture.PackageSubItem{
		{Name: "套餐鸡排", Price: 25.00},
		{Name: "套餐可乐", Price: 8.00},
	})

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 构造套餐请求
	subProducts := make([]map[string]any, 0, len(pkgResult.SubBomUUIDs))
	for _, subBomUUID := range pkgResult.SubBomUUIDs {
		subProducts = append(subProducts, map[string]any{
			"flavor_uuid":                subBomUUID,
			"num":                        1,
			"product_package_group_uuid": pkgResult.GroupUUID,
		})
	}
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  pkgResult.PackageBomUUID,
				"num":          1,
				"price":        50.00,
				"product_type": 1,
				"products":     subProducts,
			},
		},
	}

	// 8. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 9. 断言 HTTP 200
	resp.AssertOK(t)

	// 10. 解析响应
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code != 0 {
		t.Fatalf("expected success code 0 but got %d: %s (body: %s)", apiResp.Code, apiResp.Message, resp.String())
	}

	// 11. 验证返回的 sale_bill_uuid 和 sale_order_uuid 不为 0
	var result struct {
		SaleBillUuid  uint64 `json:"sale_bill_uuid"`
		SaleOrderUuid uint64 `json:"sale_order_uuid"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if result.SaleBillUuid == 0 {
		t.Error("expected sale_bill_uuid != 0, got 0")
	}
	if result.SaleOrderUuid == 0 {
		t.Error("expected sale_order_uuid != 0, got 0")
	}
}

// Test_Member_DineIn_Create_PackageAndNormal_Success 验证混合下单（普通商品+套餐）成功
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_PackageAndNormal_Success(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 创建普通商品
	normalBomUUID := fixture.SeedProductWithFlavor(t, db, "单品薯条", 15.00)

	// 6. 创建套餐商品
	pkgResult := fixture.SeedPackageProductWithSubItems(t, db, "超值套餐B", 45.00, []fixture.PackageSubItem{
		{Name: "汉堡", Price: 30.00},
		{Name: "饮料", Price: 10.00},
	})

	// 7. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 8. 构造混合请求：普通商品 + 套餐
	subProducts := make([]map[string]any, 0, len(pkgResult.SubBomUUIDs))
	for _, subBomUUID := range pkgResult.SubBomUUIDs {
		subProducts = append(subProducts, map[string]any{
			"flavor_uuid":                subBomUUID,
			"num":                        1,
			"product_package_group_uuid": pkgResult.GroupUUID,
		})
	}
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  normalBomUUID,
				"num":          2,
				"price":        15.00,
				"product_type": 0,
			},
			{
				"flavor_uuid":  pkgResult.PackageBomUUID,
				"num":          1,
				"price":        45.00,
				"product_type": 1,
				"products":     subProducts,
			},
		},
	}

	// 9. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 10. 断言 HTTP 200
	resp.AssertOK(t)

	// 11. 解析响应
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code != 0 {
		t.Fatalf("expected success code 0 but got %d: %s (body: %s)", apiResp.Code, apiResp.Message, resp.String())
	}

	// 12. 验证返回的 sale_bill_uuid 和 sale_order_uuid 不为 0
	var result struct {
		SaleBillUuid  uint64 `json:"sale_bill_uuid"`
		SaleOrderUuid uint64 `json:"sale_order_uuid"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if result.SaleBillUuid == 0 {
		t.Error("expected sale_bill_uuid != 0, got 0")
	}
	if result.SaleOrderUuid == 0 {
		t.Error("expected sale_order_uuid != 0, got 0")
	}
}

// Test_Member_DineIn_Create_PackageEmptySub_Fail 验证套餐无子商品时创建失败
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_PackageEmptySub_Fail(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 创建套餐（但请求时不传子商品）
	pkgResult := fixture.SeedPackageProductWithSubItems(t, db, "空套餐C", 30.00, []fixture.PackageSubItem{
		{Name: "子商品1", Price: 15.00},
	})

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 构造请求：套餐但不传子商品列表
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  pkgResult.PackageBomUUID,
				"num":          1,
				"price":        30.00,
				"product_type": 1,
				"products":     []map[string]any{},
			},
		},
	}

	// 8. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 9. 断言 HTTP 200
	resp.AssertOK(t)

	// 10. 验证业务错误码 != 0（套餐子商品不能为空）
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code == 0 {
		t.Error("expected error for package with empty sub-products, but got success")
	}
}

// Test_Member_DineIn_Create_IsAcceptOrder 验证普通商品堂食订单创建后 is_accept_order=0
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_IsAcceptOrder(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 创建商品
	bomUUID := fixture.SeedProductWithFlavor(t, db, "测试烤鸡", 20.00)

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 构造请求
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  bomUUID,
				"num":          1,
				"price":        20.00,
				"product_type": 0,
			},
		},
	}

	// 8. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code != 0 {
		t.Fatalf("expected success code 0 but got %d: %s", apiResp.Code, apiResp.Message)
	}

	var result struct {
		SaleBillUuid  uint64 `json:"sale_bill_uuid"`
		SaleOrderUuid uint64 `json:"sale_order_uuid"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}

	// 9. 查询 sale_order_product 表，验证 is_accept_order=0
	rows, err := db.Query(
		"SELECT is_accept_order FROM ttpos_sale_order_product WHERE sale_bill_uuid = ? AND delete_time = 0",
		result.SaleBillUuid,
	)
	if err != nil {
		t.Fatalf("failed to query sale_order_product: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var isAcceptOrder int
		if err := rows.Scan(&isAcceptOrder); err != nil {
			t.Fatalf("failed to scan is_accept_order: %v", err)
		}
		count++
		if isAcceptOrder != 0 {
			t.Errorf("sale_order_product[%d]: expected is_accept_order=0, got %d", count, isAcceptOrder)
		}
	}
	if count == 0 {
		t.Fatal("no sale_order_product records found")
	}
	t.Logf("verified %d sale_order_product records all have is_accept_order=0", count)
}

// Test_Member_DineIn_Create_Package_IsAcceptOrder 验证套餐商品堂食订单创建后 is_accept_order=0（含子商品）
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_Package_IsAcceptOrder(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入业务设置：全天营业
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 5. 创建套餐商品（1个主商品 + 2个子商品）
	pkgResult := fixture.SeedPackageProductWithSubItems(t, db, "接单验证套餐", 40.00, []fixture.PackageSubItem{
		{Name: "套餐主食", Price: 25.00},
		{Name: "套餐饮品", Price: 10.00},
	})

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 构造套餐请求
	subProducts := make([]map[string]any, 0, len(pkgResult.SubBomUUIDs))
	for _, subBomUUID := range pkgResult.SubBomUUIDs {
		subProducts = append(subProducts, map[string]any{
			"flavor_uuid":                subBomUUID,
			"num":                        1,
			"product_package_group_uuid": pkgResult.GroupUUID,
		})
	}
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  pkgResult.PackageBomUUID,
				"num":          1,
				"price":        40.00,
				"product_type": 1,
				"products":     subProducts,
			},
		},
	}

	// 8. 调用 POST /api/v1/member/order/dine_in/create
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)
	resp.AssertOK(t)

	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code != 0 {
		t.Fatalf("expected success code 0 but got %d: %s", apiResp.Code, apiResp.Message)
	}

	var result struct {
		SaleBillUuid  uint64 `json:"sale_bill_uuid"`
		SaleOrderUuid uint64 `json:"sale_order_uuid"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}

	// 9. 查询 sale_order_product 表，验证所有商品（主商品 + 子商品）的 is_accept_order=0
	rows, err := db.Query(
		"SELECT product_type, is_accept_order FROM ttpos_sale_order_product WHERE sale_bill_uuid = ? AND delete_time = 0",
		result.SaleBillUuid,
	)
	if err != nil {
		t.Fatalf("failed to query sale_order_product: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var productType, isAcceptOrder int
		if err := rows.Scan(&productType, &isAcceptOrder); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		count++
		if isAcceptOrder != 0 {
			t.Errorf("sale_order_product[%d] (product_type=%d): expected is_accept_order=0, got %d", count, productType, isAcceptOrder)
		}
	}
	// 套餐应有3条记录：1个主商品(product_type=1) + 2个子商品(product_type=2)
	if count < 3 {
		t.Errorf("expected at least 3 sale_order_product records (1 package + 2 sub-products), got %d", count)
	}
	t.Logf("verified %d sale_order_product records all have is_accept_order=0", count)
}

// Test_Member_DineIn_Create_Unauthorized 验证未登录时创建堂食订单被拒绝
// Route: POST /api/v1/member/order/dine_in/create
func Test_Member_DineIn_Create_Unauthorized(t *testing.T) {
	// 不传 token 直接调用
	reqBody := map[string]any{
		"sale_bill_uuid":  0,
		"sale_order_uuid": 0,
		"products": []map[string]any{
			{
				"flavor_uuid":  1,
				"num":          1,
				"price":        10.00,
				"product_type": 0,
			},
		},
	}

	httpClient := fixture.NewHTTPClient()
	resp := httpClient.Post(t, "/api/v1/member/order/dine_in/create", reqBody)

	// 未授权：服务端返回 HTTP 200 + 业务错误码 -102（Token 无效）
	resp.AssertOK(t)
	apiResp := resp.ParseAPIResponse(t)
	if apiResp.Code != -102 {
		t.Errorf("expected error code -102 for unauthorized, got %d: %s", apiResp.Code, apiResp.Message)
	}
}
