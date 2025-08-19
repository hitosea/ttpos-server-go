// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseForm is the golang structure of table ttpos_warehouse_form for DAO operations like Where/Data.
type WarehouseForm struct {
	g.Meta            `orm:"table:ttpos_warehouse_form, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // 库存入库单ID
	FormNo            interface{} // 编号
	Scene             interface{} // 交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库
	Num               interface{} // 数量
	Remark            interface{} // 备注
	Status            interface{} // 状态,0-success已入库 1-canceled已撤销
	ProductBomUuid    interface{} // 商品BOM表uuid
	MaterialUuid      interface{} // 材料uuid
	PurchaseOrderUuid interface{} // 采购订单uuid
	OperatorUuid      interface{} // 操作员uuid
	RevokeTime        interface{} // 撤销时间(时间戳)
	CreateTime        interface{} // 创建时间(时间戳)
	UpdateTime        interface{} // 更新时间(时间戳)
	DeleteTime        interface{} // 删除时间(时间戳)
}
