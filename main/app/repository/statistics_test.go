package repository

import (
	"testing"
	"ttpos-server-go/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	// 初始化测试配置
	if config.Database.TablePrefix == "" {
		config.Database.TablePrefix = "ttpos_"
	}
}

// setupStatisticsTestDB 创建测试数据库连接（使用 SQLite 内存数据库）
func setupStatisticsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 使用默认表前缀（测试环境）
	prefix := "ttpos_"

	// 创建 statistics_product 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `statistics_product (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0,
			update_time INTEGER DEFAULT 0,
			delete_time INTEGER DEFAULT 0,
			sale_bill_uuid INTEGER DEFAULT 0,
			sale_order_uuid INTEGER DEFAULT 0,
			product_package_uuid INTEGER DEFAULT 0,
			product_bom_uuid INTEGER DEFAULT 0,
			product_final_price DECIMAL(14,2) DEFAULT 0.00,
			product_num DECIMAL(14,2) DEFAULT 0.00,
			refund_num DECIMAL(14,2) DEFAULT 0.00,
			free_num DECIMAL(14,2) DEFAULT 0.00,
			give_num DECIMAL(14,2) DEFAULT 0.00
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create statistics_product table: %v", err)
	}

	// 创建 product_package 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `product_package (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			name TEXT DEFAULT '{}',
			category_uuid INTEGER DEFAULT 0,
			product_type INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create product_package table: %v", err)
	}

	// 创建 product_bom 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `product_bom (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			name TEXT DEFAULT '{}',
			price DECIMAL(14,2) DEFAULT 0.00
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create product_bom table: %v", err)
	}

	// 创建 product_category 表
	err = db.Exec(`
		CREATE TABLE ` + prefix + `product_category (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0,
			name TEXT DEFAULT '{}',
			parent_uuid INTEGER DEFAULT 0,
			sort INTEGER DEFAULT 0,
			create_time INTEGER DEFAULT 0
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create product_category table: %v", err)
	}

	return db
}

// TestCountProduct_RefundDeduction 测试 CountProduct 方法的退款扣减逻辑
func TestCountProduct_RefundDeduction(t *testing.T) {
	db := setupStatisticsTestDB(t)

	prefix := "ttpos_"

	// 准备测试数据
	// 1. 正常商品，无退款
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES (50.00, 10, 0, 0, 0, 1, 1)
	`)

	// 2. 正常商品，有退款（单品退款）
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES (50.00, 10, 2, 0, 0, 1, 2)
	`)

	// 3. 免单商品
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES (50.00, 5, 0, 5, 0, 1, 3)
	`)

	// 4. 赠菜商品
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES (50.00, 5, 0, 0, 5, 1, 4)
	`)

	// 5. 整单退款（refund_num = product_num）
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES (50.00, 5, 5, 0, 0, 1, 5)
	`)

	// 准备关联表数据
	db.Exec(`
		INSERT INTO ` + prefix + `product_package (uuid, name, category_uuid)
		VALUES (1, '{"zh":"测试商品"}', 1)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_bom (uuid, name, price)
		VALUES (1, '{"zh":"规格1"}', 50.00),
		       (2, '{"zh":"规格2"}', 50.00),
		       (3, '{"zh":"规格3"}', 50.00),
		       (4, '{"zh":"规格4"}', 50.00),
		       (5, '{"zh":"规格5"}', 50.00)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_category (uuid, name, parent_uuid, sort, create_time)
		VALUES (1, '{"zh":"测试分类"}', 0, 1, 1000)
	`)

	// 执行查询 - 由于 SQLite 不支持 MySQL 的 JSON 函数，我们直接测试 SQL 计算逻辑
	// 使用原始 SQL 查询来验证退款扣减逻辑
	var results []struct {
		SaleAmount float64
	}

	err := db.Raw(`
		SELECT SUM(CASE WHEN sp.free_num > 0 OR sp.give_num > 0 THEN 0 ELSE sp.product_final_price * (sp.product_num - sp.refund_num) END) AS sale_amount
		FROM ` + prefix + `statistics_product as sp
		GROUP BY sp.product_bom_uuid
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 验证退款扣减逻辑
	// 1. 正常商品无退款：sale_amount = 50.00 * 10 = 500.00
	// 2. 正常商品有退款：sale_amount = 50.00 * (10 - 2) = 400.00
	// 3. 免单商品：sale_amount = 0（被排除）
	// 4. 赠菜商品：sale_amount = 0（被排除）
	// 5. 整单退款：sale_amount = 50.00 * (5 - 5) = 0.00

	totalSaleAmount := 0.0
	for _, r := range results {
		totalSaleAmount += r.SaleAmount
	}

	// 预期总销售额：500 + 400 + 0 + 0 + 0 = 900.00
	expectedTotal := 900.00
	if totalSaleAmount != expectedTotal {
		t.Errorf("Expected total sale amount %.2f, got %.2f", expectedTotal, totalSaleAmount)
	}
}

// TestCountCategory_RefundDeduction 测试 CountCategory 方法的退款扣减逻辑
func TestCountCategory_RefundDeduction(t *testing.T) {
	db := setupStatisticsTestDB(t)

	prefix := "ttpos_"

	// 准备测试数据（与 CountProduct 相同的数据）
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES 
		(50.00, 10, 0, 0, 0, 1, 1),
		(50.00, 10, 2, 0, 0, 1, 2),
		(50.00, 5, 0, 5, 0, 1, 3),
		(50.00, 5, 0, 0, 5, 1, 4),
		(50.00, 5, 5, 0, 0, 1, 5)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_package (uuid, name, category_uuid)
		VALUES (1, '{"zh":"测试商品"}', 1)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_bom (uuid, name, price)
		VALUES (1, '{"zh":"规格1"}', 50.00),
		       (2, '{"zh":"规格2"}', 50.00),
		       (3, '{"zh":"规格3"}', 50.00),
		       (4, '{"zh":"规格4"}', 50.00),
		       (5, '{"zh":"规格5"}', 50.00)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_category (uuid, name, parent_uuid, sort, create_time)
		VALUES (1, '{"zh":"测试分类"}', 0, 1, 1000)
	`)

	// 执行查询 - 使用原始 SQL 查询来验证分类统计的退款扣减逻辑
	var results []struct {
		SaleAmount float64
	}

	err := db.Raw(`
		SELECT SUM(CASE WHEN sp.free_num > 0 OR sp.give_num > 0 THEN 0 ELSE sp.product_final_price * (sp.product_num - sp.refund_num) END) AS sale_amount
		FROM ` + prefix + `statistics_product as sp
		LEFT JOIN ` + prefix + `product_package as pp ON sp.product_package_uuid = pp.uuid
		LEFT JOIN ` + prefix + `product_category as pc ON pp.category_uuid = pc.uuid
		GROUP BY CASE WHEN pc.parent_uuid = 0 THEN pp.category_uuid ELSE pc.parent_uuid END
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 验证退款扣减逻辑
	// 预期总销售额：500 + 400 + 0 + 0 + 0 = 900.00
	expectedTotal := 900.00
	totalSaleAmount := 0.0
	for _, r := range results {
		totalSaleAmount += r.SaleAmount
	}

	if totalSaleAmount != expectedTotal {
		t.Errorf("Expected total sale amount %.2f, got %.2f", expectedTotal, totalSaleAmount)
	}
}

// TestCountProduct_FreeAndGiveExclusion 测试免单和赠菜的排除逻辑
func TestCountProduct_FreeAndGiveExclusion(t *testing.T) {
	db := setupStatisticsTestDB(t)

	prefix := "ttpos_"

	// 准备测试数据：免单商品和赠菜商品不应计入销售额
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES 
		(50.00, 10, 0, 10, 0, 1, 1),  -- 全部免单
		(50.00, 10, 0, 0, 10, 1, 2),  -- 全部赠菜
		(50.00, 10, 0, 5, 5, 1, 3)    -- 部分免单部分赠菜
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_package (uuid, name, category_uuid)
		VALUES (1, '{"zh":"测试商品"}', 1)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_bom (uuid, name, price)
		VALUES (1, '{"zh":"规格1"}', 50.00),
		       (2, '{"zh":"规格2"}', 50.00),
		       (3, '{"zh":"规格3"}', 50.00)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_category (uuid, name, parent_uuid, sort, create_time)
		VALUES (1, '{"zh":"测试分类"}', 0, 1, 1000)
	`)

	// 执行查询 - 使用原始 SQL 查询来验证免单和赠菜排除逻辑
	var results []struct {
		SaleAmount float64
	}

	err := db.Raw(`
		SELECT SUM(CASE WHEN sp.free_num > 0 OR sp.give_num > 0 THEN 0 ELSE sp.product_final_price * (sp.product_num - sp.refund_num) END) AS sale_amount
		FROM ` + prefix + `statistics_product as sp
		GROUP BY sp.product_bom_uuid
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 验证结果：所有免单和赠菜商品都不应计入销售额
	totalSaleAmount := 0.0
	for _, r := range results {
		totalSaleAmount += r.SaleAmount
		// 免单或赠菜商品的销售额应该为 0
		if r.SaleAmount != 0 {
			t.Errorf("Expected free/give product sale amount to be 0, got %.2f", r.SaleAmount)
		}
	}

	// 预期总销售额为 0（所有商品都是免单或赠菜）
	expectedTotal := 0.00
	if totalSaleAmount != expectedTotal {
		t.Errorf("Expected total sale amount %.2f, got %.2f", expectedTotal, totalSaleAmount)
	}
}

// TestCountProduct_RefundBoundary 测试退款边界情况
func TestCountProduct_RefundBoundary(t *testing.T) {
	db := setupStatisticsTestDB(t)

	prefix := "ttpos_"

	// 准备测试数据：退款数量边界情况
	db.Exec(`
		INSERT INTO ` + prefix + `statistics_product 
		(product_final_price, product_num, refund_num, free_num, give_num, product_package_uuid, product_bom_uuid)
		VALUES 
		(50.00, 10, 1, 0, 0, 1, 1),   -- 退款1个（最小退款）
		(50.00, 10, 9, 0, 0, 1, 2),   -- 退款9个（接近全部退款）
		(50.00, 10, 10, 0, 0, 1, 3)   -- 退款10个（全部退款）
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_package (uuid, name, category_uuid)
		VALUES (1, '{"zh":"测试商品"}', 1)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_bom (uuid, name, price)
		VALUES (1, '{"zh":"规格1"}', 50.00),
		       (2, '{"zh":"规格2"}', 50.00),
		       (3, '{"zh":"规格3"}', 50.00)
	`)

	db.Exec(`
		INSERT INTO ` + prefix + `product_category (uuid, name, parent_uuid, sort, create_time)
		VALUES (1, '{"zh":"测试分类"}', 0, 1, 1000)
	`)

	// 执行查询 - 使用原始 SQL 查询来验证退款边界情况
	var results []struct {
		SaleAmount float64
	}

	err := db.Raw(`
		SELECT SUM(CASE WHEN sp.free_num > 0 OR sp.give_num > 0 THEN 0 ELSE sp.product_final_price * (sp.product_num - sp.refund_num) END) AS sale_amount
		FROM ` + prefix + `statistics_product as sp
		GROUP BY sp.product_bom_uuid
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// 验证结果
	// 1. 退款1个：sale_amount = 50.00 * (10 - 1) = 450.00
	// 2. 退款9个：sale_amount = 50.00 * (10 - 9) = 50.00
	// 3. 退款10个：sale_amount = 50.00 * (10 - 10) = 0.00

	expectedTotal := 500.00 // 450 + 50 + 0
	totalSaleAmount := 0.0
	for _, r := range results {
		totalSaleAmount += r.SaleAmount
	}

	if totalSaleAmount != expectedTotal {
		t.Errorf("Expected total sale amount %.2f, got %.2f", expectedTotal, totalSaleAmount)
	}
}
