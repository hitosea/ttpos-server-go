//go:build integration

package data_manage_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

const (
	pathSetDataManage      = "/api/v1/shop/setting/data_manage/set"
	pathOrderSelectStats   = "/api/v1/shop/setting/data_manage/order_select_stats"
	testVersionHeaderKey   = "Client-Version"
	testVersionHeaderValue = "99.0.0"
)

type orderSelectStatsData struct {
	SelectedCount int64    `json:"selected_count"`
	PaidAmount    float64  `json:"paid_amount"`
	TotalCount    int64    `json:"total_count"`
	IsSelectAll   bool     `json:"is_select_all"`
	SelectedUuids []uint64 `json:"selected_uuids"`
}

// Test_P1_Shop_DataManage_OrderSelectStats_FilterOnly_AppliesDateFilter
// 验证 select_all=false 且 selected/deselected 为空时，仍会按筛选条件统计。
func Test_P1_Shop_DataManage_OrderSelectStats_FilterOnly_AppliesDateFilter(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	// 命中范围：2026-03-01 00:00:00 ~ 2026-03-03 23:59:59
	inRangeTime := mustParseDateTimeUnix("2026-03-02 12:00:00")
	outRangeTime := mustParseDateTimeUnix("2026-03-05 12:00:00")

	inRangeBillUUID := fixture.GenerateSnowflakeID()
	outRangeBillUUID := fixture.GenerateSnowflakeID()
	seedSaleBillForDataManageStats(t, db, inRangeBillUUID, inRangeTime, 100)
	seedSaleBillForDataManageStats(t, db, outRangeBillUUID, outRangeTime, 200)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().
		WithToken(token).
		WithHeaders(map[string]string{testVersionHeaderKey: testVersionHeaderValue})

	// 先持久化选中两条订单
	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{uint64(inRangeBillUUID), uint64(outRangeBillUUID)},
	}).AssertOK(t).AssertSuccess(t)

	resp := httpClient.Post(t, pathOrderSelectStats, map[string]any{
		"filters": []map[string]any{{
			"bill_type":        -1,
			"date_type":        -1,
			"deselected_uuids": []uint64{},
			"order_no":         "",
			"query_end_date":   "2026-03-03 23:59:59",
			"query_start_date": "2026-03-01 00:00:00",
			"select_all":       false,
			"selected_uuids":   []uint64{},
		}},
	})
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data orderSelectStatsData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse response data: %v", err)
	}

	if data.SelectedCount != 1 {
		t.Fatalf("expected selected_count=1, got %d", data.SelectedCount)
	}
	if data.PaidAmount != 100 {
		t.Fatalf("expected paid_amount=100, got %v", data.PaidAmount)
	}
	if data.TotalCount != 1 {
		t.Fatalf("expected total_count=1, got %d", data.TotalCount)
	}
}

// Test_P1_Shop_DataManage_OrderSelectStats_NoFilter_DefaultLast7Days
// 验证 filters 为空时默认按近7天统计。
func Test_P1_Shop_DataManage_OrderSelectStats_NoFilter_DefaultLast7Days(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	now := time.Now().Unix()
	inRangeBillUUID := fixture.GenerateSnowflakeID()
	outRangeBillUUID := fixture.GenerateSnowflakeID()
	seedSaleBillForDataManageStats(t, db, inRangeBillUUID, now, 110)
	seedSaleBillForDataManageStats(t, db, outRangeBillUUID, now-15*24*3600, 220)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().
		WithToken(token).
		WithHeaders(map[string]string{testVersionHeaderKey: testVersionHeaderValue})

	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{uint64(inRangeBillUUID), uint64(outRangeBillUUID)},
	}).AssertOK(t).AssertSuccess(t)

	resp := httpClient.Post(t, pathOrderSelectStats, map[string]any{})
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data orderSelectStatsData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse response data: %v", err)
	}

	if data.SelectedCount != 1 {
		t.Fatalf("expected selected_count=1 (default last 7 days), got %d", data.SelectedCount)
	}
	if data.PaidAmount != 110 {
		t.Fatalf("expected paid_amount=110 (default last 7 days), got %v", data.PaidAmount)
	}
}

