package model

// ReturnOrder 退货单表 ttpos_return_order
type ReturnOrder struct {
	BaseModel
	RelatedOrderType    uint    `gorm:"column:related_order_type;type:tinyint(1) unsigned;default:0;comment:关联订单类型：0-销售订单；1-充值订单;NOT NULL" json:"related_order_type"`
	RelatedOrderUuid    uint64  `gorm:"column:related_order_uuid;type:bigint(20) unsigned;default:0;comment:关联订单ID;NOT NULL" json:"related_order_uuid"`
	RelatedOrderNo      string  `gorm:"column:related_order_no;type:varchar(255);comment:关联订单号;NOT NULL" json:"related_order_no"`
	IsReverseSettlement uint    `gorm:"column:is_reverse_settlement;type:tinyint(1) unsigned;default:0;comment:是否反结账：0-否；1-是;NOT NULL" json:"is_reverse_settlement"`
	ReturnType          uint    `gorm:"column:return_type;type:tinyint(1) unsigned;default:0;comment:退货类型,1-整单退货,2-部分退货;NOT NULL" json:"return_type"`
	RefundAmount        float64 `gorm:"column:refund_amount;type:decimal(12,2);default:0.00;comment:退款金额,包括税额;NOT NULL" json:"refund_amount"`
	RefundTaxAmount     float64 `gorm:"column:refund_tax_amount;type:decimal(12,2);default:0.00;comment:退款税额;NOT NULL" json:"refund_tax_amount"`
	RefundReason        string  `gorm:"column:refund_reason;type:varchar(255);comment:退款原因;NOT NULL" json:"refund_reason"`
	RefundStatus        int     `gorm:"column:refund_status;type:int(11);default:0;comment:退款状态;NOT NULL" json:"refund_status"`

	ReturnOrderAmounts []ReturnOrderAmount `gorm:"foreignKey:ReturnOrderUuid;references:uuid"`
}

// ReturnOrderAmount 退款金额表 ttpos_return_order_amount
type ReturnOrderAmount struct {
	BaseModel
	ReturnOrderUuid   uint64  `gorm:"column:return_order_uuid;type:bigint(20) unsigned;default:0;comment:关联退货单ID;NOT NULL" json:"return_order_uuid"`
	PaymentMethodUuid uint64  `gorm:"column:payment_method_uuid;type:bigint(20) unsigned;default:0;comment:关联支付方式ID;NOT NULL" json:"payment_method_uuid"`
	Amount            float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:退款金额;NOT NULL" json:"amount"`

	ReturnOrder   *ReturnOrder   `gorm:"foreignKey:ReturnOrderUuid;references:Uuid"`   // 关联退货单
	PaymentMethod *PaymentMethod `gorm:"foreignKey:PaymentMethodUuid;references:Uuid"` // 关联支付方式
}

// ReturnOrderProduct 退货单商品表 ttpos_return_order_product
type ReturnOrderProduct struct {
	BaseModel
	// 基本信息
	SaleOrderUuid        uint64  `gorm:"column:sale_order_uuid;comment:销售订单ID" json:"sale_order_uuid"`
	SaleOrderProductUuid uint64  `gorm:"column:sale_order_product_uuid;comment:销售订单商品表ID" json:"sale_order_product_uuid"`
	ReturnOrderUuid      uint64  `gorm:"column:return_order_uuid;comment:退货单ID" json:"return_order_uuid"`
	ProductType          uint    `gorm:"column:product_type;comment:商品类型, 1-销售订单商品SaleOrderProduct 2-销售订单顾客类型SaleOrderBuffetCustomerType 3-自助餐加钟BuffetAddTimeProduct 4-自助餐加钟顾客类型BuffetAddTimeCustomerType" json:"product_type"`
	ProductUuid          uint64  `gorm:"column:product_uuid;comment:商品ID" json:"product_uuid"`
	ProductName          string  `gorm:"column:product_name;comment:商品名称" json:"product_name"`
	ProductPrice         float64 `gorm:"column:product_price;comment:商品单价" json:"product_price"`
	TaxRate              float64 `gorm:"column:tax_rate;comment:税率,根据结账时税率计算" json:"tax_rate"`
	ProductQuantity      uint    `gorm:"column:product_quantity;comment:商品数量" json:"product_quantity"`
	ProductDiscount      float64 `gorm:"column:product_discount;comment:商品折扣" json:"product_discount"`
	ProductTotalAmount   float64 `gorm:"column:product_total_amount;comment:商品总金额" json:"product_total_amount"`
}
