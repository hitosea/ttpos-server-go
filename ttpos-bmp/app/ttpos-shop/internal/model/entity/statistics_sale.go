// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StatisticsSale is the golang structure for table statistics_sale.
type StatisticsSale struct {
	Id                   uint    `json:"id"                   orm:"id"                     description:"自增ID"`      // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                   description:"UUID"`      // UUID
	SaleBillUuid         uint64  `json:"saleBillUuid"         orm:"sale_bill_uuid"         description:"销售单UUID"`   // 销售单UUID
	SaleOrderUuid        uint64  `json:"saleOrderUuid"        orm:"sale_order_uuid"        description:"销售订单UUID"`  // 销售订单UUID
	DutyNo               string  `json:"dutyNo"               orm:"duty_no"                description:"当班编号"`      // 当班编号
	DeskUuid             uint64  `json:"deskUuid"             orm:"desk_uuid"              description:"桌台UUID"`    // 桌台UUID
	MealNum              int     `json:"mealNum"              orm:"meal_num"               description:"用餐人数"`      // 用餐人数
	ProductPrice         float64 `json:"productPrice"         orm:"product_price"          description:"商品原价: 不含税"` // 商品原价: 不含税
	ProductSalePrice     float64 `json:"productSalePrice"     orm:"product_sale_price"     description:"商品销售价"`     // 商品销售价
	ProductNum           int     `json:"productNum"           orm:"product_num"            description:"商品数量"`      // 商品数量
	ProductTax           float64 `json:"productTax"           orm:"product_tax"            description:"商品税"`       // 商品税
	ServiceFee           float64 `json:"serviceFee"           orm:"service_fee"            description:"服务费"`       // 服务费
	ServiceTax           float64 `json:"serviceTax"           orm:"service_tax"            description:"服务税"`       // 服务税
	Discount             float64 `json:"discount"             orm:"discount"               description:"优惠折扣"`      // 优惠折扣
	DiscountMember       float64 `json:"discountMember"       orm:"discount_member"        description:"会员折扣"`      // 会员折扣
	GiftAmount           float64 `json:"giftAmount"           orm:"gift_amount"            description:"赠菜金额"`      // 赠菜金额
	GiftNum              int     `json:"giftNum"              orm:"gift_num"               description:"赠菜数量"`      // 赠菜数量
	FreeAmount           float64 `json:"freeAmount"           orm:"free_amount"            description:"免单金额"`      // 免单金额
	FreeNum              int     `json:"freeNum"              orm:"free_num"               description:"免单数量"`      // 免单数量
	PaymentAmount        float64 `json:"paymentAmount"        orm:"payment_amount"         description:"支付金额"`      // 支付金额
	PaymentFee           float64 `json:"paymentFee"           orm:"payment_fee"            description:"支付手续费"`     // 支付手续费
	PaymentBalance       float64 `json:"paymentBalance"       orm:"payment_balance"        description:"支付余额"`      // 支付余额
	RefundAmount         float64 `json:"refundAmount"         orm:"refund_amount"          description:"退款金额"`      // 退款金额
	RefundTax            float64 `json:"refundTax"            orm:"refund_tax"             description:"退款税额"`      // 退款税额
	RefundServiceFee     float64 `json:"refundServiceFee"     orm:"refund_service_fee"     description:"退款服务费"`     // 退款服务费
	RefundDiscount       float64 `json:"refundDiscount"       orm:"refund_discount"        description:"退款优惠折扣"`    // 退款优惠折扣
	RefundDiscountMember float64 `json:"refundDiscountMember" orm:"refund_discount_member" description:"退款会员折扣"`    // 退款会员折扣
	RefundFee            float64 `json:"refundFee"            orm:"refund_fee"             description:"退款支付手续费"`   // 退款支付手续费
	CompleteTime         uint    `json:"completeTime"         orm:"complete_time"          description:"完成时间"`      // 完成时间
	RefundTime           uint    `json:"refundTime"           orm:"refund_time"            description:"退款时间"`      // 退款时间
	CreateTime           uint    `json:"createTime"           orm:"create_time"            description:"创建时间"`      // 创建时间
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"            description:"更新时间"`      // 更新时间
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"            description:"删除时间"`      // 删除时间
}
