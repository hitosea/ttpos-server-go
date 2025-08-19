// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// StatisticsSale is the golang structure of table ttpos_statistics_sale for DAO operations like Where/Data.
type StatisticsSale struct {
	g.Meta               `orm:"table:ttpos_statistics_sale, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // UUID
	SaleBillUuid         interface{} // 销售单UUID
	SaleOrderUuid        interface{} // 销售订单UUID
	DutyNo               interface{} // 当班编号
	DeskUuid             interface{} // 桌台UUID
	MealNum              interface{} // 用餐人数
	ProductPrice         interface{} // 商品原价: 不含税
	ProductSalePrice     interface{} // 商品销售价
	ProductNum           interface{} // 商品数量
	ProductTax           interface{} // 商品税
	ServiceFee           interface{} // 服务费
	ServiceTax           interface{} // 服务税
	Discount             interface{} // 优惠折扣
	DiscountMember       interface{} // 会员折扣
	GiftAmount           interface{} // 赠菜金额
	GiftNum              interface{} // 赠菜数量
	FreeAmount           interface{} // 免单金额
	FreeNum              interface{} // 免单数量
	PaymentAmount        interface{} // 支付金额
	PaymentFee           interface{} // 支付手续费
	PaymentBalance       interface{} // 支付余额
	RefundAmount         interface{} // 退款金额
	RefundTax            interface{} // 退款税额
	RefundServiceFee     interface{} // 退款服务费
	RefundDiscount       interface{} // 退款优惠折扣
	RefundDiscountMember interface{} // 退款会员折扣
	RefundFee            interface{} // 退款支付手续费
	CompleteTime         interface{} // 完成时间
	RefundTime           interface{} // 退款时间
	CreateTime           interface{} // 创建时间
	UpdateTime           interface{} // 更新时间
	DeleteTime           interface{} // 删除时间
}
