// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseOutFormItem is the golang structure of table ttpos_warehouse_out_form_item for DAO operations like Where/Data.
type WarehouseOutFormItem struct {
	g.Meta               `orm:"table:ttpos_warehouse_out_form_item, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 出库单明细uuid
	Num                  interface{} // 数量
	Scene                interface{} // 场景,0-sales销售 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除
	Status               interface{} // 状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示
	ReduceStock          interface{} // 是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存
	RevokeTime           interface{} // 撤销时间(时间戳)
	MaterialUuid         interface{} // 材料uuid
	WarehouseOutFormUuid interface{} // 出库单uuid
	ProductBomUuid       interface{} // 商品BOM表uuid
	SaleOrderProductUuid interface{} // 销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	SaleOrderUuid        interface{} // 销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	SaleBillUuid         interface{} // 销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
