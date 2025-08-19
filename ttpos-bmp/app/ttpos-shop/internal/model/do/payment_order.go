// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PaymentOrder is the golang structure of table ttpos_payment_order for DAO operations like Where/Data.
type PaymentOrder struct {
	g.Meta               `orm:"table:ttpos_payment_order, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 支付订单ID
	PaymentMethodName    interface{} // 支付类型名称
	PaymentMethodUuid    interface{} // 支付类型ID
	PaymentFeePercent    interface{} // 支付手续费百分比,取值范围0-1
	RelatedType          interface{} // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid          interface{} // 关联的充值订单、销售订单ID
	CurrencyUnit         interface{} // 货币单位
	PaymentAmount        interface{} // 支付金额
	PaymentCommissionFee interface{} // 支付手续费,支付金额*支付手续费百分比
	Amount               interface{} // 实收金额，实收金额=支付金额+支付手续费
	TransactionNumber    interface{} // 交易号
	Status               interface{} // 支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常
	StatusReason         interface{} // 支付状态原因
	BalanceAmount        interface{} // 主账户金额,用于反结账时退款
	GiftBalanceAmount    interface{} // 赠送帐户金额,用于反结账时退款
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
