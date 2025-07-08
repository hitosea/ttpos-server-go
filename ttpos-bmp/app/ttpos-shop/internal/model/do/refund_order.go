// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RefundOrder is the golang structure of table ttpos_refund_order for DAO operations like Where/Data.
type RefundOrder struct {
	g.Meta           `orm:"table:ttpos_refund_order, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 退款单唯一标识符
	SaleOrderUuid    interface{} // 销售订单ID
	SaleOrderNo      interface{} // 销售订单号
	PaymentOrderUuid interface{} // 支付单ID
	RefundType       interface{} // 退款类型,1-反结账,2-取消付款
	Amount           interface{} // 退款金额
	Reason           interface{} // 退款原因
	Status           interface{} // 退款状态
	CreateTime       interface{} // 创建时间(时间戳)
	UpdateTime       interface{} // 更新时间(时间戳)
	DeleteTime       interface{} // 删除时间(时间戳)
}
