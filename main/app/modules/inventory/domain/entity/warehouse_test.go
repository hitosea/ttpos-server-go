package entity

import (
	"testing"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
	"ttpos-server-go/app/dto"
)

func TestNewWarehouse(t *testing.T) {
	// 准备测试数据
	name := valueobject.NewMultiLanguageName(dto.LocaleResponse{
		ZH: "测试仓库",
		EN: "Test Warehouse",
	})
	warehouseType := valueobject.NewWarehouseType("normal")
	code, _ := valueobject.NewWarehouseCode("WH01")
	status := valueobject.NewWarehouseStatus(1)
	contactInfo := valueobject.NewContactInfo("张三", "13800138000", "测试地址")

	// 创建仓库
	warehouse := NewWarehouse(name, warehouseType, code, status, contactInfo)

	// 验证
	if warehouse == nil {
		t.Fatal("warehouse should not be nil")
	}
	if warehouse.Name().ZH() != "测试仓库" {
		t.Errorf("expected name '测试仓库', got '%s'", warehouse.Name().ZH())
	}
	if !warehouse.WarehouseType().IsNormal() {
		t.Error("expected normal warehouse type")
	}
	if warehouse.Code().Value() != "WH01" {
		t.Errorf("expected code 'WH01', got '%s'", warehouse.Code().Value())
	}
	if !warehouse.Status().IsActive() {
		t.Error("expected active status")
	}
	if warehouse.IsDefault() {
		t.Error("new warehouse should not be default")
	}
}

func TestWarehouse_SetAsDefault(t *testing.T) {
	// 创建普通仓库
	warehouse := createTestWarehouse()

	// 设置为默认仓库
	err := warehouse.SetAsDefault()
	if err != nil {
		t.Fatalf("SetAsDefault failed: %v", err)
	}

	// 验证
	if !warehouse.IsDefault() {
		t.Error("warehouse should be default")
	}

	// 再次设置应该报错
	err = warehouse.SetAsDefault()
	if err == nil {
		t.Error("should return error when already default")
	}
}

func TestWarehouse_SetAsDefault_Transit(t *testing.T) {
	// 创建在途仓库
	warehouse := createTestWarehouse()
	warehouse.UpdateType(valueobject.NewWarehouseType("transit"))

	// 在途仓库不能设置为默认仓库
	err := warehouse.SetAsDefault()
	if err == nil {
		t.Error("transit warehouse should not be set as default")
	}
}

func TestWarehouse_Deactivate(t *testing.T) {
	// 创建启用的仓库
	warehouse := createTestWarehouse()

	// 禁用仓库
	err := warehouse.Deactivate()
	if err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}

	// 验证
	if !warehouse.Status().IsDisabled() {
		t.Error("warehouse should be disabled")
	}
}

func TestWarehouse_Deactivate_Default(t *testing.T) {
	// 创建默认仓库
	warehouse := createTestWarehouse()
	warehouse.SetAsDefault()

	// 默认仓库不能禁用
	err := warehouse.Deactivate()
	if err == nil {
		t.Error("default warehouse should not be deactivated")
	}
}

func TestWarehouse_Deactivate_Transit(t *testing.T) {
	// 创建在途仓库
	warehouse := createTestWarehouse()
	warehouse.UpdateType(valueobject.NewWarehouseType("transit"))

	// 在途仓库不能禁用
	err := warehouse.Deactivate()
	if err == nil {
		t.Error("transit warehouse should not be deactivated")
	}
}

func TestWarehouse_CanBeDeleted(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Warehouse)
		expectErr bool
	}{
		{
			name:      "正常仓库可以删除",
			setup:     func(w *Warehouse) {},
			expectErr: false,
		},
		{
			name: "默认仓库不能删除",
			setup: func(w *Warehouse) {
				w.SetAsDefault()
			},
			expectErr: true,
		},
		{
			name: "在途仓库不能删除",
			setup: func(w *Warehouse) {
				w.UpdateType(valueobject.NewWarehouseType("transit"))
			},
			expectErr: true,
		},
		{
			name: "有关联物品的仓库不能删除",
			setup: func(w *Warehouse) {
				w.SetHasItems(true)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warehouse := createTestWarehouse()
			tt.setup(warehouse)

			err := warehouse.CanBeDeleted()
			if (err != nil) != tt.expectErr {
				t.Errorf("CanBeDeleted() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestWarehouse_UpdateContactInfo(t *testing.T) {
	warehouse := createTestWarehouse()
	newContactInfo := valueobject.NewContactInfo("李四", "13900139000", "新地址")

	err := warehouse.UpdateContactInfo(newContactInfo)
	if err != nil {
		t.Fatalf("UpdateContactInfo failed: %v", err)
	}

	if warehouse.ContactInfo().Contact() != "李四" {
		t.Errorf("expected contact '李四', got '%s'", warehouse.ContactInfo().Contact())
	}
	if warehouse.ContactInfo().Phone() != "13900139000" {
		t.Errorf("expected phone '13900139000', got '%s'", warehouse.ContactInfo().Phone())
	}
	if warehouse.ContactInfo().Address() != "新地址" {
		t.Errorf("expected address '新地址', got '%s'", warehouse.ContactInfo().Address())
	}
}

// 辅助函数：创建测试仓库
func createTestWarehouse() *Warehouse {
	name := valueobject.NewMultiLanguageName(dto.LocaleResponse{
		ZH: "测试仓库",
		EN: "Test Warehouse",
	})
	warehouseType := valueobject.NewWarehouseType("normal")
	code, _ := valueobject.NewWarehouseCode("WH01")
	status := valueobject.NewWarehouseStatus(1)
	contactInfo := valueobject.NewContactInfo("张三", "13800138000", "测试地址")

	return NewWarehouse(name, warehouseType, code, status, contactInfo)
}


