package model

// TakeoutOrderItem 外卖订单商品表（多平台）
type TakeoutOrderItem struct {
	BaseModel
	// 关联外卖平台订单
	TakeoutOrderUuid uint64 `gorm:"column:takeout_order_uuid" json:"takeout_order_uuid"`
	Platform         string `gorm:"column:platform" json:"platform"`

	// 平台商品信息
	PlatformItemId string `gorm:"column:platform_item_id" json:"platform_item_id"`
	ItemName       string `gorm:"column:item_name;type:text" json:"item_name"`

	// TTPOS 商品信息（关联映射后）
	TtposProductUuid uint64 `gorm:"column:ttpos_product_uuid" json:"ttpos_product_uuid"`
	TtposSkuUuid     uint64 `gorm:"column:ttpos_sku_uuid" json:"ttpos_sku_uuid"`
	TtposProductType int    `gorm:"column:ttpos_product_type" json:"ttpos_product_type"` // 0-商品, 1-套餐

	// 商品数量和价格
	Quantity       int    `gorm:"column:quantity" json:"quantity"`
	Price          int64  `gorm:"column:price" json:"price"`
	Tax            int64  `gorm:"column:tax" json:"tax"`
	Specifications string `gorm:"column:specifications" json:"specifications"`

	// 关联状态
	IsMapped int `gorm:"column:is_mapped" json:"is_mapped"`

	// 平台特定数据（JSON 格式）
	PlatformData string `gorm:"column:platform_data;type:text" json:"platform_data"`

	// 关联字表结构
	TakeoutOrderItemModifiers []TakeoutOrderItemModifier `gorm:"foreignKey:TakeoutOrderItemUuid;references:Uuid"`
}

func (*TakeoutOrderItem) TableName() string {
	return "ttpos_takeout_order_item"
}

func (o *TakeoutOrderItem) SetTakeoutOrderItemModifiersNil() {
	o.TakeoutOrderItemModifiers = nil
}

func (o *TakeoutOrderItem) IsPackage() bool {
	return o.TtposProductType == 1
}
