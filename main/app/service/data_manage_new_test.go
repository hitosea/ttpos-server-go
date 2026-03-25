package service

import (
	"testing"
	"time"

	"ttpos-server-go/app/constant"
	shop_req "ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// setupDataManageTestDBWithOrders 创建包含订单相关表的测试数据库
func setupDataManageTestDBWithOrders(t *testing.T) *database.DBManager {
	t.Helper()
	dbm := setupDataManageTestDB(t)
	db := dbm.GetDB(constant.MockDB)

	// 手动创建表（SQLite 不支持 MySQL 的 unsigned 等类型，无法使用 AutoMigrate）
	for _, ddl := range []string{
		`DROP TABLE IF EXISTS ttpos_sale_bill`,
		`CREATE TABLE ttpos_sale_bill (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0, create_time INTEGER DEFAULT 0, update_time INTEGER DEFAULT 0, delete_time INTEGER DEFAULT 0,
			order_no TEXT DEFAULT '', duty_no TEXT DEFAULT '', serial_no TEXT DEFAULT '',
			status INTEGER DEFAULT 0, is_lock INTEGER DEFAULT 0, is_split_order INTEGER DEFAULT 0,
			is_order_first_pay_later INTEGER DEFAULT 0,
			bill_type INTEGER DEFAULT 0, dining_method INTEGER DEFAULT 0,
			order_source_uuid INTEGER DEFAULT 0, order_source_name TEXT DEFAULT '',
			nationality_uuid INTEGER DEFAULT 0, nationality_name TEXT DEFAULT '',
			source INTEGER DEFAULT 0, client_version TEXT DEFAULT '',
			is_buffet INTEGER DEFAULT 0, buffet_duration INTEGER DEFAULT 0, buffet_start_time INTEGER DEFAULT 0,
			delay_duration INTEGER DEFAULT 0, delay_start_time INTEGER DEFAULT 0,
			non_ordering_time INTEGER DEFAULT 0, reminder_order_time INTEGER DEFAULT 0,
			meal_num INTEGER DEFAULT 0, remark TEXT DEFAULT '', order_remark TEXT DEFAULT '', reason TEXT DEFAULT '',
			amount REAL DEFAULT 0, origin_amount REAL DEFAULT 0,
			product_amount REAL DEFAULT 0, product_original_amount REAL DEFAULT 0,
			payment_amount REAL DEFAULT 0, payment_commission_fee REAL DEFAULT 0,
			service_fee REAL DEFAULT 0, tax_fee REAL DEFAULT 0,
			custom_discount_fee REAL DEFAULT 0, member_discount_fee REAL DEFAULT 0,
			gift_amount REAL DEFAULT 0, activity_amount REAL DEFAULT 0, free_amount REAL DEFAULT 0,
			lock_time INTEGER DEFAULT 0, finish_time INTEGER DEFAULT 0,
			hide_bill_time INTEGER DEFAULT 0, production_time INTEGER DEFAULT 0, submit_pay_time INTEGER DEFAULT 0,
			cashier_name TEXT DEFAULT '', consumer_uuid INTEGER DEFAULT 0, cashier_uuid INTEGER DEFAULT 0,
			desk_uuid INTEGER DEFAULT 0, device_uuid INTEGER DEFAULT 0,
			buffet_package1_uuid INTEGER DEFAULT 0, buffet_package2_uuid INTEGER DEFAULT 0,
			buffet_package1_name TEXT DEFAULT '', buffet_package2_name TEXT DEFAULT '',
			member_sale_order_uuid INTEGER DEFAULT 0, batch_tag_uuid INTEGER DEFAULT 0,
			show_must_plan INTEGER DEFAULT 1, auto_add_must_product INTEGER DEFAULT 1,
			is_kitchen_confirm INTEGER DEFAULT 0, reverse_settle_count INTEGER DEFAULT 0
		)`,
		`DROP TABLE IF EXISTS ttpos_sale_order`,
		`CREATE TABLE ttpos_sale_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0, create_time INTEGER DEFAULT 0, update_time INTEGER DEFAULT 0, delete_time INTEGER DEFAULT 0,
			order_no TEXT DEFAULT '', status INTEGER DEFAULT 0,
			is_free INTEGER DEFAULT 0, free_reason TEXT DEFAULT '',
			cashier_name TEXT DEFAULT '', device_id TEXT DEFAULT '',
			consumer_uuid INTEGER DEFAULT 0, cashier_uuid INTEGER DEFAULT 0,
			sale_bill_uuid INTEGER DEFAULT 0, staff_shift_log_uuid INTEGER DEFAULT 0,
			product_amount REAL DEFAULT 0, product_original_amount REAL DEFAULT 0,
			service_fee REAL DEFAULT 0, tax_fee REAL DEFAULT 0,
			custom_discount_fee REAL DEFAULT 0, member_discount_fee REAL DEFAULT 0,
			origin_amount REAL DEFAULT 0, amount REAL DEFAULT 0, custom_amount REAL DEFAULT -1,
			finish_time INTEGER DEFAULT 0,
			member_discount_rate REAL DEFAULT 1, member_card_discount_rate REAL DEFAULT 1,
			custom_discount_rate REAL DEFAULT 1,
			zero_rule INTEGER DEFAULT 0, zero_fee REAL DEFAULT 0, zero_checkout_rule INTEGER DEFAULT 0,
			pay_points REAL DEFAULT 0, pay_points_amount REAL DEFAULT 0,
			points_exchange_rate REAL DEFAULT 0, auto_points_exchange INTEGER DEFAULT 0,
			full_reduction_activity_uuid INTEGER DEFAULT 0,
			coupon_amount REAL DEFAULT 0, activity_amount REAL DEFAULT 0,
			full_reduction_activity_message TEXT DEFAULT '',
			payment_amount REAL DEFAULT 0, change_amount REAL DEFAULT 0,
			zero_checkout_fee REAL DEFAULT 0, final_price REAL DEFAULT 0,
			payment_commission_fee REAL DEFAULT 0,
			gift_amount REAL DEFAULT 0, gift_points REAL DEFAULT 0,
			gift_points_rate REAL DEFAULT 0, gift_points_type INTEGER DEFAULT 0,
			member_level_name TEXT DEFAULT '', member_balance REAL DEFAULT 0,
			unit TEXT DEFAULT '0',
			erp_discount_amount REAL DEFAULT 0,
			erp_sales_invoice_name TEXT DEFAULT '',
			erp_payment_entry_names TEXT DEFAULT '',
			erp_sync_status INTEGER DEFAULT 0,
			erp_reverse_count INTEGER DEFAULT 0,
			erp_stock_deducted INTEGER DEFAULT 0
		)`,
		`DROP TABLE IF EXISTS ttpos_payment_order`,
		`CREATE TABLE ttpos_payment_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0, create_time INTEGER DEFAULT 0, update_time INTEGER DEFAULT 0, delete_time INTEGER DEFAULT 0,
			payment_method_name TEXT DEFAULT '', payment_method_uuid INTEGER DEFAULT 0,
			payment_fee_percent REAL DEFAULT 0, related_type INTEGER DEFAULT 0, related_uuid INTEGER DEFAULT 0,
			currency_unit TEXT DEFAULT '', payment_amount REAL DEFAULT 0,
			payment_commission_fee REAL DEFAULT 0, amount REAL DEFAULT 0,
			transaction_number TEXT DEFAULT '', status INTEGER DEFAULT 0,
			status_reason TEXT DEFAULT '', payment_info TEXT DEFAULT '', cancel_info TEXT DEFAULT '',
			balance_amount REAL DEFAULT 0, gift_balance_amount REAL DEFAULT 0
		)`,
		`DROP TABLE IF EXISTS ttpos_payment_method`,
		`CREATE TABLE ttpos_payment_method (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0, create_time INTEGER DEFAULT 0, update_time INTEGER DEFAULT 0, delete_time INTEGER DEFAULT 0,
			name TEXT DEFAULT '', code INTEGER DEFAULT 0, payment_name TEXT DEFAULT '', source INTEGER DEFAULT 0
		)`,
		`DROP TABLE IF EXISTS ttpos_return_order`,
		`CREATE TABLE ttpos_return_order (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0, create_time INTEGER DEFAULT 0, update_time INTEGER DEFAULT 0, delete_time INTEGER DEFAULT 0,
			related_order_type INTEGER DEFAULT 0, related_order_uuid INTEGER DEFAULT 0,
			related_order_no TEXT DEFAULT '', is_reverse_settlement INTEGER DEFAULT 0,
			return_type INTEGER DEFAULT 0, refund_amount REAL DEFAULT 0,
			unit TEXT DEFAULT '', refund_tax_amount REAL DEFAULT 0, refund_reason TEXT DEFAULT ''
		)`,
		`DROP TABLE IF EXISTS ttpos_member`,
		`CREATE TABLE ttpos_member (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid INTEGER DEFAULT 0, create_time INTEGER DEFAULT 0, update_time INTEGER DEFAULT 0, delete_time INTEGER DEFAULT 0,
			member_no TEXT DEFAULT '', nickname TEXT DEFAULT '', gender INTEGER DEFAULT 0,
			phone TEXT DEFAULT '', is_visitor INTEGER DEFAULT 0, device_id TEXT DEFAULT '',
			password TEXT DEFAULT '', birthday INTEGER DEFAULT 0, point REAL DEFAULT 0
		)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("Failed to execute DDL: %v", err)
		}
	}

	return dbm
}

// insertTestSaleBills 插入测试订单数据
func insertTestSaleBills(t *testing.T, db *gorm.DB, bills []model.SaleBill) {
	t.Helper()
	for _, bill := range bills {
		if err := db.Create(&bill).Error; err != nil {
			t.Fatalf("Failed to insert sale_bill: %v", err)
		}
	}
}

// insertTestSaleOrders 插入测试销售订单
func insertTestSaleOrders(t *testing.T, db *gorm.DB, orders []model.SaleOrder) {
	t.Helper()
	for _, order := range orders {
		if err := db.Create(&order).Error; err != nil {
			t.Fatalf("Failed to insert sale_order: %v", err)
		}
	}
}

// insertTestPaymentOrders 插入测试支付订单
func insertTestPaymentOrders(t *testing.T, db *gorm.DB, orders []model.PaymentOrder) {
	t.Helper()
	for _, order := range orders {
		if err := db.Create(&order).Error; err != nil {
			t.Fatalf("Failed to insert payment_order: %v", err)
		}
	}
}

// insertTestDataManages 插入测试数据管理记录
func insertTestDataManages(t *testing.T, db *gorm.DB, dataUuids []uint64) {
	t.Helper()
	now := time.Now().Unix()
	for i, dataUuid := range dataUuids {
		dm := model.DataManage{
			BaseModel: model.BaseModel{
				Uuid:       uint64(90000 + i),
				CreateTime: now,
				UpdateTime: now,
			},
			Type:      model.DataManageTypeOrder,
			DataUuid:  dataUuid,
			StaffUuid: 10001,
		}
		if err := db.Create(&dm).Error; err != nil {
			t.Fatalf("Failed to insert data_manage: %v", err)
		}
	}
}

// TestSaveStatus_Enable 开启数据管理
func TestSaveStatus_Enable(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	srv := newTestDataManageSrv(dbm)

	err := srv.SaveStatus(ctx, shop_req.SaveDataManageStatusReq{
		IsEnableDataManage: true,
	})
	if err != nil {
		t.Fatalf("SaveStatus enable failed: %v", err)
	}
}

// TestSaveStatus_DisableWithOrders 有已选订单时不允许关闭
func TestSaveStatus_DisableWithOrders(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	// 插入一条 data_manage 记录
	insertTestDataManages(t, db, []uint64{1001})

	err := srv.SaveStatus(ctx, shop_req.SaveDataManageStatusReq{
		IsEnableDataManage: false,
	})
	if err == nil {
		t.Fatal("Expected error when disabling with existing orders, got nil")
	}
}

// TestSaveStatus_DisableWithoutOrders 无已选订单时可以关闭
func TestSaveStatus_DisableWithoutOrders(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	srv := newTestDataManageSrv(dbm)

	err := srv.SaveStatus(ctx, shop_req.SaveDataManageStatusReq{
		IsEnableDataManage: false,
	})
	if err != nil {
		t.Fatalf("SaveStatus disable failed: %v", err)
	}
}

// TestSetStaff_SetAndClear 设置和清除操作人员
func TestSetStaff_SetAndClear(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	// 插入测试员工
	for _, uuid := range []uint64{2001, 2002, 2003} {
		db.Exec("INSERT INTO ttpos_staff (uuid, is_super, has_data_permission, delete_time) VALUES (?, 0, 0, 0)", uuid)
	}

	// 设置 2001, 2002 有权限
	err := srv.SetStaff(ctx, shop_req.SetDataManageStaffReq{
		StaffUuids: []uint64{2001, 2002},
	})
	if err != nil {
		t.Fatalf("SetStaff failed: %v", err)
	}

	// 验证权限
	var count int64
	db.Model(&model.Staff{}).Where("has_data_permission = 1 AND delete_time = 0").Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 staff with permission, got %d", count)
	}

	// 清除所有权限
	err = srv.SetStaff(ctx, shop_req.SetDataManageStaffReq{
		StaffUuids: []uint64{},
	})
	if err != nil {
		t.Fatalf("SetStaff clear failed: %v", err)
	}

	db.Model(&model.Staff{}).Where("has_data_permission = 1 AND delete_time = 0").Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 staff with permission after clear, got %d", count)
	}
}

// TestRestoreOrder_Success 成功移除已选订单
func TestRestoreOrder_Success(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	insertTestDataManages(t, db, []uint64{1001, 1002, 1003})

	err := srv.RestoreOrder(ctx, shop_req.RestoreDataManageOrderReq{
		SaleBillUuid: 1002,
	})
	if err != nil {
		t.Fatalf("RestoreOrder failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 2 {
		t.Errorf("Expected 2 records after restore, got %d", len(uuids))
	}
	for _, uid := range uuids {
		if uid == 1002 {
			t.Error("data_uuid 1002 should have been removed")
		}
	}
}

// TestRestoreOrder_NotFound 移除不存在的订单
func TestRestoreOrder_NotFound(t *testing.T) {
	dbm := setupDataManageTestDB(t)
	ctx := createDataManageTestContext(t, dbm)
	srv := newTestDataManageSrv(dbm)

	err := srv.RestoreOrder(ctx, shop_req.RestoreDataManageOrderReq{
		SaleBillUuid: 9999,
	})
	if err == nil {
		t.Fatal("Expected error for non-existent order, got nil")
	}
}

// TestGetOrderList_WithSelectedOrders 获取已选订单列表
func TestGetOrderList_WithSelectedOrders(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	// 插入 3 个已完成订单
	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, Amount: 100, PaymentAmount: 90},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, Amount: 200, PaymentAmount: 180},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now, Amount: 300, PaymentAmount: 270},
	})

	// 插入对应的 SaleOrder（GetPaymentAmount 依赖 SaleOrders）
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, PaymentAmount: 90, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, PaymentAmount: 180, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, PaymentAmount: 270, Status: 1},
	})

	// 插入支付订单
	insertTestPaymentOrders(t, db, []model.PaymentOrder{
		{BaseModel: model.BaseModel{Uuid: 3001, CreateTime: now, UpdateTime: now}, RelatedUuid: 2001, PaymentMethodName: "Cash", PaymentAmount: 90},
		{BaseModel: model.BaseModel{Uuid: 3002, CreateTime: now, UpdateTime: now}, RelatedUuid: 2002, PaymentMethodName: "Card", PaymentAmount: 180},
		{BaseModel: model.BaseModel{Uuid: 3003, CreateTime: now, UpdateTime: now}, RelatedUuid: 2003, PaymentMethodName: "Cash", PaymentAmount: 270},
	})

	// 将 1001、1002 加入 data_manage
	insertTestDataManages(t, db, []uint64{1001, 1002})

	resp, err := srv.GetOrderList(ctx, shop_req.GetDataManageOrderListReq{
		PageNo:   1,
		PageSize: 10,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderList failed: %v", err)
	}

	// 应返回 2 条已选订单
	if len(resp.List) != 2 {
		t.Errorf("Expected 2 orders in list, got %d", len(resp.List))
	}
	if resp.Meta.Total != 2 {
		t.Errorf("Expected total=2, got %d", resp.Meta.Total)
	}
	if resp.Meta.OrderCount != 2 {
		t.Errorf("Expected order_count=2, got %d", resp.Meta.OrderCount)
	}

	// 所有返回的订单应标记为已选
	for _, item := range resp.List {
		if !item.IsSelected {
			t.Errorf("Order %d should be marked as selected", item.SaleBillUuid)
		}
	}
}

// TestGetOrderList_Empty 无已选订单时返回空列表
func TestGetOrderList_Empty(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	srv := newTestDataManageSrv(dbm)

	resp, err := srv.GetOrderList(ctx, shop_req.GetDataManageOrderListReq{
		PageNo:   1,
		PageSize: 10,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderList failed: %v", err)
	}

	if len(resp.List) != 0 {
		t.Errorf("Expected empty list, got %d items", len(resp.List))
	}
	if resp.Meta.OrderCount != 0 {
		t.Errorf("Expected order_count=0, got %d", resp.Meta.OrderCount)
	}
}

// TestGetOrderSelect_WithMixedSelection 获取可选订单列表（含已选和未选）
func TestGetOrderSelect_WithMixedSelection(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	// 插入 3 个已完成订单
	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, Amount: 100, PaymentAmount: 90},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, Amount: 200, PaymentAmount: 180},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now, Amount: 300, PaymentAmount: 270},
	})

	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, PaymentAmount: 90, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, PaymentAmount: 180, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, PaymentAmount: 270, Status: 1},
	})

	insertTestPaymentOrders(t, db, []model.PaymentOrder{
		{BaseModel: model.BaseModel{Uuid: 3001, CreateTime: now, UpdateTime: now}, RelatedUuid: 2001, PaymentMethodName: "Cash", PaymentAmount: 90},
		{BaseModel: model.BaseModel{Uuid: 3002, CreateTime: now, UpdateTime: now}, RelatedUuid: 2002, PaymentMethodName: "Card", PaymentAmount: 180},
		{BaseModel: model.BaseModel{Uuid: 3003, CreateTime: now, UpdateTime: now}, RelatedUuid: 2003, PaymentMethodName: "Cash", PaymentAmount: 270},
	})

	// 只将 1001 加入 data_manage
	insertTestDataManages(t, db, []uint64{1001})

	resp, err := srv.GetOrderSelect(ctx, shop_req.GetDataManageOrderSelectReq{
		PageNo:   1,
		PageSize: 10,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderSelect failed: %v", err)
	}

	// 应返回 3 条订单（全部已完成订单）
	if len(resp.List) != 3 {
		t.Errorf("Expected 3 orders in list, got %d", len(resp.List))
	}
	if resp.Meta.Total != 3 {
		t.Errorf("Expected total=3, got %d", resp.Meta.Total)
	}

	// 验证选中状态
	selectedCount := 0
	for _, item := range resp.List {
		if item.IsSelected {
			selectedCount++
			if item.SaleBillUuid != 1001 {
				t.Errorf("Only order 1001 should be selected, but %d is selected", item.SaleBillUuid)
			}
		}
	}
	if selectedCount != 1 {
		t.Errorf("Expected 1 selected order, got %d", selectedCount)
	}
}

// TestGetOrderSelect_ExcludeUnfinished 可选订单列表不含未完成订单
func TestGetOrderSelect_ExcludeUnfinished(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, Amount: 100, PaymentAmount: 90},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 0, BillType: 0, ProductionTime: now, Amount: 200, PaymentAmount: 0}, // 未完成
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 2, BillType: 0, ProductionTime: now, Amount: 300, PaymentAmount: 0}, // 已取消
	})

	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, PaymentAmount: 90, Status: 1},
	})

	resp, err := srv.GetOrderSelect(ctx, shop_req.GetDataManageOrderSelectReq{
		PageNo:   1,
		PageSize: 10,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderSelect failed: %v", err)
	}

	// 只有 1 条已完成订单
	if len(resp.List) != 1 {
		t.Errorf("Expected 1 order (only completed), got %d", len(resp.List))
	}
}

// TestGetOrderList_PaymentAmountCalculation 验证实付金额使用 GetPaymentAmount（扣除退款）
func TestGetOrderList_PaymentAmountCalculation(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, Amount: 100, PaymentAmount: 100},
	})

	// SaleOrder 的 PaymentAmount = 100
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, PaymentAmount: 100, Status: 1},
	})

	// 存在退款 30
	db.Exec(`INSERT INTO ttpos_return_order (uuid, create_time, update_time, delete_time, related_order_type, related_order_uuid, refund_amount) VALUES (?, ?, ?, 0, 0, ?, ?)`,
		5001, now, now, 2001, 30.0)

	insertTestDataManages(t, db, []uint64{1001})

	resp, err := srv.GetOrderList(ctx, shop_req.GetDataManageOrderListReq{
		PageNo:   1,
		PageSize: 10,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderList failed: %v", err)
	}

	if len(resp.List) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(resp.List))
	}

	// GetPaymentAmount = SaleOrders.PaymentAmount 合计 - 退款金额 = 100 - 30 = 70
	if resp.List[0].PaymentAmount != 70 {
		t.Errorf("Expected PaymentAmount=70 (100-30 refund), got %v", resp.List[0].PaymentAmount)
	}
}

// TestGetOrderList_PaymentMethodDisplay 验证支付方式显示
func TestGetOrderList_PaymentMethodDisplay(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, Amount: 100, PaymentAmount: 100},
	})

	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, PaymentAmount: 50, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, PaymentAmount: 50, Status: 1},
	})

	// 两个支付订单使用不同支付方式
	insertTestPaymentOrders(t, db, []model.PaymentOrder{
		{BaseModel: model.BaseModel{Uuid: 3001, CreateTime: now, UpdateTime: now}, RelatedUuid: 2001, PaymentMethodName: "Cash", PaymentAmount: 50},
		{BaseModel: model.BaseModel{Uuid: 3002, CreateTime: now, UpdateTime: now}, RelatedUuid: 2002, PaymentMethodName: "Card", PaymentAmount: 50},
	})

	insertTestDataManages(t, db, []uint64{1001})

	resp, err := srv.GetOrderList(ctx, shop_req.GetDataManageOrderListReq{
		PageNo:   1,
		PageSize: 10,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderList failed: %v", err)
	}

	if len(resp.List) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(resp.List))
	}

	// 应包含 Cash 和 Card
	pm := resp.List[0].PaymentMethod
	if pm != "Cash, Card" && pm != "Card, Cash" {
		t.Errorf("Expected PaymentMethod to contain 'Cash' and 'Card', got '%s'", pm)
	}
}

// TestSubmitOrder_ManualSelect 手动勾选模式（select_all=false）
func TestSubmitOrder_ManualSelect(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	// 插入 3 个已完成订单
	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now},
	})

	// 初始选中 1001、1002
	insertTestDataManages(t, db, []uint64{1001, 1002})

	// 手动勾选：新增 1003，移除 1001（1002 不变）
	err := srv.SubmitOrder(ctx, shop_req.SubmitDataManageOrderReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{
				SelectedUuids:   []uint64{1003},
				DeselectedUuids: []uint64{1001},
				DateType:        -1,
				BillType:        -1,
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitOrder failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 2 {
		t.Errorf("Expected 2 records after submit, got %d", len(uuids))
	}

	uuidSet := make(map[uint64]struct{})
	for _, uid := range uuids {
		uuidSet[uid] = struct{}{}
	}
	if _, ok := uuidSet[1001]; ok {
		t.Error("data_uuid 1001 should have been removed")
	}
	if _, ok := uuidSet[1002]; !ok {
		t.Error("data_uuid 1002 should be retained")
	}
	if _, ok := uuidSet[1003]; !ok {
		t.Error("data_uuid 1003 should have been added")
	}
}

// TestSubmitOrder_SelectAll 全选模式
func TestSubmitOrder_SelectAll(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	// 插入 4 个已完成订单
	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1004, CreateTime: now, UpdateTime: now}, OrderNo: "ORD004", Status: 1, BillType: 0, ProductionTime: now},
	})

	// 初始仅选中 1001
	insertTestDataManages(t, db, []uint64{1001})

	// 全选：应该选中所有 4 个订单
	err := srv.SubmitOrder(ctx, shop_req.SubmitDataManageOrderReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{
				SelectAll: true,
				DateType:  -1,
				BillType:  -1,
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitOrder SelectAll failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 4 {
		t.Errorf("Expected 4 records after select all, got %d", len(uuids))
	}
}

// TestSubmitOrder_Deselect 案例3：从数据管理中移除 DeselectedUuids
func TestSubmitOrder_Deselect(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now},
	})

	// 初始选中 1001、1002、1003
	insertTestDataManages(t, db, []uint64{1001, 1002, 1003})

	// 案例3：SelectAll=false, DeselectedUuids=[1002] → 仅从已管理中移除 1002，不新增
	// → 最终: 1001, 1003
	err := srv.SubmitOrder(ctx, shop_req.SubmitDataManageOrderReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{
				DeselectedUuids: []uint64{1002},
				DateType:        -1,
				BillType:        -1,
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitOrder Deselect failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 2 {
		t.Errorf("Expected 2 records after deselect, got %d", len(uuids))
	}

	uuidSet := make(map[uint64]struct{})
	for _, uid := range uuids {
		uuidSet[uid] = struct{}{}
	}
	if _, ok := uuidSet[1001]; !ok {
		t.Error("data_uuid 1001 should be retained")
	}
	if _, ok := uuidSet[1002]; ok {
		t.Error("data_uuid 1002 should have been removed")
	}
	if _, ok := uuidSet[1003]; !ok {
		t.Error("data_uuid 1003 should be retained")
	}
}

// TestSubmitOrder_SelectAllWithDeselect 案例1：全选但排除部分
func TestSubmitOrder_SelectAllWithDeselect(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now},
		{BaseModel: model.BaseModel{Uuid: 1004, CreateTime: now, UpdateTime: now}, OrderNo: "ORD004", Status: 1, BillType: 0, ProductionTime: now},
	})

	// 案例1：SelectAll=true, DeselectedUuids=[1002,1004]
	// 范围=[1001,1002,1003,1004], 排除[1002,1004] → 新增 1001,1003
	err := srv.SubmitOrder(ctx, shop_req.SubmitDataManageOrderReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{
				SelectAll:       true,
				DeselectedUuids: []uint64{1002, 1004},
				DateType:        -1,
				BillType:        -1,
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitOrder SelectAll with deselect failed: %v", err)
	}

	uuids := getDataManageUuids(db)
	if len(uuids) != 2 {
		t.Errorf("Expected 2 records, got %d", len(uuids))
	}

	uuidSet := make(map[uint64]struct{})
	for _, uid := range uuids {
		uuidSet[uid] = struct{}{}
	}
	if _, ok := uuidSet[1001]; !ok {
		t.Error("data_uuid 1001 should be added")
	}
	if _, ok := uuidSet[1002]; ok {
		t.Error("data_uuid 1002 should have been excluded")
	}
	if _, ok := uuidSet[1003]; !ok {
		t.Error("data_uuid 1003 should be added")
	}
	if _, ok := uuidSet[1004]; ok {
		t.Error("data_uuid 1004 should have been excluded")
	}
}

// TestGetOrderList_Pagination 验证分页
func TestGetOrderList_Pagination(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	// 插入 5 个订单并全部加入 data_manage
	bills := make([]model.SaleBill, 5)
	dataUuids := make([]uint64, 5)
	for i := 0; i < 5; i++ {
		uuid := uint64(1001 + i)
		bills[i] = model.SaleBill{
			BaseModel:      model.BaseModel{Uuid: uuid, CreateTime: now - int64(i), UpdateTime: now},
			OrderNo:        "ORD00" + string(rune('1'+i)),
			Status:         1,
			BillType:       0,
			ProductionTime: now,
			Amount:         float64((i + 1) * 100),
		}
		dataUuids[i] = uuid
	}
	insertTestSaleBills(t, db, bills)

	orders := make([]model.SaleOrder, 5)
	for i := 0; i < 5; i++ {
		orders[i] = model.SaleOrder{
			BaseModel:    model.BaseModel{Uuid: uint64(2001 + i), CreateTime: now, UpdateTime: now},
			SaleBillUuid: uint64(1001 + i),
			Status:       1,
		}
	}
	insertTestSaleOrders(t, db, orders)

	insertTestDataManages(t, db, dataUuids)

	// 第一页，每页 2 条
	resp, err := srv.GetOrderList(ctx, shop_req.GetDataManageOrderListReq{
		PageNo:   1,
		PageSize: 2,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderList page 1 failed: %v", err)
	}

	if len(resp.List) != 2 {
		t.Errorf("Expected 2 items on page 1, got %d", len(resp.List))
	}
	if resp.Meta.Total != 5 {
		t.Errorf("Expected total=5, got %d", resp.Meta.Total)
	}
	if resp.Meta.OrderCount != 5 {
		t.Errorf("Expected order_count=5, got %d", resp.Meta.OrderCount)
	}

	// 第三页，每页 2 条，应只有 1 条
	resp, err = srv.GetOrderList(ctx, shop_req.GetDataManageOrderListReq{
		PageNo:   3,
		PageSize: 2,
		DateType: -1,
		BillType: -1,
	})
	if err != nil {
		t.Fatalf("GetOrderList page 3 failed: %v", err)
	}

	if len(resp.List) != 1 {
		t.Errorf("Expected 1 item on page 3, got %d", len(resp.List))
	}
}

// TestGetOrderSelectStats_SelectAll 案例2：全选模式统计预览
func TestGetOrderSelectStats_SelectAll(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 100},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 200},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 1, ProductionTime: now, PaymentAmount: 300},
	})
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, Status: 1},
	})

	// 全选所有 → 3条，金额 600
	stats, err := srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{SelectAll: true, DateType: -1, BillType: -1},
		},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats failed: %v", err)
	}
	if stats.SelectedCount != 3 {
		t.Errorf("Expected selected_count=3, got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 600 {
		t.Errorf("Expected paid_amount=600, got %f", stats.PaidAmount)
	}
}

// TestGetOrderSelectStats_SelectAllWithDeselect 案例1：全选后排除部分
func TestGetOrderSelectStats_SelectAllWithDeselect(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 100},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 200},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 300},
	})
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, Status: 1},
	})

	// 全选但排除 1002 → finalSet={1001,1003}，金额 400
	stats, err := srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{SelectAll: true, DeselectedUuids: []uint64{1002}, DateType: -1, BillType: -1},
		},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats failed: %v", err)
	}
	if stats.SelectedCount != 2 {
		t.Errorf("Expected selected_count=2, got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 400 {
		t.Errorf("Expected paid_amount=400, got %f", stats.PaidAmount)
	}
	if stats.IsSelectAll {
		t.Error("Expected is_select_all=false when deselecting")
	}
}

// TestGetOrderSelectStats_ManualSelect 案例4：已管理 ∪ 手动新增
func TestGetOrderSelectStats_ManualSelect(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 100},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 200},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 300},
	})
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, Status: 1},
	})

	// 已管理 1001
	insertTestDataManages(t, db, []uint64{1001})

	// 手动选 1003 → existing{1001} ∪ {1003} = {1001,1003}，2条，金额 400
	stats, err := srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{SelectedUuids: []uint64{1003}, DateType: -1, BillType: -1},
		},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats failed: %v", err)
	}
	if stats.SelectedCount != 2 {
		t.Errorf("Expected selected_count=2, got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 400 {
		t.Errorf("Expected paid_amount=400, got %f", stats.PaidAmount)
	}

	// 兜底：都为空 → 返回已管理统计 existing{1001}，1条，金额 100
	stats, err = srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats fallback failed: %v", err)
	}
	if stats.SelectedCount != 1 {
		t.Errorf("Expected selected_count=1, got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 100 {
		t.Errorf("Expected paid_amount=100, got %f", stats.PaidAmount)
	}
}

// TestGetOrderSelectStats_WithFilter 案例2：带筛选条件的全选预览
func TestGetOrderSelectStats_WithFilter(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	// 2 个餐单 + 1 个外卖
	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 100},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 200},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 1, ProductionTime: now, PaymentAmount: 300},
	})
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, Status: 1},
	})

	// 全选仅餐单（bill_type=0）→ finalSet={1001,1002}，金额 300，TotalCount=2
	stats, err := srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{SelectAll: true, DateType: -1, BillType: 0},
		},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats with filter failed: %v", err)
	}
	if stats.SelectedCount != 2 {
		t.Errorf("Expected selected_count=2 (dine-in only), got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 300 {
		t.Errorf("Expected paid_amount=300, got %f", stats.PaidAmount)
	}
	if stats.TotalCount != 2 {
		t.Errorf("Expected total_count=2 (dine-in only), got %d", stats.TotalCount)
	}
}

// TestGetOrderSelectStats_Deselect 案例3：已管理(筛选范围内) - DeselectedUuids
func TestGetOrderSelectStats_Deselect(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 100},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 200},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 300},
	})
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, Status: 1},
	})

	// 已管理 1001, 1002, 1003
	insertTestDataManages(t, db, []uint64{1001, 1002, 1003})

	// 案例3：DeselectedUuids=[1002] → existing{1001,1002,1003} - {1002} = {1001,1003}，2条，金额 400
	stats, err := srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{DeselectedUuids: []uint64{1002}, DateType: -1, BillType: -1},
		},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats failed: %v", err)
	}
	if stats.SelectedCount != 2 {
		t.Errorf("Expected selected_count=2, got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 400 {
		t.Errorf("Expected paid_amount=400, got %f", stats.PaidAmount)
	}
	if stats.TotalCount != 2 {
		t.Errorf("Expected total_count=2, got %d", stats.TotalCount)
	}
}

// TestGetOrderSelectStats_ManualSelectPartial 案例4：手动选部分，IsSelectAll=false
func TestGetOrderSelectStats_ManualSelectPartial(t *testing.T) {
	dbm := setupDataManageTestDBWithOrders(t)
	ctx := createDataManageTestContext(t, dbm)
	db := dbm.GetDB(constant.MockDB)
	srv := newTestDataManageSrv(dbm)

	now := time.Now().Unix()

	insertTestSaleBills(t, db, []model.SaleBill{
		{BaseModel: model.BaseModel{Uuid: 1001, CreateTime: now, UpdateTime: now}, OrderNo: "ORD001", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 100},
		{BaseModel: model.BaseModel{Uuid: 1002, CreateTime: now, UpdateTime: now}, OrderNo: "ORD002", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 200},
		{BaseModel: model.BaseModel{Uuid: 1003, CreateTime: now, UpdateTime: now}, OrderNo: "ORD003", Status: 1, BillType: 0, ProductionTime: now, PaymentAmount: 300},
	})
	insertTestSaleOrders(t, db, []model.SaleOrder{
		{BaseModel: model.BaseModel{Uuid: 2001, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1001, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2002, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1002, Status: 1},
		{BaseModel: model.BaseModel{Uuid: 2003, CreateTime: now, UpdateTime: now}, SaleBillUuid: 1003, Status: 1},
	})

	// 手动选 1001 → SelectedCount=1, TotalCount=1, IsSelectAll=false
	stats, err := srv.GetOrderSelectStats(ctx, shop_req.GetDataManageOrderSelectStatsReq{
		Filters: []shop_req.DataManageOrderSubmitFilter{
			{SelectedUuids: []uint64{1001}, DateType: -1, BillType: -1},
		},
	})
	if err != nil {
		t.Fatalf("GetOrderSelectStats failed: %v", err)
	}
	if stats.SelectedCount != 1 {
		t.Errorf("Expected selected_count=1, got %d", stats.SelectedCount)
	}
	if stats.PaidAmount != 100 {
		t.Errorf("Expected paid_amount=100, got %f", stats.PaidAmount)
	}
	if stats.TotalCount != 1 {
		t.Errorf("Expected total_count=1, got %d", stats.TotalCount)
	}
}
