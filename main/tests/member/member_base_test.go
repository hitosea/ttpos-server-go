//go:build integration

package member_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// Test_Member_BaseInfo_IsStoreResting_Disabled 验证门店点餐未启用时 is_store_resting=true
// Route: GET /api/v1/member/base
func Test_Member_BaseInfo_IsStoreResting_Disabled(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入门店点餐设置：is_enabled=0（未启用）
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_enabled":0,"enable_delivery":0,"enable_self_pickup":0}`)

	// 5. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 6. 调用 GET /api/v1/member/base
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Get(t, "/api/v1/member/base")

	// 8. 断言成功
	resp.AssertOK(t).AssertSuccess(t)

	// 9. 验证 is_store_resting = true
	apiResp := resp.ParseAPIResponse(t)
	var result struct {
		Member struct {
			IsStoreResting bool `json:"is_store_resting"`
		} `json:"member"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if !result.Member.IsStoreResting {
		t.Errorf("expected is_store_resting=true when store scan order is disabled, got false")
	}
}

// Test_Member_BaseInfo_IsStoreResting_Enabled_AllDay 验证门店点餐已启用且全天营业时 is_store_resting=false
// Route: GET /api/v1/member/base
func Test_Member_BaseInfo_IsStoreResting_Enabled_AllDay(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入门店点餐设置：is_enabled=1（已启用）
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_enabled":1,"enable_delivery":0,"enable_self_pickup":0}`)

	// 5. 写入业务设置：全天营业 00:00-23:59
	fixture.SeedSetting(t, db, "business", `{"opening_hours":"00:00-23:59"}`)

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 调用 GET /api/v1/member/base
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Get(t, "/api/v1/member/base")

	// 9. 断言成功
	resp.AssertOK(t).AssertSuccess(t)

	// 10. 验证 is_store_resting = false
	apiResp := resp.ParseAPIResponse(t)
	var result struct {
		Member struct {
			IsStoreResting bool `json:"is_store_resting"`
		} `json:"member"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if result.Member.IsStoreResting {
		t.Errorf("expected is_store_resting=false when store is enabled and open all day, got true")
	}
}

// Test_Member_BaseInfo_IsStoreResting_Enabled_OutsideHours 验证门店点餐已启用但不在营业时间内时 is_store_resting=true
// Route: GET /api/v1/member/base
func Test_Member_BaseInfo_IsStoreResting_Enabled_OutsideHours(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置（默认时区 Asia/Shanghai）
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入门店点餐设置：is_enabled=1（已启用）
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_enabled":1,"enable_delivery":0,"enable_self_pickup":0}`)

	// 5. 写入业务设置：营业时间设为一个必然不包含当前时间的窗口
	//    取当前上海时间 +2h 到 +3h，确保当前时间一定不在这个窗口内
	openingHours := outsideOpeningHours()
	fixture.SeedSetting(t, db, "business", fmt.Sprintf(`{"opening_hours":"%s"}`, openingHours))

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 调用 GET /api/v1/member/base
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Get(t, "/api/v1/member/base")

	// 8. 断言成功
	resp.AssertOK(t).AssertSuccess(t)

	// 9. 验证 is_store_resting = true（不在营业时间内）
	apiResp := resp.ParseAPIResponse(t)
	var result struct {
		Member struct {
			IsStoreResting bool `json:"is_store_resting"`
		} `json:"member"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if !result.Member.IsStoreResting {
		t.Errorf("expected is_store_resting=true when outside opening hours (%s), got false", openingHours)
	}
}

// Test_Member_BaseInfo_IsStoreResting_Enabled_WithinHours 验证门店点餐已启用且在营业时间内时 is_store_resting=false
// Route: GET /api/v1/member/base
func Test_Member_BaseInfo_IsStoreResting_Enabled_WithinHours(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入门店点餐设置：is_enabled=1（已启用）
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_enabled":1,"enable_delivery":0,"enable_self_pickup":0}`)

	// 5. 写入业务设置：营业时间包含当前时间（当前时间 -1h 到 +1h）
	openingHours := withinOpeningHours()
	fixture.SeedSetting(t, db, "business", fmt.Sprintf(`{"opening_hours":"%s"}`, openingHours))

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 调用 GET /api/v1/member/base
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Get(t, "/api/v1/member/base")

	// 8. 断言成功
	resp.AssertOK(t).AssertSuccess(t)

	// 9. 验证 is_store_resting = false（在营业时间内）
	apiResp := resp.ParseAPIResponse(t)
	var result struct {
		Member struct {
			IsStoreResting bool `json:"is_store_resting"`
		} `json:"member"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if result.Member.IsStoreResting {
		t.Errorf("expected is_store_resting=false when within opening hours (%s), got true", openingHours)
	}
}

// Test_Member_BaseInfo_IsStoreResting_NoSetting 验证未设置门店点餐配置时 is_store_resting=true
// Route: GET /api/v1/member/base
func Test_Member_BaseInfo_IsStoreResting_NoSetting(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 不写入 store_scan_order 设置（默认 is_enabled=0）

	// 5. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 6. 调用 GET /api/v1/member/base
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Get(t, "/api/v1/member/base")

	// 7. 断言成功
	resp.AssertOK(t).AssertSuccess(t)

	// 8. 验证 is_store_resting = true（未配置 = 未启用 = 休息中）
	apiResp := resp.ParseAPIResponse(t)
	var result struct {
		Member struct {
			IsStoreResting bool `json:"is_store_resting"`
		} `json:"member"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if !result.Member.IsStoreResting {
		t.Errorf("expected is_store_resting=true when no store_scan_order setting exists, got false")
	}
}

// Test_Member_BaseInfo_IsStoreResting_Enabled_EmptyHours 验证门店点餐已启用但未设置营业时间（视为全天）时 is_store_resting=false
// Route: GET /api/v1/member/base
func Test_Member_BaseInfo_IsStoreResting_Enabled_EmptyHours(t *testing.T) {
	// 1. 创建租户数据库
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	// 2. 创建公司和公司设置
	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))

	// 3. 创建会员
	member := fixture.SeedMember(t, db)

	// 4. 写入门店点餐设置：is_enabled=1（已启用）
	fixture.SeedSetting(t, db, "store_scan_order", `{"is_enabled":1,"enable_delivery":0,"enable_self_pickup":0}`)

	// 5. 写入业务设置：opening_hours 为空（视为全天营业）
	fixture.SeedSetting(t, db, "business", `{"opening_hours":""}`)

	// 6. 生成会员 token
	token := fixture.GenerateMemberToken(t, companyUUID, fmt.Sprintf("%d", member.UUID))

	// 7. 调用 GET /api/v1/member/base
	httpClient := fixture.NewHTTPClient().WithToken(token)
	resp := httpClient.Get(t, "/api/v1/member/base")

	// 8. 断言成功
	resp.AssertOK(t).AssertSuccess(t)

	// 9. 验证 is_store_resting = false（空营业时间 = 全天营业）
	apiResp := resp.ParseAPIResponse(t)
	var result struct {
		Member struct {
			IsStoreResting bool `json:"is_store_resting"`
		} `json:"member"`
	}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
	if result.Member.IsStoreResting {
		t.Errorf("expected is_store_resting=false when opening_hours is empty (all day), got true")
	}
}

func mustParseInt64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("mustParseInt64(%q): %v", s, err))
	}
	return n
}

// shanghaiNow returns the current time in Asia/Shanghai timezone.
func shanghaiNow() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc)
}

// outsideOpeningHours returns an opening hours string (HH:MM-HH:MM) that does NOT
// contain the current Shanghai time. It picks a 1-hour window starting 2 hours from now.
func outsideOpeningHours() string {
	now := shanghaiNow()
	start := now.Add(2 * time.Hour)
	end := now.Add(3 * time.Hour)
	return fmt.Sprintf("%02d:%02d-%02d:%02d", start.Hour(), start.Minute(), end.Hour(), end.Minute())
}

// withinOpeningHours returns an opening hours string (HH:MM-HH:MM) that contains
// the current Shanghai time. It picks a 2-hour window centered around now (-1h to +1h).
func withinOpeningHours() string {
	now := shanghaiNow()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)
	return fmt.Sprintf("%02d:%02d-%02d:%02d", start.Hour(), start.Minute(), end.Hour(), end.Minute())
}
