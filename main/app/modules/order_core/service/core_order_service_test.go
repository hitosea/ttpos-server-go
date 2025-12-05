package service

import (
	"testing"
	"ttpos-server-go/app/dto"
	core_dto "ttpos-server-go/app/modules/order_core/dto"
	"ttpos-server-go/app/modules/order_core/model"
	"ttpos-server-go/app/modules/order_core/repository"
	pkg_context "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MockContext 实现 pkg_context.Context 接口
type MockContext struct {
	pkg_context.Context
	db *gorm.DB
}

func (c *MockContext) GetDB() *gorm.DB {
	return c.db
}

func setupTestDB() *gorm.DB {
	dsn := "root:004cd912a89e4ac6@tcp(192.168.100.94:13306)/?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// 创建测试数据库
	db.Exec("CREATE DATABASE IF NOT EXISTS ttpos_test_order_core")
	db.Exec("USE ttpos_test_order_core")

	// 清理旧数据
	db.Migrator().DropTable(&model.CoreSaleBill{}, &model.CoreSaleOrder{}, &model.CoreSaleOrderProduct{})

	db.AutoMigrate(&model.CoreSaleBill{}, &model.CoreSaleOrder{}, &model.CoreSaleOrderProduct{})
	return db
}

func setupTestDB2() *gorm.DB {
	dsn := "root:004cd912a89e4ac6@tcp(192.168.100.94:13306)/?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// 创建测试数据库
	// db.Exec("CREATE DATABASE IF NOT EXISTS ttpos_test_order_core")
	db.Exec("USE shop5965986402304000")

	// 清理旧数据
	// db.Migrator().DropTable(&model.CoreSaleBill{}, &model.CoreSaleOrder{}, &model.CoreSaleOrderProduct{})

	// db.AutoMigrate(&model.CoreSaleBill{}, &model.CoreSaleOrder{}, &model.CoreSaleOrderProduct{})
	return db
}

func TestCoreOrderService_CreateOrder(t *testing.T) {
	// 初始化雪花ID
	utils.InitIdGenerator()
	db := setupTestDB2()
	svc := NewCoreOrderService(nil) // dbm ignored as we use context DB

	ctx := &MockContext{db: db}

	req := &core_dto.CreateOrderReq{
		OrderNo: "M001",
		Amount:  100.0,
		Orders: []core_dto.CreateOrder{
			{
				OrderNo: "S001",
				Amount:  100.0,
				Products: []core_dto.CreateOrderProduct{
					{
						Name:       dto.LocaleResponse{ZH: "P1", EN: "P1_EN"},
						FlavorName: dto.LocaleResponse{ZH: "P1_Flavor", EN: "P1_Flavor_EN"},
						Num:        1,
						SalePrice:  100,
						Price:      100,
						TotalPrice: 100,
					},
				},
			},
		},
	}

	resp, err := svc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if resp.BillUuid == 0 {
		t.Error("expected valid BillUuid")
	}

	// Verify DB
	repo := repository.NewCoreOrderRepo(db)
	bill, _ := repo.GetBillByUuid(resp.BillUuid)
	if bill == nil {
		t.Error("bill not found in DB")
	}
}

func TestCoreOrderService_StateTransitions(t *testing.T) {
	db := setupTestDB()
	svc := NewCoreOrderService(nil)
	ctx := &MockContext{db: db}
	repo := repository.NewCoreOrderRepo(db)

	// Create initial bill
	bill := &model.CoreSaleBill{
		BaseModel: model.BaseModel{Uuid: 100},
		Status:    0, // Pending
	}
	repo.CreateBill(bill)

	// Test MarkAsPaid (0 -> 1)
	if err := svc.MarkAsPaid(ctx, 100); err != nil {
		t.Fatalf("MarkAsPaid failed: %v", err)
	}

	saved, _ := repo.GetBillByUuid(100)
	if saved.Status != 1 {
		t.Errorf("expected status 1, got %d", saved.Status)
	}

	// Test Invalid Transition (1 -> 1 is error in my impl? No, check logic)
	// Logic: if bill.Status != 0 return error.
	if err := svc.MarkAsPaid(ctx, 100); err == nil {
		t.Error("expected error when marking paid bill as paid again")
	}

	// Test CancelOrder (1 -> 2) - Should fail because status is 1
	if err := svc.CancelOrder(ctx, 100); err == nil {
		t.Error("expected error when cancelling paid bill")
	}
}
