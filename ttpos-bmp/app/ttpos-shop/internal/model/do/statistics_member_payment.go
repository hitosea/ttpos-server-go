// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsMemberPayment is the golang structure of table ttpos_statistics_member_payment for DAO operations like Where/Data.
type StatisticsMemberPayment struct {
	g.Meta                  `orm:"table:ttpos_statistics_member_payment, do:true"`
	Id                      interface{} //
	Uuid                    interface{} // uuid
	MemberRechargeOrderUuid interface{} // 会员充值订单uuid
	DutyNo                  interface{} // 当班编号
	PaymentMethodUuid       interface{} // 支付方式uuid
	PaymentAmount           interface{} // 支付金额
	RefundAmount            interface{} // 退款金额
	CompleteTime            interface{} // 完成时间
	CreateTime              interface{} // 创建时间
	UpdateTime              interface{} // 更新时间
	DeleteTime              interface{} // 删除时间
}
