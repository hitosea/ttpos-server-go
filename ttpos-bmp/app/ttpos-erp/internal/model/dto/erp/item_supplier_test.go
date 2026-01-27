package erp_test

import (
	"testing"

	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

// TestItemSupplierStruct 测试 ItemSupplier 结构体
func TestItemSupplierStruct(t *testing.T) {
	supplier := erp.ItemSupplier{
		Name:       "supplier-001",
		Supplier:   "Test Supplier",
		Idx:        1,
		Parent:     "ITEM-001",
		Parenttype: "Item",
		Doctype:    "Item Supplier",
	}

	if supplier.Name != "supplier-001" {
		t.Errorf("Name 字段错误，期望: supplier-001, 实际: %s", supplier.Name)
	}
	if supplier.Supplier != "Test Supplier" {
		t.Errorf("Supplier 字段错误，期望: Test Supplier, 实际: %s", supplier.Supplier)
	}
	if supplier.Idx != 1 {
		t.Errorf("Idx 字段错误，期望: 1, 实际: %d", supplier.Idx)
	}
	if supplier.Parent != "ITEM-001" {
		t.Errorf("Parent 字段错误，期望: ITEM-001, 实际: %s", supplier.Parent)
	}
	if supplier.Parenttype != "Item" {
		t.Errorf("Parenttype 字段错误，期望: Item, 实际: %s", supplier.Parenttype)
	}
	if supplier.Doctype != "Item Supplier" {
		t.Errorf("Doctype 字段错误，期望: Item Supplier, 实际: %s", supplier.Doctype)
	}
}

// TestItemSupplierSlice 测试 Item 结构体中的 SupplierItems 字段
func TestItemSupplierSlice(t *testing.T) {
	item := erp.Item{
		ItemCode: "ITEM-001",
		ItemName: "Test Item",
		SupplierItems: []erp.ItemSupplier{
			{Supplier: "Supplier A", Idx: 1},
			{Supplier: "Supplier B", Idx: 2},
		},
	}

	if len(item.SupplierItems) != 2 {
		t.Errorf("SupplierItems 长度错误，期望: 2, 实际: %d", len(item.SupplierItems))
		return
	}

	if item.SupplierItems[0].Supplier != "Supplier A" {
		t.Errorf("第一个供应商错误，期望: Supplier A, 实际: %s", item.SupplierItems[0].Supplier)
	}
	if item.SupplierItems[1].Supplier != "Supplier B" {
		t.Errorf("第二个供应商错误，期望: Supplier B, 实际: %s", item.SupplierItems[1].Supplier)
	}
}

// TestItemSupplierEmptySlice 测试空供应商列表
func TestItemSupplierEmptySlice(t *testing.T) {
	item := erp.Item{
		ItemCode:      "ITEM-001",
		SupplierItems: []erp.ItemSupplier{},
	}

	if item.SupplierItems == nil {
		t.Error("SupplierItems 应该是空切片而不是 nil")
	}
	if len(item.SupplierItems) != 0 {
		t.Errorf("SupplierItems 长度错误，期望: 0, 实际: %d", len(item.SupplierItems))
	}
}

// TestItemDeliveredBySupplier 测试 DeliveredBySupplier 字段
func TestItemDeliveredBySupplier(t *testing.T) {
	testCases := []struct {
		name     string
		value    int
		expected bool
	}{
		{"零值表示false", 0, false},
		{"1表示true", 1, true},
		{"其他非零值应为false", 2, false}, // ERP 中只有 1 表示 true
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := erp.Item{
				DeliveredBySupplier: tc.value,
			}
			result := item.DeliveredBySupplier == 1
			if result != tc.expected {
				t.Errorf("DeliveredBySupplier == 1 判断错误，值: %d, 期望: %v, 实际: %v",
					tc.value, tc.expected, result)
			}
		})
	}
}
