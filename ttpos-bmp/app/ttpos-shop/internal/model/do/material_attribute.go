// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MaterialAttribute is the golang structure of table ttpos_material_attribute for DAO operations like Where/Data.
type MaterialAttribute struct {
	g.Meta                       `orm:"table:ttpos_material_attribute, do:true"`
	Id                           interface{} // 自增ID
	Uuid                         interface{} // 原料属性ID
	MaterialUuid                 interface{} // 原料ID
	HistoricalPurchaseQuantity   interface{} // 历史采购数量
	HistoricalLossReportQuantity interface{} // 历史报损数量
	HistoricalSaleQuantity       interface{} // 历史销售数量
	CreateTime                   interface{} // 创建时间(时间戳)
	UpdateTime                   interface{} // 更新时间(时间戳)
	DeleteTime                   interface{} // 删除时间(时间戳)
}
