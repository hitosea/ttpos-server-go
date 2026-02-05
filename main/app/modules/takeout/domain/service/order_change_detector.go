package service

import (
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
)

// OrderChangeDetector 订单变动检测器
type OrderChangeDetector struct{}

// NewOrderChangeDetector 创建订单变动检测器
func NewOrderChangeDetector() *OrderChangeDetector {
	return &OrderChangeDetector{}
}

// DetectChanges 检测订单菜品变动
// oldItems: 原订单菜品列表
// newItems: 新订单菜品列表
// 返回: 变动结果，包含退菜项和送厨项
func (d *OrderChangeDetector) DetectChanges(
	oldItems []model.TakeoutOrderItem,
	newItems []model.TakeoutOrderItem,
) *value_object.OrderChangeResult {
	result := value_object.NewOrderChangeResult()

	// 构建映射：以 PlatformItemId 为 key
	oldMap := d.buildItemMap(oldItems)
	newMap := d.buildItemMap(newItems)

	// 1. 检查删除和变动的菜品（遍历旧菜品）
	for platformId, oldItem := range oldMap {
		newItem, exists := newMap[platformId]
		if !exists {
			// 菜品被删除 → 全部退菜
			d.handleRemovedItem(result, oldItem)
		} else {
			// 检查变动
			d.handleItemChange(result, oldItem, newItem)
		}
	}

	// 2. 检查新增的菜品（遍历新菜品）
	for platformId, newItem := range newMap {
		if _, exists := oldMap[platformId]; !exists {
			// 新增菜品 → 送厨
			d.handleAddedItem(result, newItem)
		}
	}

	return result
}

// buildItemMap 构建菜品映射（以 PlatformItemId 为 key）
func (d *OrderChangeDetector) buildItemMap(items []model.TakeoutOrderItem) map[string]*model.TakeoutOrderItem {
	m := make(map[string]*model.TakeoutOrderItem)
	for i := range items {
		item := &items[i]
		m[item.PlatformItemId] = item
	}
	return m
}

// handleRemovedItem 处理删除的菜品
func (d *OrderChangeDetector) handleRemovedItem(
	result *value_object.OrderChangeResult,
	oldItem *model.TakeoutOrderItem,
) {
	change := value_object.ItemChange{
		ChangeType:     value_object.ChangeTypeRemoved,
		PlatformItemId: oldItem.PlatformItemId,
		ProductName:    oldItem.ItemName,
		OldQuantity:    oldItem.Quantity,
		NewQuantity:    0,
		OldItem:        d.convertToChangedItem(oldItem),
		NewItem:        nil,
	}
	result.AddReturnItem(change)
}

// handleAddedItem 处理新增的菜品
func (d *OrderChangeDetector) handleAddedItem(
	result *value_object.OrderChangeResult,
	newItem *model.TakeoutOrderItem,
) {
	change := value_object.ItemChange{
		ChangeType:     value_object.ChangeTypeAdded,
		PlatformItemId: newItem.PlatformItemId,
		ProductName:    newItem.ItemName,
		OldQuantity:    0,
		NewQuantity:    newItem.Quantity,
		OldItem:        nil,
		NewItem:        d.convertToChangedItem(newItem),
	}
	result.AddKitchenItem(change)
	result.AddChange(change)
}

// handleItemChange 处理菜品变动（数量或属性）
func (d *OrderChangeDetector) handleItemChange(
	result *value_object.OrderChangeResult,
	oldItem, newItem *model.TakeoutOrderItem,
) {
	// 新数量为 0 时视为删除
	if newItem.Quantity == 0 {
		d.handleRemovedItem(result, oldItem)
		return
	}

	// 检查数量变动
	quantityChanged := oldItem.Quantity != newItem.Quantity
	// 检查属性变动
	attributeChanged := d.hasAttributeChange(oldItem, newItem)

	if !quantityChanged && !attributeChanged {
		// 无变动
		return
	}

	// 确定变动类型
	changeType := value_object.ChangeTypeQuantity
	if attributeChanged {
		changeType = value_object.ChangeTypeAttribute
	}

	// 处理变动
	if attributeChanged {
		// 属性变动：原菜品全部退菜，新菜品全部送厨
		d.handleAttributeChange(result, oldItem, newItem, changeType)
	} else if quantityChanged {
		// 仅数量变动
		d.handleQuantityChange(result, oldItem, newItem)
	}
}

