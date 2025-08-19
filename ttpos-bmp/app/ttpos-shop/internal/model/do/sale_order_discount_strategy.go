// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderDiscountStrategy is the golang structure of table ttpos_sale_order_discount_strategy for DAO operations like Where/Data.
type SaleOrderDiscountStrategy struct {
	g.Meta        `orm:"table:ttpos_sale_order_discount_strategy, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // 销售订单优惠策略ID
	Type          interface{} // 优惠策略类型,0-整单折扣、1-会员折扣
	Name          interface{} // 优惠策略名称
	Value         interface{} // 优惠策略值
	JsonField     interface{} // JSON字段
	SaleOrderUuid interface{} // 销售订单ID
	CreateTime    interface{} // 创建时间(时间戳)
	UpdateTime    interface{} // 更新时间(时间戳)
	DeleteTime    interface{} // 删除时间(时间戳)
}
