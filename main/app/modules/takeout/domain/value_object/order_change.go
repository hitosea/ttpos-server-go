package value_object

// OrderChangeType 订单变动类型
type OrderChangeType int

const (
	ChangeTypeNone      OrderChangeType = 0 // 无变动
	ChangeTypeQuantity  OrderChangeType = 1 // 数量变动
	ChangeTypeAttribute OrderChangeType = 2 // 属性变动（口味、规格等）
	ChangeTypeAdded     OrderChangeType = 3 // 新增菜品
	ChangeTypeRemoved   OrderChangeType = 4 // 删除菜品
)

// String 返回变动类型的字符串表示
func (t OrderChangeType) String() string {
	switch t {
	case ChangeTypeNone:
		return "none"
	case ChangeTypeQuantity:
		return "quantity"
	case ChangeTypeAttribute:
		return "attribute"
	case ChangeTypeAdded:
		return "added"
	case ChangeTypeRemoved:
		return "removed"
	default:
		return "unknown"
	}
}

// ChangedItemModifier 变动菜品的修饰符信息
type ChangedItemModifier struct {
	PlatformModifierId      string  // 平台修饰符 ID
	ModifierName            string  // 修饰符名称
	TtposModifierType       string  // TTPOS 修饰符类型
	TtposModifierUuid       uint64  // TTPOS 修饰符 UUID（用于库存处理）
	TtposProductPackageUuid uint64  // TTPOS 商品套餐 UUID（commodity 类型的修饰符使用）
	TtposFlavorProductBomUuid uint64  // TTPOS 规格配方 UUID（flavor 类型的修饰符使用）
	Quantity                int     // 数量
	Price                   float64 // 价格
}

// ChangedItem 变动菜品信息（轻量级，避免循环导入）
type ChangedItem struct {
	Uuid                    uint64                // 菜品 UUID
	PlatformItemId          string                // 平台菜品 ID
	ItemName                string                // 菜品名称
	TtposProductPackageUuid uint64                // TTPOS 商品套餐 UUID
	TtposProductType        int                   // 0-商品, 1-套餐
	Quantity                int                   // 数量
	Price                   float64               // 单价
	Modifiers               []ChangedItemModifier // 修饰符列表
}

// ItemChange 单个菜品变动信息
type ItemChange struct {
	ChangeType     OrderChangeType // 变动类型
	PlatformItemId string          // 平台菜品 ID（用于匹配）
	ProductName    string          // 菜品名称

	// 数量信息
	OldQuantity int // 原数量
	NewQuantity int // 新数量

	// 完整菜品信息（用于打印和库存处理）
	OldItem *ChangedItem // 原菜品信息（退菜用）
	NewItem *ChangedItem // 新菜品信息（送厨用）
}

// GetQuantityDiff 获取数量差值（正数表示增加，负数表示减少）
func (c *ItemChange) GetQuantityDiff() int {
	return c.NewQuantity - c.OldQuantity
}

// IsIncrease 数量是否增加
func (c *ItemChange) IsIncrease() bool {
	return c.NewQuantity > c.OldQuantity
}

// IsDecrease 数量是否减少
func (c *ItemChange) IsDecrease() bool {
	return c.NewQuantity < c.OldQuantity
}

// OrderChangeResult 订单变动结果
type OrderChangeResult struct {
	HasChange    bool         // 是否有变动
	ReturnItems  []ItemChange // 需要退菜的项（打印退菜单 + 归还库存）
	KitchenItems []ItemChange // 需要送厨的项（打印送厨单 + 扣减库存）
	AllChanges   []ItemChange // 所有变动项（用于日志）
}

// NewOrderChangeResult 创建空的变动结果
func NewOrderChangeResult() *OrderChangeResult {
	return &OrderChangeResult{
		HasChange:    false,
		ReturnItems:  make([]ItemChange, 0),
		KitchenItems: make([]ItemChange, 0),
		AllChanges:   make([]ItemChange, 0),
	}
}

// AddReturnItem 添加退菜项
func (r *OrderChangeResult) AddReturnItem(item ItemChange) {
	r.ReturnItems = append(r.ReturnItems, item)
	r.AllChanges = append(r.AllChanges, item)
	r.HasChange = true
}

// AddKitchenItem 添加送厨项
func (r *OrderChangeResult) AddKitchenItem(item ItemChange) {
	r.KitchenItems = append(r.KitchenItems, item)
	r.HasChange = true
}

// AddChange 添加变动记录（仅用于日志，不影响打印和库存）
func (r *OrderChangeResult) AddChange(item ItemChange) {
	r.AllChanges = append(r.AllChanges, item)
	r.HasChange = true
}

// GetReturnItemCount 获取退菜项数量
func (r *OrderChangeResult) GetReturnItemCount() int {
	return len(r.ReturnItems)
}

// GetKitchenItemCount 获取送厨项数量
func (r *OrderChangeResult) GetKitchenItemCount() int {
	return len(r.KitchenItems)
}

// GetTotalReturnQuantity 获取退菜总数量
func (r *OrderChangeResult) GetTotalReturnQuantity() int {
	total := 0
	for _, item := range r.ReturnItems {
		total += item.OldQuantity
	}
	return total
}

// GetTotalKitchenQuantity 获取送厨总数量
func (r *OrderChangeResult) GetTotalKitchenQuantity() int {
	total := 0
	for _, item := range r.KitchenItems {
		total += item.NewQuantity
	}
	return total
}
