// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsMember is the golang structure of table ttpos_statistics_member for DAO operations like Where/Data.
type StatisticsMember struct {
	g.Meta                  `orm:"table:ttpos_statistics_member, do:true"`
	Id                      interface{} // 自增ID
	Uuid                    interface{} // UUID
	MemberRechargeOrderUuid interface{} // 会员充值订单uuid
	DutyNo                  interface{} // 当班编号
	RechargeAmount          interface{} // 充值金额
	GiveAmount              interface{} // 赠送金额
	GivePoint               interface{} // 赠送积分
	PaymentAmount           interface{} // 支付金额
	PaymentFee              interface{} // 支付手续费
	RefundAmount            interface{} // 退款金额
	RefundFee               interface{} // 退款手续费
	CompleteTime            interface{} // 完成时间
	RefundTime              interface{} // 完成时间
	CreateTime              interface{} // 创建时间
	UpdateTime              interface{} // 更新时间
	DeleteTime              interface{} // 删除时间
}
