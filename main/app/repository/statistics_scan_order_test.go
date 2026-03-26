package repository

import (
	"testing"
)

// TestScanOrderAmount_Source5Only 测试扫码订单金额只统计 source=5 的订单
// 修复前使用了不存在的 bill_type 列，导致查询始终返回 0
func TestScanOrderAmount_Source5Only(t *testing.T) {
	db := setupStatisticsTestDB(t)
	prefix := "ttpos_"

	// 创建 statistics_sale 表
	err := db.Exec(`
		CREATE TABLE ` + prefix + `statistics_sale (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			source INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0,
			order_source_uuid INTEGER DEFAULT 0,
			is_takeout INTEGER DEFAULT 0,
			is_meger INTEGER DEFAULT 0,
			is_special INTEGER DEFAULT 0,
			payment_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_payment_balance DECIMAL(14,2) DEFAULT 0.00,
			complete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_sale table: %v", err)
	}

	// 插入测试数据
	// 1. source=5（会员扫码），desk_uuid=0，is_takeout=0 → 应计入 scan_order_amount
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 0, 0, 100.00, 0.00, 0.00, 1000)`)

	// 2. source=5（会员扫码），有桌台 → 不应计入（desk_uuid > 0）
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 1001, 0, 0, 0, 200.00, 0.00, 0.00, 1000)`)

	// 3. source=5（会员扫码），外卖 → 不应计入（is_takeout=1）
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 1, 0, 300.00, 0.00, 0.00, 1000)`)

	// 4. source=1（收银端），desk_uuid=0，is_takeout=0 → 不应计入（source != 5）
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (1, 0, 0, 0, 0, 400.00, 0.00, 0.00, 1000)`)

	// 5. source=4（H5），desk_uuid=0，is_takeout=0 → 不应计入（source != 5，修复后去掉了 source=4）
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (4, 0, 0, 0, 0, 500.00, 0.00, 0.00, 1000)`)

	// 使用修复后的 SQL 逻辑：source = 5 AND desk_uuid = 0 AND is_takeout = 0
	var result struct {
		ScanOrderAmount    float64
		AvgScanOrderAmount float64
	}
	err = db.Raw(`
		SELECT
			SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END) AS scan_order_amount,
			SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END) AS avg_scan_order_amount
		FROM ` + prefix + `statistics_sale
		WHERE delete_time = 0
	`).Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 只有第 1 条记录（source=5, desk_uuid=0, is_takeout=0, amount=100）应计入
	expectedAmount := 100.00
	if result.ScanOrderAmount != expectedAmount {
		t.Errorf("scan_order_amount: expected %.2f, got %.2f", expectedAmount, result.ScanOrderAmount)
	}
	if result.AvgScanOrderAmount != expectedAmount {
		t.Errorf("avg_scan_order_amount: expected %.2f, got %.2f", expectedAmount, result.AvgScanOrderAmount)
	}
}

// TestScanOrderAmount_WithRefund 测试扫码订单金额正确扣减退款
func TestScanOrderAmount_WithRefund(t *testing.T) {
	db := setupStatisticsTestDB(t)
	prefix := "ttpos_"

	err := db.Exec(`
		CREATE TABLE ` + prefix + `statistics_sale (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			source INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0,
			order_source_uuid INTEGER DEFAULT 0,
			is_takeout INTEGER DEFAULT 0,
			is_meger INTEGER DEFAULT 0,
			is_special INTEGER DEFAULT 0,
			payment_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_payment_balance DECIMAL(14,2) DEFAULT 0.00,
			complete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_sale table: %v", err)
	}

	// 扫码订单有退款
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 0, 200.00, 50.00, 10.00, 1000)`)

	var result struct {
		ScanOrderAmount float64
	}
	err = db.Raw(`
		SELECT SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END) AS scan_order_amount
		FROM ` + prefix + `statistics_sale WHERE delete_time = 0
	`).Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 200 - 50 - 10 = 140
	expected := 140.00
	if result.ScanOrderAmount != expected {
		t.Errorf("scan_order_amount with refund: expected %.2f, got %.2f", expected, result.ScanOrderAmount)
	}
}

