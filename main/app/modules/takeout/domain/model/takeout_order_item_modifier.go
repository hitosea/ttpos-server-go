package model

// TakeoutOrderItemModifier 外卖订单商品修饰符表（多平台）
//
// 此表对应 TTPOS 的三种修饰符类型：
//   - 规格（Flavor）：如大杯、中杯、小杯
//   - 加料（Sauce）：如珍珠、椰果、布丁
//   - 属性（Attribute）：如冰度、糖度
type TakeoutOrderItemModifier struct {
	BaseModel
	TakeoutOrderItemUuid uint64 `gorm:"column:takeout_order_item_uuid" json:"takeout_order_item_uuid"`
	Platform             string `gorm:"column:platform" json:"platform"`

	// 平台修饰符信息
	PlatformModifierId string `gorm:"column:platform_modifier_id" json:"platform_modifier_id"`
	ModifierName       string `gorm:"column:modifier_name;type:text" json:"modifier_name"`

	// TTPOS 修饰符信息（关联映射后）
	TtposModifierUuid uint64 `gorm:"column:ttpos_modifier_uuid" json:"ttpos_modifier_uuid"` // TTPOS 修饰符UUID（规格/加料/属性值的UUID）
	TtposModifierType string `gorm:"column:ttpos_modifier_type" json:"ttpos_modifier_type"` // TTPOS 修饰符类型：flavor=规格, sauce=加料, attr=属性, commodity=套餐商品

	// 数量和价格
	Quantity int     `gorm:"column:quantity" json:"quantity"`
	Price    float64 `gorm:"column:price;type:decimal(10,4)" json:"price"` // 价格(元,4位小数)
	Tax      float64 `gorm:"column:tax;type:decimal(10,4)" json:"tax"`     // 税费(元,4位小数)

	// 关联状态
	IsMapped int `gorm:"column:is_mapped" json:"is_mapped"`

	// 临时字段（不映射到数据库）
	TtposSkuName string `gorm:"-" json:"-"` // TTPOS BOM 的 sku_name（临时存储，用于库存不足提示）
}

func (*TakeoutOrderItemModifier) TableName() string {
	return "ttpos_takeout_order_item_modifier"
}

func (o *TakeoutOrderItemModifier) IsCommodity() bool {
	return o.TtposModifierType == "commodity"
}

func (o *TakeoutOrderItemModifier) IsPackage() bool {
	return o.TtposModifierType == "package"
}

func (o *TakeoutOrderItemModifier) IsFlavor() bool {
	return o.TtposModifierType == "flavor"
}

func (o *TakeoutOrderItemModifier) IsSauce() bool {
	return o.TtposModifierType == "sauce"
}

func (o *TakeoutOrderItemModifier) IsAttr() bool {
	return o.TtposModifierType == "attr"
}