// handleAttributeChange 处理属性变动
// 属性变动时，原菜品需要退菜，新菜品需要送厨
func (d *OrderChangeDetector) handleAttributeChange(
	result *value_object.OrderChangeResult,
	oldItem, newItem *model.TakeoutOrderItem,
	changeType value_object.OrderChangeType,
) {
	// 原菜品退菜
	returnChange := value_object.ItemChange{
		ChangeType:     changeType,
		PlatformItemId: oldItem.PlatformItemId,
		ProductName:    oldItem.ItemName,
		OldQuantity:    oldItem.Quantity,
		NewQuantity:    newItem.Quantity,
		OldItem:        d.convertToChangedItem(oldItem),
		NewItem:        d.convertToChangedItem(newItem),
	}
	result.AddReturnItem(returnChange)

	// 新菜品送厨
	kitchenChange := value_object.ItemChange{
		ChangeType:     changeType,
		PlatformItemId: newItem.PlatformItemId,
		ProductName:    newItem.ItemName,
		OldQuantity:    oldItem.Quantity,
		NewQuantity:    newItem.Quantity,
		OldItem:        d.convertToChangedItem(oldItem),
		NewItem:        d.convertToChangedItem(newItem),
	}
	result.AddKitchenItem(kitchenChange)
}

// handleQuantityChange 处理数量变动
// 采用"退旧增新"模式：无论数量增减，都将旧商品退菜，新商品送厨
// 这样厨房能清楚看到哪些要退、哪些要新做
func (d *OrderChangeDetector) handleQuantityChange(
	result *value_object.OrderChangeResult,
	oldItem, newItem *model.TakeoutOrderItem,
) {
	// 旧菜品退菜
	returnChange := value_object.ItemChange{
		ChangeType:     value_object.ChangeTypeQuantity,
		PlatformItemId: oldItem.PlatformItemId,
		ProductName:    oldItem.ItemName,
		OldQuantity:    oldItem.Quantity,
		NewQuantity:    newItem.Quantity,
		OldItem:        d.convertToChangedItem(oldItem),
		NewItem:        d.convertToChangedItem(newItem),
	}
	result.AddReturnItem(returnChange)

	// 新菜品送厨
	kitchenChange := value_object.ItemChange{
		ChangeType:     value_object.ChangeTypeQuantity,
		PlatformItemId: newItem.PlatformItemId,
		ProductName:    newItem.ItemName,
		OldQuantity:    oldItem.Quantity,
		NewQuantity:    newItem.Quantity,
		OldItem:        d.convertToChangedItem(oldItem),
		NewItem:        d.convertToChangedItem(newItem),
	}
	result.AddKitchenItem(kitchenChange)
}

// hasAttributeChange 检查属性是否变动
// 比较规格/备注和修饰符列表是否一致
func (d *OrderChangeDetector) hasAttributeChange(
	oldItem, newItem *model.TakeoutOrderItem,
) bool {
	// 检查规格/备注是否变动
	if oldItem.Specifications != newItem.Specifications {
		return true
	}

	oldModifiers := oldItem.TakeoutOrderItemModifiers
	newModifiers := newItem.TakeoutOrderItemModifiers

	// 数量不同，肯定有变化
	if len(oldModifiers) != len(newModifiers) {
		return true
	}

	// 构建旧修饰符映射
	oldModifierMap := make(map[string]*model.TakeoutOrderItemModifier)
	for i := range oldModifiers {
		m := &oldModifiers[i]
		oldModifierMap[m.PlatformModifierId] = m
	}

	// 比较每个修饰符
	for i := range newModifiers {
		newM := &newModifiers[i]
		oldM, exists := oldModifierMap[newM.PlatformModifierId]
		if !exists {
			// 新增修饰符
			return true
		}
		// 比较数量和价格
		if oldM.Quantity != newM.Quantity || oldM.Price != newM.Price {
			return true
		}
	}

	return false
}

// convertToChangedItem 将 TakeoutOrderItem 转换为 ChangedItem
func (d *OrderChangeDetector) convertToChangedItem(item *model.TakeoutOrderItem) *value_object.ChangedItem {
	if item == nil {
		return nil
	}

	modifiers := make([]value_object.ChangedItemModifier, 0, len(item.TakeoutOrderItemModifiers))
	for _, m := range item.TakeoutOrderItemModifiers {
		modifiers = append(modifiers, value_object.ChangedItemModifier{
			Uuid:                      m.Uuid,
			PlatformModifierId:        m.PlatformModifierId,
			ModifierName:              m.ModifierName,
			TtposModifierType:         m.TtposModifierType,
			TtposModifierUuid:         m.TtposModifierUuid,
			TtposProductPackageUuid:   m.TtposProductPackageUuid,
			TtposFlavorProductBomUuid: m.TtposFlavorProductBomUuid,
			TtposFlavorName:           m.TtposFlavorName,
			Quantity:                  m.Quantity,
			Price:                     m.Price,
		})
	}

	return &value_object.ChangedItem{
		Uuid:                    item.Uuid,
		PlatformItemId:          item.PlatformItemId,
		ItemName:                item.ItemName,
		TtposProductPackageUuid: item.TtposProductPackageUuid,
		TtposProductType:        item.TtposProductType,
		Quantity:                item.Quantity,
		Price:                   item.Price,
		Modifiers:               modifiers,
	}
}
