// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsPayment is the golang structure of table ttpos_statistics_payment for DAO operations like Where/Data.
type StatisticsPayment struct {
	g.Meta            `orm:"table:ttpos_statistics_payment, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // UUID
	SaleBillUuid      interface{} // 销售单UUID
	SaleOrderUuid     interface{} // 销售订单UUID
	DutyNo            interface{} // 当班编号
	DeskUuid          interface{} // 桌台UUID
	PaymentMethodUuid interface{} // 支付方式UUID
	PaymentAmount     interface{} // 支付金额
	RefundAmount      interface{} // 退款金额
	CompleteTime      interface{} // 完成时间
	CreateTime        interface{} // 创建时间
	UpdateTime        interface{} // 更新时间
	DeleteTime        interface{} // 删除时间
}
