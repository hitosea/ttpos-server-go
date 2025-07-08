// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LlPaymentOrder is the golang structure of table ttpos_ll_payment_order for DAO operations like Where/Data.
type LlPaymentOrder struct {
	g.Meta            `orm:"table:ttpos_ll_payment_order, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // UUID
	PaymentOrderUuid  interface{} // 自己系统的支付订单ID
	PaymentMethodUuid interface{} // 支付方式ID
	RelatedType       interface{} // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid       interface{} // 关联的充值订单、销售订单ID
	MerchantId        interface{} // lianlian商户号
	MerchantOrderId   interface{} // 自己系统的为支付生成的订单号
	OrderId           interface{} // lianlian订单ID
	OrderType         interface{} // 订单类型
	OrderStatus       interface{} // lianlian订单状态 PI-初始化(未访问支付页操作) WP-等待支付 PS-支付成功 PF-支付失败 PE-支付已过期
	OrderAmount       interface{} // lianlian订单金额
	OrderCurrency     interface{} // lianlian订单货币
	CommissionFee     interface{} // 支付手续费,支付金额*支付手续费百分比
	FullName          interface{} // 订单人名称
	OrderDesc         interface{} // 订单描述
	LinkUrl           interface{} // lianlian订单支付链接
	MerchantUserId    interface{} // 自己系统的用户ID
	LlCreateTime      interface{} // lianlian订单创建时间
	ExpiredTime       interface{} // 过期时间
	PayTime           interface{} // 支付时间
	CreateTime        interface{} // 创建时间
	UpdateTime        interface{} // 更新时间
	DeleteTime        interface{} // 删除时间(时间戳)
}
