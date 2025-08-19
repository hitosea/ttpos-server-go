// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseFormItem is the golang structure of table ttpos_warehouse_form_item for DAO operations like Where/Data.
type WarehouseFormItem struct {
	g.Meta               `orm:"table:ttpos_warehouse_form_item, do:true"`
	Id                   interface{} // 自增ID
	Uuid                 interface{} // 入库单明细uuid
	Num                  interface{} // 入库数量
	Scene                interface{} // 场景,0-采购 1-添加入库 2-调整入库 3-退菜入库、反结账入库,这个场景不显示在入库记录页面
	AddStock             interface{} // 是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存
	MaterialUuid         interface{} // 材料uuid
	ProductBomUuid       interface{} // 商品BOM表uuid
	WarehouseFormUuid    interface{} // 入库单uuid
	SaleOrderProductUuid interface{} // 销售订单商品uuid,用于退菜入库
	SaleBillUuid         interface{} // 销售账单uuid,用于退菜入库
	CreateTime           interface{} // 创建时间(时间戳)
	UpdateTime           interface{} // 更新时间(时间戳)
	DeleteTime           interface{} // 删除时间(时间戳)
}
