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
	TtposProductPackageUuid uint64 `gorm:"column:ttpos_product_package_uuid" json:"ttpos_product_package_uuid"` // TTPOS 商品套餐UUID(来自 ttpos_product_package)
	TtposProductType        int    `gorm:"column:ttpos_product_type" json:"ttpos_product_type"`                 // 0-商品, 1-套餐
	TtposItemName           string `gorm:"column:ttpos_item_name;type:text" json:"ttpos_item_name"`             // TTPOS 商品名称(来自 ttpos_product_package)
	TtposItemErpCode        string `gorm:"column:ttpos_item_erp_code" json:"ttpos_item_erp_code"`               // TTPOS 商品 ERP 编码(来自 ProductBom.ErpCode)

	// 商品数量和价格
	Quantity       int     `gorm:"column:quantity" json:"quantity"`
	Price          float64 `gorm:"column:price;type:decimal(20,4)" json:"price"` // 单价(元,4位小数)
	Tax            float64 `gorm:"column:tax;type:decimal(20,4)" json:"tax"`     // 税费(元,4位小数)
	Specifications string  `gorm:"column:specifications" json:"specifications"`

	// 关联状态
	IsMapped int `gorm:"column:is_mapped" json:"is_mapped"`

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

// 获取商品价格
func (o *TakeoutOrderItem) GetTotalPrice() float64 {
	return o.Price * float64(o.Quantity)
}
