// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberRechargeOrder is the golang structure of table ttpos_member_recharge_order for DAO operations like Where/Data.
type MemberRechargeOrder struct {
	g.Meta           `orm:"table:ttpos_member_recharge_order, do:true"`
	Id               interface{} // 自增ID
	Uuid             interface{} // 充值订单ID
	OrderNo          interface{} // 充值订单编号
	DutyNo           interface{} // 当班编号
	Status           interface{} // 状态,0-pending待支付 1-paid已支付 2-canceled已取消
	Amount           interface{} // 交易金额=充值金额+手续费
	RefundMoney      interface{} // 退款金额，不大于amount
	ChargeDue        interface{} // 找零
	RechargeAmount   interface{} // 充值金额
	RefundAmount     interface{} // 退款充值金额，不大于recharge_amount
	GiftAmount       interface{} // 赠送金额
	GiftPoint        interface{} // 赠送积分
	MemberUuid       interface{} // 会员ID
	StaffUuid        interface{} // 员工ID
	PaymentTime      interface{} // 支付时间(时间戳)
	Balance          interface{} // 充值前会员余额
	BalanceRecharged interface{} // 充值后会员余额
	CreateTime       interface{} // 创建时间(时间戳)
	UpdateTime       interface{} // 更新时间(时间戳)
	DeleteTime       interface{} // 删除时间(时间戳)
}
