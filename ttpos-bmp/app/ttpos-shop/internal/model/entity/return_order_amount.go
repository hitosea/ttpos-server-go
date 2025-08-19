// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ReturnOrderAmount is the golang structure for table return_order_amount.
type ReturnOrderAmount struct {
	Id                    uint    `json:"id"                    orm:"id"                       description:"自增ID"`                     // 自增ID
	Uuid                  uint64  `json:"uuid"                  orm:"uuid"                     description:"退货金额唯一标识符"`                // 退货金额唯一标识符
	ReturnOrderUuid       uint64  `json:"returnOrderUuid"       orm:"return_order_uuid"        description:"关联退货单ID"`                  // 关联退货单ID
	PaymentMethodUuid     uint64  `json:"paymentMethodUuid"     orm:"payment_method_uuid"      description:"关联支付方式ID"`                 // 关联支付方式ID
	PaymentOrderUuid      uint64  `json:"paymentOrderUuid"      orm:"payment_order_uuid"       description:"关联支付单ID,用于判断支付单的钱还有多少未退"`  // 关联支付单ID,用于判断支付单的钱还有多少未退
	Amount                float64 `json:"amount"                orm:"amount"                   description:"退款金额"`                     // 退款金额
	RefundStatus          int     `json:"refundStatus"          orm:"refund_status"            description:"退款状态 0-退款中 1-退款成功 2-退款失败"` // 退款状态 0-退款中 1-退款成功 2-退款失败
	LlReturnOrderId       string  `json:"llReturnOrderId"       orm:"ll_return_order_id"       description:"连连退款订单ID, 用来重新发起退款"`       // 连连退款订单ID, 用来重新发起退款
	MerchantRefundOrderNo string  `json:"merchantRefundOrderNo" orm:"merchant_refund_order_no" description:"商户退款单号"`                   // 商户退款单号
	CreateTime            uint    `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`                // 创建时间(时间戳)
	UpdateTime            uint    `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`                // 更新时间(时间戳)
	DeleteTime            uint    `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`                // 删除时间(时间戳)
}
