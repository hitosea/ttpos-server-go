// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrder is the golang structure of table ttpos_sale_order for DAO operations like Where/Data.
type SaleOrder struct {
	g.Meta                 `orm:"table:ttpos_sale_order, do:true"`
	Id                     interface{} // 自增ID
	Uuid                   interface{} // 销售订单ID
	OrderNo                interface{} // 订单编号
	Status                 interface{} // 订单状态, 0-未结账 1-已结账
	MemberDiscountFee      interface{} // 总会员折扣金额。总会员折扣金额=(订单商品.会员折扣金额)之和
	CustomDiscountFee      interface{} // 总自定义折扣金额。总自定义折扣金额=(订单商品.自定义折扣金额)之和
	ZeroFee                interface{} // 优惠折扣抹零金额。
	ProductAmount          interface{} // 商品金额，订单商品的最终单价(折后价)之和。商品已含税时，该金额包括了税费。当商品未含税时，该金额不包括税费
	ProductOriginalAmount  interface{} // 原始商品金额(折前价)。 商品原始金额=订单商品的销售价(折前价)之和。
	ServiceFee             interface{} // 服务费固定服务费时，服务费=固定服务费；按比例收服务费时，服务费=(订单商品.总服务费)之和
	TaxFee                 interface{} // 税费。税费=(订单商品.总税费)之和
	Amount                 interface{} // 应收金额。商品未含税时，总金额=商品金额+服务费+税费。商品已含税时，总金额=商品金额（含商品消费税）+服务费+税费（只有服务费税）
	OriginAmount           interface{} // 原始应收金额。原始应收金额=商品金额+服务费+消费税。商品未含税时，原始应收金额=商品金额+服务费+消费税（商品消费税税费+服务费税费）。商品已含税时，原始应收金额=商品金额（包含商品消费税税费）+服务费+服务费税费。
	IsFree                 interface{} // 是否免单, 0-否 1-是
	FreeReason             interface{} // 免单原因
	MemberDiscountRate     interface{} // 会员折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1
	MemberCardDiscountRate interface{} // 会员卡折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1
	CustomDiscountRate     interface{} // 自定义折扣率(0-100%)，默认100%，取值范围0-1，如折扣率为10%，则取值为0.1
	CustomAmount           interface{} // 整单改价金额。改价后，应收金额=整单改价金额，前端优先显示改价后的金额，改价金额不能为负数。当为-1时，表示不改价，显示amount改收金额
	ZeroRule               interface{} // 优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数
	ZeroCheckoutRule       interface{} // 结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元
	PaymentAmount          interface{} // 已支付金额,关联付款单的支付金额之和。
	ChangeAmount           interface{} // 找零金额,结账完成后才记录
	ZeroCheckoutFee        interface{} // 结账抹零金额。
	FinalPrice             interface{} // 最终应收金额。最终应收金额=应收金额+手续费-结账抹零金额
	PaymentCommissionFee   interface{} // 支付手续费,关联付款单的支付手续费之和
	GiftAmount             interface{} // 赠菜金额,(销售订单赠菜商品.总最终单价)之和
	GiftPoints             interface{} // 赠送积分. 赠送积分=应收金额amount*积分赠送比例.
	GiftPointsRate         interface{} // 赠送积分比例. 取值范围0-1。结账后记录，不受后台改变
	MemberBalance          interface{} // 会员余额.会员消费本单后剩余的余额
	CashierName            interface{} // 收银员名称
	ConsumerUuid           interface{} // 消费者ID
	CashierUuid            interface{} // 收银员ID
	SaleBillUuid           interface{} // 销售账单ID
	FinishTime             interface{} // 完成时间(时间戳),结账时间
	CreateTime             interface{} // 创建时间(时间戳)
	UpdateTime             interface{} // 更新时间(时间戳)
	DeleteTime             interface{} // 删除时间(时间戳)
}
