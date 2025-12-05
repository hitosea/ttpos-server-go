package repository

import (
	"testing"
	"ttpos-server-go/main/app/modules/order_core/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&model.CoreSaleBill{}, &model.CoreSaleOrder{}, &model.CoreSaleOrderProduct{})
	return db
}

func TestCoreOrderRepo_CreateBill(t *testing.T) {
	db := setupTestDB()
	repo := NewCoreOrderRepo(db)

	bill := &model.CoreSaleBill{
		BaseModel: model.BaseModel{
			Uuid:       1,
			CreateTime: 100,
			UpdateTime: 100,
		},
		OrderNo: "B001",
		Status:  0,
		Amount:  100.0,
	}

	if err := repo.CreateBill(bill); err != nil {
		t.Fatalf("failed to create bill: %v", err)
	}

	saved, err := repo.GetBillByUuid(1)
	if err != nil {
		t.Fatalf("failed to get bill: %v", err)
	}
	if saved.OrderNo != "B001" {
		t.Errorf("expected OrderNo B001, got %s", saved.OrderNo)
	}
}

func TestCoreOrderRepo_UpdateBillStatus(t *testing.T) {
	db := setupTestDB()
	repo := NewCoreOrderRepo(db)

	bill := &model.CoreSaleBill{
		BaseModel: model.BaseModel{Uuid: 1},
		Status:    0,
	}
	repo.CreateBill(bill)

	if err := repo.UpdateBillStatus(1, 1); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	saved, _ := repo.GetBillByUuid(1)
	if saved.Status != 1 {
		t.Errorf("expected status 1, got %d", saved.Status)
	}
}

