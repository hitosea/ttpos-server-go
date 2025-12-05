package model

// CoreSaleBill 核心销售账单模型 `ttpos_sale_bill`
type CoreSaleBill struct {
	BaseModel
	OrderNo  string  `gorm:"column:order_no" json:"order_no"`
	Status   uint    `gorm:"column:status" json:"status"` // 0-待付款、1-已完成、2-已取消
	BillType uint    `gorm:"column:bill_type" json:"bill_type"` // 0-桌台订单、1-点餐订单、2-会员端订单
	Amount   float64 `gorm:"column:amount" json:"amount"`
}

// TableName 指定表名
func (*CoreSaleBill) TableName() string {
	return "ttpos_sale_bill"
}
