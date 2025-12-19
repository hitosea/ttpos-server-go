package grab

import (
	"testing"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// ============================================================================
// UpdateMenuItemReq DTO 测试
// ============================================================================

// TestUpdateMenuItemReq_ToSDKUpdateMenuItem 测试 DTO 转换为 SDK 请求
func TestUpdateMenuItemReq_ToSDKUpdateMenuItem(t *testing.T) {
	testCases := []struct {
		name     string
		req      *UpdateMenuItemReq
		validate func(t *testing.T, item *grabfood.UpdateMenuItem)
	}{
		{
			name: "基本字段转换",
			req: &UpdateMenuItemReq{
				MerchantID: "M-12345",
				ItemID:     "ITEM-001",
			},
			validate: func(t *testing.T, item *grabfood.UpdateMenuItem) {
				if item.GetMerchantID() != "M-12345" {
					t.Errorf("MerchantID = %s, want M-12345", item.GetMerchantID())
				}
				if item.GetId() != "ITEM-001" {
					t.Errorf("Id = %s, want ITEM-001", item.GetId())
				}
				if item.GetField() != MenuItemUpdateFieldItem {
					t.Errorf("Field = %s, want %s", item.GetField(), MenuItemUpdateFieldItem)
				}
			},
		},
		{
			name: "包含价格和状态",
			req: &UpdateMenuItemReq{
				MerchantID:      "M-12345",
				ItemID:          "ITEM-002",
				Price:           ptrInt64(1000), // 10.00
				AvailableStatus: MenuAvailableStatusAvailable,
			},
			validate: func(t *testing.T, item *grabfood.UpdateMenuItem) {
				if item.GetPrice() != 1000 {
					t.Errorf("Price = %d, want 1000", item.GetPrice())
				}
				if item.GetAvailableStatus() != MenuAvailableStatusAvailable {
					t.Errorf("AvailableStatus = %s, want %s", item.GetAvailableStatus(), MenuAvailableStatusAvailable)
				}
			},
		},
		{
			name: "包含库存",
			req: &UpdateMenuItemReq{
				MerchantID:      "M-12345",
				ItemID:          "ITEM-003",
				MaxStock:        ptrInt64(50),
				AvailableStatus: MenuAvailableStatusUnavailable,
			},
			validate: func(t *testing.T, item *grabfood.UpdateMenuItem) {
				if item.GetMaxStock() != 50 {
					t.Errorf("MaxStock = %d, want 50", item.GetMaxStock())
				}
			},
		},
		{
			name: "包含高级定价",
			req: &UpdateMenuItemReq{
				MerchantID: "M-12345",
				ItemID:     "ITEM-004",
				AdvancedPricings: []UpdateAdvancedPricingReq{
					{Key: "DELIVERY.STANDARD.GRABFOOD", Price: 1500},
					{Key: "PICKUP.STANDARD.GRABFOOD", Price: 1200},
				},
			},
			validate: func(t *testing.T, item *grabfood.UpdateMenuItem) {
				pricings := item.GetAdvancedPricings()
				if len(pricings) != 2 {
					t.Errorf("AdvancedPricings len = %d, want 2", len(pricings))
					return
				}
				if pricings[0].GetKey() != "DELIVERY.STANDARD.GRABFOOD" {
					t.Errorf("AdvancedPricings[0].Key = %s, want DELIVERY.STANDARD.GRABFOOD", pricings[0].GetKey())
				}
				if pricings[0].GetPrice() != 1500 {
					t.Errorf("AdvancedPricings[0].Price = %d, want 1500", pricings[0].GetPrice())
				}
			},
		},
		{
			name: "包含购买能力配置",
			req: &UpdateMenuItemReq{
				MerchantID: "M-12345",
				ItemID:     "ITEM-005",
				Purchasabilities: []UpdatePurchasabilityReq{
					{Key: "DELIVERY.STANDARD.GRABFOOD", Purchasable: true},
					{Key: "PICKUP.STANDARD.GRABFOOD", Purchasable: false},
				},
			},
			validate: func(t *testing.T, item *grabfood.UpdateMenuItem) {
				purchasabilities := item.GetPurchasabilities()
				if len(purchasabilities) != 2 {
					t.Errorf("Purchasabilities len = %d, want 2", len(purchasabilities))
					return
				}
				if !purchasabilities[0].GetPurchasable() {
					t.Error("Purchasabilities[0].Purchasable = false, want true")
				}
				if purchasabilities[1].GetPurchasable() {
					t.Error("Purchasabilities[1].Purchasable = true, want false")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.req.ToSDKUpdateMenuItem()
			tc.validate(t, result)
		})
	}
}

// ============================================================================
// UpdateMenuModifierReq DTO 测试
// ============================================================================

// TestUpdateMenuModifierReq_ToSDKUpdateMenuModifier 测试修饰符 DTO 转换
func TestUpdateMenuModifierReq_ToSDKUpdateMenuModifier(t *testing.T) {
	testCases := []struct {
		name     string
		req      *UpdateMenuModifierReq
		validate func(t *testing.T, modifier *grabfood.UpdateMenuModifier)
	}{
		{
			name: "基本字段转换",
			req: &UpdateMenuModifierReq{
				MerchantID:   "M-12345",
				ModifierID:   "MOD-001",
				ModifierName: "Extra Cheese",
			},
			validate: func(t *testing.T, modifier *grabfood.UpdateMenuModifier) {
				if modifier.GetMerchantID() != "M-12345" {
					t.Errorf("MerchantID = %s, want M-12345", modifier.GetMerchantID())
				}
				if modifier.GetId() != "MOD-001" {
					t.Errorf("Id = %s, want MOD-001", modifier.GetId())
				}
				if modifier.GetName() != "Extra Cheese" {
					t.Errorf("Name = %s, want Extra Cheese", modifier.GetName())
				}
				if modifier.GetField() != MenuItemUpdateFieldModifier {
					t.Errorf("Field = %s, want %s", modifier.GetField(), MenuItemUpdateFieldModifier)
				}
			},
		},
		{
			name: "包含价格和状态",
			req: &UpdateMenuModifierReq{
				MerchantID:      "M-12345",
				ModifierID:      "MOD-002",
				ModifierName:    "Large Size",
				Price:           ptrInt64(500), // 5.00
				AvailableStatus: MenuAvailableStatusAvailable,
			},
			validate: func(t *testing.T, modifier *grabfood.UpdateMenuModifier) {
				if modifier.GetPrice() != 500 {
					t.Errorf("Price = %d, want 500", modifier.GetPrice())
				}
				if modifier.GetAvailableStatus() != MenuAvailableStatusAvailable {
					t.Errorf("AvailableStatus = %s, want %s", modifier.GetAvailableStatus(), MenuAvailableStatusAvailable)
				}
			},
		},
		{
			name: "免费修饰符",
			req: &UpdateMenuModifierReq{
				MerchantID:   "M-12345",
				ModifierID:   "MOD-003",
				ModifierName: "Free Sauce",
				Price:        ptrInt64(0),
				IsFree:       ptrBool(true),
			},
			validate: func(t *testing.T, modifier *grabfood.UpdateMenuModifier) {
				if modifier.GetPrice() != 0 {
					t.Errorf("Price = %d, want 0", modifier.GetPrice())
				}
				if !modifier.GetIsFree() {
					t.Error("IsFree = false, want true")
				}
			},
		},
		{
			name: "包含高级定价",
			req: &UpdateMenuModifierReq{
				MerchantID:   "M-12345",
				ModifierID:   "MOD-004",
				ModifierName: "Premium Topping",
				AdvancedPricings: []UpdateAdvancedPricingReq{
					{Key: "DELIVERY.STANDARD.GRABFOOD", Price: 300},
				},
			},
			validate: func(t *testing.T, modifier *grabfood.UpdateMenuModifier) {
				pricings := modifier.GetAdvancedPricings()
				if len(pricings) != 1 {
					t.Errorf("AdvancedPricings len = %d, want 1", len(pricings))
					return
				}
				if pricings[0].GetPrice() != 300 {
					t.Errorf("AdvancedPricings[0].Price = %d, want 300", pricings[0].GetPrice())
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.req.ToSDKUpdateMenuModifier()
			tc.validate(t, result)
		})
	}
}

// ============================================================================
// 常量测试
// ============================================================================

// TestMenuUpdateConstants 测试常量定义
func TestMenuUpdateConstants(t *testing.T) {
	// 测试字段类型常量
	if MenuItemUpdateFieldItem != "ITEM" {
		t.Errorf("MenuItemUpdateFieldItem = %s, want ITEM", MenuItemUpdateFieldItem)
	}
	if MenuItemUpdateFieldModifier != "MODIFIER" {
		t.Errorf("MenuItemUpdateFieldModifier = %s, want MODIFIER", MenuItemUpdateFieldModifier)
	}

	// 测试状态常量
	if MenuAvailableStatusAvailable != "AVAILABLE" {
		t.Errorf("MenuAvailableStatusAvailable = %s, want AVAILABLE", MenuAvailableStatusAvailable)
	}
	if MenuAvailableStatusUnavailable != "UNAVAILABLE" {
		t.Errorf("MenuAvailableStatusUnavailable = %s, want UNAVAILABLE", MenuAvailableStatusUnavailable)
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

// ptrInt64 返回 int64 指针
func ptrInt64(v int64) *int64 {
	return &v
}

// ptrBool 返回 bool 指针
func ptrBool(v bool) *bool {
	return &v
}
