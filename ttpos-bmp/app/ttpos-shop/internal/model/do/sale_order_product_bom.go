// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductBom is the golang structure of table ttpos_sale_order_product_bom for DAO operations like Where/Data.
type SaleOrderProductBom struct {
	g.Meta               `orm:"table:ttpos_sale_order_product_bom, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 销售订单商品规格或小料ID
	Name                 interface{} // 规格或小料名称,不随后台更新
	Price                interface{} // 单价,不随后台更新，记录加购时的价格。结账时要校验价格是否变动
	SaleOrderUuid        interface{} // 销售订单ID
	SaleOrderProductUuid interface{} // 销售订单商品ID
	ProductBomUuid       interface{} // 商品BOM ID
	IsFlavorBom          interface{} // 是否为规格商品BOM, 0-否 1-是
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
