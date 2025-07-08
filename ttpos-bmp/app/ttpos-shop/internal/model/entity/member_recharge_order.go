// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberRechargeOrder is the golang structure for table member_recharge_order.
type MemberRechargeOrder struct {
	Id               uint    `json:"id"               orm:"id"                description:"自增ID"`                                    // 自增ID
	Uuid             uint64  `json:"uuid"             orm:"uuid"              description:"充值订单ID"`                                  // 充值订单ID
	OrderNo          string  `json:"orderNo"          orm:"order_no"          description:"充值订单编号"`                                  // 充值订单编号
	DutyNo           string  `json:"dutyNo"           orm:"duty_no"           description:"当班编号"`                                    // 当班编号
	Status           int     `json:"status"           orm:"status"            description:"状态,0-pending待支付 1-paid已支付 2-canceled已取消"` // 状态,0-pending待支付 1-paid已支付 2-canceled已取消
	Amount           float64 `json:"amount"           orm:"amount"            description:"交易金额=充值金额+手续费"`                           // 交易金额=充值金额+手续费
	RefundMoney      float64 `json:"refundMoney"      orm:"refund_money"      description:"退款金额，不大于amount"`                          // 退款金额，不大于amount
	ChargeDue        float64 `json:"chargeDue"        orm:"charge_due"        description:"找零"`                                      // 找零
	RechargeAmount   float64 `json:"rechargeAmount"   orm:"recharge_amount"   description:"充值金额"`                                    // 充值金额
	RefundAmount     float64 `json:"refundAmount"     orm:"refund_amount"     description:"退款充值金额，不大于recharge_amount"`               // 退款充值金额，不大于recharge_amount
	GiftAmount       float64 `json:"giftAmount"       orm:"gift_amount"       description:"赠送金额"`                                    // 赠送金额
	GiftPoint        float64 `json:"giftPoint"        orm:"gift_point"        description:"赠送积分"`                                    // 赠送积分
	MemberUuid       uint64  `json:"memberUuid"       orm:"member_uuid"       description:"会员ID"`                                    // 会员ID
	StaffUuid        uint64  `json:"staffUuid"        orm:"staff_uuid"        description:"员工ID"`                                    // 员工ID
	PaymentTime      uint    `json:"paymentTime"      orm:"payment_time"      description:"支付时间(时间戳)"`                               // 支付时间(时间戳)
	Balance          float64 `json:"balance"          orm:"balance"           description:"充值前会员余额"`                                 // 充值前会员余额
	BalanceRecharged float64 `json:"balanceRecharged" orm:"balance_recharged" description:"充值后会员余额"`                                 // 充值后会员余额
	CreateTime       uint    `json:"createTime"       orm:"create_time"       description:"创建时间(时间戳)"`                               // 创建时间(时间戳)
	UpdateTime       uint    `json:"updateTime"       orm:"update_time"       description:"更新时间(时间戳)"`                               // 更新时间(时间戳)
	DeleteTime       uint    `json:"deleteTime"       orm:"delete_time"       description:"删除时间(时间戳)"`                               // 删除时间(时间戳)
}
