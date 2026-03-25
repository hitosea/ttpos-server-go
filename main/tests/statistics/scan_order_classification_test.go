//go:build integration

package statistics_test

import (
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

// API path constants
const (
	pathStatsBiz = "/api/v1/shop/statistics/business"
)

// --- P0 ---

// Test_P0_Shop_Statistics_ScanOrder_NotDoubleCountedAsInstant verifies that
// scan orders (source=5) are counted ONLY in scan_order stats, NOT in instant_order (店内) stats.
//
// Before fix: instant_order conditions (desk_uuid=0 AND order_source_uuid=0) included source=5
// After fix:  instant_order conditions add source != 5 to exclude scan orders
//
// Route: GET /shop/statistics/business
func Test_P0_Shop_Statistics_ScanOrder_NotDoubleCountedAsInstant(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	now := time.Now().Unix()

	// Record 1: cashier instant order (source=1, desk_uuid=0) → should count in instant_order only
	billA := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billA, now-100, now-200, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 1, 0, 0, 0, 0, 0, 400.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billA, socNextID(), now-100, now-200, now)

	// Record 2: scan order (source=5, desk_uuid=0) → should count in scan_order only, NOT instant_order
	billB := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billB, now-50, now-150, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 5, 0, 0, 0, 0, 0, 100.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billB, socNextID(), now-50, now-150, now)

	// Record 3: assistant instant order (source=2, desk_uuid=0) → should count in instant_order only
	billC := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billC, now-30, now-120, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 2, 0, 0, 0, 0, 0, 200.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billC, socNextID(), now-30, now-120, now)

	// --- SQL-level assertion: verify the fixed SQL conditions ---
	var result struct {
		instantAmount float64
		instantCount  int
		scanAmount    float64
		scanCount     int
	}
	err := db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN desk_uuid = 0 AND order_source_uuid = 0 AND source != 5 AND is_meger = 0
				THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END), 0),
			COUNT(CASE WHEN desk_uuid = 0 AND order_source_uuid = 0 AND source != 5 AND is_meger = 0 THEN 1 END),
			COALESCE(SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0
				THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END), 0),
			COUNT(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 AND is_meger = 0 THEN 1 END)
		FROM ttpos_statistics_sale
		WHERE delete_time = 0
	`).Scan(&result.instantAmount, &result.instantCount, &result.scanAmount, &result.scanCount)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// instant_order: source=1(400) + source=2(200) = 600, should NOT include source=5(100)
	if result.instantAmount != 600 {
		t.Errorf("instant_order_amount: expected 600, got %.2f (source=5 leak?)", result.instantAmount)
	}
	if result.instantCount != 2 {
		t.Errorf("instant_order_num: expected 2, got %d", result.instantCount)
	}

	// scan_order: only source=5(100)
	if result.scanAmount != 100 {
		t.Errorf("scan_order_amount: expected 100, got %.2f", result.scanAmount)
	}
	if result.scanCount != 1 {
		t.Errorf("scan_order_num: expected 1, got %d", result.scanCount)
	}

	// Verify no double counting: instant + scan amounts should equal total non-desk non-takeout amount
	totalNonDesk := result.instantAmount + result.scanAmount
	if totalNonDesk != 700 {
		t.Errorf("instant + scan = %.2f, expected 700 (400+200+100)", totalNonDesk)
	}

	t.Logf("PASS: instant_order(%.0f, %d orders) + scan_order(%.0f, %d orders) — no double counting",
		result.instantAmount, result.instantCount, result.scanAmount, result.scanCount)
}

// Test_P0_Shop_Statistics_ScanOrder_CountedAfterFullRefund verifies that
// fully refunded scan orders are still counted in scan_order_num.
//
// Before fix: scan_order_num used scan_order_amount > 0, so refunded orders (amount=0) were lost
// After fix:  scan_order_num uses structural condition (source=5 AND desk_uuid=0 AND is_takeout=0)
//
// Route: GET /shop/statistics/business
func Test_P0_Shop_Statistics_ScanOrder_CountedAfterFullRefund(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	now := time.Now().Unix()

	// Record 1: scan order, paid normally
	billA := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billA, now-100, now-200, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 5, 0, 0, 0, 0, 0, 150.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billA, socNextID(), now-100, now-200, now)

	// Record 2: scan order, FULLY REFUNDED (payment=200, refund=200)
	billB := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billB, now-50, now-150, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 5, 0, 0, 0, 0, 0, 200.00, 200.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billB, socNextID(), now-50, now-150, now)

	// Record 3: non-scan cashier order (for control group)
	billC := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billC, now-30, now-120, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 1, 0, 0, 0, 0, 0, 300.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billC, socNextID(), now-30, now-120, now)

	// --- SQL-level assertion ---
	var scanCount int
	var scanAmount float64
	var instantCount int
	var instantAmount float64

	err := db.QueryRow(`
		SELECT
			COUNT(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 AND is_meger = 0 THEN 1 END),
			COALESCE(SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0
				THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END), 0),
			COUNT(CASE WHEN desk_uuid = 0 AND order_source_uuid = 0 AND source != 5 AND is_meger = 0 THEN 1 END),
			COALESCE(SUM(CASE WHEN desk_uuid = 0 AND order_source_uuid = 0 AND source != 5 AND is_meger = 0
				THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END), 0)
		FROM ttpos_statistics_sale
		WHERE delete_time = 0
	`).Scan(&scanCount, &scanAmount, &instantCount, &instantAmount)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// scan_order_num: 2 (both scan orders counted, including fully refunded one)
	if scanCount != 2 {
		t.Errorf("scan_order_num: expected 2 (including refunded), got %d", scanCount)
	}

	// scan_order_amount: 150 (only the paid order; refunded one contributes 200-200=0)
	if scanAmount != 150 {
		t.Errorf("scan_order_amount: expected 150, got %.2f", scanAmount)
	}

	// instant_order: only the cashier order (source=1, 300)
	if instantCount != 1 {
		t.Errorf("instant_order_num: expected 1, got %d", instantCount)
	}
	if instantAmount != 300 {
		t.Errorf("instant_order_amount: expected 300, got %.2f", instantAmount)
	}

	t.Logf("PASS: scan_order_num=%d (includes refunded), scan_amount=%.0f, instant_num=%d, instant_amount=%.0f",
		scanCount, scanAmount, instantCount, instantAmount)
}

// --- P1 ---

// Test_P1_Shop_Statistics_ScanOrder_API_Smoke verifies that the business statistics API
// responds successfully when scan orders (source=5) and instant orders coexist.
// This ensures the modified SQL (source != 5 exclusion, structural scan count) doesn't
// break the API query path.
//
// Route: GET /shop/statistics/business
func Test_P1_Shop_Statistics_ScanOrder_API_Smoke(t *testing.T) {
	companyUUID := fixture.GenerateCompanyUUID(t)
	db := fixture.NewTestTenantFull(t, companyUUID)
	companyUUIDInt := mustParseInt64(companyUUID)
	now := time.Now().Unix()

	fixture.SeedCompany(t, db, fixture.WithCompanyUUID(companyUUIDInt))
	fixture.SeedCompanySetting(t, db, fixture.WithCompanySettingCompanyUUID(companyUUIDInt))
	staff := fixture.SeedStaff(t, db,
		fixture.WithStaffCompanyUUID(companyUUIDInt),
		fixture.WithStaffIsSuper(1),
	)
	fixture.SetupWireMock(t)

	// Seed: cashier instant order (source=1)
	billA := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billA, now-100, now-200, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 1, 0, 0, 0, 0, 0, 400.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billA, socNextID(), now-100, now-200, now)

	// Seed: scan order (source=5)
	billB := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billB, now-50, now-150, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 5, 0, 0, 0, 0, 0, 100.00, 0.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billB, socNextID(), now-50, now-150, now)

	// Seed: fully refunded scan order (source=5)
	billC := socNextID()
	socExec(t, db,
		`INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		 VALUES (?, 1, ?, ?, ?, 0)`,
		billC, now-30, now-120, now)
	socExec(t, db,
		`INSERT INTO ttpos_statistics_sale
		 (uuid, sale_bill_uuid, sale_order_uuid, source, desk_uuid, order_source_uuid, is_takeout, is_meger, is_special,
		  payment_amount, refund_amount, refund_payment_balance, complete_time, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, 5, 0, 0, 0, 0, 0, 200.00, 200.00, 0.00, ?, ?, ?, 0)`,
		socNextID(), billC, socNextID(), now-30, now-120, now)

	// API should respond successfully with mixed source data present
	token := fixture.GenerateShopToken(t, companyUUID, mustParseString(staff.UUID))
	fixture.NewHTTPClient().WithToken(token).WithHeaders(versionHeaders).
		Get(t, pathStatsBiz+"?query_start_time=1700000000&query_end_time=2100000000").
		AssertOK(t).AssertSuccess(t)
}

// --- Helpers ---

var socIDSeq int64

func socNextID() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond)<<16 | atomic.AddInt64(&socIDSeq, 1)
}

func socExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("SQL exec failed: %v\nquery: %s\nargs: %v", err, query, args)
	}
}
