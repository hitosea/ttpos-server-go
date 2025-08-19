// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnOrderAmount is the golang structure of table ttpos_return_order_amount for DAO operations like Where/Data.
type ReturnOrderAmount struct {
	g.Meta                `orm:"table:ttpos_return_order_amount, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 退货金额唯一标识符
	ReturnOrderUuid       interface{} // 关联退货单ID
	PaymentMethodUuid     interface{} // 关联支付方式ID
	PaymentOrderUuid      interface{} // 关联支付单ID,用于判断支付单的钱还有多少未退
	Amount                interface{} // 退款金额
	RefundStatus          interface{} // 退款状态 0-退款中 1-退款成功 2-退款失败
	LlReturnOrderId       interface{} // 连连退款订单ID, 用来重新发起退款
	MerchantRefundOrderNo interface{} // 商户退款单号
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
