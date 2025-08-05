// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StatisticsMember is the golang structure for table statistics_member.
type StatisticsMember struct {
	Id                      uint    `json:"id"                      orm:"id"                         description:"自增ID"`       // 自增ID
	Uuid                    uint64  `json:"uuid"                    orm:"uuid"                       description:"UUID"`       // UUID
	MemberRechargeOrderUuid uint64  `json:"memberRechargeOrderUuid" orm:"member_recharge_order_uuid" description:"会员充值订单uuid"` // 会员充值订单uuid
	DutyNo                  string  `json:"dutyNo"                  orm:"duty_no"                    description:"当班编号"`       // 当班编号
	RechargeAmount          float64 `json:"rechargeAmount"          orm:"recharge_amount"            description:"充值金额"`       // 充值金额
	GiveAmount              float64 `json:"giveAmount"              orm:"give_amount"                description:"赠送金额"`       // 赠送金额
	GivePoint               int     `json:"givePoint"               orm:"give_point"                 description:"赠送积分"`       // 赠送积分
	PaymentAmount           float64 `json:"paymentAmount"           orm:"payment_amount"             description:"支付金额"`       // 支付金额
	PaymentFee              float64 `json:"paymentFee"              orm:"payment_fee"                description:"支付手续费"`      // 支付手续费
	RefundAmount            float64 `json:"refundAmount"            orm:"refund_amount"              description:"退款金额"`       // 退款金额
	RefundFee               float64 `json:"refundFee"               orm:"refund_fee"                 description:"退款手续费"`      // 退款手续费
	CompleteTime            uint    `json:"completeTime"            orm:"complete_time"              description:"完成时间"`       // 完成时间
	RefundTime              uint    `json:"refundTime"              orm:"refund_time"                description:"完成时间"`       // 完成时间
	CreateTime              uint    `json:"createTime"              orm:"create_time"                description:"创建时间"`       // 创建时间
	UpdateTime              uint    `json:"updateTime"              orm:"update_time"                description:"更新时间"`       // 更新时间
	DeleteTime              uint    `json:"deleteTime"              orm:"delete_time"                description:"删除时间"`       // 删除时间
}
