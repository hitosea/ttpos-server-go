// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionOrder is the golang structure of table ttpos_production_order for DAO operations like Where/Data.
type ProductionOrder struct {
	g.Meta        `orm:"table:ttpos_production_order, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // 生产订单ID
	DeskUuid      interface{} // 桌台ID
	SaleOrderUuid interface{} // 销售订单ID
	SaleBillUuid  interface{} // 销售账单ID
	CreateTime    interface{} // 创建时间(时间戳)
	UpdateTime    interface{} // 更新时间(时间戳)
	DeleteTime    interface{} // 删除时间(时间戳)
}