// Test_P1_Shop_DataManage_OrderSelectStats_MultiFilters
// 验证跨时间范围累积选择：多个 filters 合并统计。
func Test_P1_Shop_DataManage_OrderSelectStats_MultiFilters(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db,
		fixture.WithCompanySettingCompanyUUID(companyUUIDInt),
		fixture.WithCompanySettingEnableDataManagement(1),
	)
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)

	fixture.SetupWireMock(t)

	marchTime := mustParseDateTimeUnix("2026-03-15 12:00:00")
	febTime := mustParseDateTimeUnix("2026-02-15 12:00:00")
	marchBill1 := fixture.GenerateSnowflakeID()
	marchBill2 := fixture.GenerateSnowflakeID()
	febBill1 := fixture.GenerateSnowflakeID()
	febBill2 := fixture.GenerateSnowflakeID()
	seedSaleBillForDataManageStats(t, db, marchBill1, marchTime, 100)
	seedSaleBillForDataManageStats(t, db, marchBill2, marchTime, 200)
	seedSaleBillForDataManageStats(t, db, febBill1, febTime, 300)
	seedSaleBillForDataManageStats(t, db, febBill2, febTime, 400)

	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	httpClient := fixture.NewHTTPClient().
		WithToken(token).
		WithHeaders(map[string]string{testVersionHeaderKey: testVersionHeaderValue})

	httpClient.Post(t, pathSetDataManage, map[string]any{
		"is_enable_data_manage": true,
		"staff_uuids":           []uint64{},
		"sale_bill_uuids":       []uint64{},
	}).AssertOK(t).AssertSuccess(t)

	// 跨时间范围：3月选2单 + 2月选2单 → 共4单，金额 1000
	resp := httpClient.Post(t, pathOrderSelectStats, map[string]any{
		"filters": []map[string]any{
			{
				"select_all":       false,
				"selected_uuids":   []uint64{uint64(marchBill1), uint64(marchBill2)},
				"deselected_uuids": []uint64{},
				"date_type":        -1,
				"bill_type":        -1,
				"query_start_date": "2026-03-01 00:00:00",
				"query_end_date":   "2026-03-31 23:59:59",
			},
			{
				"select_all":       false,
				"selected_uuids":   []uint64{uint64(febBill1), uint64(febBill2)},
				"deselected_uuids": []uint64{},
				"date_type":        -1,
				"bill_type":        -1,
				"query_start_date": "2026-02-01 00:00:00",
				"query_end_date":   "2026-02-28 23:59:59",
			},
		},
	})
	resp.AssertOK(t).AssertSuccess(t)

	apiResp := resp.ParseAPIResponse(t)
	var data orderSelectStatsData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		t.Fatalf("failed to parse response data: %v", err)
	}
	if data.SelectedCount != 4 {
		t.Fatalf("expected selected_count=4, got %d", data.SelectedCount)
	}
	if data.PaidAmount != 1000 {
		t.Fatalf("expected paid_amount=1000, got %v", data.PaidAmount)
	}
	if len(data.SelectedUuids) != 4 {
		t.Fatalf("expected 4 selected_uuids, got %d", len(data.SelectedUuids))
	}
}

func seedSaleBillForDataManageStats(tb testing.TB, db *sql.DB, saleBillUUID int64, createTime int64, paymentAmount float64) {
	tb.Helper()
	_, err := db.Exec(`
		INSERT INTO ttpos_sale_bill
			(uuid, status, bill_type, production_time, payment_amount, create_time, update_time, delete_time)
		VALUES (?, 1, 0, ?, ?, ?, ?, 0)
	`, saleBillUUID, createTime, paymentAmount, createTime, createTime)
	if err != nil {
		tb.Fatalf("failed to seed sale bill: %v", err)
	}
}

func mustParseDateTimeUnix(s string) int64 {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.UTC)
	if err != nil {
		panic(err)
	}
	return t.Unix()
}

func mustParseInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid int64 %q: %v", s, err))
	}
	return v
}

func mustParseString(i int64) string {
	return strconv.FormatInt(i, 10)
}
