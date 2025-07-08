// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductAttribute is the golang structure of table ttpos_sale_order_product_attribute for DAO operations like Where/Data.
type SaleOrderProductAttribute struct {
	g.Meta               `orm:"table:ttpos_sale_order_product_attribute, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 商品属性ID
	Name                 interface{} // 商品属性名称,不随后台更新
	SaleOrderUuid        interface{} // 销售订单ID
	SaleOrderProductUuid interface{} // 销售订单商品ID
	ProductAttributeUuid interface{} // 商品属性ID
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