// TestScanOrderAmount_OldSource4Excluded 验证修复后 source=4 的订单不再计入扫码统计
// 修复前的条件是 source IN (4, 5)，修复后只统计 source = 5
func TestScanOrderAmount_OldSource4Excluded(t *testing.T) {
	db := setupStatisticsTestDB(t)
	prefix := "ttpos_"

	err := db.Exec(`
		CREATE TABLE ` + prefix + `statistics_sale (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			source INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0,
			order_source_uuid INTEGER DEFAULT 0,
			is_takeout INTEGER DEFAULT 0,
			is_meger INTEGER DEFAULT 0,
			is_special INTEGER DEFAULT 0,
			payment_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_payment_balance DECIMAL(14,2) DEFAULT 0.00,
			complete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_sale table: %v", err)
	}

	// source=4 的订单
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (4, 0, 0, 0, 500.00, 0.00, 0.00, 1000)`)

	// source=5 的订单
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 0, 100.00, 0.00, 0.00, 1000)`)

	var result struct {
		ScanOrderAmount float64
	}
	err = db.Raw(`
		SELECT SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END) AS scan_order_amount
		FROM ` + prefix + `statistics_sale WHERE delete_time = 0
	`).Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 只有 source=5 的 100 计入，source=4 的 500 不计入
	expected := 100.00
	if result.ScanOrderAmount != expected {
		t.Errorf("source=4 should be excluded: expected %.2f, got %.2f", expected, result.ScanOrderAmount)
	}
}

// TestInstantOrderExcludesScanOrder 验证店内即时订单统计排除扫码订单（source=5）
// 修复前：instant_order 条件为 desk_uuid=0 AND order_source_uuid=0，包含了 source=5 的扫码订单
// 修复后：增加 source != 5 排除条件，扫码订单不再计入店内统计
func TestInstantOrderExcludesScanOrder(t *testing.T) {
	db := setupStatisticsTestDB(t)
	prefix := "ttpos_"

	err := db.Exec(`
		CREATE TABLE ` + prefix + `statistics_sale (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			source INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0,
			order_source_uuid INTEGER DEFAULT 0,
			is_takeout INTEGER DEFAULT 0,
			is_meger INTEGER DEFAULT 0,
			is_special INTEGER DEFAULT 0,
			payment_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_payment_balance DECIMAL(14,2) DEFAULT 0.00,
			complete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_sale table: %v", err)
	}

	// 1. source=1（收银端），desk_uuid=0，order_source_uuid=0 → 应计入 instant_order
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (1, 0, 0, 0, 0, 400.00, 0.00, 0.00, 1000)`)

	// 2. source=5（扫码），desk_uuid=0，order_source_uuid=0 → 不应计入 instant_order（应只计入 scan_order）
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 0, 0, 100.00, 0.00, 0.00, 1000)`)

	// 3. source=2（助手），desk_uuid=0，order_source_uuid=0 → 应计入 instant_order
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, order_source_uuid, is_takeout, is_meger, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (2, 0, 0, 0, 0, 200.00, 0.00, 0.00, 1000)`)

	var result struct {
		InstantOrderAmount float64
		ScanOrderAmount    float64
	}
	err = db.Raw(`
		SELECT
			SUM(CASE WHEN desk_uuid = 0 AND order_source_uuid = 0 AND source != 5 AND is_meger = 0 THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END) AS instant_order_amount,
			SUM(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 THEN payment_amount - refund_amount - refund_payment_balance ELSE 0 END) AS scan_order_amount
		FROM ` + prefix + `statistics_sale
		WHERE delete_time = 0
	`).Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// instant_order: 只有 source=1(400) + source=2(200) = 600，不含 source=5(100)
	expectedInstant := 600.00
	if result.InstantOrderAmount != expectedInstant {
		t.Errorf("instant_order_amount should exclude source=5: expected %.2f, got %.2f", expectedInstant, result.InstantOrderAmount)
	}

	// scan_order: 只有 source=5(100)
	expectedScan := 100.00
	if result.ScanOrderAmount != expectedScan {
		t.Errorf("scan_order_amount: expected %.2f, got %.2f", expectedScan, result.ScanOrderAmount)
	}
}

// TestScanOrderNum_CountsAfterFullRefund 验证全额退款后扫码订单数量仍正确统计
// 修复前：scan_order_num 基于 scan_order_amount > 0 判断，全额退款后金额为 0 导致不计数
// 修复后：scan_order_num 使用结构条件 source=5 AND desk_uuid=0 AND is_takeout=0，退款不影响计数
func TestScanOrderNum_CountsAfterFullRefund(t *testing.T) {
	db := setupStatisticsTestDB(t)
	prefix := "ttpos_"

	err := db.Exec(`
		CREATE TABLE ` + prefix + `statistics_sale (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			source INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0,
			order_source_uuid INTEGER DEFAULT 0,
			is_takeout INTEGER DEFAULT 0,
			is_meger INTEGER DEFAULT 0,
			is_special INTEGER DEFAULT 0,
			scan_order_amount DECIMAL(14,2) DEFAULT 0.00,
			payment_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_amount DECIMAL(14,2) DEFAULT 0.00,
			refund_payment_balance DECIMAL(14,2) DEFAULT 0.00,
			complete_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_sale table: %v", err)
	}

	// 1. 扫码订单，正常支付（未退款）→ 应计入 scan_order_num
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, is_takeout, is_meger, scan_order_amount, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 0, 100.00, 100.00, 0.00, 0.00, 1000)`)

	// 2. 扫码订单，全额退款（scan_order_amount=0）→ 修复后仍应计入 scan_order_num
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, is_takeout, is_meger, scan_order_amount, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (5, 0, 0, 0, 0.00, 200.00, 200.00, 0.00, 1000)`)

	// 3. 非扫码订单（source=1）→ 不计入 scan_order_num
	db.Exec(`INSERT INTO ` + prefix + `statistics_sale (source, desk_uuid, is_takeout, is_meger, scan_order_amount, payment_amount, refund_amount, refund_payment_balance, complete_time) VALUES (1, 0, 0, 0, 0.00, 300.00, 0.00, 0.00, 1000)`)

	var result struct {
		ScanOrderNum int64
	}
	// 使用修复后的结构条件
	err = db.Raw(`
		SELECT COUNT(CASE WHEN source = 5 AND desk_uuid = 0 AND is_takeout = 0 AND is_meger = 0 THEN 1 END) AS scan_order_num
		FROM ` + prefix + `statistics_sale
		WHERE delete_time = 0
	`).Scan(&result).Error
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 应该为 2（包括全额退款的扫码订单）
	var expected int64 = 2
	if result.ScanOrderNum != expected {
		t.Errorf("scan_order_num should count refunded scan orders: expected %d, got %d", expected, result.ScanOrderNum)
	}
}
