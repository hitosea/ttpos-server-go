package grab

import (
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// ============================================================================
// Update Menu Record (PUT /partner/v1/merchants/menu/record)
// ============================================================================

// MenuItemUpdateField 菜单项更新字段类型
const (
	MenuItemUpdateFieldItem     = "ITEM"     // 商品
	MenuItemUpdateFieldModifier = "MODIFIER" // 修饰符
)

// MenuAvailableStatus 菜单可用状态
const (
	MenuAvailableStatusAvailable   = "AVAILABLE"         // 可用
	MenuAvailableStatusUnavailable = "UNAVAILABLE"       // 不可用
	MenuAvailableStatusHide        = "UNAVAILABLEHIDE"   // 隐藏
	MenuAvailableStatusUnavailHide = "UNAVAILABLEHIDDEN" // 隐藏（别名）
)

// UpdateMenuItemReq 更新菜单项请求
type UpdateMenuItemReq struct {
	MerchantID      string `json:"merchant_id" v:"required#商户ID不能为空"` // Grab MerchantID
	ItemID          string `json:"item_id" v:"required#商品ID不能为空"`     // 商品ID (partner item id)
	Price           *int64 `json:"price,omitempty"`                   // 价格 (minor unit，单位：分)
	AvailableStatus string `json:"available_status,omitempty"`        // 可用状态: AVAILABLE, UNAVAILABLE, UNAVAILABLEHIDE
	MaxStock        *int64 `json:"max_stock,omitempty"`               // 库存数量，设为 0 时需同时设置 AvailableStatus 为 UNAVAILABLE
	// 高级定价配置
	AdvancedPricings []UpdateAdvancedPricingReq `json:"advanced_pricings,omitempty"`
	// 购买能力配置
	Purchasabilities []UpdatePurchasabilityReq `json:"purchasabilities,omitempty"`
}

// UpdateMenuModifierReq 更新修饰符请求
type UpdateMenuModifierReq struct {
	MerchantID      string `json:"merchant_id" v:"required#商户ID不能为空"`    // Grab MerchantID
	ModifierID      string `json:"modifier_id" v:"required#修饰符ID不能为空"`   // 修饰符ID (partner modifier id)
	ModifierName    string `json:"modifier_name" v:"required#修饰符名称不能为空"` // 修饰符名称 (用于定位记录)
	Price           *int64 `json:"price,omitempty"`                      // 价格 (minor unit，单位：分)
	AvailableStatus string `json:"available_status,omitempty"`           // 可用状态: AVAILABLE, UNAVAILABLE
	IsFree          *bool  `json:"is_free,omitempty"`                    // 是否免费
	// 高级定价配置
	AdvancedPricings []UpdateAdvancedPricingReq `json:"advanced_pricings,omitempty"`
}

// UpdateAdvancedPricingReq 高级定价请求
type UpdateAdvancedPricingReq struct {
	Key   string `json:"key"`   // 定价键 (格式: serviceType.orderType.channel)
	Price int64  `json:"price"` // 价格 (minor unit)
}

// UpdatePurchasabilityReq 购买能力请求
type UpdatePurchasabilityReq struct {
	Key         string `json:"key"`         // 购买能力键 (格式: serviceType.orderType.channel)
	Purchasable bool   `json:"purchasable"` // 是否可购买
}

// ToSDKUpdateMenuItem 转换为 SDK UpdateMenuItem 请求
func (r *UpdateMenuItemReq) ToSDKUpdateMenuItem() *grabfood.UpdateMenuItem {
	item := grabfood.NewUpdateMenuItem(r.MerchantID, MenuItemUpdateFieldItem, r.ItemID)

	if r.Price != nil {
		item.SetPrice(*r.Price)
	}
	if r.AvailableStatus != "" {
		item.SetAvailableStatus(r.AvailableStatus)
	}
	if r.MaxStock != nil {
		item.SetMaxStock(*r.MaxStock)
	}

	// 转换高级定价配置
	if len(r.AdvancedPricings) > 0 {
		var advancedPricings []grabfood.UpdateAdvancedPricing
		for _, ap := range r.AdvancedPricings {
			pricing := grabfood.NewUpdateAdvancedPricing()
			pricing.SetKey(ap.Key)
			pricing.SetPrice(ap.Price)
			advancedPricings = append(advancedPricings, *pricing)
		}
		item.SetAdvancedPricings(advancedPricings)
	}

	// 转换购买能力配置
	if len(r.Purchasabilities) > 0 {
		var purchasabilities []grabfood.UpdatePurchasability
		for _, p := range r.Purchasabilities {
			purchasability := grabfood.NewUpdatePurchasability()
			purchasability.SetKey(p.Key)
			purchasability.SetPurchasable(p.Purchasable)
			purchasabilities = append(purchasabilities, *purchasability)
		}
		item.SetPurchasabilities(purchasabilities)
	}

	return item
}

// ToSDKUpdateMenuModifier 转换为 SDK UpdateMenuModifier 请求
func (r *UpdateMenuModifierReq) ToSDKUpdateMenuModifier() *grabfood.UpdateMenuModifier {
	modifier := grabfood.NewUpdateMenuModifier(r.MerchantID, MenuItemUpdateFieldModifier, r.ModifierID, r.ModifierName)

	if r.Price != nil {
		modifier.SetPrice(*r.Price)
	}
	if r.AvailableStatus != "" {
		modifier.SetAvailableStatus(r.AvailableStatus)
	}
	if r.IsFree != nil {
		modifier.SetIsFree(*r.IsFree)
	}

	// 转换高级定价配置
	if len(r.AdvancedPricings) > 0 {
		var advancedPricings []grabfood.UpdateAdvancedPricing
		for _, ap := range r.AdvancedPricings {
			pricing := grabfood.NewUpdateAdvancedPricing()
			pricing.SetKey(ap.Key)
			pricing.SetPrice(ap.Price)
			advancedPricings = append(advancedPricings, *pricing)
		}
		modifier.SetAdvancedPricings(advancedPricings)
	}

	return modifier
}

// ============================================================================
// Batch Update Menu (POST /partner/v1/batch/menu)
// ============================================================================

// BatchUpdateMenuReq 批量更新菜单请求
type BatchUpdateMenuReq struct {
	MerchantID   string       `json:"merchant_id" v:"required#商户ID不能为空"`                                    // Grab MerchantID
	Field        string       `json:"field" v:"required|in:ITEM,MODIFIER#字段类型不能为空|字段类型必须是 ITEM 或 MODIFIER"` // 字段类型: ITEM (商品) 或 MODIFIER (修饰符)
	MenuEntities []MenuEntity `json:"menu_entities" v:"required#菜单实体列表不能为空"`                                // 菜单实体列表 (1-200 个)
}

// MenuEntity 菜单实体（用于批量更新）
type MenuEntity struct {
	ID               string                     `json:"id" v:"required#菜单项ID不能为空"`   // 菜单项ID或修饰符ID (partner item id 或 partner modifier id)
	Price            *int64                     `json:"price,omitempty"`             // 价格 (minor unit，单位：分)
	AvailableStatus  string                     `json:"available_status,omitempty"`  // 可用状态: AVAILABLE, UNAVAILABLE, UNAVAILABLEHIDE (商品), AVAILABLE, UNAVAILABLE (修饰符)
	MaxStock         *int64                     `json:"max_stock,omitempty"`         // 库存数量 (仅商品支持)
	AdvancedPricings []UpdateAdvancedPricingReq `json:"advanced_pricings,omitempty"` // 高级定价配置
	Purchasabilities []UpdatePurchasabilityReq  `json:"purchasabilities,omitempty"`  // 购买能力配置 (仅商品支持)
}

// BatchUpdateMenuResp 批量更新菜单响应
type BatchUpdateMenuResp struct {
	MerchantID string            `json:"merchant_id"`      // 商户ID
	Status     string            `json:"status"`           // 更新状态: success, partial_success, fail
	Errors     []MenuEntityError `json:"errors,omitempty"` // 错误列表 (仅在 partial_success 或 fail 时返回)
}

// MenuEntityError 菜单实体错误信息
type MenuEntityError struct {
	ID           string `json:"id"`            // 菜单项ID或修饰符ID
	ErrorCode    string `json:"error_code"`    // 错误代码
	ErrorMessage string `json:"error_message"` // 错误消息
}
