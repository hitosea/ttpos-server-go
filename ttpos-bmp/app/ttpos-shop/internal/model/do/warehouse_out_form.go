// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseOutForm is the golang structure of table ttpos_warehouse_out_form for DAO operations like Where/Data.
type WarehouseOutForm struct {
	g.Meta              `orm:"table:ttpos_warehouse_out_form, do:true"`
	Id                  interface{} // 自增ID
	Uuid                interface{} // 出库单uuid
	FormNo              interface{} // 编号
	Scene               interface{} // 出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库
	Remark              interface{} // 备注
	Status              interface{} // 状态,0-success已出库 1-canceled已撤销
	OperatorUuid        interface{} // 操作员uuid
	AssociatedOrderUuid interface{} // 关联订单uuid
	RevokeTime          interface{} // 撤销时间(时间戳)
	CreateTime          interface{} // 创建时间(时间戳)
	UpdateTime          interface{} // 更新时间(时间戳)
	DeleteTime          interface{} // 删除时间(时间戳)
}
