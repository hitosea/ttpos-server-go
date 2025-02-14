package model

// 退货单表
type ReturnOrder struct {
	BaseModel

	// 基本信息
	SaleOrderUuid   uint64  `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	SaleOrderNo     string  `gorm:"column:sale_order_no;comment:销售订单号" json:"sale_order_no"`
	ReturnType      uint    `gorm:"column:return_type;comment:退货类型,1-整单退货,2-部分退货" json:"return_type"`
	RefundAmount    float64 `gorm:"column:refund_amount;comment:退款金额,包括税额" json:"refund_amount"`
	RefundTaxAmount float64 `gorm:"column:refund_tax_amount;comment:退款税额" json:"refund_tax_amount"`
	RefundReason    string  `gorm:"column:refund_reason;comment:退款原因" json:"refund_reason"`
	RefundStatus    uint    `gorm:"column:refund_status;comment:退款状态" json:"refund_status"`
}

// 退货单表
type ReturnOrderProduct struct {
	BaseModel

	// 基本信息
	ReturnOrderUuid    uint64  `gorm:"column:return_order_uuid;comment:退货单ID" json:"return_order_uuid"`
	ProductType        uint    `gorm:"column:product_type;comment:商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct 4-自助餐加钟顾客类型BuffetAddTimeCustomerType" json:"product_type"`
	ProductUuid        uint64  `gorm:"column:product_uuid;comment:商品ID" json:"product_uuid"`
	ProductName        string  `gorm:"column:product_name;comment:商品名称" json:"product_name"`
	ProductPrice       float64 `gorm:"column:product_price;comment:商品单价" json:"product_price"`
	TaxRate            float64 `gorm:"column:tax_rate;comment:税率,根据结账时税率计算" json:"tax_rate"`
	ProductQuantity    uint    `gorm:"column:product_quantity;comment:商品数量" json:"product_quantity"`
	ProductDiscount    float64 `gorm:"column:product_discount;comment:商品折扣" json:"product_discount"`
	ProductTotalAmount float64 `gorm:"column:product_total_amount;comment:商品总金额" json:"product_total_amount"`
}
