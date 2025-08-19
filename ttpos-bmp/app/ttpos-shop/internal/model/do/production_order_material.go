// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionOrderMaterial is the golang structure of table ttpos_production_order_material for DAO operations like Where/Data.
type ProductionOrderMaterial struct {
	g.Meta                     `orm:"table:ttpos_production_order_material, do:true"`
	Id                         interface{} // 自增ID
	Uuid                       interface{} // 生产订单原料ID
	Name                       interface{} // 原料名称,不随后台改变
	MaterialUuid               interface{} // 原料ID
	Num                        interface{} // 原料数量
	IsProductBom               interface{} // 是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1
	Unit                       interface{} // 单位,不随后台改变
	ProductionOrderProductUuid interface{} // 生产订单商品ID
	SaleOrderProductUuid       interface{} // 销售订单商品ID
	CreateTime                 interface{} // 创建时间(时间戳)
	UpdateTime                 interface{} // 更新时间(时间戳)
	DeleteTime                 interface{} // 删除时间(时间戳)
}
