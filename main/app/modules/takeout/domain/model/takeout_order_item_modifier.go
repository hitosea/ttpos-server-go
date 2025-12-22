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
	PlatformModifierId   string `gorm:"column:platform_modifier_id" json:"platform_modifier_id"`
	PlatformModifierName string `gorm:"column:platform_modifier_name" json:"platform_modifier_name"`

	// TTPOS 修饰符信息（关联映射后）
	TtposModifierUuid uint64 `gorm:"column:ttpos_modifier_uuid" json:"ttpos_modifier_uuid"` // TTPOS 修饰符UUID（规格/加料/属性值的UUID）
	TtposModifierType string `gorm:"column:ttpos_modifier_type" json:"ttpos_modifier_type"` // TTPOS 修饰符类型：flavor=规格, sauce=加料, attr=属性

	// 数量和价格
	Quantity int   `gorm:"column:quantity" json:"quantity"`
	Price    int64 `gorm:"column:price" json:"price"`
	Tax      int64 `gorm:"column:tax" json:"tax"`

	// 关联状态
	IsMapped int `gorm:"column:is_mapped" json:"is_mapped"`

	// 平台特定数据（JSON 格式）
	PlatformData string `gorm:"column:platform_data;type:text" json:"platform_data"`
}

func (*TakeoutOrderItemModifier) TableName() string {
	return "ttpos_takeout_order_item_modifier"
}
