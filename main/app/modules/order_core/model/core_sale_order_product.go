package model

// CoreSaleOrderProduct 核心销售订单商品模型 `ttpos_sale_order_product`
type CoreSaleOrderProduct struct {
	BaseModel
	SaleBillUuid  uint64  `gorm:"column:sale_bill_uuid;index" json:"sale_bill_uuid"`
	SaleOrderUuid uint64  `gorm:"column:sale_order_uuid;index" json:"sale_order_uuid"`
	Name          string  `gorm:"column:name" json:"name"`
	FlavorName    string  `gorm:"column:flavor_name" json:"flavor_name"`
	Num           float64 `gorm:"column:num" json:"num"`
	Status        uint    `gorm:"column:status" json:"status"` // 状态, 0-未送厨 1-已送厨
	SalePrice     float64 `gorm:"column:sale_price" json:"sale_price"`
	Price         float64 `gorm:"column:price" json:"price"`
	TotalPrice    float64 `gorm:"column:total_price" json:"total_price"`
}

// TableName 指定表名
func (*CoreSaleOrderProduct) TableName() string {
	return "ttpos_sale_order_product"
}
