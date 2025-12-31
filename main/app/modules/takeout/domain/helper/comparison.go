package helper

import (
	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// IsInt64PtrEqual 比较两个 *int64 指针是否相等
// 用于比较可能为 nil 的整数指针值
func IsInt64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// HasMenuItemChanged 检查 Grab 菜单商品是否发生变更
// 比较商品的价格和可用状态
func HasMenuItemChanged(oldItem, newItem *grabfood.MenuItem) bool {
	// 比较价格变更
	if oldItem.Price != newItem.Price {
		return true
	}
	// 比较状态变更
	if oldItem.AvailableStatus != newItem.AvailableStatus {
		return true
	}
	return false
}

// HasMenuModifierChanged 检查 Grab 菜单修饰符是否发生变更
// 比较修饰符的价格和可用状态（价格字段为指针类型）
func HasMenuModifierChanged(oldModifier, newModifier *grabfood.MenuModifier) bool {
	// 比较价格变更（需要处理指针类型）
	if !IsInt64PtrEqual(oldModifier.Price, newModifier.Price) {
		return true
	}
	// 比较状态变更
	if oldModifier.AvailableStatus != newModifier.AvailableStatus {
		return true
	}
	return false
}
