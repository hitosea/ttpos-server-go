// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionOrderProduct is the golang structure of table ttpos_production_order_product for DAO operations like Where/Data.
type ProductionOrderProduct struct {
	g.Meta                `orm:"table:ttpos_production_order_product, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 生产订单商品ID
	Name                  interface{} // 名称
	Num                   interface{} // 商品数量
	FlavorName            interface{} // 规格名称,不随后台改变
	ProductAttributeNames interface{} // 商品属性名称,多个属性名用逗号分隔,不随后台改变
	ProductSaucesNames    interface{} // 商品加料名称,多个加料名用逗号分隔,不随后台改变
	Status                interface{} // 状态, 0-待制作 1-制作中 2-已完成 3-已退菜
	Remark                interface{} // 商品备注
	HasMaterial           interface{} // 是否无原料, 0-无原料,商品没有关联原料 1-有原料
	SaleBillUuid          interface{} // 销售账单ID
	ProductPackageUuid    interface{} // 商品包ID
	SaleOrderProductUuid  interface{} // 销售订单商品ID
	ProductionOrderUuid   interface{} // 生产订单ID
	FirstCategoryUuid     interface{} // 一级分类ID
	FinishedTime          interface{} // 完成时间(时间戳)
	CreateTime            interface{} // 创建时间(时间戳),送厨时间
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
