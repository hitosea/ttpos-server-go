package model

// CoreSaleOrder 核心销售订单模型 `ttpos_sale_order`
type CoreSaleOrder struct {
	BaseModel
	SaleBillUuid uint64  `gorm:"column:sale_bill_uuid;index" json:"sale_bill_uuid"`
	OrderNo      string  `gorm:"column:order_no" json:"order_no"`
	Status       uint    `gorm:"column:status" json:"status"` // 0-未结账 1-已结账
	Amount       float64 `gorm:"column:amount" json:"amount"`
}

// TableName 指定表名
func (*CoreSaleOrder) TableName() string {
	return "ttpos_sale_order"
}
