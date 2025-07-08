// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PaymentOrder is the golang structure for table payment_order.
type PaymentOrder struct {
	Id                   uint    `json:"id"                   orm:"id"                     description:"自增ID"`                           // 自增ID
	Uuid                 uint64  `json:"uuid"                 orm:"uuid"                   description:"支付订单ID"`                         // 支付订单ID
	PaymentMethodName    string  `json:"paymentMethodName"    orm:"payment_method_name"    description:"支付类型名称"`                         // 支付类型名称
	PaymentMethodUuid    uint64  `json:"paymentMethodUuid"    orm:"payment_method_uuid"    description:"支付类型ID"`                         // 支付类型ID
	PaymentFeePercent    float64 `json:"paymentFeePercent"    orm:"payment_fee_percent"    description:"支付手续费百分比,取值范围0-1"`               // 支付手续费百分比,取值范围0-1
	RelatedType          int     `json:"relatedType"          orm:"related_type"           description:"关联订单类型：0-销售订单；1-充值订单"`           // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid          uint64  `json:"relatedUuid"          orm:"related_uuid"           description:"关联的充值订单、销售订单ID"`                 // 关联的充值订单、销售订单ID
	CurrencyUnit         string  `json:"currencyUnit"         orm:"currency_unit"          description:"货币单位"`                           // 货币单位
	PaymentAmount        float64 `json:"paymentAmount"        orm:"payment_amount"         description:"支付金额"`                           // 支付金额
	PaymentCommissionFee float64 `json:"paymentCommissionFee" orm:"payment_commission_fee" description:"支付手续费,支付金额*支付手续费百分比"`            // 支付手续费,支付金额*支付手续费百分比
	Amount               float64 `json:"amount"               orm:"amount"                 description:"实收金额，实收金额=支付金额+支付手续费"`           // 实收金额，实收金额=支付金额+支付手续费
	TransactionNumber    string  `json:"transactionNumber"    orm:"transaction_number"     description:"交易号"`                            // 交易号
	Status               uint    `json:"status"               orm:"status"                 description:"支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常"` // 支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常
	StatusReason         string  `json:"statusReason"         orm:"status_reason"          description:"支付状态原因"`                         // 支付状态原因
	BalanceAmount        float64 `json:"balanceAmount"        orm:"balance_amount"         description:"主账户金额,用于反结账时退款"`                 // 主账户金额,用于反结账时退款
	GiftBalanceAmount    float64 `json:"giftBalanceAmount"    orm:"gift_balance_amount"    description:"赠送帐户金额,用于反结账时退款"`                // 赠送帐户金额,用于反结账时退款
	CreateTime           uint    `json:"createTime"           orm:"create_time"            description:"创建时间(时间戳)"`                      // 创建时间(时间戳)
	UpdateTime           uint    `json:"updateTime"           orm:"update_time"            description:"更新时间(时间戳)"`                      // 更新时间(时间戳)
	DeleteTime           uint    `json:"deleteTime"           orm:"delete_time"            description:"删除时间(时间戳)"`                      // 删除时间(时间戳)
}
