// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StatisticsMemberPayment is the golang structure for table statistics_member_payment.
type StatisticsMemberPayment struct {
	Id                      uint    `json:"id"                      orm:"id"                         description:""`           //
	Uuid                    uint64  `json:"uuid"                    orm:"uuid"                       description:"uuid"`       // uuid
	MemberRechargeOrderUuid uint64  `json:"memberRechargeOrderUuid" orm:"member_recharge_order_uuid" description:"会员充值订单uuid"` // 会员充值订单uuid
	DutyNo                  string  `json:"dutyNo"                  orm:"duty_no"                    description:"当班编号"`       // 当班编号
	PaymentMethodUuid       uint64  `json:"paymentMethodUuid"       orm:"payment_method_uuid"        description:"支付方式uuid"`   // 支付方式uuid
	PaymentAmount           float64 `json:"paymentAmount"           orm:"payment_amount"             description:"支付金额"`       // 支付金额
	RefundAmount            float64 `json:"refundAmount"            orm:"refund_amount"              description:"退款金额"`       // 退款金额
	CompleteTime            int     `json:"completeTime"            orm:"complete_time"              description:"完成时间"`       // 完成时间
	CreateTime              int     `json:"createTime"              orm:"create_time"                description:"创建时间"`       // 创建时间
	UpdateTime              int     `json:"updateTime"              orm:"update_time"                description:"更新时间"`       // 更新时间
	DeleteTime              int     `json:"deleteTime"              orm:"delete_time"                description:"删除时间"`       // 删除时间
}
